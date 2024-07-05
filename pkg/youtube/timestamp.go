package youtube

import (
	"errors"
	"regexp"
	"strconv"
	"time"
)

type Timestamp struct {
	Hour   int
	Minute int
	Second int
}

func (ts Timestamp) ToMsDuration() int64 {
	hour := int64(ts.Hour) * time.Hour.Milliseconds()
	minute := int64(ts.Minute) * time.Minute.Milliseconds()
	second := int64(ts.Second) * time.Second.Milliseconds()

	return hour + minute + second
}

func ParseTimestamp(input string) (Timestamp, error) {

	re := regexp.MustCompile(`^(?:([01]\d|2[0-3]):)?([0-5]\d):([0-5]\d)$`)
	matches := re.FindStringSubmatch(input)
	if matches == nil {
		return Timestamp{}, errors.New("invalid timestamp string")
	}

	// due to the regex, matches will always have at least 3 and at most 4 elements
	// matches[0] is always the full match
	switch len(matches) {
	case 3:
		minute, _ := strconv.Atoi(matches[1])
		second, _ := strconv.Atoi(matches[2])
		return Timestamp{Minute: minute, Second: second}, nil
	case 4:
		hour, _ := strconv.Atoi(matches[1])
		minute, _ := strconv.Atoi(matches[2])
		second, _ := strconv.Atoi(matches[3])
		return Timestamp{Hour: hour, Minute: minute, Second: second}, nil
	default:
		panic("unreachable")
	}
}
