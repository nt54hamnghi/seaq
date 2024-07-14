package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_resolveVideoId(t *testing.T) {
	var testCases = []struct {
		name string
		src  string
		want videoId
	}{
		{name: "validUrl", src: YouTubeWatchUrl + "?v=SL_YMm9C6tw", want: "SL_YMm9C6tw"},
		{name: "validVid", src: "SL_YMm9C6tw", want: "SL_YMm9C6tw"},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vid, err := resolveVideoId(tc.src)
			asserts.Nil(err)
			asserts.Equal(vid, tc.want)
		})
	}
}

func Test_resolveVideoId_Error(t *testing.T) {
	var testCases = []struct {
		name string
		src  string
		err  error
	}{
		{name: "tooShort", src: "short", err: ErrInvalidVideoId},
		{name: "tooLong", src: "videoIdIsTooLong", err: ErrInvalidVideoId},
		{name: "invalidChars", src: "!!!!!!!!!!!", err: ErrInvalidVideoId},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := resolveVideoId(tc.src)
			asserts.Equal(err, tc.err)
		})
	}
}

func Test_extractVideoId(t *testing.T) {
	var testCases = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid",
			url:  YouTubeWatchUrl + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "multipleParams",
			url:  YouTubeWatchUrl + "?v=SL_YMm9C6tw&t=5s",
			want: "SL_YMm9C6tw",
		},
		{
			name: "multipleVids",
			url:  YouTubeWatchUrl + "?v=SL_YMm9C6tw&v=12345678910",
			want: "SL_YMm9C6tw",
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractVideoId(tc.url)
			asserts.Nil(err)
			asserts.Equal(actual, tc.want)
		})
	}
}
func Test_extractVideoId_Error(t *testing.T) {
	var testCases = []struct {
		name string
		url  string
		err  error
	}{
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
			_, err := extractVideoId(tc.url)
			asserts.Equal(err, tc.err)
		})
	}
}
