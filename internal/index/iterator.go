package index

import (
	"bytes"
	"sort"

	"github.com/dgraph-io/badger/v3"
)

// Iterators provides a sorted list of iterators that facilitates concurrently walking N iterators.
type Iterators []*Iterator

// Len returns the number of iterators in the collection.
func (i Iterators) Len() int {
	return len(i)
}

// AllSame compares the document ids of the first and last iterators and returns if they're the same.
func (i Iterators) AllSame() bool {
	if len(i) == 0 {
		return false
	}

	first := i[0]
	last := i[len(i)-1]

	if !first.Valid() || !last.Valid() {
		return false
	}

	return first.Current().Document == last.Current().Document
}

// Push inserts the provided iterator back into the collection. This is a `O(log n)` operation.
func (i *Iterators) Push(iterator *Iterator) {
	o := *i
	n := len(o)

	idx := sort.Search(n, func(pos int) bool {
		return bytes.Compare(iterator.iter.Item().Key(), o[pos].iter.Item().Key()) <= 0
	})

	if idx == n {
		*i = append(o, iterator)
	} else {
		*i = append(o[:idx], append([]*Iterator{iterator}, o[idx:]...)...)
	}
}

// Pop returns the smallest element in the collection. This is a `O(C)` operation.
func (i *Iterators) Pop() *Iterator {
	o := *i
	if len(o) == 0 {
		return nil
	}

	*i = o[1:]
	return o[0]
}

// Close cleans up the underlying iterators.
func (i *Iterators) Close() {
	o := *i

	for _, iter := range o {
		_ = iter.Close()
	}
}

// Iterator wraps a badger iterator with a Query to filter keys from the underlying store.
type Iterator struct {
	iter *badger.Iterator

	prefix []byte
	end    []byte

	cache *TermKey
}

// Current returns the current key for the underlying iterator.
func (i *Iterator) Current() *TermKey {
	if i.cache == nil {
		cache := &TermKey{}
		_ = cache.UnmarshalBinary(i.iter.Item().Key())
		i.cache = cache
	}

	return i.cache
}

// Next advances the iterator and clears the cached value.
func (i *Iterator) Next() {
	i.iter.Next()
	i.cache = nil
}

// Valid ensures the iterator is still valid.
func (i *Iterator) Valid() bool {
	valid := i.iter.Valid()
	if !valid {
		return valid
	}

	key := i.iter.Item().Key()
	if len(i.end) > 0 {
		// 0 if a==b, -1 if a < b, and +1 if a > b
		v := bytes.Compare(key, i.end)

		return v < 0
	}

	return bytes.HasPrefix(key, i.prefix)
}

// Close closes the underlying iterator
func (i *Iterator) Close() error {
	i.iter.Close()
	return nil
}
