package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVideoId(t *testing.T) {
	var testCases = []struct {
		name     string
		src      string
		err      error
		expected videoId
	}{
		{"validUrl", YouTubeWatchUrl + "?v=SL_YMm9C6tw", nil, "SL_YMm9C6tw"},
		{"validVid", "SL_YMm9C6tw", nil, "SL_YMm9C6tw"},
		{"tooShort", "short", ErrInvalidVideoId, ""},
		{"tooLong", "videoIdIsTooLong", ErrInvalidVideoId, ""},
		{"invalidChars", "!!!!!!!!!!!", ErrInvalidVideoId, ""},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vid, err := resolveVideoId(tc.src)
			if tc.err != nil {
				asserts.Equal(err, tc.err)
			} else {
				asserts.Equal(vid, tc.expected)
			}
		})
	}
}

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
