package youtube

import (
	"context"

	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type YouTubeLoader struct {
	source          string
	start           *Timestamp
	end             *Timestamp
	includeMetadata bool
}

type YoutubeLoaderOption func(*YouTubeLoader)

func WithSource(src string) YoutubeLoaderOption {
	return func(o *YouTubeLoader) {
		o.source = src
	}
}

func WithStart(start *Timestamp) YoutubeLoaderOption {
	return func(o *YouTubeLoader) {
		o.start = start
	}
}

func WithEnd(end *Timestamp) YoutubeLoaderOption {
	return func(o *YouTubeLoader) {
		o.end = end
	}
}

func WithMetadata(includeMetadata bool) YoutubeLoaderOption {
	return func(o *YouTubeLoader) {
		o.includeMetadata = includeMetadata
	}
}

func NewYouTubeCaption(opts ...YoutubeLoaderOption) *YouTubeLoader {
	loader := &YouTubeLoader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l YouTubeLoader) Load(ctx context.Context) ([]schema.Document, error) {
	vid, err := resolveVideoId(l.source)
	if err != nil {
		return nil, err
	}

	tasks := []pool.Task[schema.Document]{
		func() (schema.Document, error) {
			return fetchCaptionAsDocument(ctx, vid, &l)
		},
	}

	if l.includeMetadata {
		tasks = append(tasks, func() (schema.Document, error) {
			return fetchMetadtaAsDocument(ctx, vid)
		})
	}

	docs := make([]schema.Document, 0, 2)

	for _, r := range pool.OrderedGo(tasks) {
		if r.Err != nil {
			return nil, r.Err
		}
		docs = append(docs, r.Output)
	}

	return docs, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l YouTubeLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
