package youtube

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveVideoId(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want videoID
	}{
		{
			name: "directVideoId",
			src:  "dQw4w9WgXcQ",
			want: "dQw4w9WgXcQ",
		},
		{
			name: "watchUrl",
			src:  YouTubeWatchURL + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "shortsUrl",
			src:  YouTubeShortURL + "/6hz4edk5uh0",
			want: "6hz4edk5uh0",
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

func Test_extractVideoId(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		want    string
		wantErr error
	}{
		{
			name: "watch",
			url:  YouTubeWatchURL + "?v=SL_YMm9C6tw",
			want: "SL_YMm9C6tw",
		},
		{
			name: "shorts",
			url:  YouTubeShortURL + "/6hz4edk5uh0",
			want: "6hz4edk5uh0",
		},
		{
			name:    "invalidHost",
			url:     "https://www.google.com",
			wantErr: ErrInvalidYouTubeURL,
		},
		{
			name:    "invalidPath",
			url:     "https://www.youtube.com/tv",
			wantErr: errors.New("only /watch and /shorts are supported"),
		},
		{
			name:    "videoIDTooShort",
			url:     "https://www.youtube.com/watch?v=short",
			wantErr: ErrInvalidVideoID,
		},
		{
			name:    "videoIDTooLong",
			url:     "https://www.youtube.com/watch?v=this-is-definitely-too-long",
			wantErr: ErrInvalidVideoID,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			actual, err := extractVideoID(tc.url)
			r.Equal(tc.wantErr, err)
			r.Equal(tc.want, actual)
		})
	}
}

func Test_fromWatchURL(t *testing.T) {
	testCases := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{
			name:   "validVideoID",
			rawURL: "https://www.youtube.com/watch?v=SL_YMm9C6tw",
			want:   "SL_YMm9C6tw",
		},
		{
			name:   "multipleParams",
			rawURL: "https://www.youtube.com/watch?v=SL_YMm9C6tw&t=5s&list=abc",
			want:   "SL_YMm9C6tw",
		},
		{
			name:   "duplicateVParam",
			rawURL: "https://www.youtube.com/watch?v=SL_YMm9C6tw&v=12345678910",
			want:   "SL_YMm9C6tw",
		},
		{
			name:    "missingVParam",
			rawURL:  "https://www.youtube.com/watch?t=5s",
			wantErr: ErrVideoIDNotFoundInURL,
		},
		{
			name:    "emptyVParam",
			rawURL:  "https://www.youtube.com/watch?v=",
			wantErr: ErrVideoIDNotFoundInURL,
		},
		{
			name:    "invalidVideoID",
			rawURL:  "https://www.youtube.com/watch?v=this-is-invalid",
			wantErr: ErrInvalidVideoID,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			u, err := url.Parse(tc.rawURL)
			r.NoError(err)

			actual, err := fromWatchURL(u)
			r.Equal(tc.wantErr, err)
			r.Equal(tc.want, actual)
		})
	}
}

func Test_fromShortsURL(t *testing.T) {
	testCases := []struct {
		name    string
		rawURL  string
		want    string
		wantErr error
	}{
		{
			name:   "validVideoID",
			rawURL: "https://www.youtube.com/shorts/6hz4edk5uh0",
			want:   "6hz4edk5uh0",
		},
		{
			name:   "multipleSegments",
			rawURL: "https://www.youtube.com/shorts/6hz4edk5uh0/extra/path",
			want:   "6hz4edk5uh0",
		},
		{
			name:   "withQueryParams",
			rawURL: "https://www.youtube.com/shorts/6hz4edk5uh0?t=5s",
			want:   "6hz4edk5uh0",
		},
		{
			name:    "missingVideoID",
			rawURL:  "https://www.youtube.com/shorts/",
			wantErr: ErrVideoIDNotFoundInURL,
		},
		{
			name:    "emptyPath",
			rawURL:  "https://www.youtube.com/shorts",
			wantErr: ErrVideoIDNotFoundInURL,
		},
		{
			name:    "invalidVideoID",
			rawURL:  "https://www.youtube.com/shorts/this-is-invalid",
			wantErr: ErrInvalidVideoID,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			u, err := url.Parse(tc.rawURL)
			r.NoError(err)

			actual, err := fromShortsURL(u)
			r.Equal(tc.wantErr, err)
			r.Equal(tc.want, actual)
		})
	}
}
