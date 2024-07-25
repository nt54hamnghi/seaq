package html

import (
	"bytes"
	"net/url"
	"runtime"
	"sync"

	"github.com/gobwas/glob"
	"github.com/gocolly/colly"
	"golang.org/x/net/publicsuffix"
)

type crawler struct {
	Url      string
	MaxPages int
	urlGlob  glob.Glob
	mutex    sync.Mutex
	visited  int
}

func newCrawler(dest string, maxPages int) (*crawler, error) {
	url, err := url.ParseRequestURI(dest)
	if err != nil {
		return nil, err
	}

	etldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(url.Hostname())
	if err != nil {
		return nil, err
	}

	gl := glob.MustCompile(`https://*.` + glob.QuoteMeta(etldPlusOne) + `*`)

	return &crawler{
		Url:      url.String(),
		MaxPages: maxPages,
		urlGlob:  gl,
		mutex:    sync.Mutex{},
		visited:  0,
	}, nil
}

type Content struct {
	Title    string
	Url      string
	Markdown string
}

func (crw *crawler) crawl(scr scraper) ([]Content, error) {
	contentList := make([]Content, 0)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Configure the collector
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: runtime.NumCPU(),
	})

	c.OnRequest(func(r *colly.Request) {
		// TODO: add logging
	})

	c.OnError(func(r *colly.Response, err error) {
		// TODO: add logging
	})

	// On every a element with a href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		crw.mutex.Lock()
		defer crw.mutex.Unlock()

		if crw.visited >= crw.MaxPages-1 {
			return
		}

		link := e.Request.AbsoluteURL(e.Attr("href"))
		if crw.urlGlob.Match(link) {
			crw.visited++
			c.Visit(link)
		}
	})

	c.OnHTML("title", func(e *colly.HTMLElement) {
		e.Response.Ctx.Put("title", e.Text)
	})

	// Run after OnHTML, as a final step of scraping
	c.OnScraped(func(r *colly.Response) {
		scraped, err := scrapeFromReader(scr, bytes.NewReader(r.Body))
		if err != nil {
			return
		}

		content := Content{
			Title:    r.Ctx.Get("title"),
			Url:      r.Request.URL.String(),
			Markdown: scraped,
		}

		contentList = append(contentList, content)
	})

	c.Visit(crw.Url)
	c.Wait()

	return contentList, nil
}
