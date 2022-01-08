package index_test

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/require"

	"github.com/mjpitz/homestead/internal/index"
)

func TestIndex(t *testing.T) {
	docs := []interface{}{
		Document{String: "aab"},
		Document{String: "abb"},
		Document{String: "baa"},
		Document{String: "bac"},
		Document{String: "cab"},
	}

	idx, err := index.Open("", false)
	require.NoError(t, err)

	defer idx.Close()

	err = idx.Index(docs...)
	require.NoError(t, err)

	require.Len(t, idx.Schema(), 13)

	results := idx.Get(Document{}, 1, 2, 3)
	require.Len(t, results, 3)
	require.Equal(t, "abb", results[0].(*Document).String)
	require.Equal(t, "baa", results[1].(*Document).String)
	require.Equal(t, "bac", results[2].(*Document).String)

	testCases := []struct {
		Query   index.Query
		Results []uint64
	}{
		{index.Query{"String", "<=", "baa"}, []uint64{0x0, 0x1, 0x2}},
		{index.Query{"String", "=<", "baa"}, []uint64{0x0, 0x1, 0x2}},
		{index.Query{"String", "<", "baa"}, []uint64{0x0, 0x1}},
		{index.Query{"String", "=", "baa"}, []uint64{0x2}},
		{index.Query{"String", ">", "baa"}, []uint64{0x3, 0x4}},
		{index.Query{"String", "=>", "baa"}, []uint64{0x2, 0x3, 0x4}},
		{index.Query{"String", ">=", "baa"}, []uint64{0x2, 0x3, 0x4}},
	}

	for _, testCase := range testCases {
		t.Log(testCase.Query.Operator)
		require.Equal(t, testCase.Results, idx.Query(testCase.Query))
	}

	require.Equal(t, []uint64{0x2}, idx.Query(
		index.Query{"String", "=<", "baa"},
		index.Query{"String", "=>", "baa"},
	))
}

func TestIndexComp(t *testing.T) {
	clock := clockwork.NewFakeClockAt(time.Now())

	docs := []interface{}{
		Document{String: "aab"},
		Document{String: "abb"},
		Document{String: "baa"},
		Document{String: "bac"},
		Document{String: "cab"},
	}

	for i := 0; i < len(docs); i++ {
		doc := docs[i].(Document)
		doc.Time = clock.Now()
		clock.Advance(time.Second)
		docs[i] = doc
	}

	idx, err := index.Open("", false)
	require.NoError(t, err)

	defer idx.Close()

	err = idx.Index(docs...)
	require.NoError(t, err)

	start := docs[1].(Document).Time
	end := docs[4].(Document).Time

	require.Equal(t, []uint64{0x1, 0x2, 0x3}, idx.Query(
		index.Query{"Time", "<", end},
		index.Query{"Time", ">=", start},
	))
}
