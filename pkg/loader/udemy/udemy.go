package udemy

import (
	"context"
	"errors"

	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Loader struct {
	url    string
	client *udemyClient
}

type Option func(*Loader)

func WithURL(url string) Option {
	return func(o *Loader) {
		o.url = url
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
		caption, err := lec.getCaption()
		if err != nil {
			return nil, err
		}
		return caption.download(ctx)
	case article:
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
