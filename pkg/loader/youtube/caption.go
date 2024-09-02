package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/util/httpx"
	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/tmc/langchaingo/schema"
)

// region: --- errors

var ErrCaptionTracksNotFound = errors.New("caption tracks not found")

// endregion: --- errors

// region: --- consts

const (
	YouTubeWatchUrl = "https://www.youtube.com/watch"
)

// endregion: --- consts

// region: --- helpers

// TESTME
type retryFunc[T any] func() (T, error)

func retry[T any](n int, delay time.Duration, f retryFunc[T]) (T, error) {
	var (
		result T
		err    error
	)

	for i := 0; i < n; i++ {
		result, err = f()
		if err == nil {
			return result, nil
		}
		time.Sleep(delay)
	}

	return result, err
}

// endregion: --- helpers

func fetchCaptionAsDocument(ctx context.Context, vid videoId, opt *youtubeFilter) (schema.Document, error) {
	cap, err := fetchCaption(ctx, vid, opt)
	if err != nil {
		return schema.Document{}, err
	}
	return schema.Document{
		PageContent: cap,
		Metadata: map[string]any{
			"videoId": vid,
			"type":    "caption",
			"start":   opt.start,
			"end":     opt.end,
		},
	}, nil
}

func fetchCaption(ctx context.Context, vid videoId, opt *youtubeFilter) (cap string, err error) {
	// load caption tracks by sending a GET request to the YouTube watch URL
	captionTracks, err := loadCaptionTracks(ctx, vid)
	if err != nil {
		return cap, fmt.Errorf("failed to fetch caption tracks: %w", err)
	}

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

// TESTME
func loadCaptionTracks(ctx context.Context, vid videoId) ([]captionTrack, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	url := YouTubeWatchUrl + "?v=" + vid
	client := &http.Client{Jar: jar}

	captionTracks, err := retry(2, 100*time.Millisecond, func() ([]captionTrack, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// check if the response is successful
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		// the Raw HTML content contains a list of available caption tracks
		return extractCaptionTracks(resp.Body)
	})

	if err != nil {
		return nil, err
	}

	processCaptionTracks(captionTracks)
	return captionTracks, nil
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

	return httpx.GetAs[caption](ctx, ct.BaseURL, nil)
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

func (c *caption) filter(opt *youtubeFilter) {
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
	nThreads := pool.GetThreadCount(len(events))

	res := pool.BatchReduce(nThreads, events, func(es []event) string {
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
