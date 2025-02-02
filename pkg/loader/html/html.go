package html

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type Loader struct {
	url      string
	selector string
	auto     bool
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

func WithAuto(auto bool) Option {
	return func(l *Loader) {
		l.auto = auto
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
	var s scraper

	switch {
	case l.selector != "":
		s = selectorScraper{selector: l.selector}
	case l.auto:
		s = autoScraper{}
	default:
		s = pageScraper{}
	}

	content, err := scrapeFromURL(ctx, l.url, s)
	if err != nil {
		return nil, err
	}

	return []schema.Document{
		{
			PageContent: content,
			Metadata: map[string]any{
				"url":      l.url,
				"selector": l.selector,
			},
		},
	}, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l Loader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

type RecursiveLoader struct {
	*Loader
	maxPages int
}

type RecursiveOption func(*RecursiveLoader)

func WithPageLoader(loader *Loader) RecursiveOption {
	return func(l *RecursiveLoader) {
		l.Loader = loader
	}
}

func WithMaxPages(maxPages int) RecursiveOption {
	return func(l *RecursiveLoader) {
		l.maxPages = maxPages
	}
}

func NewRecursiveLoader(opts ...RecursiveOption) *RecursiveLoader {
	loader := &RecursiveLoader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l RecursiveLoader) Load(_ context.Context) ([]schema.Document, error) {
	var s scraper
	switch {
	case l.selector != "":
		s = selectorScraper{selector: l.selector}
	case l.auto:
		s = autoScraper{}
	default:
		s = pageScraper{}
	}

	crawler, err := newCrawler(l.url, l.maxPages)
	if err != nil {
		return nil, err
	}

	contents, err := crawler.crawl(s)
	if err != nil {
		return nil, err
	}

	docs := make([]schema.Document, 0, len(contents))

	for _, ct := range contents {
		docs = append(docs, schema.Document{
			PageContent: ct.Markdown,
			Metadata: map[string]any{
				"url":      ct.URL,
				"title":    ct.Title,
				"selector": l.selector,
				"auto":     l.auto,
			},
		})
	}

	return docs, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l RecursiveLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
