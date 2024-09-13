package x

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrInvalidXURL          = errors.New("invalid X url")
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
	parsedUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return "", ErrInvalidXURL
	}

	// TODO: add check for scheme
	if !strings.HasSuffix(parsedUrl.Hostname(), "x.com") {
		return "", ErrInvalidXURL
	}

	path := parsedUrl.Path
	if path == "" {
		return "", ErrTweetIdNotFoundInURL
	}

	segments := strings.Split(path, "/")
	if len(segments) != 4 {
		return "", ErrTweetIdNotFoundInURL
	}

	if segments[2] != "status" {
		return "", ErrInvalidXURL
	}

	id := segments[3]
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", ErrInvalidTweetId
	}

	return id, nil
}
