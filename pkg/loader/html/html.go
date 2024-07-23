package html

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type HtmlLoader struct {
	url       string
	selector  string
	auto      bool
	recursive bool
	maxPages  int
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

func WithRecursive(recursive bool, maxPages int) HtmlOption {
	return func(l *HtmlLoader) {
		l.recursive = recursive
		l.maxPages = maxPages
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
func (h HtmlLoader) Load(ctx context.Context) ([]schema.Document, error) {
	var s scraper
	if h.selector != "" {
		s = selectorScraper{selector: h.selector}
	} else if h.auto {
		s = autoScraper{}
	} else {
		s = pageScraper{}
	}

	var content string
	var err error

	if !h.recursive {
		content, err = scrapeFromUrl(ctx, h.url, s)
		if err != nil {
			return nil, err
		}
	} else {
		crw, err := newCrawler(h.url, h.maxPages)
		if err != nil {
			return nil, err
		}

		contentList, err := crw.crawl(s)
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		enc.Encode(contentList)
		content = buf.String()
	}

	return []schema.Document{
		{
			PageContent: content,
			Metadata: map[string]any{
				"url":      h.url,
				"selector": h.selector,
			},
		},
	}, nil
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (h HtmlLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := h.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
