package html

import (
	"bytes"
	"net/url"
	"runtime"
	"sync"

	"github.com/gobwas/glob"
	"github.com/gocolly/colly"
	"github.com/nt54hamnghi/seaq/pkg/util/set"
	"golang.org/x/net/publicsuffix"
)

// crawler manages the web crawling process for a specific domain.
// It tracks visited pages and ensures crawling stays within
// the specified domain and page limit.
type crawler struct {
	URL      string
	MaxPages int
	urlGlob  glob.Glob
	mutex    sync.Mutex
	visited  set.Set[string]
}

// newCrawler creates a new crawler for the given destination URL.
func newCrawler(dest string, maxPages int) (*crawler, error) {
	url, err := url.ParseRequestURI(dest)
	if err != nil {
		return nil, err
	}

	// Get the effective top-level domain plus one.
	// For example, for "foo.bar.golang.org", the eTLD+1 is "golang.org".
	etldPlusOne, err := publicsuffix.EffectiveTLDPlusOne(url.Hostname())
	if err != nil {
		return nil, err
	}

	// Create a glob pattern that matches URLs under the same domain,
	// including any subdomains and paths
	gl := glob.MustCompile(`https://*.` + glob.QuoteMeta(etldPlusOne) + `*`)

	return &crawler{
		URL:      url.String(),
		MaxPages: maxPages,
		urlGlob:  gl,
		mutex:    sync.Mutex{},
		visited:  set.New[string](),
	}, nil
}

// Content represents a scraped web page with its title, URL, and markdown content.
type Content struct {
	Title    string
	URL      string
	Markdown string
}

func (crw *crawler) crawl(scr scraper) ([]Content, error) {
	ch := make(chan Content, crw.MaxPages)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.MaxDepth(1),
		colly.Async(true),
	)

	// Configure the collector
	err := c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: runtime.NumCPU(),
	})
	if err != nil {
		return nil, err
	}

	// TODO: add logging
	// c.OnRequest(func(r *colly.Request) {})

	// TODO: add logging
	// c.OnError(func(r *colly.Response, err error) {})

	// On every anchor element with a href attribute
	// invoke the callback function
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		crw.mutex.Lock()
		defer crw.mutex.Unlock()

		if len(crw.visited) >= crw.MaxPages {
			return
		}

		// Get the absolute URL of the link
		link := e.Request.AbsoluteURL(e.Attr("href"))

		// Recursively visit the link
		// if it matches the glob pattern and `MaxPages` is not reached
		if crw.urlGlob.Match(link) && !crw.visited.Contains(link) {
			crw.visited.Add(link)

			// Ignore the link, if failed to visit
			_ = c.Visit(link)
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
			URL:      r.Request.URL.String(),
			Markdown: scraped,
		}

		ch <- content
	})

	// Start the collector
	// Add initial URL to visited set before starting crawl.
	// No lock needed here - concurrent goroutines start only after c.Visit() is called.
	crw.visited.Add(crw.URL)
	if err = c.Visit(crw.URL); err != nil {
		return nil, err
	}

	go func() {
		defer close(ch)
		c.Wait()
	}()

	// Collect all content from channel into slice
	slc := make([]Content, 0, crw.MaxPages)
	for content := range ch {
		slc = append(slc, content)
	}

	return slc, nil
}
