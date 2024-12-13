package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/nt54hamnghi/hiku/pkg/util/reqx"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
)

const YouTubeWatchURL = "https://www.youtube.com/watch"

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

func fetchCaptionAsDocuments(ctx context.Context, vid videoID, filter *filter) ([]schema.Document, error) {
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

type baseURL struct {
	url.URL
}

// baseURL implements the json.Unmarshaler interface
// it marshals a JSON URL string into an url.URL object
func (u *baseURL) UnmarshalJSON(data []byte) error {
	// Unmarshal the data into a string
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Parse URL
	parsed, err := reqx.ParseURL("www.youtube.com")(str)
	if err != nil {
		return err
	}

	// Assign the parsed URL to the baseURL
	*u = baseURL{URL: *parsed}

	return nil
}

func (u *baseURL) setQuery(key, value string) {
	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
}

// captionTrack represents metadata for a single caption track.
// Reverse-engineered from a YouTube response.
type captionTrack struct {
	BaseURL        *baseURL `json:"baseUrl"`      // URL to fetch the caption track.
	VssID          string   `json:"vssId"`        // type and language (e.g., "a.en" for English automatic captions, ".en" for English manual captions).
	LanguageCode   string   `json:"languageCode"` // language code in ISO 639-1 format (e.g., "ar" for Arabic, "en" for English).
	Kind           string   `json:"kind"`         // type of the caption track (e.g., "asr" for auto-generated captions).
	IsTranslatable bool     `json:"isTranslatable"`
}

// asJSON3 adds "fmt=json3" to the base URL's query parameters.
func (ct *captionTrack) asJSON3() {
	ct.BaseURL.setQuery("fmt", "json3")
}

// toEnglish adds "tlang=en" to the base URL's query parameters
// if the captionTrack is translatable.
func (ct *captionTrack) toEnglish() error {
	if !ct.IsTranslatable {
		return errors.New("caption track is not translatable")
	}
	ct.BaseURL.setQuery("tlang", "en")
	return nil
}

// loadCaptionTracks fetches the HTML page and extracts available caption tracks.
// A GET request to a YouTube watch URL returns an HTML page that
// contains a list of available caption tracks, stored in a JSON object.
func loadCaptionTracks(ctx context.Context, vid videoID) ([]captionTrack, error) {
	const (
		maxAttempts = 2
		maxDelay    = 100 * time.Millisecond
	)

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	req := reqx.WithClient(&http.Client{Jar: jar})
	vidURL := YouTubeWatchURL + "?v=" + vid

	tracks, err := retry(maxAttempts, maxDelay, func() ([]captionTrack, error) {
		res, err := req(ctx, http.MethodGet, vidURL, nil, nil)
		if err != nil || !res.IsSuccess() {
			return nil, err
		}
		defer res.Body.Close()

		return extractCaptionTracks(res.Body)
	})
	if err != nil {
		return nil, err
	}

	return tracks, nil
}

// extractCaptionTracks reads a HTML body and returns a list of available caption tracks.
func extractCaptionTracks(body io.Reader) ([]captionTrack, error) {
	// regex pattern to capture the caption tracks JSON object
	re := regexp.MustCompile(`"captionTracks":\s*(\[[^\]]+\])`)
	// a YouTube HTML is around 1.4MB, allocate 4MB to avoid resizing
	buf := make([]byte, 0, 1<<22)
	// each iteration reads 256KB
	tmp := make([]byte, 1<<18)

	// read until regex matches, to avoid reading the entire body
	for {
		// override the entire tmp buffer
		n, err := io.ReadFull(body, tmp)
		switch err {
		// if no error or read less than tmp's size, append to the buf
		case nil, io.ErrUnexpectedEOF:
			buf = append(buf, tmp[:n]...)
		case io.EOF:
			return nil, errors.New("caption tracks not found")
		default:
			return nil, err
		}

		matches := re.FindSubmatch(buf)
		// if less than 1 match + 1 capture group, keep reading
		if len(matches) < 2 {
			continue
		}

		var tracks []captionTrack
		if err := json.Unmarshal(matches[1], &tracks); err != nil {
			return nil, fmt.Errorf("malformed caption tracks: %w", err)
		}
		return tracks, nil
	}
}

// selectCaptionTrack selects based on `VssID`:
// 1. ".en" for user-added captions in English
// 2. "a.en" for auto-generated captions in English
// 3. English translatable track as fallback
func selectCaptionTrack(tracks []captionTrack) (*captionTrack, error) {
	// Priority: .en > a.en > translatable track
	var fallback *captionTrack

	for _, track := range tracks {
		// Priority 1: Manual captions (`.en`)
		if track.VssID == ".en" {
			return &track, nil
		}
		// Priority 2: Auto-generated captions (`a.en`)
		if track.VssID == "a.en" {
			return &track, nil
		}
		// Priority 3: Translatable track (as fallback)
		if fallback == nil && track.IsTranslatable {
			fallback = &track
		}
	}

	// No suitable track found
	if fallback == nil {
		return nil, fmt.Errorf("no English caption track found")
	}

	// Attempt to translate the fallback track into English
	if err := fallback.toEnglish(); err != nil {
		return nil, err
	}

	return fallback, nil
}

// loadCaption returns a caption from a list of available caption tracks.
// Only English caption tracks are supported.
func loadCaption(ctx context.Context, tracks []captionTrack) (caption, error) {
	if len(tracks) == 0 {
		return caption{}, errors.New("caption tracks list is empty")
	}

	ct, err := selectCaptionTrack(tracks)
	if err != nil {
		return caption{}, err
	}

	ct.asJSON3()
	return reqx.GetAs[caption](ctx, ct.BaseURL.String(), nil)
}

// caption represents a collection of caption events.
// Reverse-engineered from a YouTube response.
type caption struct {
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

func (e event) AsDuration() time.Duration {
	return time.Duration(e.TStartMs) * time.Millisecond
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

func (c *caption) filter(opt *filter) {
	if opt == nil {
		return
	}
	if opt.start != nil {
		c.Events = timestamp.After(*opt.start, c.Events)
	}
	if opt.end != nil {
		c.Events = timestamp.Before(*opt.end, c.Events)
	}
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
