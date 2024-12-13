package udemy

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func Test_parseUdemyURL(t *testing.T) {
	testCases := []struct {
		name       string
		rawURL     string
		wantCourse string
		wantID     int
		wantErr    bool
	}{
		{
			name:       "valid URL",
			rawURL:     "https://www.udemy.com/course/test-course/learn/lecture/12345",
			wantCourse: "test-course",
			wantID:     12345,
			wantErr:    false,
		},
		{
			name:    "invalid Host",
			rawURL:  "https://www.example.com/course/test-course/learn/lecture/12345",
			wantErr: true,
		},
		{
			name:    "non-numeric lectureID",
			rawURL:  "https://www.udemy.com/course/test-course/learn/lecture/abc",
			wantErr: true,
		},
		{
			name:    "missing lectureID",
			rawURL:  "https://www.udemy.com/course/test-course/learn/lecture/",
			wantErr: true,
		},
		{
			name:    "missing courseName",
			rawURL:  "https://www.udemy.com/course//learn/lecture/12345",
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			courseName, lectureID, err := parseUdemyURL(tt.rawURL)
			if tt.wantErr {
				r.Error(err)
				return
			}

			r.NoError(err)
			r.Equal(tt.wantCourse, courseName)
			r.Equal(tt.wantID, lectureID)
		})
	}
}

func Test_assetType_MarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		a       assetType
		want    string
		wantErr error
	}{
		{
			name: "video",
			a:    video,
			want: `"Video"`,
		},
		{
			name: "article",
			a:    article,
			want: `"Article"`,
		},
		{
			name:    "unknown",
			a:       assetType(999),
			wantErr: errors.New("unknown asset type"),
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := tt.a.MarshalJSON()
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.JSONEq(tt.want, string(got))
		})
	}
}

func Test_assetType_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name    string
		data    string
		want    assetType
		wantErr error
	}{
		{
			name: "video",
			data: `"Video"`,
			want: video,
		},
		{
			name: "article",
			data: `"Article"`,
			want: article,
		},
		{
			name:    "unknown",
			data:    `"Unknown"`,
			wantErr: errors.New(`unsupported asset type: "Unknown"`),
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			var got assetType
			err := got.UnmarshalJSON([]byte(tt.data))
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func Test_lecture_findCaption(t *testing.T) {
	testCases := []struct {
		name      string
		lecture   lecture
		localeIDs []string
		want      caption
		wantErr   error
	}{
		{
			name: "caption found",
			lecture: lecture{
				Asset: asset{
					Captions: []caption{
						{LocaleID: "en_US"},
						{LocaleID: "fr_FR"},
					},
				},
			},
			localeIDs: []string{"en_US", "en_GB"},
			want:      caption{LocaleID: "en_US"},
		},
		{
			name: "caption not found",
			lecture: lecture{
				Asset: asset{
					Captions: []caption{
						{LocaleID: "fr_FR"},
					},
				},
			},
			localeIDs: []string{"en_US"},
			wantErr:   fmt.Errorf("no caption found for locale IDs: %v", []string{"en_US"}),
		},
		{
			name: "no captions available",
			lecture: lecture{
				Asset: asset{
					Captions: []caption{},
				},
			},
			localeIDs: []string{"en_US"},
			wantErr:   fmt.Errorf("no caption found for locale IDs: %v", []string{"en_US"}),
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := tt.lecture.findCaption(tt.localeIDs...)

			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func Test_event_toDocument(t *testing.T) {
	testCases := []struct {
		name    string
		item    *astisub.Item
		want    schema.Document
		wantErr error
	}{
		{
			name:    "nil",
			item:    nil,
			wantErr: errors.New("nil event"),
		},
		{
			name: "valid",
			item: &astisub.Item{
				StartAt: 1000 * time.Millisecond,
				EndAt:   5000 * time.Millisecond,
				Lines: []astisub.Line{
					{Items: []astisub.LineItem{{Text: "Hello, world!"}}},
				},
				Comments: []string{"Test comment"},
			},
			want: schema.Document{
				PageContent: "Hello, world!",
				Metadata: map[string]any{
					"Comment": []string{"Test comment"},
					"StartAt": 1000 * time.Millisecond,
					"EndAt":   5000 * time.Millisecond,
				},
				Score: 0,
			},
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			event := event{Item: tt.item}
			got, err := event.toDocument()
			if tt.wantErr != nil {
				r.Equal(tt.wantErr, err)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func Test_caption_filter(t *testing.T) {
	events := []event{
		{&astisub.Item{StartAt: 0}},
		{&astisub.Item{StartAt: time.Second * 1}},
		{&astisub.Item{StartAt: time.Second * 2}},
		{&astisub.Item{StartAt: time.Second * 10}},
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
				events[2], // 2 seconds
				events[3], // 10 seconds
			},
		},
		{
			name: "valid end",
			filterOpts: &filter{
				end: timestamp.Timestamp{Second: 1},
			},
			want: []event{
				events[0], // 0 seconds
				events[1], // 1 second
			},
		},
		{
			name: "valid start and end",
			filterOpts: &filter{
				start: timestamp.Timestamp{Second: 1},
				end:   timestamp.Timestamp{Second: 2},
			},
			want: []event{
				events[1], // 1 second
				events[2], // 2 seconds
			},
		},
	}

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{events: events}
			c.filter(tt.filterOpts)

			a.Equal(tt.want, c.events)
		})
	}
}
