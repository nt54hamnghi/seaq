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
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/nt54hamnghi/hiku/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
)

// region: --- errors

var ErrCaptionTracksNotFound = errors.New("caption tracks not found")

// endregion: --- errors

// region: --- consts

const (
	YouTubeWatchURL = "https://www.youtube.com/watch"
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

func fetchCaptionAsDocuments(ctx context.Context, vid videoID, filter *youtubeFilter) ([]schema.Document, error) {
	// fetch available caption tracks from a YouTube video ID
	tracks, err := loadCaptionTracks(ctx, vid)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available caption tracks: %w", err)
	}

	// fetch caption from the available caption tracks
	caption, err := loadCaption(ctx, tracks)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch caption: %w", err)
	}

	// filter the caption based on the start and end time
	if filter != nil {
		caption.filter(filter)
	}

	// convert the caption to a list of documents
	// TODO: add video ID to the metadata
	return caption.getFullCaption(), nil
}

// captionTrack represents metadata for a single caption track.
// It contains information about the track's location, language, type, and available translation.
type captionTrack struct {
	// NOTE:
	// YouTube doesn't provide a public API for caption tracks.
	// This struct is reverse-engineered from a YouTube response.
	// Thus, they're subject to change without notice.

	BaseURL            string `json:"baseUrl"`            // URL to fetch the caption track
	LanguageCode       string `json:"languageCode"`       // 2-letter language code
	Kind               string `json:"kind,omitempty"`     // empty if user-added caption, "asr" if Automatic Speech Recognition, etc.
	HasAutoTranslation bool   `json:"hasAutoTranslation"` // added property to check if the caption track has an English auto-translation
}

// TESTME

// loadCaptionTracks fetches the HTML content and extracts the available caption tracks.
func loadCaptionTracks(ctx context.Context, vid videoID) ([]captionTrack, error) {
	// NOTE:
	// A GET request to a YouTube watch URL returns an HTML page that
	// contains a list of available caption tracks, stored in a JSON object.

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	url := YouTubeWatchURL + "?v=" + vid
	client := &http.Client{Jar: jar}

	captionTracks, err := retry(2, 100*time.Millisecond, func() ([]captionTrack, error) {
		resp, err := reqx.DoWith(ctx, client, http.MethodGet, url, nil, nil)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err = resp.ExpectSuccess(); err != nil {
			return nil, err
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

// extractCaptionTracks read the HTML body
// and returns the list of caption tracks available for a YouTube video
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

// asJSON3 adds "&fmt=json3" to the base URL of the caption track
func (ct *captionTrack) asJSON3() {
	url := ct.BaseURL
	if !strings.Contains(url, "?") {
		url += "?"
	}
	url += "&fmt=json3" // TODO: prepend "&" might be incorrect here

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

		c.asJSON3()
		// if en or en-US caption track is found, skip the auto-translation check
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

// loadCaption returns the caption of a YouTube video from a list of available caption tracks.
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
		}
		if t.HasAutoTranslation {
			ct = &t
			break
		}
	}

	if ct == nil {
		return caption{}, errors.New("no English caption track found")
	}

	return reqx.GetAs[caption](ctx, ct.BaseURL, nil)
}

// caption represents a collection of caption events.
type caption struct {
	// NOTE:
	// YouTube doesn't provide a public API for caption tracks.
	// This struct is reverse-engineered from a YouTube response.
	// Thus, they're subject to change without notice.

	Events []event `json:"events"`
}

// event represents a single caption event containing multiple segments.
// It includes timing information and a list of segments.
type event struct {
	ID          int       `json:"id,omitempty"`
	TStartMs    int64     `json:"tStartMs"`
	DDurationMs int64     `json:"dDurationMs,omitempty"` // duration of the caption event in milliseconds
	Segs        []segment `json:"segs,omitempty"`
}

type segment struct {
	AcAsrConf int    `json:"acAsrConf"` // confidence of the ASR caption
	Utf8      string `json:"utf8"`      // caption text in UTF-8
}

func (e event) toDocument() (schema.Document, error) {
	if len(e.Segs) == 0 {
		return schema.Document{}, errors.New("no segments found")
	}

	doc := schema.Document{
		Metadata: map[string]any{
			"startMs":    e.TStartMs,
			"durationMs": e.DDurationMs,
			"type":       "caption",
		},
	}

	// return early if there's only one segment
	// to avoid unnecessary allocations for the strings.Builder
	if len(e.Segs) == 1 {
		doc.PageContent = e.Segs[0].Utf8
		return doc, nil
	}

	var content strings.Builder
	for _, seg := range e.Segs {
		if seg.Utf8 != "" {
			content.WriteString(seg.Utf8)
		}
	}

	doc.PageContent = content.String()
	return doc, nil
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
func (c *caption) getFullCaption() []schema.Document {
	events := c.Events
	nThreads := pool.GetThreadCount(len(events))

	res := pool.BatchReduce(nThreads, events, func(es []event) []schema.Document {
		res := make([]schema.Document, 0, len(es))
		for _, e := range es {
			if d, err := e.toDocument(); err == nil {
				res = append(res, d)
			}
		}
		return res
	})

	return slices.Concat(res...)
}
