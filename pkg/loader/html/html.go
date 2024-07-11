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
	return func(o *HtmlLoader) {
		o.url = url
	}
}

func WithSelector(selector string) HtmlOption {
	return func(o *HtmlLoader) {
		o.selector = selector
	}
}

func WithAuto(auto bool) HtmlOption {
	return func(o *HtmlLoader) {
		o.auto = auto
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

	page, err := scrapeUrl(ctx, h.url, s)
	if err != nil {
		return nil, err
	}

	return []schema.Document{
		{
			PageContent: page,
			Metadata: map[string]any{
				"url": h.url,
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
