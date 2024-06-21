package youtube

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/nt54hamnghi/hiku/pkg/util"
)

// region: --- errors

var ErrYoutubeApiKeyNotSet = errors.New("YOUTUBE_API_KEY is not set")

// endregion: --- errors

const YoutubeApiUrl = "https://youtube.googleapis.com/youtube/v3/videos"

func FetchMetadata(ctx context.Context, src string) (*Snippet, error) {
	vid, err := extractVideoId(src)
	if err != nil {
		return nil, err
	}
	return fetchMetadtaWithVideoId(ctx, vid)
}

func fetchMetadtaWithVideoId(ctx context.Context, vid string) (*Snippet, error) {
	url, err := buildSnippetRequestUrl(vid)
	if err != nil {
		return nil, err
	}

	data, err := util.Get[youtubeVideoListResponse](ctx, url, nil)
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
		Snippet Snippet `json:"snippet"`
	} `json:"items"`
}

type Snippet struct {
	Title        string   `json:"title"`
	ChannelTitle string   `json:"channelTitle"`
	Description  string   `json:"description"`
	Tags         []string `json:"-"` //`json:"tags"`
}

// Implement the String method
func (s *Snippet) String() string {
	return fmt.Sprintf(`
---
Title: %s
Channel: %s
Description:
%s
---
`, s.Title, s.ChannelTitle, s.Description)

}

func buildSnippetRequestUrl(videoId string) (string, error) {
	apiKey := os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		return "", ErrYoutubeApiKeyNotSet
	}

	url := fmt.Sprintf("%s?part=snippet&id=%s&key=%s", YoutubeApiUrl, videoId, apiKey)
	return url, nil
}
