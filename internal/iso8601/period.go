package iso8601

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Period struct {
	Time     time.Time
	Duration time.Duration
}

func (t *Period) UnmarshalJSON(data []byte) (err error) {
	return t.UnmarshalText(data[1 : len(data)-1])
}

func (t *Period) UnmarshalText(text []byte) (err error) {
	if tzPos := bytes.IndexByte(text, '+'); tzPos > -1 {
		// replace '+' with 'Z'
		text[tzPos] = 'Z'
	} else if tzNeg := bytes.IndexByte(text, '-'); tzNeg > -1 {
		// insert 'Z' before '-'
		text = append(text[:tzNeg], append([]byte{'Z'}, text[tzNeg:]...)...)
	}

	parts := bytes.SplitN(text, []byte{'/'}, 2)

	timePortion := parts[0]
	timePortion = bytes.TrimSuffix(timePortion, []byte("00:00"))

	t.Time, err = time.Parse(time.RFC3339, string(timePortion))
	t.Duration, err = ParseDuration(string(parts[1]))

	return err
}

func ParseDuration(str string) (time.Duration, error) {
	val := []byte(str)
	if val[0] != 'P' {
		return 0, fmt.Errorf("invalid duration")
	}

	parts := bytes.SplitN(val[1:], []byte{'T'}, 2)
	duration := time.Duration(0)

	matcher := regexp.MustCompile("(?P<measure>\\d+)(?P<unit>[YMDHSymdhs])")
	{
		matches := matcher.FindAllSubmatch(parts[0], -1)

		for _, match := range matches {
			i, err := strconv.ParseInt(string(match[1]), 10, 32)
			if err != nil {
				return 0, err
			}

			base := time.Duration(i)
			switch string(match[2]) {
			case "Y", "y":
				duration += base * 365 * 24 * time.Hour
			case "M", "m":
				duration += base * 30 * 24 * time.Hour
			case "D", "d":
				duration += base * 24 * time.Hour
			default:
				return 0, fmt.Errorf("unrecognized symbol: %s", match[i])
			}
		}
	}

	{
		matches := matcher.FindAllSubmatch(parts[1], -1)

		for _, match := range matches {
			i, err := strconv.ParseInt(string(match[1]), 10, 32)
			if err != nil {
				return 0, err
			}

			base := time.Duration(i)
			switch string(match[2]) {
			case "H", "h":
				duration += base * time.Hour
			case "M", "m":
				duration += base * time.Minute
			case "S", "s":
				duration += base * time.Second
			default:
				return 0, fmt.Errorf("unrecognized symbol: %s", match[i])
			}
		}
	}

	return duration, nil
}
