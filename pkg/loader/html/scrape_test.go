package html

import (
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func Test_findMainContent(t *testing.T) {
	testCases := []struct {
		name    string
		html    string
		want    []string
		wantErr error
	}{
		{
			name: "main priority",
			html: `
			<main><p>main tag</p></main>
			<div id="content"><p>content id</p></div>
			<article><p>article tag</p></article>
			<section><p>section tag</p></section>
			`,
			want:    []string{"<p>main tag</p>"},
			wantErr: nil,
		},
		{
			name: "content priority",
			html: `
			<div id="content"><p>content id</p></div>
			<article><p>article tag</p></article>
			<section><p>section tag</p></section>
			`,
			want:    []string{"<p>content id</p>"},
			wantErr: nil,
		},
		{
			name: "article priority",
			html: `
			<article><p>article tag</p></article>
			<section><p>section tag</p></section>
			`,
			want:    []string{"<p>article tag</p>"},
			wantErr: nil,
		},
		{
			name:    "section priority",
			html:    `<section><p>section tag</p></section>`,
			want:    []string{"<p>section tag</p>"},
			wantErr: nil,
		},
		{
			name:    "no content",
			html:    `<div>Some text without main containers</div>`,
			want:    nil,
			wantErr: errors.New("no content found"),
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			doc, _ := goquery.NewDocumentFromReader(strings.NewReader(tt.html))

			got, err := findMainContent(doc)
			if tt.wantErr != nil {
				r.EqualError(err, tt.wantErr.Error())
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func Test_scrapeFromReader(t *testing.T) {
	tests := []struct {
		name    string
		html    string
		scraper scraper
		want    string
		wantErr error
	}{
		{
			name: "auto scraper",
			html: `
			<main><p>main tag</p></main>
			<div id="content"><p>content id</p></div>
			<article><p>article tag</p></article>
			<section><p>section tag</p></section>
			`,
			scraper: autoScraper{},
			want:    "main tag",
			wantErr: nil,
		},
		{
			name: "page scraper",
			html: `
			<div id="first"><p>First</p></div>
			<div id="second"><p>Second</p></div>
			<div id="third"><p>Third</p></div>
			`,
			scraper: pageScraper{},
			want:    "First\n\nSecond\n\nThird",
			wantErr: nil,
		},
		{
			name: "selector scraper",
			html: `
			<div id="selected"><p>Selected</p></div>
			<div id="ignored"><p>Ignored</p></div>
			`,
			scraper: selectorScraper{selector: "div#selected"},
			want:    "Selected",
			wantErr: nil,
		},
		{
			name:    "mock scraper",
			html:    "<div>content</div>",
			scraper: &mockScraper{},
			want:    "# Header 1\n\n## Header 2\n\n### Header 3",
			wantErr: nil,
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			reader := strings.NewReader(tt.html)
			got, err := scrapeFromReader(tt.scraper, reader)

			if tt.wantErr != nil {
				r.EqualError(err, tt.wantErr.Error())
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

// mockScraper implements the scraper interface for testing
type mockScraper struct{}

func (m *mockScraper) scrape(*goquery.Document) ([]string, error) {
	return []string{"<h1>Header 1</h1>", "<h2>Header 2</h2>", "<h3>Header 3</h3>"}, nil
}
