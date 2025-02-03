package firecrawl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

const FirecrawlURL = "https://api.firecrawl.dev/v1/scrape"

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
	apiKey, err := env.FirecrawlAPIKey()
	if err != nil {
		return nil, err
	}

	headers := make(http.Header)
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	headers.Set("Content-Type", "application/json")

	body := map[string]any{
		"url":     l.url,
		"formats": []string{"markdown"},
	}

	if l.selector != "" {
		body["includeTags"] = []string{l.selector}
	}

	res, err := reqx.PostAs[firecrawlResponse](ctx, FirecrawlURL, headers, body)
	if err != nil {
		return nil, err
	}

	return []schema.Document{res.toDocument()}, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

type firecrawlResponse struct {
	Data struct {
		Markdown string `json:"markdown"`
		Metadata struct {
			Description   string `json:"description"`
			Language      string `json:"language"`
			PublishedTime string `json:"publishedTime"`
			Title         string `json:"title"`
			URL           string `json:"url"`
		} `json:"metadata"`
	} `json:"data"`
}

func (r firecrawlResponse) toDocument() schema.Document {
	return schema.Document{
		PageContent: r.Data.Markdown,
		Metadata: map[string]any{
			"url":         r.Data.Metadata.URL,
			"title":       r.Data.Metadata.Title,
			"description": r.Data.Metadata.Description,
		},
	}
}
