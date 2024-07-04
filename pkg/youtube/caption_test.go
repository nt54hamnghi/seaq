package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractCaptionTracks(t *testing.T) {
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

	asserts := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			actual, err := extractCaptionTracks(body)

			asserts.Nil(err)
			asserts.Equal(actual, tt.want)
		})
	}
}

func TestExtractCaptionTracks_Fail(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(tt.body)
			_, err := extractCaptionTracks(body)
			asserts.Equal(err, tt.err)
		})
	}
}

func Ok(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestProcessCaptionTracks(t *testing.T) {
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
		asserts.True(strings.Contains(r.BaseURL, "fmt=json3"))

		if strings.Contains(r.BaseURL, "not-found") {
			asserts.False(r.HasAutoTranslation)
			asserts.False(strings.Contains(r.BaseURL, "tlang=en"))
		} else {
			asserts.True(r.HasAutoTranslation)
			asserts.True(strings.Contains(r.BaseURL, "tlang=en"))
		}
	}
}

func TestFetchCaption(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	var testCases = []struct {
		name         string
		captionTrack []captionTrack
		expected     caption
	}{
		{
			name: "valid",
			captionTrack: []captionTrack{
				{
					BaseURL:      server.URL + "?fmt=json3",
					LanguageCode: "en",
				},
			},
			expected: caption{Events: []event{
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
			expected: caption{Events: []event{
				{DDurationMs: 1249390, ID: 1},
			}},
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, err := fetchCaption(context.TODO(), tc.captionTrack)
			asserts.Nil(err)
			asserts.Equal(c, tc.expected)
		})
	}
}

func TestFetchCaption_Fail(t *testing.T) {
	var testCases = []struct {
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
		t.Run(tc.name, func(t *testing.T) {
			_, err := fetchCaption(context.TODO(), tc.captionTrack)
			asserts.Equal(err, tc.err)
		})
	}
}
