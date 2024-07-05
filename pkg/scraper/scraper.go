package scraper

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

type Scraper interface {
	Scrape(*goquery.Document) ([]string, error)
}

// ContentScraper scrapes the main content of a webpage
// It uses a waterfall approach to find the content.
// The search order is as follows: content id, main tag, article tag, section tag
type ContentScraper struct{}

func New() *ContentScraper {
	return &ContentScraper{}
}

func (s ContentScraper) Scrape(doc *goquery.Document) ([]string, error) {
	return findContent(doc)
}

type FullPageScraper struct{}

func (s FullPageScraper) Scrape(doc *goquery.Document) ([]string, error) {
	return combine(doc.Selection.Contents()), nil
}

func WithFullPage() *FullPageScraper {
	return &FullPageScraper{}
}

type SelectorScraper struct {
	selector string
}

func WithSelector(selector string) (*SelectorScraper, error) {
	return &SelectorScraper{selector: selector}, nil
}

func (s SelectorScraper) Scrape(doc *goquery.Document) ([]string, error) {
	return findSelector(s.selector, doc)
}

func ScrapeUrl(ctx context.Context, url string, scr Scraper) (string, error) {
	htmlBytes, err := util.GetRaw(ctx, url, nil)
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(htmlBytes))
	if err != nil {
		return "", err
	}

	contents, err := scr.Scrape(doc)
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
