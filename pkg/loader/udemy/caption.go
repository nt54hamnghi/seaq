package udemy

// TESTME: this entire package
// TODO: add comments to functions and types

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"strconv"
	"time"

	"github.com/asticode/go-astisub"
	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/nt54hamnghi/hiku/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
)

const (
	UdemyAPIURL = "https://www.udemy.com/api-2.0"
)

var englishLocalIDs = []string{"en_US", "en_GB"}

// region: --- helpers

func parseUdemyURL(rawURL string) (courseName string, lectureID int, err error) {
	udemyURL, err := reqx.ParseURL("www.udemy.com")(rawURL)
	if err != nil {
		return
	}

	// extract course name and lecture id
	matches, err := reqx.ParsePath(udemyURL.Path, "/course/{courseName}/learn/lecture/{lectureId}")
	if err != nil {
		return
	}

	courseName = matches["courseName"]
	lectureID, err = strconv.Atoi(matches["lectureId"])
	if err != nil {
		return "", 0, fmt.Errorf("invalid lecture ID: %s", matches["lectureId"])
	}

	return
}

func requestAPI[T any](ctx context.Context, u *udemyClient, target string) (T, error) {
	var t T

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return t, err
	}

	res, err := u.Do(req)
	if err != nil {
		return t, err
	}
	defer res.Body.Close()

	return reqx.Into[T](
		&reqx.Response{
			Response: res,
			Request:  res.Request,
		},
	)
}

// endregion: --- helpers

type udemyClient struct {
	*http.Client
}

func newUdemyClient() *udemyClient {
	jar, _ := cookiejar.New(nil)
	return &udemyClient{
		Client: &http.Client{
			Jar:     jar,
			Timeout: 30 * time.Second,
		},
	}
}

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

func (l *lecture) findCaption(localeIDs ...string) (caption, error) {
	captions := l.Asset.Captions

	if len(captions) == 0 {
		return caption{}, errors.New("no captions available")
	}

	idx := slices.IndexFunc(captions,
		func(c caption) bool {
			for _, id := range localeIDs {
				if c.LocaleID == id {
					return true
				}
			}
			return false
		},
	)
	if idx == -1 {
		return caption{}, fmt.Errorf("no caption found for locale ID: %q", localeIDs)
	}

	return captions[idx], nil
}

func (l *lecture) getCaption() (caption, error) {
	return l.findCaption(englishLocalIDs...)
}

func (l *lecture) getArticle() (string, error) {
	article := l.Asset.Body
	if article == "" {
		return "", errors.New("no article available")
	}

	return article, nil
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
}

func (c *caption) download(ctx context.Context) ([]schema.Document, error) {
	// download caption content
	resp, err := reqx.Get(ctx, c.URL, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	sub, err := astisub.ReadFromWebVTT(resp.Body)
	if err != nil {
		return nil, err
	}

	return pool.OrderedRun(sub.Items, subtitleToDocument)
}

func subtitleToDocument(subtitle *astisub.Item) (schema.Document, error) {
	if subtitle == nil {
		return schema.Document{}, errors.New("nil subtitle")
	}

	return schema.Document{
		PageContent: subtitle.String(),
		Metadata: map[string]any{
			"Comment": subtitle.Comments,
			"StartAt": subtitle.StartAt,
			"EndAt":   subtitle.EndAt,
		},
		Score: 0,
	}, nil
}

func (u *udemyClient) searchCourse(ctx context.Context, courseName string) (course, error) {
	query := url.Values{}
	query.Add("fields[course]", "@min")

	target := fmt.Sprintf("%s/courses/%s?%s", UdemyAPIURL, courseName, query.Encode())

	return requestAPI[course](ctx, u, target)
}

// searchLecture searches for a lecture by course ID and lecture ID
func (u *udemyClient) searchLecture(ctx context.Context, courseID, lectureID int) (lecture, error) {
	query := url.Values{}
	query.Add("fields[lecture]", "description,asset")
	query.Add("fields[asset]", "asset_type,captions,body")

	target := fmt.Sprintf(
		"%s/users/me/subscribed-courses/%d/lectures/%d/?%s",
		UdemyAPIURL, courseID, lectureID, query.Encode(),
	)

	return requestAPI[lecture](ctx, u, target)
}

// searchLectureByURL searches for a lecture by URL
func (u *udemyClient) searchLectureByURL(ctx context.Context, url string) (lecture, error) {
	// parse course name and lecture ID from URL
	courseName, lectureID, err := parseUdemyURL(url)
	if err != nil {
		log.Fatal(err)
	}

	// search course to retrieve course ID
	course, err := u.searchCourse(ctx, courseName)
	if err != nil {
		return lecture{}, fmt.Errorf("failed to search course: %w", err)
	}

	// search lecture to retrieve caption download URL
	return u.searchLecture(ctx, course.ID, lectureID)
}
