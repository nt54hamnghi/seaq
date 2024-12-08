package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/schema"
)

func Test_baseURL_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    string
		wantErr bool
	}{
		{
			name:    "valid",
			json:    `{"baseUrl":"https://www.youtube.com/watch?v=dQw4w9WgXcQ"}`,
			want:    "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			wantErr: false,
		},
		{
			name:    "not www.youtube.com",
			json:    `{"baseUrl":"https://www.example.com"}`,
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty",
			json:    "",
			want:    "",
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(*testing.T) {
			var u struct {
				BaseURL *baseURL `json:"baseUrl"`
			}
			err := json.Unmarshal([]byte(tc.json), &u)

			if tc.wantErr {
				r.Error(err)
			} else {
				r.NoError(err)
				r.Equal(tc.want, u.BaseURL.String())
			}
		})
	}
}

func Test_baseURL_setQuery(t *testing.T) {
	tests := []struct {
		name    string
		baseURL *baseURL
		key     string
		value   string
		want    string
	}{
		{
			name:    "add",
			baseURL: &baseURL{URL: url.URL{Scheme: "https", Host: "www.youtube.com"}},
			key:     "fmt",
			value:   "json3",
			want:    "fmt=json3",
		},
		{
			name:    "replace",
			baseURL: &baseURL{URL: url.URL{Scheme: "https", Host: "www.youtube.com", RawQuery: "tlang=es"}},
			key:     "tlang",
			value:   "en",
			want:    "tlang=en",
		},
	}

	r := require.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(*testing.T) {
			tc.baseURL.setQuery(tc.key, tc.value)

			r.Contains(tc.baseURL.String(), tc.want)
		})
	}
}

func Test_captionTrack_toEnglish(t *testing.T) {
	tests := []struct {
		name    string
		track   captionTrack
		wantErr error
	}{
		{
			name: "translatable",
			track: captionTrack{
				BaseURL:        &baseURL{URL: url.URL{Scheme: "https", Host: "www.youtube.com", Path: "/watch"}},
				IsTranslatable: true,
			},
			wantErr: nil,
		},
		{
			name:    "not translatable",
			track:   captionTrack{IsTranslatable: false},
			wantErr: errors.New("caption track is not translatable"),
		},
	}

	r := require.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(*testing.T) {
			err := tc.track.toEnglish()
			if tc.wantErr != nil {
				r.Equal(tc.wantErr, err)
			} else {
				r.NoError(err)
				r.Contains(tc.track.BaseURL.RawQuery, "tlang=en")
			}
		})
	}
}

func Test_extractCaptionTracks(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    []captionTrack
		wantErr string
	}{
		{
			name: "valid",
			body: `{"captionTracks":[{"baseUrl":"https://www.youtube.com","kind":"asr","languageCode":"en"}]}`,
			want: []captionTrack{
				{
					BaseURL: &baseURL{
						URL: url.URL{Scheme: "https", Host: "www.youtube.com"},
					},
					LanguageCode: "en",
					Kind:         "asr",
				},
			},
		},
		{
			name: "multiple tracks",
			body: `{"captionTracks":[
                {"baseUrl":"https://www.youtube.com","kind":"asr","languageCode":"en"},
                {"baseUrl":"https://www.youtube.com","kind":"asr","languageCode":"es"}
            ]}`,
			want: []captionTrack{
				{
					BaseURL: &baseURL{
						URL: url.URL{Scheme: "https", Host: "www.youtube.com"},
					},
					LanguageCode: "en",
					Kind:         "asr",
				},
				{
					BaseURL: &baseURL{
						URL: url.URL{Scheme: "https", Host: "www.youtube.com"},
					},
					LanguageCode: "es",
					Kind:         "asr",
				},
			},
		},
		{
			name:    "empty body",
			body:    "",
			wantErr: "caption tracks not found",
		},
		{
			name:    "malformed JSON",
			body:    `{"captionTracks":[INVALID]}`,
			wantErr: "malformed caption tracks",
		},
		{
			name:    "no caption tracks",
			body:    `{"something":"else"}`,
			wantErr: "caption tracks not found",
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			body := strings.NewReader(tt.body)
			actual, err := extractCaptionTracks(body)

			if tt.wantErr != "" {
				r.Error(err)
				r.Contains(err.Error(), tt.wantErr)
			} else {
				r.NoError(err)
				r.Equal(tt.want, actual)
			}
		})
	}
}

func Test_selectCaptionTrack(t *testing.T) {
	baseURL := &baseURL{URL: url.URL{Scheme: "https", Host: "www.youtube.com", Path: "/watch"}}

	tests := []struct {
		name    string
		tracks  []captionTrack
		want    *captionTrack
		wantErr bool
	}{
		{
			name: "manual captions",
			tracks: []captionTrack{
				{BaseURL: baseURL, VssID: ".en"},
				{BaseURL: baseURL, VssID: "a.en"},
			},
			want:    &captionTrack{BaseURL: baseURL, VssID: ".en"},
			wantErr: false,
		},
		{
			name: "auto-generated captions",
			tracks: []captionTrack{
				{BaseURL: baseURL, VssID: "a.en"},
				{BaseURL: baseURL, VssID: ".es"},
			},
			want:    &captionTrack{BaseURL: baseURL, VssID: "a.en"},
			wantErr: false,
		},
		{
			name: "translatable track",
			tracks: []captionTrack{
				{BaseURL: baseURL, VssID: ".fr", IsTranslatable: false},
				{BaseURL: baseURL, VssID: ".es", IsTranslatable: true},
			},
			want:    &captionTrack{BaseURL: baseURL, VssID: ".es", IsTranslatable: true},
			wantErr: false,
		},
		{
			name: "no suitable track",
			tracks: []captionTrack{
				{BaseURL: baseURL, VssID: ".es"},
			},
			want:    nil,
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := selectCaptionTrack(tt.tracks)
			if tt.wantErr {
				r.Error(err)
			} else {
				r.NoError(err)
				r.Equal(tt.want, got)
			}
		})
	}
}

func Test_loadCaption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"events":[{"dDurationMs":1249390,"id":1}]}`))
	})

	server := httptest.NewServer(mux)
	serverHost := server.Listener.Addr().String()
	defer server.Close()

	tests := []struct {
		name         string
		captionTrack []captionTrack
		want         caption
		wantErr      error
	}{
		{
			name: "valid",
			captionTrack: []captionTrack{
				{
					BaseURL: &baseURL{URL: url.URL{Scheme: "http", Host: serverHost}},
					VssID:   ".en",
				},
			},
			want: caption{Events: []event{
				{DDurationMs: 1249390, ID: 1},
			}},
			wantErr: nil,
		},
		{
			name:         "empty",
			captionTrack: []captionTrack{},
			want:         caption{},
			wantErr:      errors.New("caption tracks list is empty"),
		},
	}

	r := require.New(t)

	for _, tc := range tests {
		t.Run(tc.name, func(*testing.T) {
			c, err := loadCaption(context.TODO(), tc.captionTrack)

			if tc.wantErr != nil {
				r.Equal(tc.wantErr, err)
			} else {
				r.NoError(err)
				r.Equal(tc.want, c)
			}
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

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{Events: tt.events}
			c.filterStart(tt.start)

			a.Equal(tt.want, c.Events)
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

	a := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			c := &caption{Events: tt.events}
			c.filterEnd(tt.end)

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
