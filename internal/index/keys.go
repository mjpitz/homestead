package index

import (
	"bytes"
	"encoding/binary"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/mjpitz/myago/encoding"
)

const (
	// DocumentPrefix keys the internal_id of the system to the underlying document.
	DocumentPrefix = 'd'
	// SchemaPrefix stores field names and types
	SchemaPrefix = 's'
	// TermPrefix keys field-term-document tuple to the score of the result.
	TermPrefix = 't'
	// Separator separates portions of the field key
	Separator = ':'
)

// DocumentKey represents a key to a document within the index.
type DocumentKey uint64

func (k DocumentKey) MarshalBinary() ([]byte, error) {
	key := make([]byte, 10)
	key[0] = DocumentPrefix
	key[1] = Separator
	binary.BigEndian.PutUint64(key[2:], uint64(k))
	return key, nil
}

func (k *DocumentKey) UnmarshalBinary(data []byte) error {
	if data[0] != DocumentPrefix || data[1] != Separator {
		return os.ErrInvalid
	}

	*k = DocumentKey(binary.BigEndian.Uint64(data[2:]))
	return nil
}

// TermKey represents various metadata that forms a key for a given documents field. This is structured to allow tf-idf
// later on. For now, it's a rather convenient structure for storing our information. This key space exists entirely in
// memory so there's currently no need to fetch values.
type TermKey struct {
	Name     string
	Term     interface{}
	Document DocumentKey
}

// Prefix returns the longest running prefix for the given key. When all values are non-0, this returns the equivalent
// of MarshalBinary.
func (k TermKey) Prefix() []byte {
	buf := bytes.NewBuffer([]byte{TermPrefix, Separator})
	buf.WriteString(k.Name)
	buf.WriteByte(Separator)

	if k.Term != nil {
		enc := encoding.MsgPack.Encoder(buf).(*msgpack.Encoder)
		enc.SetOmitEmpty(false)
		_ = enc.Encode(k.Term)
		buf.WriteByte(Separator)
	}

	return buf.Bytes()
}

func (k TermKey) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{TermPrefix, Separator})
	buf.WriteString(k.Name)
	buf.WriteByte(Separator)

	enc := encoding.MsgPack.Encoder(buf).(*msgpack.Encoder)
	enc.SetOmitEmpty(false)
	err := enc.Encode(k.Term)
	if err != nil {
		return nil, err
	}
	buf.WriteByte(Separator)

	documentKey, _ := k.Document.MarshalBinary()
	buf.Write(documentKey)

	return buf.Bytes(), nil
}

func (k *TermKey) UnmarshalBinary(data []byte) error {
	if data[0] != TermPrefix || data[1] != Separator {
		return os.ErrInvalid
	}

	var err error
	buf := bytes.NewBuffer(data[2:])
	k.Name, err = buf.ReadString(Separator)
	if err != nil {
		return err
	}
	k.Name = strings.TrimSuffix(k.Name, ":")

	err = encoding.MsgPack.Decoder(buf).Decode(&k.Term)
	if err != nil {
		return err
	}

	_, err = buf.ReadByte()
	if err != nil {
		return err
	}

	return (&k.Document).UnmarshalBinary(buf.Bytes())
}

// TermSet defines a generic collection of field keys.
type TermSet []TermKey

// Terms provides a convenience function that will extract a TermSet from the provided interface{} using golang's
// reflect library.
func Terms(v interface{}) TermSet {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return nil
	}

	terms := make([]TermKey, 0, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		field := val.Field(i)

		name := strings.Split(fieldType.Tag.Get("json"), ",")[0]
		if name == "" {
			name = fieldType.Name
		} else if name == "-" {
			continue
		}

		terms = append(terms, TermKey{
			Name: name,
			Term: field.Interface(),
		})
	}

	return terms
}

// SchemaKey maps a given field to it's corresponding type.
type SchemaKey string

func (k SchemaKey) MarshalBinary() ([]byte, error) {
	return append([]byte{SchemaPrefix, Separator}, []byte(k)...), nil
}

func (k *SchemaKey) UnmarshalBinary(data []byte) error {
	if data[0] != SchemaPrefix || data[1] != Separator {
		return os.ErrInvalid
	}

	*k = SchemaKey(data[2:])
	return nil
}

// Schema returns a collection of fields that describe the provided document.
func Schema(v interface{}) []Field {
	val := reflect.Indirect(reflect.ValueOf(v))
	if val.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]Field, 0, val.NumField())

	for i := 0; i < val.NumField(); i++ {
		fieldType := val.Type().Field(i)
		field := val.Field(i)

		name := strings.Split(fieldType.Tag.Get("json"), ",")[0]
		if name == "" {
			name = fieldType.Name
		} else if name == "-" {
			continue
		}

		kind := "string"

		switch field.Interface().(type) {
		case *time.Time, time.Time:
			kind = "time"
		default:
			switch field.Kind() {
			case reflect.Bool:
				kind = "number"
			case reflect.Float32, reflect.Float64:
				kind = "number"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				kind = "number"
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				kind = "number"
			}
		}

		fields = append(fields, Field{
			Text: SchemaKey(name),
			Type: kind,
		})
	}

	return fields
}
