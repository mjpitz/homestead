package index

import (
	"bytes"
	"fmt"
	"log"
	"reflect"

	"github.com/dgraph-io/badger/v3"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/mjpitz/myago/encoding"
)

var docIdSequence = []byte("!:doc_id_sequence")

// Open attempts to open the InvertedIndex at the provided path.
func Open(path string, readonly bool) (*InvertedIndex, error) {
	opts := badger.DefaultOptions(path).
		WithInMemory(len(path) == 0).
		WithReadOnly(readonly).
		WithLoggingLevel(badger.WARNING)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	var seq *badger.Sequence
	if !readonly {
		seq, err = db.GetSequence(docIdSequence, 100)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return &InvertedIndex{
		db:  db,
		seq: seq,
	}, nil
}

// Index defines an abstraction
type Index interface {
	Schema() []Field
	Query(query ...Query) []uint64
	Get(base interface{}, ids ...uint64) []interface{}
	Index(docs ...interface{}) error
	Close() error
}

// InvertedIndex provides an Index whose data is backed by an inverted index. Currently, entire copies of documents are
// stored but an argument could be made to optionally store those later on. Based on what I've seen in bleve, it seems
// like most, if not all numeric fields are stored. Longer text fields are not. Given that this solution isn't focusing
// on the broader search space and instead only on cares about a subset of the functionality for analytical purposes,
// ignoring this for now seems negligible.
type InvertedIndex struct {
	db  *badger.DB
	seq *badger.Sequence
}

// Schema returns the set of fields belonging to the database.
func (i *InvertedIndex) Schema() []Field {
	txn := i.db.NewTransaction(false)
	defer txn.Discard()

	opts := badger.IteratorOptions{
		PrefetchSize:   100,
		PrefetchValues: true,
		Prefix:         []byte{SchemaPrefix, Separator},
	}

	iter := txn.NewIterator(opts)
	defer iter.Close()

	iter.Seek(opts.Prefix)

	fields := make([]Field, 0)
	for iter.Valid() {
		data, _ := iter.Item().ValueCopy(nil)

		field := Field{Type: string(data)}
		_ = field.Text.UnmarshalBinary(iter.Item().Key())

		fields = append(fields, field)
		iter.Next()
	}

	return fields
}

// Query returns a list of document ids that matched the provided query.
func (i *InvertedIndex) Query(query ...Query) []uint64 {
	txn := i.db.NewTransaction(false)
	defer txn.Discard()

	newIter := func(q Query) *Iterator {
		var prefix, end []byte

		term := TermKey{Name: q.Field, Term: q.Value}
		termKey, _ := term.MarshalBinary()

		termPrefix := term.Prefix()
		fieldPrefix := TermKey{Name: q.Field}.Prefix()
		prefix = fieldPrefix[:]

		iter := txn.NewIterator(badger.IteratorOptions{
			PrefetchSize: 100,
			Prefix:       fieldPrefix[:],
		})

		switch q.Operator {
		case ">=", "=>":
			iter.Seek(termPrefix)
		case ">":
			iter.Seek(termPrefix)

			for iter.Valid() && bytes.HasPrefix(iter.Item().Key(), termPrefix) {
				iter.Next()
			}
		case "=":
			iter.Seek(termPrefix)
			prefix = termPrefix[:]
		case "<":
			iter.Seek(fieldPrefix)
			prefix = nil
			end = termKey
		case "<=", "=<":
			iter.Seek(fieldPrefix)
			prefix = nil
			end = termPrefix[:]
			end[len(end)-1]++ // turns ':' into ':'++
		}

		return &Iterator{
			iter:   iter,
			prefix: prefix,
			end:    end,
		}
	}

	iter := &Iterators{}
	defer iter.Close()

	for _, q := range query {
		iter.Push(newIter(q))
	}

	matched := make([]uint64, 0)
	if iter.Len() == 0 {
		return matched
	}

	for {
		if iter.AllSame() {
			matched = append(matched, uint64((*iter)[0].Current().Document))
		}

		i := iter.Pop()
		i.Next()

		if !i.Valid() {
			i.iter.Close()
			return matched
		}

		iter.Push(i)
	}
}

func (i *InvertedIndex) Get(base interface{}, ids ...uint64) []interface{} {
	result := make([]interface{}, 0, len(ids))

	txn := i.db.NewTransaction(false)
	defer txn.Discard()

	for _, id := range ids {
		docKey, _ := DocumentKey(id).MarshalBinary()

		item, err := txn.Get(docKey)
		if err != nil {
			log.Println(err)
			continue
		}

		doc := reflect.New(reflect.Indirect(reflect.ValueOf(base)).Type()).Interface()

		err = item.Value(func(val []byte) error {
			dec := encoding.MsgPack.Decoder(bytes.NewReader(val)).(*msgpack.Decoder)
			dec.SetCustomStructTag("json")
			return dec.Decode(doc)
		})
		if err != nil {
			log.Println(err)
			continue
		}

		result = append(result, doc)
	}

	return result
}

// Index stores the provided documents in the underlying kv-store.
func (i *InvertedIndex) Index(docs ...interface{}) error {
	if i.seq == nil {
		return fmt.Errorf("index read only")
	}

	kvs := make([]kv, 0)

	for _, doc := range docs {
		id, err := i.seq.Next()
		if err != nil {
			return err
		}

		docID := DocumentKey(id)
		docKey, _ := docID.MarshalBinary()

		buf := bytes.NewBuffer(nil)
		enc := encoding.MsgPack.Encoder(buf).(*msgpack.Encoder)
		enc.SetCustomStructTag("json")
		enc.SetOmitEmpty(false)

		err = enc.Encode(doc)
		if err != nil {
			return err
		}

		kvs = append(kvs, kv{docKey, buf.Bytes()})

		for _, field := range Schema(doc) {
			fieldKey, _ := field.Text.MarshalBinary()

			kvs = append(kvs, kv{fieldKey, []byte(field.Type)})
		}

		for _, term := range Terms(doc) {
			term.Document = docID
			termKey, _ := term.MarshalBinary()

			kvs = append(kvs, kv{key: termKey})
		}
	}

	return i.db.Update(func(txn *badger.Txn) error {
		for _, kv := range kvs {
			err := txn.Set(kv.key, kv.value)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Close ensures the underlying sequence is released and the database is closed.
func (i *InvertedIndex) Close() error {
	if i.seq != nil {
		_ = i.seq.Release()
	}

	return i.db.Close()
}

// Field provides a loose reference to a field within the index and its type.
type Field struct {
	Type string
	Text SchemaKey
}

// Query defines the components needed to filter data from the underlying index.
type Query struct {
	Field    string
	Operator string
	Value    interface{}
}

// kv defines a general key-value structure used to simplify the Index operation.
type kv struct {
	key   []byte
	value []byte
}
