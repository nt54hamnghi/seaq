package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	ErrInvalidYouTubeURL    = errors.New("invalid YouTube url")
	ErrInvalidVideoID       = errors.New("invalid YouTube video ID")
	ErrVideoIDNotFoundInURL = errors.New("video id not found in YouTube url")
)

type videoID = string

var videoIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func ResolveVideoID(src string) (videoID, error) {
	if videoIDRegex.MatchString(src) {
		return src, nil
	}

	return extractVideoID(src)
}

// extractVideoID returns the video ID of a YouTube watch URL
func extractVideoID(rawURL string) (videoID, error) {
	query, found := strings.CutPrefix(rawURL, YouTubeWatchURL+"?")
	if !found {
		return "", ErrInvalidYouTubeURL
	}

	// parse the query string
	q, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}

	vid, ok := q["v"]
	if !ok || len(vid) == 0 || vid[0] == "" {
		return "", ErrVideoIDNotFoundInURL
	}

	if !videoIDRegex.MatchString(vid[0]) {
		return "", ErrInvalidVideoID
	}

	return vid[0], nil
}
