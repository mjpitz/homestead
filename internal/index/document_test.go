package index_test

import (
	"time"
)

// Document provides a generic document containing all the possible field types a document can have.
type Document struct {
	Time    time.Time
	String  string
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float64 float64
	Float32 float32
	Bool    bool
}
