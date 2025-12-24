package x

import (
	"errors"
	"strconv"

	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
)

var (
	ErrInvalidTweetID       = errors.New("invalid tweet ID")
	ErrTweetIDNotFoundInURL = errors.New("tweet id not found in X url")
)

type TweetID = string

func ResolveTweetID(src string) (TweetID, error) {
	// tweetId is a unique unsigned 64-bit integer
	// https://developer.x.com/en/docs/x-ids
	if _, err := strconv.ParseUint(src, 10, 64); err == nil {
		return src, nil
	}

	return extractTweetID(src)
}

// extractTweetID returns the tweet ID of a provided URL
func extractTweetID(rawURL string) (string, error) {
	xURL, err := reqx.ParseURL("x.com")(rawURL)
	if err != nil {
		return "", err
	}

	matches, err := reqx.ParsePath(xURL.Path, "/{username}/status/{tweetId}")
	if err != nil {
		return "", ErrTweetIDNotFoundInURL
	}

	id := matches["tweetId"]
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", ErrInvalidTweetID
	}

	return id, nil
}
