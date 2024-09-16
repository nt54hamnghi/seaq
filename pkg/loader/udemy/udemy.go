package udemy

import (
	"context"

	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type UdemyLoader struct {
	url    string
	client *udemyClient
}

type UdemyLoaderOption func(*UdemyLoader)

func WithUrl(url string) UdemyLoaderOption {
	return func(o *UdemyLoader) {
		o.url = url
	}
}

func NewUdemyLoader(opts ...UdemyLoaderOption) (*UdemyLoader, error) {
	accessToken, err := env.UdemyAccessToken()
	if err != nil {
		return nil, err
	}

	client := newUdemyClient()
	client.setAccessToken(accessToken)
	loader := &UdemyLoader{
		client: client,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

// Load loads from a source and returns documents.
func (l UdemyLoader) Load(ctx context.Context) ([]schema.Document, error) {
	caption, err := l.client.getCaption(l.url)
	if err != nil {
		return nil, err
	}

	return caption.download(ctx)
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l UdemyLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
