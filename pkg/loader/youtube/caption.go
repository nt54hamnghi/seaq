package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/nt54hamnghi/hiku/pkg/util"
)

// region: --- errors

var ErrCaptionTracksNotFound = errors.New("caption tracks not found")

// endregion: --- errors

// region: --- consts

const (
	YouTubeWatchUrl = "https://www.youtube.com/watch"
)

// endregion: --- consts

func FetchCaption(ctx context.Context, src string, options ...Option) (string, error) {
	opt := &option{}
	for _, of := range options {
		of(opt)
	}

	vid, err := resolveVideoId(src)
	if err != nil {
		return "", err
	}
	return fetchCaptionWithVideoId(ctx, vid, opt)
}

func fetchCaptionWithVideoId(ctx context.Context, vid videoId, opt *option) (cap string, err error) {
	// get the raw HTML content of the YouTube video page
	resp, err := http.Get(YouTubeWatchUrl + "?v=" + vid)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("response status code: %d", resp.StatusCode)
	}

	// the Raw HTML content contains a list of available caption tracks
	captionTracks, err := extractCaptionTracks(resp.Body)
	if err != nil {
		return cap, fmt.Errorf("failed to extract caption tracks: %w", err)
	}
	processCaptionTracks(captionTracks)

	// fetch the caption of the YouTube video
	// only support English captions
	// prioritize user-added caption over ASR (Automatic Speech Recognition) caption
	caption, err := loadCaption(ctx, captionTracks)
	if err != nil {
		return cap, fmt.Errorf("failed to fetch caption: %w", err)
	}

	if opt != nil {
		caption.filter(opt)
	}

	return caption.getFullCaption(), nil
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
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return captionTracks, nil

}

// asJson3 adds "&fmt=json3" to the base URL of the caption track
func (ct *captionTrack) asJson3() {
	url := ct.BaseURL
	if !strings.Contains(url, "?") {
		url += "?"
	}
	url += "&fmt=json3"

	ct.BaseURL = url
}

// hasTranslation checks if the caption track has an English auto-translation
// if true, add "&tlang=en" to the base URL and set HasAutoTranslation to true
func (ct *captionTrack) hasTranslation() bool {
	url := ct.BaseURL
	if !strings.Contains(url, "?") {
		url += "?"
	}
	url += "&tlang=en"

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	ct.BaseURL = url
	ct.HasAutoTranslation = true

	return true
}

// processCaptionTracks add "&fmt=json3" to the base URL of each caption track
// it also checks if the caption track has an English auto-translation
// if true, add "&tlang=en" to the base URL of the caption track and set HasAutoTranslation to true
func processCaptionTracks(captionTracks []captionTrack) {
	wg := sync.WaitGroup{}

	for i := 0; i < len(captionTracks); i++ {
		c := &captionTracks[i]

		c.asJson3()
		if c.LanguageCode == "en" || c.LanguageCode == "en-US" {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			c.hasTranslation()
		}()
	}

	wg.Wait()
}

// loadCaption returns the caption of a YouTube video.
// The returned caption is a list of events.
// Each event contains a list of segments, and each segment has a caption text.
// It only supports English captions and prioritizes user-added caption
// over ASR (Automatic Speech Recognition) caption.
func loadCaption(ctx context.Context, tracks []captionTrack) (caption, error) {
	if len(tracks) == 0 {
		return caption{}, errors.New("caption tracks must not be empty")
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
		return caption{}, errors.New("no English caption track found")
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
	DDurationMs int64 `json:"dDurationMs,omitempty"`
	ID          int   `json:"id,omitempty"`
	TStartMs    int64 `json:"tStartMs"`
	Segs        []struct {
		// confidence of the ASR caption
		AcAsrConf int `json:"acAsrConf"`
		// caption text in UTF-8
		Utf8 string `json:"utf8"`
		// timestamp offset in milliseconds from TStartMs
		TOffsetMs int `json:"tOffsetMs,omitempty"`
	} `json:"segs,omitempty"`
}

func (c *caption) filter(opt *option) {
	if opt.start != nil {
		c.filterStart(opt.start)
	}
	if opt.end != nil {
		c.filterEnd(opt.end)
	}
}

func (c *caption) filterStart(start *Timestamp) {
	if start == nil {
		return
	}

	startMs := start.ToMsDuration()
	newEvents := make([]event, 0, len(c.Events))
	for _, e := range c.Events {
		if e.TStartMs >= startMs {
			newEvents = append(newEvents, e)
		}
	}

	c.Events = newEvents
}

func (c *caption) filterEnd(end *Timestamp) {
	if end == nil {
		return
	}

	endMs := end.ToMsDuration()
	newEvents := make([]event, 0, len(c.Events))
	for _, e := range c.Events {
		if e.TStartMs <= endMs {
			newEvents = append(newEvents, e)
		}
	}

	c.Events = newEvents
}

// getFullCaption returns the full caption text of a YouTube video.
func (c *caption) getFullCaption() string {
	events := c.Events
	nThreads := util.GetThreadCount(len(events))

	res := util.BatchReduce(nThreads, events, func(es []event) string {
		var res string
		for i := 0; i < len(es); i++ {
			segs := es[i].Segs
			for j := 0; j < len(segs); j++ {
				txt := segs[j].Utf8
				if txt != "" {
					res += txt
				}
			}
		}
		return strings.TrimSpace(res)
	})

	return strings.Join(res, "")
}

type Option func(*option)

type option struct {
	start *Timestamp
	end   *Timestamp
}

func WithStart(start *Timestamp) Option {
	return func(o *option) {
		o.start = start
	}
}

func WithEnd(end *Timestamp) Option {
	return func(o *option) {
		o.end = end
	}
}
