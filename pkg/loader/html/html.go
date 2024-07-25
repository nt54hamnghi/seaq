package html

import (
	"context"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type HtmlLoader struct {
	url      string
	selector string
	auto     bool
}

type HtmlOption func(*HtmlLoader)

func WithUrl(url string) HtmlOption {
	return func(l *HtmlLoader) {
		l.url = url
	}
}

func WithSelector(selector string) HtmlOption {
	return func(l *HtmlLoader) {
		l.selector = selector
	}
}

func WithAuto(auto bool) HtmlOption {
	return func(l *HtmlLoader) {
		l.auto = auto
	}
}

func NewHtmlLoader(opts ...HtmlOption) *HtmlLoader {
	loader := &HtmlLoader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l HtmlLoader) Load(ctx context.Context) ([]schema.Document, error) {
	var s scraper
	if l.selector != "" {
		s = selectorScraper{selector: l.selector}
	} else if l.auto {
		s = autoScraper{}
	} else {
		s = pageScraper{}
	}

	content, err := scrapeFromUrl(ctx, l.url, s)
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
func (l HtmlLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}

type RecursiveHtmlLoader struct {
	*HtmlLoader
	maxPages int
}

type RecursiveHtmlOption func(*RecursiveHtmlLoader)

func WithHtmlLoader(loader *HtmlLoader) RecursiveHtmlOption {
	return func(l *RecursiveHtmlLoader) {
		l.HtmlLoader = loader
	}
}

func WithMaxPages(maxPages int) RecursiveHtmlOption {
	return func(l *RecursiveHtmlLoader) {
		l.maxPages = maxPages
	}
}

func NewRecursiveHtmlLoader(opts ...RecursiveHtmlOption) *RecursiveHtmlLoader {
	loader := &RecursiveHtmlLoader{}
	for _, opt := range opts {
		opt(loader)
	}
	return loader
}

// Load loads from a source and returns documents.
func (l RecursiveHtmlLoader) Load(ctx context.Context) ([]schema.Document, error) {
	var s scraper
	if l.selector != "" {
		s = selectorScraper{selector: l.selector}
	} else if l.auto {
		s = autoScraper{}
	} else {
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
				"url":      ct.Url,
				"title":    ct.Title,
				"selector": l.selector,
			},
		})
	}

	return docs, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l RecursiveHtmlLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
