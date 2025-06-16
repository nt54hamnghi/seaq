package html

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
)

// scraper defines the interface for HTML content extraction.
// Implementations should extract relevant content from a goquery.Document
// and return it as a slice of HTML strings.
type scraper interface {
	// scrape extracts content from the provided document.
	// Returns a slice of HTML strings and any error encountered during extraction.
	scrape(*goquery.Document) ([]string, error)
}

// scrapeFromURL fetches HTML content from a URL
// and converts it to markdown using the provided scraper.
func scrapeFromURL(ctx context.Context, url string, scr scraper) (string, error) {
	res, err := reqx.Get(ctx, url, nil)
	if err != nil {
		return "", err
	}

	if err := res.ExpectContentType("text/html"); err != nil {
		return "", err
	}

	htmlBytes, err := res.Bytes()
	if err != nil {
		return "", err
	}

	return scrapeFromReader(scr, bytes.NewReader(htmlBytes))
}

// scrapeFromReader parses HTML content from a reader
// and converts it to markdown using the provided scraper.
func scrapeFromReader(scr scraper, r io.Reader) (string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", err
	}

	contents, err := scr.scrape(doc)
	if err != nil {
		return "", err
	}

	html := strings.Join(contents, "\n")
	markdown, err := html2md(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

// autoScraper scrapes the main content of a webpage
// It uses a waterfall approach to find the main content.
// The search order is: main tag, content IDs (#content, #primary, #main), article tag, section tag
type autoScraper struct{}

// scrape implements the scraper interface
func (s autoScraper) scrape(doc *goquery.Document) ([]string, error) {
	return findMainContent(doc)
}

// pageScraper scrapes content from a webpage without any filtering.
type pageScraper struct{}

// scrape implements the scraper interface
func (s pageScraper) scrape(doc *goquery.Document) ([]string, error) {
	return collect(doc.Contents()), nil
}

// selectorScraper scrapes content from a webpage using a CSS selector.
type selectorScraper struct {
	selector string
}

// scrape implements the scraper interface
func (s selectorScraper) scrape(doc *goquery.Document) ([]string, error) {
	return findSelector(s.selector, doc)
}

// findMainContent attempts to locate the main content of a webpage by searching for
// common content-containing elements in order of specificity: main tag,
// content IDs (#content, #primary, #main), article tag, and section tag.
// Returns the first matching content or an error if no content is found.
func findMainContent(doc *goquery.Document) ([]string, error) {
	selectors := []string{
		"main",                          // semantic main element
		"#content", "#primary", "#main", // common content IDs
		"article", // semantic article element
		"section", // semantic section element
	}

	for _, selector := range selectors {
		if s := doc.Find(selector); s.Length() != 0 {
			return collect(s), nil
		}
	}

	return nil, errors.New("no content found")
}

// findSelector attempts to locate content using a CSS selector.
// It returns the content if found, or an error if the selector is not found.
func findSelector(selector string, doc *goquery.Document) ([]string, error) {
	res := doc.Find(selector)
	if res.Length() == 0 {
		return nil, fmt.Errorf("selector '%s' not found", selector)
	}
	return collect(res), nil
}

// collect extracts HTML content from each element and returns them as a slice of strings.
// It skips any elements that fail HTML extraction.
func collect(selection *goquery.Selection) []string {
	res := make([]string, 0, selection.Length())
	selection.Each(func(_ int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			return
		}
		res = append(res, html)
	})
	return res
}

// html2md converts HTML content to Markdown format.
// The HTML is first sanitized to remove potentially unsafe content
// before conversion.
func html2md(rawHTML string) (string, error) {
	safeHTML := sanitizeHTML(rawHTML)
	return htmltomarkdown.ConvertString(safeHTML)
}

// sanitizeHTML removes unsafe HTML elements and attributes, keeping only
// safe link elements and their href attributes.
func sanitizeHTML(html string) string {
	// create a User Generated Content policy
	policy := bluemonday.UGCPolicy()
	// only allow href attribute on anchor tags
	policy.AllowAttrs("href").OnElements("a")
	// ensure all URLs are parseable with `url.Parse`
	// only allows mailto, http, and https schemes
	// allows relative URLs
	policy.AllowStandardURLs()

	return policy.Sanitize(html)
}
