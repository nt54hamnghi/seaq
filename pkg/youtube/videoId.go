package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var ErrInvalidVideoId = errors.New("invalid video ID")

type videoId = string

var videoIdRe = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func resolveVideoId(src string) (videoId, error) {
	vid, err := extractVideoId(src)
	if err != nil {
		if !errors.Is(err, ErrInValidYouTubeURL) {
			return "", err
		}
		if !checkVideoId(src) {
			return "", ErrInvalidVideoId
		}
		return src, nil
	}

	return vid, nil
}

func checkVideoId(vid string) bool {
	return videoIdRe.MatchString(vid)
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

	return vid[0], nil
}
