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

type scraper interface {
	scrape(*goquery.Document) ([]string, error)
}

// autoScraper scrapes the main content of a webpage
// It uses a waterfall approach to find the content.
// The search order is: content id, main tag, article tag, section tag
type autoScraper struct{}

func (s autoScraper) scrape(doc *goquery.Document) ([]string, error) {
	return findContent(doc)
}

type pageScraper struct{}

func (s pageScraper) scrape(doc *goquery.Document) ([]string, error) {
	return collect(doc.Selection.Contents()), nil
}

// selectorScraper scrapes content from a webpage using a CSS selector.
// The selector field specifies which elements to extract content from.
type selectorScraper struct {
	selector string
}

func (s selectorScraper) scrape(doc *goquery.Document) ([]string, error) {
	return findSelector(s.selector, doc)
}

func scrapeFromURL(ctx context.Context, url string, scr scraper) (string, error) {
	resp, err := reqx.Get(ctx, url, nil)
	if err != nil {
		return "", err
	}

	if err := resp.ExpectContentType("text/html"); err != nil {
		return "", err
	}

	htmlBytes, err := resp.Bytes()
	if err != nil {
		return "", err
	}

	return scrapeFromReader(scr, bytes.NewReader(htmlBytes))
}

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

func findContent(doc *goquery.Document) ([]string, error) {
	for _, tag := range []string{"#content", "main", "article", "section"} {
		if res := doc.Find(tag); res.Length() != 0 {
			return collect(res), nil
		}
	}

	return nil, errors.New("no content found")
}

func findSelector(selector string, doc *goquery.Document) ([]string, error) {
	res := doc.Find(selector)
	if res.Length() == 0 {
		return nil, fmt.Errorf("selector '%s' not found", selector)
	}
	return collect(res), nil
}

func collect(selection *goquery.Selection) []string {
	res := make([]string, 0, selection.Length())
	selection.Contents().Each(func(_ int, s *goquery.Selection) {
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
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("href").OnElements("a")
	policy.RequireParseableURLs(true)
	policy.RequireNoFollowOnFullyQualifiedLinks(true)

	return policy.Sanitize(html)
}
