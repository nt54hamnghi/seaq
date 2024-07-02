package youtube

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractVideoId(t *testing.T) {
	var testCases = []struct {
		name     string
		url      string
		expected string
		err      error
	}{
		{
			name:     "valid",
			url:      YouTubeWatchUrl + "?v=SL_YMm9C6tw",
			expected: "SL_YMm9C6tw",
		},
		{
			name:     "multipleParams",
			url:      YouTubeWatchUrl + "?v=SL_YMm9C6tw&t=5s",
			expected: "SL_YMm9C6tw",
		},
		{
			name:     "multipleVids",
			url:      YouTubeWatchUrl + "?v=SL_YMm9C6tw&v=12345678910",
			expected: "SL_YMm9C6tw",
		},
		{
			name: "empty",
			url:  "",
			err:  ErrInValidYouTubeURL,
		},
		{
			name: "missingVid",
			url:  YouTubeWatchUrl + "?t=5s",
			err:  ErrVideoIdNotFoundInURL,
		},
		{
			name: "noValue",
			url:  YouTubeWatchUrl + "?v=",
			err:  ErrVideoIdNotFoundInURL,
		},
		{
			name: "invalid",
			url:  "https://www.google.com/watch?v=SL_YMm9C6tw&v=12345678910",
			err:  ErrInValidYouTubeURL,
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractVideoId(tc.url)
			if tc.err != nil {
				asserts.Equal(err, tc.err)
			} else {
				asserts.Equal(actual, tc.expected)
			}
		})
	}
}

func TestExtractCaptionTracks(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected []captionTrack
		err      error
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
			expected: []captionTrack{
				{
					BaseURL:      "https://www.example.com",
					LanguageCode: "en",
					Kind:         "asr",
				},
			},
		},
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
			actual, err := extractCaptionTracks(body)

			if err != nil {
				asserts.Equal(err, tt.err)
			} else {
				asserts.Equal(actual, tt.expected)
			}
		})
	}
}

func Ok(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestProcessCaptionTracks(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", Ok)
	// mux.HandleFunc("/not-found", http.NotFound)

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
