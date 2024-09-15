package x

import (
	"errors"
	"strconv"

	"github.com/nt54hamnghi/hiku/pkg/util/reqx"
)

var (
	ErrInvalidTweetId       = errors.New("invalid tweet ID")
	ErrTweetIdNotFoundInURL = errors.New("tweet id not found in X url")
)

type tweetId = string

func ResolveTweetId(src string) (tweetId, error) {
	// tweetId is a unique unsigned 64-bit integer
	// https://developer.x.com/en/docs/x-ids
	if _, err := strconv.ParseUint(src, 10, 64); err == nil {
		return src, nil
	}

	return extractTweetId(src)
}

// extractTweetId returns the tweet ID of a provided URL
func extractTweetId(rawUrl string) (string, error) {
	xUrl, err := reqx.ParseUrl("x.com")(rawUrl)
	if err != nil {
		return "", err
	}

	matches, err := reqx.ParsePath(xUrl.Path, "/{username}/status/{tweetId}")
	if err != nil {
		return "", ErrTweetIdNotFoundInURL
	}

	id := matches["tweetId"]
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", ErrInvalidTweetId
	}

	return id, nil
}
