package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveVideoId(t *testing.T) {
	var testCases = []struct {
		name string
		src  string
		want videoId
	}{
		{
			name: "validVid",
			src:  "SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "validUrl",
			src:  YouTubeWatchUrl + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vid, err := ResolveVideoId(tc.src)
			asserts.Nil(err)
			asserts.Equal(vid, tc.want)
		})
	}
}

func TestResolveVideoId_Error(t *testing.T) {
	var testCases = []struct {
		name string
		src  string
		err  error
	}{
		{name: "tooShort", src: "short", err: ErrInvalidYouTubeURL},
		{name: "tooLong", src: "looooooooong", err: ErrInvalidYouTubeURL},
		{name: "invalidChars", src: "!!!!!!!!!!!", err: ErrInvalidYouTubeURL},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ResolveVideoId(tc.src)
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
			err:  ErrInvalidYouTubeURL,
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
			name: "videoIdTooShort",
			url:  YouTubeWatchUrl + "?v=short",
			err:  ErrInvalidVideoId,
		},
		{
			name: "videoIdTooLong",
			url:  YouTubeWatchUrl + "?v=looooooooong",
			err:  ErrInvalidVideoId,
		},
		{
			name: "invalidChars",
			url:  YouTubeWatchUrl + "?v=!!!!!!!!!!!",
			err:  ErrInvalidVideoId,
		},
		{
			name: "invalidUrl",
			url:  "https://www.google.com/watch?v=SL_YMm9C6tw&v=12345678910",
			err:  ErrInvalidYouTubeURL,
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
