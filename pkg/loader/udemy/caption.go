package udemy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/nt54hamnghi/hiku/pkg/util/reqx"
	"github.com/nt54hamnghi/hiku/pkg/util/timestamp"
	"github.com/tmc/langchaingo/schema"
)

const (
	UdemyAPIURL = "https://www.udemy.com/api-2.0"
)

var englishLocalIDs = []string{"en_US", "en_GB"}

// region: --- helpers

// parseUdemyURL extracts the course name and lecture ID from a Udemy URL.
// Expected format: https://www.udemy.com/course/{courseName}/learn/lecture/{lectureId}
func parseUdemyURL(rawURL string) (courseName string, lectureID int, err error) {
	udemyURL, err := reqx.ParseURL("www.udemy.com")(rawURL)
	if err != nil {
		return "", 0, err
	}

	// extract course name and lecture id
	matches, err := reqx.ParsePath(udemyURL.Path, "/course/{courseName}/learn/lecture/{lectureId}")
	if err != nil {
		return "", 0, err
	}

	courseName = matches["courseName"]
	lectureID, err = strconv.Atoi(matches["lectureId"])
	if err != nil {
		return "", 0, fmt.Errorf("invalid lecture ID: %s", matches["lectureId"])
	}

	return courseName, lectureID, nil
}

func requestAPI[T any](ctx context.Context, u *udemyClient, target string) (T, error) {
	return reqx.WithClientAs[T](u.Client)(ctx, http.MethodGet, target, nil, nil)
}

// endregion: --- helpers

// udemyClient is a client for the Udemy API
type udemyClient struct {
	*http.Client
}

// newUdemyClient creates a new Udemy client with a cookie jar for storing access tokens
func newUdemyClient() *udemyClient {
	jar, _ := cookiejar.New(nil)
	return &udemyClient{
		Client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

// setAccessToken sets the access token in the udemyClient's cookie jar
func (u *udemyClient) setAccessToken(token string) {
	cookies := []*http.Cookie{
		{
			Name:     "access_token",
			Value:    token,
			Path:     "/",
			Domain:   "www.udemy.com",
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}

	u.setCookies(cookies)
}

// setCookies sets the cookies in the udemyClient's cookie jar
func (u *udemyClient) setCookies(cookies []*http.Cookie) {
	url := &url.URL{
		Scheme: "https",
		Host:   "www.udemy.com",
	}

	u.Jar.SetCookies(url, cookies)
}

type course struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type lecture struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Asset       asset  `json:"asset"`
}

type assetType int

const (
	video assetType = iota
	article
)

func (a assetType) MarshalJSON() ([]byte, error) {
	switch a {
	case video:
		return json.Marshal("Video")
	case article:
		return json.Marshal("Article")
	default:
		return nil, errors.New("unknown asset type")
	}
}

func (a *assetType) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	switch str {
	case "Video":
		*a = video
	case "Article":
		*a = article
	default:
		return fmt.Errorf("unsupported asset type: %q", str)
	}
	return nil
}

type asset struct {
	ID       int       `json:"id"`
	Type     assetType `json:"asset_type"`
	Body     string    `json:"body,omitempty"`
	Captions []caption `json:"captions,omitempty"`
}

// findCaption returns a caption matching any of the given locale IDs,
// or an error if no caption is found.
func (l *lecture) findCaption(localeIDs ...string) (caption, error) {
	for _, c := range l.Asset.Captions {
		for _, id := range localeIDs {
			if c.LocaleID == id {
				return c, nil
			}
		}
	}

	return caption{}, fmt.Errorf("no caption found for locale IDs: %v", localeIDs)
}

// getCaption returns a caption in English.
func (l *lecture) getCaption() (caption, error) {
	return l.findCaption(englishLocalIDs...)
}

// getArticle returns the lecture's article content.
func (l *lecture) getArticle() (string, error) {
	if l.Asset.Body == "" {
		return "", errors.New("no article available")
	}
	return l.Asset.Body, nil
}

type caption struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Created    time.Time `json:"created"`
	FileName   string    `json:"file_name"`
	Status     int       `json:"status"`
	URL        string    `json:"url"`
	Source     string    `json:"source"`
	LocaleID   string    `json:"locale_id"`
	VideoLabel string    `json:"video_label"`
	AssetID    int       `json:"asset_id"`
	events     []event
}

type event struct {
	*astisub.Item
}

func (c *caption) filter(opt *filter) {
	if opt == nil {
		return
	}

	if !opt.start.IsZero() {
		c.events = timestamp.After(opt.start, c.events)
	}
	if !opt.end.IsZero() {
		c.events = timestamp.Before(opt.end, c.events)
	}
}

// download retrieves and parses the caption's WebVTT content.
func (c *caption) download(ctx context.Context) error {
	// TESTME: download
	res, err := reqx.Get(ctx, c.URL, nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	sub, err := astisub.ReadFromWebVTT(res.Body)
	if err != nil {
		return err
	}

	c.events = make([]event, 0, len(sub.Items))
	for _, item := range sub.Items {
		if item != nil {
			c.events = append(c.events, event{Item: item})
		}
	}

	return nil
}

// toDocuments converts a caption into a list of documents.
func (c *caption) toDocuments() ([]schema.Document, error) {
	return pool.OrderedRun(c.events, event.toDocument)
}

func (e event) AsDuration() time.Duration {
	return e.StartAt
}

// toDocument converts an event into a document.
func (e event) toDocument() (schema.Document, error) {
	if e.Item == nil {
		return schema.Document{}, errors.New("nil event")
	}

	return schema.Document{
		PageContent: e.String(),
		Metadata: map[string]any{
			"Comment": e.Comments,
			"StartAt": e.StartAt,
			"EndAt":   e.EndAt,
		},
		Score: 0,
	}, nil
}

// searchLectureByURL searches for a lecture by URL
func (u *udemyClient) searchLectureByURL(ctx context.Context, url string) (lecture, error) {
	// parse course name and lecture ID from URL
	courseName, lectureID, err := parseUdemyURL(url)
	if err != nil {
		return lecture{}, fmt.Errorf("parse Udemy URL: %w", err)
	}

	// search course to retrieve course ID
	course, err := u.searchCourse(ctx, courseName)
	if err != nil {
		return lecture{}, fmt.Errorf("search course: %w", err)
	}

	// search lecture to retrieve caption download URL
	return u.searchLecture(ctx, course.ID, lectureID)
}

func (u *udemyClient) searchCourse(ctx context.Context, courseName string) (course, error) {
	query := url.Values{
		"fields[course]": {"@min"},
	}

	target := fmt.Sprintf("%s/courses/%s?%s", UdemyAPIURL, courseName, query.Encode())

	return requestAPI[course](ctx, u, target)
}

// searchLecture searches for a lecture by course ID and lecture ID
func (u *udemyClient) searchLecture(ctx context.Context, courseID, lectureID int) (lecture, error) {
	query := url.Values{
		"fields[lecture]": {"description,asset"},
		"fields[asset]":   {"asset_type,captions,body"},
	}

	target := fmt.Sprintf(
		"%s/users/me/subscribed-courses/%d/lectures/%d/?%s",
		UdemyAPIURL, courseID, lectureID, query.Encode(),
	)

	return requestAPI[lecture](ctx, u, target)
}
