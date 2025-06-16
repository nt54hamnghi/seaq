package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/nt54hamnghi/seaq/pkg/util/pool"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/nt54hamnghi/seaq/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
)

const YouTubeWatchURL = "https://www.youtube.com/watch"

func getCaptionAsDocuments(ctx context.Context, vid videoID, filter *filter) ([]schema.Document, error) {
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
	return caption.toDocuments(), nil
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

	if !opt.start.IsZero() {
		c.Events = timestamp.After(opt.start, c.Events)
	}

	if !opt.end.IsZero() {
		c.Events = timestamp.Before(opt.end, c.Events)
	}
}

// toDocuments returns the full caption text of a YouTube video.
func (c *caption) toDocuments() []schema.Document {
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
