package timestamp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimestamp_AsDuration(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp Timestamp
		want      time.Duration
	}{
		{
			name:      "zero",
			timestamp: Timestamp{},
			want:      time.Duration(0),
		},
		{
			name:      "onlySecond",
			timestamp: Timestamp{Second: 22},
			want:      time.Second * 22,
		},
		{
			name:      "onlyMinute",
			timestamp: Timestamp{Minute: 22},
			want:      time.Minute * 22,
		},
		{
			name:      "onlyHour",
			timestamp: Timestamp{Hour: 22},
			want:      time.Hour * 22,
		},
		{
			name:      "combined",
			timestamp: Timestamp{Hour: 22, Minute: 22, Second: 22},
			want:      time.Hour*22 + time.Minute*22 + time.Second*22,
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := tt.timestamp.AsDuration()
			a.Equal(tt.want, got)
		})
	}
}

func Test_parseMinutesSeconds(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Timestamp
		wantErr error
	}{
		{
			name:  "valid MM:SS",
			input: "59:59",
			want:  Timestamp{Minute: 59, Second: 59},
		},
		{
			name:  "valid M:SS",
			input: "5:59",
			want:  Timestamp{Minute: 5, Second: 59},
		},
		{
			name:  "zero values",
			input: "0:00",
			want:  Timestamp{Minute: 0, Second: 0},
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

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := parseMinutesSeconds(tt.input)
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func Test_parseHoursMinutesSeconds(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Timestamp
		wantErr error
	}{
		{
			name:  "valid HH:MM:SS",
			input: "23:59:59",
			want:  Timestamp{Hour: 23, Minute: 59, Second: 59},
		},
		{
			name:  "valid H:MM:SS",
			input: "5:59:59",
			want:  Timestamp{Hour: 5, Minute: 59, Second: 59},
		},
		{
			name:  "zero values",
			input: "0:00:00",
			want:  Timestamp{Hour: 0, Minute: 0, Second: 0},
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

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := parseHoursMinutesSeconds(tt.input)
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    Timestamp
		wantErr error
	}{
		{
			name:  "one colon - valid",
			input: "59:59",
			want:  Timestamp{Minute: 59, Second: 59},
		},
		{
			name:  "two colons - valid",
			input: "23:59:59",
			want:  Timestamp{Hour: 23, Minute: 59, Second: 59},
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

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := ParseTimestamp(tt.input)
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

type second int64

func (t second) AsDuration() time.Duration {
	return time.Duration(t) * time.Second
}

func TestBefore(t *testing.T) {
	testCases := []struct {
		name string
		ts   Timestamp
		secs []second
		want []second
	}{
		{
			name: "empty sequence",
			ts:   Timestamp{Minute: 1},
			secs: []second{},
			want: nil,
		},
		{
			name: "all before",
			ts:   Timestamp{Second: 5},
			secs: []second{1, 2, 3},
			want: []second{1, 2, 3},
		},
		{
			name: "some before",
			ts:   Timestamp{Second: 2},
			secs: []second{1, 2, 3},
			want: []second{1, 2},
		},
		{
			name: "none before",
			ts:   Timestamp{Second: 3},
			secs: []second{4, 5, 6},
			want: nil,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := Before(tt.ts, tt.secs)
			r.Equal(tt.want, got)
		})
	}
}

func TestAfter(t *testing.T) {
	testCases := []struct {
		name string
		ts   Timestamp
		secs []second
		want []second
	}{
		{
			name: "empty sequence",
			ts:   Timestamp{Minute: 1},
			secs: []second{},
			want: nil,
		},
		{
			name: "all after",
			ts:   Timestamp{Second: 1},
			secs: []second{1, 2, 3},
			want: []second{1, 2, 3},
		},
		{
			name: "some after",
			ts:   Timestamp{Second: 2},
			secs: []second{1, 2, 3},
			want: []second{2, 3},
		},
		{
			name: "none after",
			ts:   Timestamp{Second: 5},
			secs: []second{1, 2, 3},
			want: nil,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := After(tt.ts, tt.secs)
			r.Equal(tt.want, got)
		})
	}
}

func TestTimestamp_String(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp Timestamp
		want      string
	}{
		{
			name:      "zero timestamp",
			timestamp: Timestamp{},
			want:      "",
		},
		{
			name:      "without hour",
			timestamp: Timestamp{Minute: 1, Second: 30},
			want:      "01:30",
		},
		{
			name:      "with hour",
			timestamp: Timestamp{Hour: 1, Minute: 1, Second: 30},
			want:      "01:01:30",
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got := tt.timestamp.String()
			a.Equal(tt.want, got)
		})
	}
}

func TestTimestamp_Set(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  Timestamp
	}{
		{
			name:  "empty string",
			input: "",
			want:  Timestamp{},
		},
		{
			name:  "normal",
			input: "01:02:03",
			want:  Timestamp{Hour: 1, Minute: 2, Second: 3},
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			var ts Timestamp
			err := ts.Set(tt.input)
			r.NoError(err)
			r.Equal(tt.want, ts)
		})
	}
}
