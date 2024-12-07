package youtube

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidTimestamp     = errors.New("invalid timestamp string")
	minutesSecondsRegex     = regexp.MustCompile(`^([0-5]?[0-9]):([0-5][0-9])$`)
	hoursMinutesSecondRegex = regexp.MustCompile(`^([0-9]|[01][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])$`)
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

// ParseTimestamp parses a string in the format "M:SS" or "MM:SS" or "H:MM:SS" or "HH:MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func ParseTimestamp(input string) (*Timestamp, error) {
	switch c := strings.Count(input, ":"); c {
	case 1:
		return parseMinutesSeconds(input)
	case 2:
		return parseHoursMinutesSeconds(input)
	default:
		return nil, ErrInvalidTimestamp
	}
}

// parseMinutesSeconds parses a string in the format "M:SS" or "MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func parseMinutesSeconds(input string) (*Timestamp, error) {
	matches := minutesSecondsRegex.FindStringSubmatch(input)
	if matches == nil {
		return nil, ErrInvalidTimestamp
	}

	// if matches, there will always be 3 elements
	// full match + 2 groups

	minute, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minute: %w", err)
	}

	second, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse second: %w", err)
	}

	return &Timestamp{Minute: minute, Second: second}, nil
}

// parseHoursMinutesSeconds parses a string in the format "H:MM:SS" or "HH:MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func parseHoursMinutesSeconds(input string) (*Timestamp, error) {
	matches := hoursMinutesSecondRegex.FindStringSubmatch(input)
	if matches == nil {
		return nil, ErrInvalidTimestamp
	}

	// if matches, there will always be 4 elements
	// full match + 3 groups
	hour, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to parse hour: %w", err)
	}

	minute, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse minute: %w", err)
	}

	second, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse second: %w", err)
	}

	return &Timestamp{Hour: hour, Minute: minute, Second: second}, nil
}
