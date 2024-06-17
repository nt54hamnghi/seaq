package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"

	"github.com/nt54hamnghi/hoc/util"
)

const (
	YouTubeWatchUrl = "https://www.youtube.com/watch"
)

func WithVideoUrl(ctx context.Context, rawUrl string) (cap string, err error) {
	query, found := strings.CutPrefix(rawUrl, YouTubeWatchUrl+"?")
	if !found {
		return "", fmt.Errorf("invalid YouTube video URL")
	}

	// parse the query string
	q, err := url.ParseQuery(query)
	if err != nil {
		return
	}

	vid, ok := q["v"]
	if !ok {
		return "", fmt.Errorf("YouTube URL does not contain video ID query parameter")
	}

	return WithVideoId(ctx, vid[0])
}

func WithVideoId(ctx context.Context, vid string) (cap string, err error) {
	rawUrl := YouTubeWatchUrl + "?v=" + vid

	// get the raw HTML content of the YouTube video page
	resp, err := http.Get(rawUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	// the Raw HTML content contains a list of available caption tracks in JSON format
	captionJson, err := extractCaptionTracks(resp.Body)
	if err != nil {
		return cap, fmt.Errorf("failed to extract caption tracks: %w", err)
	}

	// fetch the caption of the YouTube video
	// only support English captions and prioritize user-added caption over ASR (Automatic Speech Recognition) caption
	caption, err := fetchCaption(ctx, captionJson)
	if err != nil {
		return cap, fmt.Errorf("failed to fetch caption: %w", err)
	}

	return caption.getFullCaption(), nil

}

// YouTube doesn't provide a public API for caption tracks.
// This struct, its fields, and their meanings are reverse-engineered from a YouTube response.
// Thus, they're subject to change without notice.
type CaptionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind,omitempty"`
}

// extractCaptionTracks returns the list of caption tracks available for a YouTube video
func extractCaptionTracks(body io.Reader) ([]CaptionTrack, error) {
	// compile the regex and panic if the pattern is invalid
	re := regexp.MustCompile(`"captionTracks":\s*(\[[^\]]+\])`)

	var found bool
	var match string
	b := make([]byte, 0, 1<<22)
	temp := make([]byte, 1<<18)

	for {
		n, err := io.ReadFull(body, temp)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		b = append(b, temp[:n]...)

		// the first element is the entire match
		// the rest are the group matches
		matches := re.FindStringSubmatch(string(b))

		if len(matches) >= 2 {
			found, match = true, matches[1]
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("caption tracks not found")
	}

	var captionTracks []CaptionTrack
	err := json.Unmarshal([]byte(match), &captionTracks)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal caption tracks: %w", err)
	}

	return captionTracks, nil

}

// fetchCaption returns the caption of a YouTube video.
// The returned caption is a list of events.
// Each event contains a list of segments, and each segment has a caption text.
// It only supports English captions and prioritizes user-added caption
// over ASR (Automatic Speech Recognition) caption.
func fetchCaption(ctx context.Context, tracks []CaptionTrack) (caption, error) {
	var caption caption

	if len(tracks) == 0 {
		return caption, fmt.Errorf("caption tracks must not be empty")
	}

	var ct *CaptionTrack

	for _, t := range tracks {
		if t.LanguageCode != "en" {
			continue
		}
		if t.Kind == "" || t.Kind == "asr" {
			ct = &t
			break
		}
	}

	res, err := util.GetRaw(ctx, ct.BaseURL+"&fmt=json3", nil)
	if err != nil {
		return caption, err
	}

	if err := json.Unmarshal(res, &caption); err != nil {
		return caption, err
	}

	return caption, nil
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
	nThreads := getThreadCount(len(events))

	res := util.BatchProcess(nThreads, events, func(es []event) string {
		var res string
		for i := 0; i < len(es); i++ {
			segs := es[i].Segs
			for j := 0; j < len(segs); j++ {
				txt := segs[j].Utf8
				if txt == "\n" {
					txt = " "
				}
				res += txt
			}
		}
		return res
	})

	return strings.Join(res, "")

}

func getThreadCount(taskCount int) int {
	numCpu := runtime.NumCPU()
	return int(math.Min(float64(taskCount), float64(numCpu)))
}
