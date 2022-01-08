package index_test

import (
	"encoding"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mjpitz/homestead/internal/index"
)

func TestKeys(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name      string
		In        interface{}
		Out       interface{}
		Marshaled string
		Error     string
	}{
		{
			Name: "TermKey - int64",
			In: index.TermKey{
				Name:     "test",
				Term:     int64(0),
				Document: 1,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743ad300000000000000003a643a0000000000000001",
		},
		{
			Name:      "DocumentKey",
			In:        index.DocumentKey(100),
			Out:       new(index.DocumentKey),
			Marshaled: "643a0000000000000064",
		},
		{
			Name: "TermKey - float",
			In: index.TermKey{
				Name:     "test",
				Term:     float64(1.064243),
				Document: 100,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743acb3ff10723aafff36b3a643a0000000000000064",
		},
		{
			Name: "TermKey - float",
			In: index.TermKey{
				Name:     "test",
				Term:     float64(10.64243),
				Document: 1,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743acb402548ec95bff0453a643a0000000000000001",
		},
		{
			Name: "TermKey - float",
			In: index.TermKey{
				Name:     "test",
				Term:     float64(1.064243),
				Document: 100,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743acb3ff10723aafff36b3a643a0000000000000064",
		},
		{
			Name: "TermKey - int64",
			In: index.TermKey{
				Name:     "test",
				Term:     int64(10),
				Document: 1,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743ad3000000000000000a3a643a0000000000000001",
		},
		{
			Name: "TermKey - int64",
			In: index.TermKey{
				Name:     "test",
				Term:     int64(1000),
				Document: 100,
			},
			Out:       new(index.TermKey),
			Marshaled: "743a746573743ad300000000000003e83a643a0000000000000064",
		},
	}

	for _, testCase := range testCases {
		m, ok := testCase.In.(encoding.BinaryMarshaler)
		require.True(t, ok, "in not of type BinaryMarshaler")

		data, err := m.MarshalBinary()
		require.NoError(t, err)

		require.Equal(t, testCase.Marshaled, hex.EncodeToString(data))

		u, ok := testCase.Out.(encoding.BinaryUnmarshaler)
		require.True(t, ok, "out not of type BinaryUnmarshaler")

		err = u.UnmarshalBinary(data)
		require.NoError(t, err)

		// pass out back through the marshaler to ensure equality

		m, ok = testCase.Out.(encoding.BinaryMarshaler)
		require.True(t, ok, "out not of type BinaryMarshaler")

		data, err = m.MarshalBinary()
		require.NoError(t, err)

		require.Equal(t, testCase.Marshaled, hex.EncodeToString(data))
	}
}

func TestFields(t *testing.T) {
	t.Parallel()

	expected := index.TermSet{
		{Name: "Time", Term: time.Time{}},
		{Name: "String", Term: ""},
		{Name: "Int8", Term: int8(0)},
		{Name: "Int16", Term: int16(0)},
		{Name: "Int32", Term: int32(0)},
		{Name: "Int64", Term: int64(0)},
		{Name: "Uint8", Term: uint8(0)},
		{Name: "Uint16", Term: uint16(0)},
		{Name: "Uint32", Term: uint32(0)},
		{Name: "Uint64", Term: uint64(0)},
		{Name: "Float64", Term: float64(0)},
		{Name: "Float32", Term: float32(0)},
		{Name: "Bool", Term: false},
	}

	require.Equal(t, expected, index.Terms(&Document{}))
}

func TestSchema(t *testing.T) {
	t.Parallel()

	expected := []index.Field{
		{Text: "Time", Type: "time"},
		{Text: "String", Type: "string"},
		{Text: "Int8", Type: "number"},
		{Text: "Int16", Type: "number"},
		{Text: "Int32", Type: "number"},
		{Text: "Int64", Type: "number"},
		{Text: "Uint8", Type: "number"},
		{Text: "Uint16", Type: "number"},
		{Text: "Uint32", Type: "number"},
		{Text: "Uint64", Type: "number"},
		{Text: "Float64", Type: "number"},
		{Text: "Float32", Type: "number"},
		{Text: "Bool", Type: "number"},
	}

	require.Equal(t, expected, index.Schema(&Document{}))
}
