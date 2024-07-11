package html

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nt54hamnghi/hiku/pkg/util"
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
	return combine(doc.Selection.Contents()), nil
}

type selectorScraper struct {
	selector string
}

func (s selectorScraper) scrape(doc *goquery.Document) ([]string, error) {
	return findSelector(s.selector, doc)
}

func scrapeUrl(ctx context.Context, url string, scr scraper) (string, error) {
	htmlBytes, err := util.GetRaw(ctx, url, nil)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return "", err
	}

	contents, err := scr.scrape(doc)
	if err != nil {
		return "", err
	}

	html := strings.Join(contents, "\n")
	markdown, err := html2md([]byte(html))
	if err != nil {
		return "", err
	}

	return string(markdown), nil

}

func findContent(doc *goquery.Document) ([]string, error) {
	for _, tag := range []string{"#content", "main", "article", "section"} {
		if res := doc.Find(tag); res.Length() != 0 {
			return combine(res), nil
		}
	}

	return nil, errors.New("no content found")
}

func findSelector(selector string, doc *goquery.Document) ([]string, error) {
	res := doc.Find(selector)
	if res.Length() == 0 {
		return nil, fmt.Errorf("selector '%s' not found", selector)
	}
	return combine(res), nil
}

func combine(selection *goquery.Selection) []string {
	res := make([]string, 0, selection.Length())
	selection.Contents().Each(func(i int, s *goquery.Selection) {
		html, err := s.Html()
		if err != nil {
			return
		}
		res = append(res, sanitizeHTML(html))
	})
	return res
}

func html2md(safeHTML []byte) ([]byte, error) {
	converter := md.NewConverter("", true, nil)
	// converter.Use(plugin.Table())

	return converter.ConvertBytes(safeHTML)
}

func sanitizeHTML(html string) string {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("href").OnElements("a")
	policy.RequireParseableURLs(true)
	policy.RequireNoFollowOnFullyQualifiedLinks(true)

	return policy.Sanitize(html)
}
