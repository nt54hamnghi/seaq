package youtube

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveVideoId(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want videoID
	}{
		{
			name: "validVid",
			src:  "SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "validUrl",
			src:  YouTubeWatchURL + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			vid, err := ResolveVideoID(tc.src)
			r.NoError(err)
			r.Equal(tc.want, vid)
		})
	}
}

func TestResolveVideoId_Error(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		err  error
	}{
		{name: "tooShort", src: "short", err: ErrInvalidYouTubeURL},
		{name: "tooLong", src: "looooooooong", err: ErrInvalidYouTubeURL},
		{name: "invalidChars", src: "!!!!!!!!!!!", err: ErrInvalidYouTubeURL},
	}

	a := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			_, err := ResolveVideoID(tc.src)
			a.Equal(err, tc.err)
		})
	}
}

func Test_extractVideoId(t *testing.T) {
	testCases := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid",
			url:  YouTubeWatchURL + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "multipleParams",
			url:  YouTubeWatchURL + "?v=SL_YMm9C6tw&t=5s",
			want: "SL_YMm9C6tw",
		},
		{
			name: "multipleVids",
			url:  YouTubeWatchURL + "?v=SL_YMm9C6tw&v=12345678910",
			want: "SL_YMm9C6tw",
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			actual, err := extractVideoID(tc.url)
			r.NoError(err)
			r.Equal(tc.want, actual)
		})
	}
}

func Test_extractVideoId_Error(t *testing.T) {
	testCases := []struct {
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
			url:  YouTubeWatchURL + "?t=5s",
			err:  ErrVideoIDNotFoundInURL,
		},
		{
			name: "noValue",
			url:  YouTubeWatchURL + "?v=",
			err:  ErrVideoIDNotFoundInURL,
		},
		{
			name: "videoIdTooShort",
			url:  YouTubeWatchURL + "?v=short",
			err:  ErrInvalidVideoID,
		},
		{
			name: "videoIdTooLong",
			url:  YouTubeWatchURL + "?v=looooooooong",
			err:  ErrInvalidVideoID,
		},
		{
			name: "invalidChars",
			url:  YouTubeWatchURL + "?v=!!!!!!!!!!!",
			err:  ErrInvalidVideoID,
		},
		{
			name: "invalidUrl",
			url:  "https://www.google.com/watch?v=SL_YMm9C6tw&v=12345678910",
			err:  ErrInvalidYouTubeURL,
		},
	}

	a := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			_, err := extractVideoID(tc.url)
			a.Equal(err, tc.err)
		})
	}
}
