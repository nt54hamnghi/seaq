package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInValidYouTubeURL    = errors.New("invalid YouTube url")
	ErrInvalidVideoId       = errors.New("invalid video ID")
	ErrVideoIdNotFoundInURL = errors.New("video id not found in YouTube url")
)

type videoId = string

var videoIdRe = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func resolveVideoId(src string) (videoId, error) {
	if videoIdRe.MatchString(src) {
		return src, nil
	}

	return extractVideoId(src)
}

// extractVideoId returns the video ID of a YouTube watch URL
func extractVideoId(rawUrl string) (videoId, error) {
	query, found := strings.CutPrefix(rawUrl, YouTubeWatchUrl+"?")
	if !found {
		return "", ErrInValidYouTubeURL
	}

	// parse the query string
	q, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}

	vid, ok := q["v"]
	if !ok || len(vid) == 0 || vid[0] == "" {
		return "", ErrVideoIdNotFoundInURL
	}

	if !videoIdRe.MatchString(vid[0]) {
		return "", ErrInvalidVideoId
	}

	return vid[0], nil
}
