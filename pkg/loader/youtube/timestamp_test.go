package youtube

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_ToMsDuration(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp Timestamp
		want      int64
	}{
		{
			name:      "zero",
			timestamp: Timestamp{},
			want:      0,
		},
		{
			name:      "onlySecond",
			timestamp: Timestamp{Second: 22},
			want:      (time.Second * 22).Milliseconds(),
		},
		{
			name:      "onlyMinute",
			timestamp: Timestamp{Minute: 22},
			want:      (time.Minute * 22).Milliseconds(),
		},
		{
			name:      "onlyHour",
			timestamp: Timestamp{Hour: 22},
			want:      (time.Hour * 22).Milliseconds(),
		},
		{
			name:      "combined",
			timestamp: Timestamp{Hour: 22, Minute: 22, Second: 22},
			want:      80542000,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			msDuration := tt.timestamp.ToMsDuration()
			asserts.Equal(tt.want, msDuration)
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  *Timestamp
	}{
		{
			name:  "minute:second",
			input: "11:11",
			want:  &Timestamp{Minute: 11, Second: 11},
		},
		{
			name:  "hour:minute:second",
			input: "23:59:59",
			want:  &Timestamp{Hour: 23, Minute: 59, Second: 59},
		},
	}

	requires := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := ParseTimestamp(tt.input)
			requires.NoError(err)
			requires.NotNil(got)
			requires.Equal(*tt.want, *got)
		})
	}
}

func TestParseTimestamp_Error(t *testing.T) {
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
		t.Run(tt.name, func(*testing.T) {
			_, err := ParseTimestamp(tt.input)
			asserts.Equal(err, errors.New("invalid timestamp string"))
		})
	}
}
