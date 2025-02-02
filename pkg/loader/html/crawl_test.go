package html

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gobwas/glob"
	"github.com/nt54hamnghi/seaq/pkg/util/set"
	"github.com/stretchr/testify/suite"
)

type CrawlSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestCrawlSuite(t *testing.T) {
	suite.Run(t, &CrawlSuite{})
}

func (s *CrawlSuite) SetupSuite() {
	mux := http.NewServeMux()

	// single page
	mux.HandleFunc("/index", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<html>
			<head>
				<title>Index</title>
			</head>
			<body>
				<p>This is the index page.</p>
			</body>
		</html>`)
	})

	// with anchor tags
	mux.HandleFunc("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<html>
			<head>
				<title>Docs</title>
			</head>
			<body>
				<p>This is the docs page.</p>
				<a href="/index">Index</a>
			</body>
		</html>`)
	})

	// with repeated anchor tags to the same page
	mux.HandleFunc("/about", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<html>
			<head>
				<title>About</title>
			</head>
			<body>
				<p>This is the about page.</p>
				<p>
					<a href="/index">Index</a>
					<a href="/index">Index</a>
				</p>
			</body>
		</html>`)
	})

	s.server = httptest.NewServer(mux)
}

func (s *CrawlSuite) TearDownSuite() {
	s.server.Close()
}

// newCrawler is a helper method to create a crawler with a composed URL,
// specified maxPages, and the default glob pattern for URLs.
func (s *CrawlSuite) newTestCrawler(path string, maxPages int) *crawler {
	return &crawler{
		URL:      s.server.URL + path,
		MaxPages: maxPages,
		urlGlob:  glob.MustCompile("http://*"),
		mutex:    sync.Mutex{},
		visited:  set.New[string](),
	}
}

func (s *CrawlSuite) TestCrawl__singlePage() {
	scr := &selectorScraper{selector: "body"}
	crw := s.newTestCrawler("/index", 2)

	r := s.Require()

	contents, err := crw.crawl(scr)
	r.NoError(err)
	r.Len(contents, 1)

	content := contents[0]
	r.Equal("Index", content.Title)
	r.Equal(s.server.URL+"/index", content.URL)
	r.Equal("This is the index page.", content.Markdown)
}

func (s *CrawlSuite) TestCrawl__withAnchorTags() {
	scr := &selectorScraper{selector: "body"}
	crw := s.newTestCrawler("/docs", 2)

	r := s.Require()

	contents, err := crw.crawl(scr)
	r.NoError(err)
	r.Len(contents, 2)

	content := contents[0]
	r.Equal("Docs", content.Title)
	r.Equal(s.server.URL+"/docs", content.URL)
	r.Equal("This is the docs page.\n\n[Index](/index)", content.Markdown)

	content = contents[1]
	r.Equal("Index", content.Title)
	r.Equal(s.server.URL+"/index", content.URL)
	r.Equal("This is the index page.", content.Markdown)
}

func (s *CrawlSuite) TestCrawl__withMaxPages() {
	scr := &selectorScraper{selector: "body"}
	crw := s.newTestCrawler("/docs", 1)

	r := s.Require()

	contents, err := crw.crawl(scr)
	r.NoError(err)
	r.Len(contents, 1)

	content := contents[0]
	r.Equal("Docs", content.Title)
	r.Equal(s.server.URL+"/docs", content.URL)
	r.Equal("This is the docs page.\n\n[Index](/index)", content.Markdown)
}

func (s *CrawlSuite) TestCrawl__withRepeatedAnchorTags() {
	scr := &selectorScraper{selector: "body"}
	crw := s.newTestCrawler("/about", 2)

	r := s.Require()

	contents, err := crw.crawl(scr)
	r.NoError(err)
	r.Len(contents, 2)

	content := contents[0]
	r.Equal("About", content.Title)
	r.Equal(s.server.URL+"/about", content.URL)
	r.Equal("This is the about page.\n\n[Index](/index) [Index](/index)", content.Markdown)

	content = contents[1]
	r.Equal("Index", content.Title)
	r.Equal(s.server.URL+"/index", content.URL)
	r.Equal("This is the index page.", content.Markdown)
}
