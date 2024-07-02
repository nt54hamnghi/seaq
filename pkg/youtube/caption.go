package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/nt54hamnghi/hiku/pkg/util"
)

// region: --- errors

var ErrInValidYouTubeURL = errors.New("invalid YouTube video URL")
var ErrVideoIdNotFoundInURL = errors.New("YouTube URL does not contain video ID query parameter")

var ErrCaptionTracksNotFound = errors.New("caption tracks not found")

// endregion: --- errors

// region: --- consts

const (
	YouTubeWatchUrl = "https://www.youtube.com/watch"
)

// endregion: --- consts

func FetchCaption(ctx context.Context, src string) (string, error) {
	vid, err := extractVideoId(src)
	if err != nil {
		if errors.Is(err, ErrInValidYouTubeURL) {
			vid = src
		} else {
			return "", err
		}
	}
	return fetchCaptionWithVideoId(ctx, vid)
}

type videoId = string

func fetchCaptionWithVideoId(ctx context.Context, vid videoId) (cap string, err error) {
	// get the raw HTML content of the YouTube video page
	resp, err := http.Get(YouTubeWatchUrl + "?v=" + vid)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// the Raw HTML content contains a list of available caption tracks
	captionTracks, err := extractCaptionTracks(resp.Body)

	if err != nil {
		return cap, fmt.Errorf("failed to extract caption tracks: %w", err)
	}
	processCaptionTracks(captionTracks)

	// fetch the caption of the YouTube video
	// only support English captions
	// prioritize user-added caption over ASR (Automatic Speech Recognition) caption
	caption, err := fetchCaption(ctx, captionTracks)
	if err != nil {
		return cap, fmt.Errorf("failed to fetch caption: %w", err)
	}

	return caption.getFullCaption(), nil
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

// YouTube doesn't provide a public API for caption tracks.
// This struct, its fields, and their meanings are reverse-engineered from a YouTube response.
// Thus, they're subject to change without notice.
type captionTrack struct {
	BaseURL            string `json:"baseUrl"`            // the URL to fetch the caption track
	LanguageCode       string `json:"languageCode"`       // 2-letter language code
	Kind               string `json:"kind,omitempty"`     // empty if user-added caption, "asr" if Automatic Speech Recognition, etc.
	HasAutoTranslation bool   `json:"hasAutoTranslation"` // added property to check if the caption track has an English auto-translation
}

// extractCaptionTracks returns the list of caption tracks available for a YouTube video
func extractCaptionTracks(body io.Reader) ([]captionTrack, error) {
	// compile the regex and panic if the pattern is invalid
	re := regexp.MustCompile(`"captionTracks":\s*(\[[^\]]+\])`)

	var found bool
	var match string
	buf := make([]byte, 0, 1<<22) // a YouTube HTML is around 1.4MB, allocate 4MB to avoid resizing
	tmp := make([]byte, 1<<18)    // each iteration reads 256KB

	// read until the regex pattern is found
	// to avoid reading the entire body ahead of time
outerLoop:
	for {
		// override the entire tmp buffer
		n, err := io.ReadFull(body, tmp)

		switch err {
		case nil, io.ErrUnexpectedEOF:
			// FIXME: expensive, can be optimized
			buf = append(buf, tmp[:n]...)
		case io.EOF:
			break outerLoop
		default:
			return nil, err
		}

		// the first element is the entire match
		// the rest are the captured groups
		matches := re.FindStringSubmatch(string(buf))

		// if we find the match and at least one group
		if len(matches) >= 2 {
			found, match = true, matches[1]
			break
		}
	}

	if !found {
		return nil, ErrCaptionTracksNotFound
	}

	var captionTracks []captionTrack
	err := json.Unmarshal([]byte(match), &captionTracks)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	return captionTracks, nil

}

// processCaptionTracks add "&fmt=json3" to the base URL of each caption track
// it also checks if the caption track has an English auto-translation
// if true, add "&tlang=en" to the base URL of the caption track and set HasAutoTranslation to true
func processCaptionTracks(captionTracks []captionTrack) {
	wg := sync.WaitGroup{}
	for i := 0; i < len(captionTracks); i++ {
		wg.Add(1)
		c := &captionTracks[i]

		rawUrl, err := url.Parse(c.BaseURL)
		if err != nil {
			continue // TODO: reconsider the error handling
		}

		query := rawUrl.Query()
		query.Add("fmt", "json3")
		rawUrl.RawQuery = query.Encode()

		c.BaseURL = rawUrl.String()
		if c.LanguageCode == "en" {
			continue
		}

		go func() {
			defer wg.Done()

			// FIXME: expensive, can be optimized with http pool
			query.Add("tlang", "en")
			rawUrl.RawQuery = query.Encode()
			resp, err := http.Get(rawUrl.String())

			if err != nil || resp.StatusCode != http.StatusOK {
				return
			}
			defer resp.Body.Close()

			c.BaseURL = rawUrl.String()
			c.HasAutoTranslation = true
		}()
	}
	wg.Wait()
}

// fetchCaption returns the caption of a YouTube video.
// The returned caption is a list of events.
// Each event contains a list of segments, and each segment has a caption text.
// It only supports English captions and prioritizes user-added caption
// over ASR (Automatic Speech Recognition) caption.
func fetchCaption(ctx context.Context, tracks []captionTrack) (caption, error) {
	if len(tracks) == 0 {
		return caption{}, fmt.Errorf("caption tracks must not be empty")
	}

	var ct *captionTrack

	for _, t := range tracks {
		if t.LanguageCode == "en" && (t.Kind == "asr" || t.Kind == "") {
			ct = &t
			break
		} else if t.HasAutoTranslation {
			ct = &t
			break
		}
	}

	if ct == nil {
		return caption{}, fmt.Errorf("no English caption track found")
	}

	return util.Get[caption](ctx, ct.BaseURL, nil)
}

// YouTube doesn't provide a public API for caption tracks.
// This struct, its fields, and their meanings are reverse-engineered from a YouTube response.
// Thus, they're subject to change without notice.
type caption struct {
	Events []event `json:"events"`
}

type event struct {
	// duration of the caption event in milliseconds
	DDurationMs int `json:"dDurationMs,omitempty"`
	ID          int `json:"id,omitempty"`
	TStartMs    int `json:"tStartMs"`
	Segs        []struct {
		// confidence of the ASR caption
		AcAsrConf int `json:"acAsrConf"`
		// caption text in UTF-8
		Utf8 string `json:"utf8"`
		// timestamp offset in milliseconds from TStartMs
		TOffsetMs int `json:"tOffsetMs,omitempty"`
	} `json:"segs,omitempty"`
}

// getFullCaption returns the full caption text of a YouTube video.
func (c *caption) getFullCaption() string {

	events := c.Events
	nThreads := util.GetThreadCount(len(events))

	res := util.BatchProcess(nThreads, events, func(es []event) string {
		var res string
		for i := 0; i < len(es); i++ {
			segs := es[i].Segs
			for j := 0; j < len(segs); j++ {
				res += segs[j].Utf8
			}
		}
		return res
	})

	return strings.Join(res, "")

}
