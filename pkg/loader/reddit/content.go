package reddit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
)

// parseRedditURL validates and normalizes a Reddit URL to use the JSON API endpoint.
//
// It automatically appends the .json suffix to the URL path if it's not present.
//
// Returns an error if the URL is invalid or not a Reddit URL.
func parseRedditURL(rawURL string) (*url.URL, error) {
	redditURL, err := reqx.ParseURL("www.reddit.com")(rawURL)
	if err != nil {
		return nil, err
	}

	// if the URL already has .json suffix, return it as is
	if strings.HasSuffix(redditURL.EscapedPath(), "/.json") {
		return redditURL, nil
	}

	return redditURL.JoinPath(".json"), nil
}

// parseRedditPath extracts Reddit URL components into a structured map.
//
// It supports two Reddit URL formats:
//   - Post (titled): /r/{subreddit}/comments/{post-id}/{title}
//   - Post (un-titled): /r/{subreddit}/comments/{post-id}
//   - Comment: /r/{subreddit}/comments/{post-id}/comment/{comment-id}
//
// The function automatically handles URLs with or without the .json suffix.
//
// Returns a map containing the extracted components:
//   - For posts: "subreddit", "post-id", "title"
//   - For comments: "subreddit", "post-id", "comment-id"
//
// Returns an error if the URL doesn't match either supported format.
func parseRedditPath(url *url.URL) (map[string]string, error) {
	path := strings.TrimSuffix(url.EscapedPath(), "/.json")

	var m map[string]string

	// parse as post path (titled version) first
	m, err := reqx.ParsePath(path, "/r/{subreddit}/comments/{post-id}/{title}")
	if err == nil {
		return m, nil
	}

	// if failed, try to parse as post path (un-titled version)
	m, err = reqx.ParsePath(path, "/r/{subreddit}/comments/{post-id}")
	if err == nil {
		return m, nil
	}

	// if failed, try to parse as comment path
	m, err = reqx.ParsePath(path, "/r/{subreddit}/comments/{post-id}/comment/{comment-id}")
	if err == nil {
		return m, nil
	}

	return nil, errors.New("invalid path for Reddit URL")
}

func getRedditContentAsDocuments(ctx context.Context, target *url.URL) ([]schema.Document, error) {
	ls, err := reqx.GetAs[[]*listing](ctx, target.String(),
		map[string][]string{
			// TODO: use a more generic user agent
			"User-Agent": {"User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:139.0) Gecko/20100101 Firefox/139.0"},
		})
	if err != nil {
		return nil, err
	}

	// response from .json should returns an array of 2 listings:
	// 1. post
	// 2. comment tree
	if len(ls) != 2 {
		return nil, errors.New("listing has unexpected length")
	}

	pathMap, err := parseRedditPath(target)
	if err != nil {
		return nil, err
	}

	var children []*thing
	if _, ok := pathMap["comment-id"]; !ok {
		// post is the first listing
		children = ls[0].Data.Children
	} else {
		// comment tree is the second listing
		children = ls[1].Data.Children
	}

	if len(children) == 0 {
		return nil, errors.New("no children found")
	}

	return []schema.Document{children[0].Data.toDocument()}, nil
}

// listing represents a Reddit listing used to paginate content.
type listing struct {
	Kind string `json:"kind"` // Kind is always "Listing"
	Data struct {
		Before   *string  `json:"before"`
		After    *string  `json:"after"`
		Children []*thing `json:"children"`
	} `json:"data"`
}

func (l *listing) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Use temporary struct to defer child data unmarshaling.
	// Children data varies by kind (t1=comment, t3=link, etc), so we
	// unmarshal child.data as json.RawMessage first, then unmarshal
	// based on the kind field.
	var temp struct {
		Kind string `json:"kind"`
		Data struct {
			Before   *string `json:"before"`
			After    *string `json:"after"`
			Children []*struct {
				Kind string          `json:"kind"`
				Data json.RawMessage `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy simple fields
	l.Kind = temp.Kind
	l.Data.Before = temp.Data.Before
	l.Data.After = temp.Data.After

	// Unmarshal each child's data based on its kind
	var children []*thing
	for _, item := range temp.Data.Children {
		var obj redditObject

		switch item.Kind {
		case "t1":
			var c comment
			if err := json.Unmarshal(item.Data, &c); err != nil {
				return err
			}
			obj = &c
		case "t3":
			var l link
			if err := json.Unmarshal(item.Data, &l); err != nil {
				return err
			}
			obj = &l
		default:
			obj = &unsupported{}
		}

		children = append(children, &thing{
			Kind: item.Kind,
			Data: obj,
		})
	}

	l.Data.Children = children

	return nil
}

// thing is the base Reddit object with a kind identifier and data payload.
type thing struct {
	Kind string       `json:"kind"`
	Data redditObject `json:"data"`
}

// redditObject defines the interface for Reddit data objects (comments, links, etc).
type redditObject interface {
	kind() string
	toDocument() schema.Document
}

// link represents a Reddit post/submission (kind "t3").
type link struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"` // fullname (e.g. "t3_15bfi0")
	Author      string  `json:"author"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Domain      string  `json:"domain"`
	SelfText    string  `json:"selftext"`
	Subreddit   string  `json:"subreddit"`    // subreddit name without /r/ prefix
	SubredditID string  `json:"subreddit_id"` // subreddit's fullname
	Created     float32 `json:"created"`
	Over18      bool    `json:"over_18"`
}

func (l *link) kind() string {
	return "t3"
}

func (l *link) toDocument() schema.Document {
	content := strings.Join([]string{l.Title, l.URL, l.SelfText}, "\n\n")
	return schema.Document{
		PageContent: content,
		Metadata: map[string]any{
			"id":           l.ID,
			"name":         l.Name,
			"author":       l.Author,
			"domain":       l.Domain,
			"subreddit":    l.Subreddit,
			"subreddit_id": l.SubredditID,
			"created":      l.Created,
			"over_18":      l.Over18,
		},
	}
}

// comment represents a Reddit comment (kind "t1").
type comment struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"` // fullname (e.g. "t1_c3v7f8u")
	Author      string  `json:"author"`
	Body        string  `json:"body"`
	ParentID    string  `json:"parent_id"`    // ID of parent comment or link
	Subreddit   string  `json:"subreddit"`    // subreddit name without /r/ prefix
	SubredditID string  `json:"subreddit_id"` // subreddit's fullname
	Created     float32 `json:"created"`
	// Disable recursive replies for now
	// Replies     replies `json:"replies"`
}

func (c *comment) kind() string {
	return "t1"
}

func (c *comment) toDocument() schema.Document {
	return schema.Document{
		PageContent: c.Body,
		Metadata: map[string]any{
			"id":           c.ID,
			"name":         c.Name,
			"author":       c.Author,
			"parent_id":    c.ParentID,
			"subreddit":    c.Subreddit,
			"subreddit_id": c.SubredditID,
			"created":      c.Created,
		},
	}
}

// replies handles comment replies, which can be either an empty string or a listing.
type replies struct {
	*listing
}

func (rp *replies) UnmarshalJSON(data []byte) error {
	// try to unmarshal as string first, if success, set to nil
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		if str != "" {
			return fmt.Errorf("unexpected non-empty string replies")
		}
		rp.listing = nil
		return nil
	}

	// it's not a string, try as listing
	var l listing
	if err := json.Unmarshal(data, &l); err != nil {
		return err
	}
	rp.listing = &l
	return nil
}

// unsupported represents unsupported Reddit object types (t2, t4, t5, t6, etc).
// This loader only handles comments (t1) and links (t3).
type unsupported struct{}

func (u *unsupported) kind() string {
	return "unknown"
}

func (u *unsupported) toDocument() schema.Document {
	return schema.Document{}
}
