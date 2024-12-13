package timestamp

import (
	"errors"
	"fmt"
	"iter"
	"regexp"
	"slices"
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

// IsZero returns true if the timestamp is zero (00:00:00).
func (ts Timestamp) IsZero() bool {
	return ts.Hour == 0 && ts.Minute == 0 && ts.Second == 0
}

// String returns string representation of the timestamp.
// Implement pflag.Value interface, to use with cobra.
func (ts *Timestamp) String() string {
	if ts.IsZero() {
		return ""
	}

	if ts.Hour > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", ts.Hour, ts.Minute, ts.Second)
	}

	return fmt.Sprintf("%02d:%02d", ts.Minute, ts.Second)
}

// Set parses the input string into a Timestamp.
// Implement pflag.Value interface, to use with cobra.
func (ts *Timestamp) Set(value string) error {
	if value == "" {
		return nil
	}

	parsed, err := ParseTimestamp(value)
	if err != nil {
		return err
	}

	*ts = parsed
	return nil
}

// Type returns the description of the flag type.
// Implement pflag.Value interface, to use with cobra.
func (ts *Timestamp) Type() string {
	return "timestamp (HH:MM:SS or MM:SS)"
}

// AsDuration converts the timestamp into a time.Duration.
func (ts Timestamp) AsDuration() time.Duration {
	hour := time.Duration(ts.Hour) * time.Hour
	minute := time.Duration(ts.Minute) * time.Minute
	second := time.Duration(ts.Second) * time.Second

	return hour + minute + second
}

// ParseTimestamp parses a string in the format "M:SS" or "MM:SS" or "H:MM:SS" or "HH:MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func ParseTimestamp(input string) (Timestamp, error) {
	switch c := strings.Count(input, ":"); c {
	case 1:
		return parseMinutesSeconds(input)
	case 2:
		return parseHoursMinutesSeconds(input)
	default:
		return Timestamp{}, ErrInvalidTimestamp
	}
}

// parseMinutesSeconds parses a string in the format "M:SS" or "MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func parseMinutesSeconds(input string) (Timestamp, error) {
	matches := minutesSecondsRegex.FindStringSubmatch(input)
	if matches == nil {
		return Timestamp{}, ErrInvalidTimestamp
	}

	// if matches, there will always be 3 elements
	// full match + 2 groups

	minute, err := strconv.Atoi(matches[1])
	if err != nil {
		return Timestamp{}, fmt.Errorf("failed to parse minute: %w", err)
	}

	second, err := strconv.Atoi(matches[2])
	if err != nil {
		return Timestamp{}, fmt.Errorf("failed to parse second: %w", err)
	}

	return Timestamp{Minute: minute, Second: second}, nil
}

// parseHoursMinutesSeconds parses a string in the format "H:MM:SS" or "HH:MM:SS" into a Timestamp.
// It returns an error if the input format is invalid or the values are out of range.
func parseHoursMinutesSeconds(input string) (Timestamp, error) {
	matches := hoursMinutesSecondRegex.FindStringSubmatch(input)
	if matches == nil {
		return Timestamp{}, ErrInvalidTimestamp
	}

	// if matches, there will always be 4 elements
	// full match + 3 groups
	hour, err := strconv.Atoi(matches[1])
	if err != nil {
		return Timestamp{}, fmt.Errorf("failed to parse hour: %w", err)
	}

	minute, err := strconv.Atoi(matches[2])
	if err != nil {
		return Timestamp{}, fmt.Errorf("failed to parse minute: %w", err)
	}

	second, err := strconv.Atoi(matches[3])
	if err != nil {
		return Timestamp{}, fmt.Errorf("failed to parse second: %w", err)
	}

	return Timestamp{Hour: hour, Minute: minute, Second: second}, nil
}

// AsDuration represents types that can be converted to time.Duration.
type AsDuration interface {
	AsDuration() time.Duration
}

// Before returns a new slice containing
// elements that occur before the specified timestamp.
func Before[D AsDuration](ts Timestamp, d []D) []D {
	return slices.Collect(KeepBefore(ts, TimeSequence(d)))
}

// After returns a new slice containing
// elements that occur after the specified timestamp.
func After[D AsDuration](ts Timestamp, d []D) []D {
	return slices.Collect(KeepAfter(ts, TimeSequence(d)))
}

// TimeSequence returns an iterator that yields elements in the slice.
func TimeSequence[D AsDuration](s []D) iter.Seq[D] {
	return func(yield func(D) bool) {
		for _, v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// KeepBefore returns a sequence with elements ending at or before the timestamp.
func KeepBefore[D AsDuration](limit Timestamp, seq iter.Seq[D]) iter.Seq[D] {
	return func(yield func(D) bool) {
		for v := range seq {
			if v.AsDuration() <= limit.AsDuration() {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// KeepAfter returns a sequence with elements starting at or after the timestamp.
func KeepAfter[D AsDuration](limit Timestamp, seq iter.Seq[D]) iter.Seq[D] {
	return func(yield func(D) bool) {
		for v := range seq {
			if v.AsDuration() >= limit.AsDuration() {
				if !yield(v) {
					return
				}
			}
		}
	}
}
