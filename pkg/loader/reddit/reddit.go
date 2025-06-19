package reddit

import (
	"encoding/json"
	"fmt"
)

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
		var obj object

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
	Kind string `json:"kind"`
	Data object `json:"data"`
}

// object defines the interface for Reddit data objects (comments, links, etc).
type object interface {
	Kind() string
	Content() string
}

// link represents a Reddit post/submission (kind "t3").
type link struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"` // fullname (e.g. "t3_15bfi0")
	Author       string  `json:"author"`
	Title        string  `json:"title"`
	URL          string  `json:"url"`
	SelfText     string  `json:"selftext"`
	SelfTextHTML string  `json:"selftext_html"`
	Subreddit    string  `json:"subreddit"`    // subreddit name without /r/ prefix
	SubredditID  string  `json:"subreddit_id"` // subreddit's fullname
	Created      float32 `json:"created"`
	Over18       bool    `json:"over_18"`
}

func (l *link) Content() string {
	panic("todo")
}

func (l *link) Kind() string {
	return "t3"
}

// comment represents a Reddit comment (kind "t1").
type comment struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"` // fullname (e.g. "t1_c3v7f8u")
	Author      string  `json:"author"`
	Body        string  `json:"body"`
	BodyHTML    string  `json:"body_html"`
	ParentID    string  `json:"parent_id"`    // ID of parent comment or link
	Subreddit   string  `json:"subreddit"`    // subreddit name without /r/ prefix
	SubredditID string  `json:"subreddit_id"` // subreddit's fullname
	Created     float32 `json:"created"`
	Replies     replies `json:"replies"`
}

func (c *comment) Content() string {
	panic("todo")
}

func (c *comment) Kind() string {
	return "t1"
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

func (u *unsupported) Content() string {
	return ""
}

func (u *unsupported) Kind() string {
	return "unknown"
}
