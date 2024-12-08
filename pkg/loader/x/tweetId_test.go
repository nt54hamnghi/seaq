package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testTweetID = "1567638108937801728"
	testXUrl    = "https://x.com/LogarithmicRex/status/1567638108937801728"
)

func TestResolveTweetId(t *testing.T) {
	testCases := []struct {
		name string
		src  string
		want tweetID
	}{
		{
			name: "validId",
			src:  testTweetID,
			want: testTweetID,
		},
		{
			name: "validUrl",
			src:  testXUrl,
			want: testTweetID,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			vid, err := ResolveTweetID(tc.src)
			r.NoError(err)
			r.Equal(tc.want, vid)
		})
	}
}

func Test_extractTweetId(t *testing.T) {
	testCases := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid",
			url:  testXUrl,
			want: testTweetID,
		},
		{
			name: "withQueryParams",
			url:  testXUrl + "?page=0",
			want: testTweetID,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			actual, err := extractTweetID(tc.url)
			r.NoError(err)
			r.Equal(tc.want, actual)
		})
	}
}

func Test_extractTweetId_Error(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "noPath",
			url:     "https://x.com",
			wantErr: ErrTweetIDNotFoundInURL,
		},
		{
			name:    "ShortPath",
			url:     "https://x.com/LogarithmicRex",
			wantErr: ErrTweetIDNotFoundInURL,
		},
		{
			name:    "wrongPath",
			url:     "https://x.com/LogarithmicRex/invalid/1567638108937801728",
			wantErr: ErrTweetIDNotFoundInURL,
		},
		{
			name:    "invalidId",
			url:     "https://x.com/LogarithmicRex/status/invalid",
			wantErr: ErrInvalidTweetID,
		},
	}

	a := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			_, err := extractTweetID(tc.url)
			a.Equal(tc.wantErr, err)
		})
	}
}
