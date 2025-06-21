package reddit

import (
	"context"

	"github.com/nt54hamnghi/seaq/pkg/loader/cache"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Loader struct {
	url string
}

type Option func(*Loader)

func WithURL(url string) Option {
	return func(o *Loader) {
		o.url = url
	}
}

func NewRedditLoader(opts ...Option) (*Loader, error) {
	loader := &Loader{}
	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// Load loads from a source and returns documents.
func (l Loader) Load(ctx context.Context) ([]schema.Document, error) {
	redditURL, err := parseRedditURL(l.url)
	if err != nil {
		return nil, err
	}
	return getRedditContentAsDocuments(ctx, redditURL)
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

func (l Loader) Hash() ([]byte, error) {
	data := map[string]any{
		"type": "reddit",
		"url":  l.url,
	}
	return cache.MarshalAndHash(data)
}

func (l Loader) Type() string {
	return "reddit"
}
