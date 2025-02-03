package jina

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const JinaReaderURL = "https://r.jina.ai"

type Loader struct {
	url      string
	selector string
}

type Option func(*Loader)

func WithURL(url string) Option {
	return func(l *Loader) {
		l.url = url
	}
}

func WithSelector(selector string) Option {
	return func(l *Loader) {
		l.selector = selector
	}
}

func NewLoader(opts ...Option) *Loader {
	loader := &Loader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l Loader) Load(ctx context.Context) ([]schema.Document, error) {
	target := fmt.Sprintf("%s/%s", JinaReaderURL, l.url)

	apiKey, err := env.JinaAPIKey()
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	headers.Set("Accept", "application/json")
	if l.selector != "" {
		headers.Set("X-Target-Selector", l.selector)
	}

	res, err := reqx.GetAs[jinaResponse](ctx, target, headers)
	if err != nil {
		return nil, err
	}

	doc := res.toDocument()
	doc.Metadata["selector"] = l.selector
	return []schema.Document{doc}, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

type jinaResponse struct {
	Code   int `json:"code"`
	Status int `json:"status"`
	Data   struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
		Content     string `json:"content"`
		Usage       struct {
			Tokens int `json:"tokens"`
		} `json:"usage"`
	} `json:"data"`
}

func (r jinaResponse) toDocument() schema.Document {
	return schema.Document{
		PageContent: r.Data.Content,
		Metadata: map[string]any{
			"url":         r.Data.URL,
			"title":       r.Data.Title,
			"description": r.Data.Description,
			"engine":      "jina",
		},
	}
}
