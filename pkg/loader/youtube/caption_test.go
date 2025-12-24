package youtube

import (
	"errors"
	"testing"

	"github.com/nt54hamnghi/seaq/pkg/util/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func Test_caption_filter(t *testing.T) {
	events := []event{
		{ID: 1, TStartMs: 0},
		{ID: 2, TStartMs: 1000},
		{ID: 3, TStartMs: 2000},
		{ID: 4, TStartMs: 10000},
	}

	testCases := []struct {
		name       string
		filterOpts *filter
		want       []event
	}{
		{
			name:       "nil filter",
			filterOpts: nil,
			want:       events,
		},
		{
			name: "valid start",
			filterOpts: &filter{
				start: timestamp.Timestamp{Second: 2},
			},
			want: []event{
				{ID: 3, TStartMs: 2000},
				{ID: 4, TStartMs: 10000},
			},
		},
		{
			name: "valid end",
			filterOpts: &filter{
				end: timestamp.Timestamp{Second: 1},
			},
			want: []event{
				{ID: 1, TStartMs: 0},
				{ID: 2, TStartMs: 1000},
			},
		},
		{
			name: "valid start and end",
			filterOpts: &filter{
				start: timestamp.Timestamp{Second: 1},
				end:   timestamp.Timestamp{Second: 2},
			},
			want: []event{
				{ID: 2, TStartMs: 1000},
				{ID: 3, TStartMs: 2000},
			},
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{Events: events}
			c.filter(tt.filterOpts)

			a.Equal(tt.want, c.Events)
		})
	}
}

func Test_event_toDocument(t *testing.T) {
	testCases := []struct {
		name  string
		event event
		want  schema.Document
	}{
		{
			name: "valid",
			event: event{
				Segs: []segment{
					{Utf8: "hello"},
					{Utf8: " world"},
				},
				TStartMs:    240,
				DDurationMs: 4880,
			},
		},
		{
			name: "withEmptySegment",
			event: event{
				Segs: []segment{
					{Utf8: "hello"},
					{Utf8: ""},
					{Utf8: " world"},
				},
				TStartMs:    240,
				DDurationMs: 4880,
			},
		},
	}

	r := require.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := tt.event.toDocument()
			r.NoError(err)
			r.Equal("hello world", got.PageContent)
			r.Equal(int64(240), got.Metadata["startMs"])
			r.Equal(int64(4880), got.Metadata["durationMs"])
		})
	}
}

func Test_event_toDocument_Error(t *testing.T) {
	testCases := []struct {
		name    string
		event   event
		wantErr error
	}{
		{
			name: "noSegment",
			event: event{
				Segs:        []segment{},
				TStartMs:    240,
				DDurationMs: 4880,
			},
			wantErr: errors.New("no segments found"),
		},
	}

	a := assert.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			_, err := tt.event.toDocument()
			a.Equal(tt.wantErr, err)
		})
	}
}
