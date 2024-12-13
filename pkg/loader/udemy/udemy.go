package udemy

import (
	"context"
	"errors"

	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type filter struct {
	start timestamp.Timestamp
	end   timestamp.Timestamp
}

type Loader struct {
	filter
	url    string
	client *udemyClient
}

type Option func(*Loader)

func WithURL(url string) Option {
	return func(o *Loader) {
		o.url = url
	}
}

func WithStart(start timestamp.Timestamp) Option {
	return func(o *Loader) {
		o.start = start
	}
}

func WithEnd(end timestamp.Timestamp) Option {
	return func(o *Loader) {
		o.end = end
	}
}

func NewUdemyLoader(opts ...Option) (*Loader, error) {
	accessToken, err := env.UdemyAccessToken()
	if err != nil {
		return nil, err
	}

	client := newUdemyClient()
	client.setAccessToken(accessToken)
	loader := &Loader{
		client: client,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// Load loads from a source and returns documents.
func (l Loader) Load(ctx context.Context) ([]schema.Document, error) {
	lec, err := l.client.searchLectureByURL(ctx, l.url)
	if err != nil {
		return nil, err
	}

	switch lec.Asset.Type {
	case video:
		// get caption from the lecture
		caption, err := lec.getCaption()
		if err != nil {
			return nil, err
		}
		// download the caption
		if err := caption.download(ctx); err != nil {
			return nil, err
		}
		// filter the caption based on the start and end time
		caption.filter(&l.filter)
		return caption.toDocuments()
	case article:
		// get article from the lecture
		article, err := lec.getArticle()
		if err != nil {
			return nil, err
		}
		return []schema.Document{{PageContent: article}}, nil
	default:
		return nil, errors.New("unknown asset type")
	}
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
