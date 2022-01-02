package iso8601_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mjpitz/homestead/internal/iso8601"
)

func TestPeriod(t *testing.T) {
	var period iso8601.Period

	err := (&period).UnmarshalText([]byte("2021-12-30T15:00:00+00:00/PT1H"))
	require.NoError(t, err)

	require.Equal(t, 2021, period.Time.Year())
	require.Equal(t, time.Month(12), period.Time.Month())
	require.Equal(t, 30, period.Time.Day())
	require.Equal(t, 15, period.Time.Hour())
	require.Equal(t, 0, period.Time.Minute())
	require.Equal(t, 0, period.Time.Second())

	require.Equal(t, "1h0m0s", period.Duration.String())
}
