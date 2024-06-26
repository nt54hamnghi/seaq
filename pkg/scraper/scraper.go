package scraper

import (
	"bytes"
	"context"
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
// It uses a waterfall approach to find the content:
//
// - It first tries to find the `main` tag
//
// - If not found, it tries to find the `article` tag
//
// - If not found, it tries to find the `section` tag
//
// - If none of the above are found, it returns an error
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

type TagScraper struct {
	Tag string
}

func WithTag(tag string) (*TagScraper, error) {
	if !supportedTags[tag] {
		return nil, fmt.Errorf("tag '%s' is not supported", tag)
	}
	return &TagScraper{Tag: tag}, nil
}

func (s TagScraper) Scrape(doc *goquery.Document) ([]string, error) {
	return findTag(s.Tag, doc)
}

func Scrape(ctx context.Context, url string, scr Scraper) (string, error) {
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
	if res, _ := findTag("main", doc); res != nil {
		return res, nil
	}

	if res, _ := findTag("article", doc); res != nil {
		return res, nil
	}

	return findTag("section", doc)
}

// List of valid HTML tags that typically contain text or are semantic
var supportedTags = map[string]bool{
	"p":          true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"li":         true,
	"span":       true,
	"a":          true,
	"blockquote": true,
	"pre":        true,
	"code":       true,
	"strong":     true,
	"em":         true,
	"article":    true,
	"section":    true,
	"header":     true,
	"footer":     true,
	"aside":      true,
	"main":       true,

	// TODO: reconsider these tags
	// "div":        true,
	// "nav":        true,
}

func findTag(tag string, doc *goquery.Document) ([]string, error) {
	if !supportedTags[tag] {
		return nil, fmt.Errorf("tag '%s' is not supported", tag)
	}

	tags := doc.Find(tag)
	if tags.Length() == 0 {
		return nil, fmt.Errorf("tag '%s' not found", tag)
	}

	return combine(tags), nil
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
