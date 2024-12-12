package youtube

import (
	"context"
	"slices"

	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type filter struct {
	start *timestamp.Timestamp
	end   *timestamp.Timestamp
}

type Loader struct {
	filter
	videoID         string
	includeMetadata bool
}

type Option func(*Loader)

func WithVideoID(src string) Option {
	return func(o *Loader) {
		o.videoID = src
	}
}

func WithStart(start *timestamp.Timestamp) Option {
	return func(o *Loader) {
		o.start = start
	}
}

func WithEnd(end *timestamp.Timestamp) Option {
	return func(o *Loader) {
		o.end = end
	}
}

func WithMetadata(includeMetadata bool) Option {
	return func(o *Loader) {
		o.includeMetadata = includeMetadata
	}
}

func NewYouTubeLoader(opts ...Option) *Loader {
	loader := &Loader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l Loader) Load(ctx context.Context) ([]schema.Document, error) {
	tasks := []pool.Task[[]schema.Document]{
		func() ([]schema.Document, error) {
			return fetchCaptionAsDocuments(ctx, l.videoID, &l.filter)
		},
	}

	if l.includeMetadata {
		tasks = append(tasks, func() ([]schema.Document, error) {
			doc, err := fetchMetadtaAsDocument(ctx, l.videoID)
			return []schema.Document{doc}, err
		})
	}

	docs := make([][]schema.Document, 0, 2)

	for _, r := range pool.OrderedGo(tasks) {
		if r.Err != nil {
			return nil, r.Err
		}
		docs = append(docs, r.Output)
	}

	return slices.Concat(docs...), nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
