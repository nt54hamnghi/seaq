package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/nt54hamnghi/seaq/pkg/env"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
	"github.com/tmc/langchaingo/schema"
)

// region: --- errors

var ErrYoutubeAPIKeyNotSet = errors.New("YOUTUBE_API_KEY is not set")

// endregion: --- errors

const YoutubeAPIURL = "https://youtube.googleapis.com/youtube/v3/videos"

func getMetadataAsDocument(ctx context.Context, vid VideoID) (schema.Document, error) {
	snippet, err := fetchMetadta(ctx, vid)
	if err != nil {
		return schema.Document{}, err
	}

	content, err := json.MarshalIndent(snippet, "", "\t")
	if err != nil {
		return schema.Document{}, err
	}

	return schema.Document{
		PageContent: string(content),
		Metadata: map[string]interface{}{
			"videoId": vid,
			"type":    "metadata",
		},
	}, nil
}

func fetchMetadta(ctx context.Context, vid string) (*snippet, error) {
	url, err := buildSnippetRequestURL(vid)
	if err != nil {
		return nil, err
	}

	data, err := reqx.GetAs[youtubeVideoListResponse](ctx, url, nil)
	if err != nil {
		return nil, err
	}

	items := data.Items
	if len(items) != 1 {
		return nil, fmt.Errorf("unexpected number of items: %d", len(items))
	}

	return &items[0].Snippet, nil
}

type youtubeVideoListResponse struct {
	Items []struct {
		ID      string  `json:"id"`
		Snippet snippet `json:"snippet"`
	} `json:"items"`
}

type snippet struct {
	Title        string   `json:"title"`
	ChannelTitle string   `json:"channelTitle"`
	Description  string   `json:"description"`
	Tags         []string `json:"-"` // `json:"tags"`
}

// Implement the String method
func (s *snippet) String() string {
	return fmt.Sprintf(`
---
Title: %s
Channel: %s
Description:
%s
---
`, s.Title, s.ChannelTitle, s.Description)
}

// API docs: https://developers.google.com/youtube/v3/docs/videos/list
func buildSnippetRequestURL(vid VideoID) (string, error) {
	apiKey, err := env.YoutubeAPIKey()
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s?part=snippet&id=%s&key=%s", YoutubeAPIURL, vid, apiKey)
	return url, nil
}
