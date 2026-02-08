package youtube

import (
	"errors"
	"net/url"
	"regexp"
	"strings"

	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
)

var (
	ErrInvalidYouTubeURL    = errors.New("invalid YouTube url")
	ErrInvalidVideoID       = errors.New("invalid YouTube video ID")
	ErrVideoIDNotFoundInURL = errors.New("video id not found in YouTube url")
)

type VideoID = string

var videoIDRegex = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

// ResolveVideoID extracts a YouTube video ID from a raw video ID or YouTube URL.
// If src matches the video ID pattern `^[A-Za-z0-9_-]{11}$`, it's returned as is.
// Otherwise, it attempts to extract the video ID from YouTube URLs (watch/short URLs).
//
// Example usage:
//
//	// Direct video ID
//	id, err := ResolveVideoID("dQw4w9WgXcQ")
//
//	// Watch URL
//	id, err := ResolveVideoID("https://www.youtube.com/watch?v=dQw4w9WgXcQ")
//
//	// Short URL
//	id, err := ResolveVideoID("https://www.youtube.com/shorts/dQw4w9WgXcQ")
func ResolveVideoID(src string) (VideoID, error) {
	// check if src is exactly a valid video ID (11 chars matching the pattern)
	if videoIDRegex.MatchString(src) {
		return src, nil
	}

	return extractVideoID(src)
}

// extractVideoID returns the video ID of a YouTube watch URL
func extractVideoID(rawURL string) (VideoID, error) {
	u, err := reqx.ParseURL("www.youtube.com")(rawURL)
	if err != nil {
		return "", ErrInvalidYouTubeURL
	}

	path := u.EscapedPath()
	switch {
	case strings.HasPrefix(path, "/watch"):
		return fromWatchURL(u)
	case strings.HasPrefix(path, "/shorts"):
		return fromShortsURL(u)
	default:
		return "", errors.New("only /watch and /shorts are supported")
	}
}

func fromWatchURL(u *url.URL) (VideoID, error) {
	// Get the first value of the "v" query parameter
	vid := u.Query().Get("v")

	if vid == "" {
		return "", ErrVideoIDNotFoundInURL
	}

	if !videoIDRegex.MatchString(vid) {
		return "", ErrInvalidVideoID
	}

	return vid, nil
}

func fromShortsURL(u *url.URL) (VideoID, error) {
	path := u.EscapedPath()
	segments := strings.Split(path, "/")
	if len(segments) < 3 {
		return "", ErrVideoIDNotFoundInURL
	}

	// empty string is at index 0
	// "shorts" is at index 1
	// video ID is at index 2
	vid := segments[2]
	if vid == "" {
		return "", ErrVideoIDNotFoundInURL
	}

	if !videoIDRegex.MatchString(vid) {
		return "", ErrInvalidVideoID
	}

	return vid, nil
}
