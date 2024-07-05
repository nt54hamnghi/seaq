package youtube

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimestamp_ToMsDuration(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp Timestamp
		expected  int64
	}{
		{
			name:      "zero",
			timestamp: Timestamp{},
			expected:  0,
		},
		{
			name:      "onlySecond",
			timestamp: Timestamp{Second: 22},
			expected:  (time.Second * 22).Milliseconds(),
		},
		{
			name:      "onlyMinute",
			timestamp: Timestamp{Minute: 22},
			expected:  (time.Minute * 22).Milliseconds(),
		},
		{
			name:      "onlyHour",
			timestamp: Timestamp{Hour: 22},
			expected:  (time.Hour * 22).Milliseconds(),
		},
		{
			name:      "combined",
			timestamp: Timestamp{Hour: 22, Minute: 22, Second: 22},
			expected:  80542000,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			msDuration := tt.timestamp.ToMsDuration()
			asserts.Equal(tt.expected, msDuration)
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Timestamp
	}{
		{
			name:  "minute:second",
			input: "11:11",
			want:  Timestamp{Minute: 11, Second: 11},
		},
		{
			name:  "hour:minute:second",
			input: "23:59:59",
			want:  Timestamp{Hour: 23, Minute: 59, Second: 59},
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestamp(tt.input)
			asserts.Nil(err)
			asserts.Equal(tt.want, got)
		})
	}
}

func TestParseTimestamp_Fail(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty",
			input: "",
		},
		{
			name:  "invalid",
			input: "not-a-timestamp",
		},
		{
			name:  "leadingColon",
			input: ":22:22",
		},
		{
			name:  "trailingColon",
			input: "22:22:",
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTimestamp(tt.input)
			asserts.Equal(err, errors.New("invalid timestamp string"))
		})
	}
}
