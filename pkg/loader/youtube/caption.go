package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/nt54hamnghi/seaq/pkg/util/pool"
	"github.com/nt54hamnghi/seaq/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
)

const (
	YouTubeShortURL = "https://www.youtube.com/shorts"
	YouTubeWatchURL = "https://www.youtube.com/watch"
)

func getCaptionAsDocuments(ctx context.Context, vid VideoID, filter *filter) ([]schema.Document, error) {
	caption, err := downloadYouTubeCaptions(ctx, vid)
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

func downloadYouTubeCaptions(ctx context.Context, vid string) (caption, error) {
	var sub caption
	appName := "seaq"

	tmpDir, err := os.MkdirTemp("", appName+"-*")
	if err != nil {
		return sub, fmt.Errorf("failed to create temporary directory to save the caption: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// TODO: enable EJS, with "--js-runtimes node" and "--remote-components ejs:github"
	// see: https://github.com/yt-dlp/yt-dlp/wiki/EJS
	outFile := fmt.Sprintf("%s/%s.en.json3", tmpDir, appName)
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--skip-download",
		"--write-subs",
		"--write-auto-subs",
		"--sub-langs", "en",
		"--sub-format", "json3",
		"--quiet",
		"-P", tmpDir,
		"--output", appName,
		vid,
	)

	// the command doesn't write anything to stdout
	_, err = cmd.Output()
	if err != nil {
		prefix := "failed to download YouTube captions"

		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			rootMsg := string(exitErr.Stderr)
			return sub, fmt.Errorf("%s: %w: %s", prefix, exitErr, rootMsg)
		}

		return sub, fmt.Errorf("%s: %w", prefix, err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		return sub, err
	}

	if err = json.Unmarshal(data, &sub); err != nil {
		return sub, err
	}

	return sub, nil
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
