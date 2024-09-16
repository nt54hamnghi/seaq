package udemy

// TESTME: this entire package
// TODO: add comments to functions and types

import (
	"context"
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
	UDEMY_API_URL = "https://www.udemy.com/api-2.0"
)

var (
	englishLocalIds = []string{"en_US", "en_GB"}
)

// region: --- helpers

func parseUdemyUrl(rawUrl string) (courseName string, lectureId int, err error) {
	udemyUrl, err := reqx.ParseUrl("www.udemy.com")(rawUrl)
	if err != nil {
		return
	}

	// extract course name and lecture id
	matches, err := reqx.ParsePath(udemyUrl.Path, "/course/{courseName}/learn/lecture/{lectureId}")
	if err != nil {
		return
	}

	courseName = matches["courseName"]
	lectureId, err = strconv.Atoi(matches["lectureId"])
	if err != nil {
		return "", 0, fmt.Errorf("invalid lecture ID: %s", matches["lectureId"])
	}

	return
}

func requestAPI[T any](u *udemyClient, target string) (T, error) {
	var t T

	resp, err := u.Get(target)
	if err != nil {
		return t, err
	}
	defer resp.Body.Close()

	return reqx.Into[T](
		&reqx.Response{
			Response: resp,
			Request:  resp.Request,
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

func (u *udemyClient) searchCourse(courseName string) (course, error) {
	query := url.Values{}
	query.Add("fields[course]", "@min")

	target := fmt.Sprintf("%s/courses/%s?%s", UDEMY_API_URL, courseName, query.Encode())

	return requestAPI[course](u, target)
}

type lecture struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Asset       asset  `json:"asset"`
}

type asset struct {
	ID       int       `json:"id"`
	Captions []caption `json:"captions"`
}

func (l *lecture) findCaption(localeIDs ...string) (caption, error) {
	captions := l.Asset.Captions

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

func (u *udemyClient) searchLecture(courseId, lectureId int) (lecture, error) {
	query := url.Values{}
	query.Add("fields[lecture]", "description,asset")
	query.Add("fields[asset]", "captions")

	target := fmt.Sprintf(
		"%s/users/me/subscribed-courses/%d/lectures/%d/?%s",
		UDEMY_API_URL, courseId, lectureId, query.Encode(),
	)

	return requestAPI[lecture](u, target)
}

func (u *udemyClient) getCaption(url string) (caption, error) {
	courseName, lectureId, err := parseUdemyUrl(url)
	if err != nil {
		log.Fatal(err)
	}

	// search course to retrieve course ID
	course, err := u.searchCourse(courseName)
	if err != nil {
		return caption{}, fmt.Errorf("failed to search course: %w", err)
	}

	// search lecture to retrieve caption download URL
	lecture, err := u.searchLecture(course.ID, lectureId)
	if err != nil {
		return caption{}, fmt.Errorf("failed to search lecture: %w", err)
	}

	return lecture.findCaption(englishLocalIds...)
}
