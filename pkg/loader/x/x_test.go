package x

import (
	"errors"
	"testing"

	x "github.com/imperatrona/twitter-scraper"
	"github.com/stretchr/testify/assert"
	"github.com/tmc/langchaingo/schema"
)

func Test_toDocument_tweet(t *testing.T) {
	var nilQuote map[string]any

	var testCases = []struct {
		name  string
		tweet *x.Tweet
		want  schema.Document
	}{
		{
			name: "no quoted status",
			tweet: &x.Tweet{
				ID:       "1",
				Username: "author1",
				Text:     "sample",
				Mentions: []x.Mention{},
				Hashtags: []string{},
			},
			want: schema.Document{
				PageContent: "sample",
				Metadata: map[string]any{
					"ID":       "1",
					"author":   "author1",
					"Mentions": []x.Mention{},
					"Hashtags": []string{},
					"Quote":    nilQuote,
				},
				Score: 0,
			},
		},
		{
			name: "quoted status present",
			tweet: &x.Tweet{
				ID:       "1",
				Username: "author1",
				Text:     "sample",
				Mentions: []x.Mention{},
				Hashtags: []string{},
				QuotedStatus: &x.Tweet{
					ID:       "2",
					Name:     "author2",
					Text:     "quoted",
					Mentions: []x.Mention{},
					Hashtags: []string{},
				},
			},
			want: schema.Document{
				PageContent: "sample",
				Metadata: map[string]any{
					"ID":       "1",
					"author":   "author1",
					"Mentions": []x.Mention{},
					"Hashtags": []string{},
					"Quote": map[string]any{
						"ID":       "2",
						"Author":   "author2",
						"Text":     "quoted",
						"Mentions": []x.Mention{},
						"Hashtags": []string{},
					},
				},
				Score: 0,
			},
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := toDocument(tt.tweet)
			asserts.Equal(tt.want, got)
		})
	}
}

func Test_toDocument_tweet_Error(t *testing.T) {
	var testCases = []struct {
		name    string
		tweet   *x.Tweet
		wantErr error
	}{
		{
			name:    "nil tweet",
			tweet:   nil,
			wantErr: errors.New("nil tweet"),
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toDocument(tt.tweet)
			asserts.NotNil(err)
		})
	}
}
