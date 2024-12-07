package youtube

import (
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

func Test_parseMinutesSeconds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Timestamp
		wantErr error
	}{
		{
			name:  "valid MM:SS",
			input: "59:59",
			want:  &Timestamp{Minute: 59, Second: 59},
		},
		{
			name:  "valid M:SS",
			input: "5:59",
			want:  &Timestamp{Minute: 5, Second: 59},
		},
		{
			name:  "zero values",
			input: "0:00",
			want:  &Timestamp{Minute: 0, Second: 0},
		},
		{
			name:    "invalid minutes too high",
			input:   "60:00",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid seconds too high",
			input:   "59:60",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format no colon",
			input:   "5959",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format empty",
			input:   "",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format too many colons",
			input:   "59:59:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid single digit seconds",
			input:   "59:5",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid leading colon",
			input:   ":59:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid trailing colon",
			input:   "59:59:",
			wantErr: ErrInvalidTimestamp,
		},
	}

	requires := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := parseMinutesSeconds(tt.input)
			if tt.wantErr != nil {
				requires.Equal(tt.wantErr, err)
				requires.Nil(got)
				return
			}

			requires.NoError(err)
			requires.Equal(tt.want, got)
		})
	}
}

func Test_parseHoursMinutesSeconds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Timestamp
		wantErr error
	}{
		{
			name:  "valid HH:MM:SS",
			input: "23:59:59",
			want:  &Timestamp{Hour: 23, Minute: 59, Second: 59},
		},
		{
			name:  "valid H:MM:SS",
			input: "5:59:59",
			want:  &Timestamp{Hour: 5, Minute: 59, Second: 59},
		},
		{
			name:  "zero values",
			input: "0:00:00",
			want:  &Timestamp{Hour: 0, Minute: 0, Second: 0},
		},
		{
			name:    "invalid hours too high",
			input:   "24:00:00",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid minutes too high",
			input:   "23:60:00",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid seconds too high",
			input:   "23:59:60",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format no colons",
			input:   "235959",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format empty",
			input:   "",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format single colon",
			input:   "23:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid format too many colons",
			input:   "23:59:59:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid single digit minutes",
			input:   "23:5:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid single digit seconds",
			input:   "23:59:5",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid leading colon",
			input:   ":23:59:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "invalid trailing colon",
			input:   "23:59:59:",
			wantErr: ErrInvalidTimestamp,
		},
	}

	requires := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := parseHoursMinutesSeconds(tt.input)
			if tt.wantErr != nil {
				requires.Equal(tt.wantErr, err)
				requires.Nil(got)
				return
			}

			requires.NoError(err)
			requires.Equal(tt.want, got)
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Timestamp
		wantErr error
	}{
		{
			name:  "one colon - valid",
			input: "59:59",
			want:  &Timestamp{Minute: 59, Second: 59},
		},
		{
			name:  "two colons - valid",
			input: "23:59:59",
			want:  &Timestamp{Hour: 23, Minute: 59, Second: 59},
		},
		{
			name:    "no colons",
			input:   "2359",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "too many colons",
			input:   "23::59::59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "leading colon",
			input:   ":23:59:59",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "trailing colon",
			input:   "23:59:59:",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrInvalidTimestamp,
		},
		{
			name:    "just colons",
			input:   ":::",
			wantErr: ErrInvalidTimestamp,
		},
	}

	requires := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := ParseTimestamp(tt.input)
			if tt.wantErr != nil {
				requires.Equal(tt.wantErr, err)
				requires.Nil(got)
				return
			}

			requires.NoError(err)
			requires.Equal(tt.want, got)
		})
	}
}
