package x

import (
	"context"
	"errors"

	x "github.com/imperatrona/twitter-scraper"
	"github.com/nt54hamnghi/hiku/pkg/env"
	"github.com/nt54hamnghi/hiku/pkg/util/pool"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type XLoader struct {
	tweetId string
	xNgin   *x.Scraper
	noReply bool
}

type XLoaderOption func(*XLoader)

func WithTweetId(id string) XLoaderOption {
	return func(o *XLoader) {
		o.tweetId = id
	}
}

func WithoutReply(noReply bool) XLoaderOption {
	return func(o *XLoader) {
		o.noReply = noReply
	}
}

func NewXLoader(opts ...XLoaderOption) (*XLoader, error) {
	authToken, err := env.XAuthToken()
	if err != nil {
		return nil, err
	}

	csrfToken, err := env.XCSRFToken()
	if err != nil {
		return nil, err
	}

	xngin := x.New()
	xngin.SetAuthToken(x.AuthToken{
		Token:     authToken,
		CSRFToken: csrfToken,
	})

	// After setting AuthToken, IsLoggedIn must be called to verify the token.
	if !xngin.IsLoggedIn() {
		return nil, errors.New("invalid AuthToken")
	}

	loader := &XLoader{
		xNgin:   xngin,
		noReply: false,
	}

	for _, opt := range opts {
		opt(loader)
	}

	return loader, nil
}

func tweetToDocument(tweet *x.Tweet) (schema.Document, error) {
	if tweet == nil {
		return schema.Document{}, errors.New("nil tweet")
	}

	var quote map[string]any
	if tweet.QuotedStatus != nil {
		quote = map[string]any{
			"ID":       tweet.QuotedStatus.ID,
			"Author":   tweet.QuotedStatus.Name,
			"Text":     tweet.QuotedStatus.Text,
			"Mentions": tweet.QuotedStatus.Mentions,
			"Hashtags": tweet.QuotedStatus.Hashtags,
		}
	}

	return schema.Document{
		PageContent: tweet.Text,
		Metadata: map[string]any{
			"ID":       tweet.ID,
			"author":   tweet.Username,
			"Mentions": tweet.Mentions,
			"Hashtags": tweet.Hashtags,
			"Quote":    quote,
		},
		Score: 0,
	}, nil
}

func (l XLoader) getTweet() (schema.Document, error) {
	var doc schema.Document

	tweet, err := l.xNgin.GetTweet(l.tweetId)
	if err != nil {
		return doc, err
	}

	return tweetToDocument(tweet)
}

func (l XLoader) getThread() ([]schema.Document, error) {
	thread, _, err := l.xNgin.GetTweetReplies(l.tweetId, "")
	if err != nil {
		return nil, err
	}

	return pool.OrderedRun(thread, tweetToDocument)
}

// Load loads from a source and returns documents.
func (l XLoader) Load(ctx context.Context) ([]schema.Document, error) {
	if l.noReply {
		tweet, err := l.getTweet()
		if err != nil {
			return nil, err
		}
		return []schema.Document{tweet}, nil
	}

	return l.getThread()
}

// LoadAndSplit loads from a source and splits the documents using a text splitter.
func (l XLoader) LoadAndSplit(ctx context.Context, splitter textsplitter.TextSplitter) ([]schema.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}
	return textsplitter.SplitDocuments(splitter, docs)
}
