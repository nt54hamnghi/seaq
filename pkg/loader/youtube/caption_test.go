package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func Test_extractCaptionTracks(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []captionTrack
	}{
		{
			name: "valid",
			body: `
			{
				"captionTracks": [
					{
						"baseUrl": "https://www.example.com",
						"kind": "asr",
						"languageCode": "en"
					}
				]
			}
			`,
			want: []captionTrack{
				{
					BaseURL:      "https://www.example.com",
					LanguageCode: "en",
					Kind:         "asr",
				},
			},
		},
	}

	requires := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			body := strings.NewReader(tt.body)
			actual, err := extractCaptionTracks(body)

			requires.NoError(err)
			requires.Equal(tt.want, actual)
		})
	}
}

func Test_extractCaptionTracks_Error(t *testing.T) {
	tests := []struct {
		name string
		body string
		err  error
	}{
		{
			name: "empty",
			body: "",
			err:  ErrCaptionTracksNotFound,
		},
	}

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			body := strings.NewReader(tt.body)
			_, err := extractCaptionTracks(body)
			asserts.Equal(err, tt.err)
		})
	}
}

func Ok(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func Test_processCaptionTracks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", Ok)
	mux.HandleFunc("/not-found", http.NotFound)

	server := httptest.NewServer(mux)
	defer server.Close()

	tracks := []captionTrack{
		{BaseURL: server.URL + "/ok"},
		{BaseURL: server.URL + "/not-found"},
	}

	processCaptionTracks(tracks)

	asserts := assert.New(t)
	for _, r := range tracks {
		asserts.Contains(r.BaseURL, "fmt=json3")

		if strings.Contains(r.BaseURL, "not-found") {
			asserts.False(r.HasAutoTranslation)
			asserts.NotContains(r.BaseURL, "tlang=en")
		} else {
			asserts.True(r.HasAutoTranslation)
			asserts.Contains(r.BaseURL, "tlang=en")
		}
	}
}

func Test_loadCaption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// nolint: errcheck
		w.Write([]byte(`
		{
			"events": [
				{
					"dDurationMs": 1249390,
					"id": 1
				}
			]
		}
		`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	testCases := []struct {
		name         string
		captionTrack []captionTrack
		want         caption
	}{
		{
			name: "valid",
			captionTrack: []captionTrack{
				{
					BaseURL:      server.URL + "?fmt=json3",
					LanguageCode: "en",
				},
			},
			want: caption{Events: []event{
				{DDurationMs: 1249390, ID: 1},
			}},
		},
		{
			name: "hasTranslation",
			captionTrack: []captionTrack{
				{
					BaseURL:            server.URL + "?fmt=json3&tlang=en",
					LanguageCode:       "en",
					HasAutoTranslation: true,
				},
			},
			want: caption{Events: []event{
				{DDurationMs: 1249390, ID: 1},
			}},
		},
	}

	requires := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			c, err := loadCaption(context.TODO(), tc.captionTrack)
			requires.NoError(err)
			requires.Equal(tc.want, c)
		})
	}
}

func Test_loadCaption_Error(t *testing.T) {
	testCases := []struct {
		name         string
		captionTrack []captionTrack
		err          error
	}{
		{
			name:         "empty",
			captionTrack: []captionTrack{},
			err:          errors.New("caption tracks must not be empty"),
		},
		{
			name: "noEnglish",
			captionTrack: []captionTrack{
				{LanguageCode: "es"},
				{LanguageCode: "en", Kind: "unknown"},
			},
			err: errors.New("no English caption track found"),
		},
		{
			name: "noEnglishTranslation",
			captionTrack: []captionTrack{
				{LanguageCode: "es", HasAutoTranslation: false},
			},
			err: errors.New("no English caption track found"),
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			_, err := loadCaption(context.TODO(), tc.captionTrack)
			asserts.Equal(err, tc.err)
		})
	}
}

func Test_caption_filterStart(t *testing.T) {
	testCases := []struct {
		name   string
		events []event
		start  *Timestamp
		want   []event
	}{
		{
			name:   "empty",
			events: []event{},
			start:  &Timestamp{Second: 1},
			want:   []event{},
		},
		{
			name: "noStart",
			events: []event{
				{ID: 1, TStartMs: 0},
			},
			start: nil,
			want: []event{
				{ID: 1, TStartMs: 0},
			},
		},
		{
			name: "valid",
			events: []event{
				{ID: 1, TStartMs: 0},
				{ID: 2, TStartMs: 1000},
				{ID: 3, TStartMs: 2000},
			},
			start: &Timestamp{Second: 1},
			want: []event{
				{ID: 2, TStartMs: 1000},
				{ID: 3, TStartMs: 2000},
			},
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{Events: tt.events}
			c.filterStart(tt.start)

			asserts.Equal(tt.want, c.Events)
		})
	}
}

func Test_caption_filterEnd(t *testing.T) {
	testCases := []struct {
		name   string
		events []event
		end    *Timestamp
		want   []event
	}{
		{
			name:   "empty",
			events: []event{},
			end:    &Timestamp{Second: 1},
			want:   []event{},
		},
		{
			name: "noStart",
			events: []event{
				{ID: 1, TStartMs: 0},
			},
			end: nil,
			want: []event{
				{ID: 1, TStartMs: 0},
			},
		},
		{
			name: "valid",
			events: []event{
				{ID: 1, TStartMs: 0},
				{ID: 2, TStartMs: 1000},
				{ID: 3, TStartMs: 10000},
			},
			end: &Timestamp{Second: 1},
			want: []event{
				{ID: 1, TStartMs: 0},
				{ID: 2, TStartMs: 1000},
			},
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{Events: tt.events}
			c.filterEnd(tt.end)

			asserts.Equal(tt.want, c.Events)
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

	requires := require.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := tt.event.toDocument()
			requires.NoError(err)
			requires.Equal("hello world", got.PageContent)
			requires.Equal(int64(240), got.Metadata["startMs"])
			requires.Equal(int64(4880), got.Metadata["durationMs"])
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

	asserts := assert.New(t)
	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			_, err := tt.event.toDocument()
			asserts.Equal(tt.wantErr, err)
		})
	}
}
