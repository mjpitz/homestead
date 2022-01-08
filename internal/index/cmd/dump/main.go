package main

import (
	"fmt"
	"os"

	"github.com/dgraph-io/badger/v3"

	"github.com/mjpitz/homestead/internal/index"
)

func main() {
	opts := badger.DefaultOptions(os.Args[1]).
		WithReadOnly(true)

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	txn := db.NewTransaction(true)
	defer txn.Discard()

	iterOps := badger.IteratorOptions{PrefetchSize: 100}
	iter := txn.NewIterator(iterOps)
	defer iter.Close()

	iter.Seek(nil)
	for iter.Valid() {
		key := iter.Item().Key()
		val, _ := iter.Item().ValueCopy(nil)

		if len(key) < 2 || key[1] != index.Separator {
			iter.Next()
			continue
		}

		switch key[0] {
		case index.SchemaPrefix:
			var k index.SchemaKey
			_ = (&k).UnmarshalBinary(key)
			fmt.Printf("%s\t%s\n", k, string(val))

		case index.TermPrefix:
			k := index.TermKey{}
			_ = (&k).UnmarshalBinary(key)
			fmt.Printf("%d\t%s\t%v\n", k.Document, k.Name, k.Term)
		}

		iter.Next()
	}
}
