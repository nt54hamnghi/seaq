package x

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testTweetId = "1567638108937801728"
	testXUrl    = "https://x.com/LogarithmicRex/status/1567638108937801728"
)

func TestResolveTweetId(t *testing.T) {
	var testCases = []struct {
		name string
		src  string
		want tweetId
	}{
		{
			name: "validId",
			src:  testTweetId,
			want: testTweetId,
		},
		{
			name: "validUrl",
			src:  testXUrl,
			want: testTweetId,
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vid, err := ResolveTweetId(tc.src)
			asserts.Nil(err)
			asserts.Equal(vid, tc.want)
		})
	}
}

func Test_extractTweetId(t *testing.T) {
	var testCases = []struct {
		name string
		url  string
		want string
	}{
		{
			name: "valid",
			url:  testXUrl,
			want: testTweetId,
		},
		{
			name: "withQueryParams",
			url:  testXUrl + "?v=SL_YMm9C6tw&t=5s",
			want: testTweetId,
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := extractTweetId(tc.url)
			asserts.Nil(err)
			asserts.Equal(actual, tc.want)
		})
	}
}

func Test_extractTweetId_Error(t *testing.T) {
	var testCases = []struct {
		name    string
		url     string
		wantErr error
	}{
		{
			name:    "unparsableURL",
			url:     "",
			wantErr: ErrInvalidXURL,
		},
		{
			name:    "not_x.com",
			url:     "https://www.google.com/LogarithmicRex/status/1567638108937801728",
			wantErr: ErrInvalidXURL,
		},
		{
			name:    "noPath",
			url:     "https://x.com",
			wantErr: ErrTweetIdNotFoundInURL,
		},
		{
			name:    "ShortPath",
			url:     "https://x.com/LogarithmicRex",
			wantErr: ErrTweetIdNotFoundInURL,
		},
		{
			name:    "wrongPath",
			url:     "https://x.com/LogarithmicRex/invalid/1567638108937801728",
			wantErr: ErrInvalidXURL,
		},
		{
			name:    "invalidId",
			url:     "https://x.com/LogarithmicRex/status/invalid",
			wantErr: ErrInvalidTweetId,
		},
	}

	asserts := assert.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := extractTweetId(tc.url)
			asserts.Equal(err, tc.wantErr)
		})
	}
}
