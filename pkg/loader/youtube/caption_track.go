package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
)

const (
	youtubePlayerAPI = "https://www.youtube.com/youtubei/v1/player"
	youtubeTVURL     = "https://www.youtube.com/tv"
)

func loadCaptionTracks(ctx context.Context, vid videoID) ([]captionTrack, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	body, err := probeTvPlayerAPI(ctx, client)
	if err != nil {
		return nil, err
	}

	itc, err := extractInnerTubeContext(body)
	if err != nil {
		return nil, err
	}

	cfg := itc.toPlayerConfig(vid)
	res, err := fetchPlayerResponse(ctx, client, cfg)
	if err != nil {
		return nil, err
	}

	return res.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks, nil
}

// probeTvPlayerAPI sends a GET request to the YouTube TV player API and returns the body.
// The body is a HTML document that contains a script tag with the INNERTUBE_CONTEXT
func probeTvPlayerAPI(ctx context.Context, client *http.Client) (io.ReadCloser, error) {
	resp, err := reqx.WithClient(client)(ctx,
		http.MethodGet,
		youtubeTVURL,
		map[string][]string{
			"User-Agent":      {"Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version"},
			"Accept":          {"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"},
			"Accept-Language": {"en-us,en;q=0.5"},
			"Sec-Fetch-Mode":  {"navigate"},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// extractInnerTubeContext reads a HTML body
// and returns INNERTUBE_CONTEXT, which is the context for the InnerTube API (Youtube's internal API)
func extractInnerTubeContext(body io.Reader) (*innerTubeContext, error) {
	// TESTME:
	html, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	var buf string
	re := regexp.MustCompile(`"INNERTUBE_CONTEXT":\s*(\{.*?\}),"INNERTUBE_CONTEXT_CLIENT_NAME`)

	scripts := html.Find("script")
	scripts.EachWithBreak(func(_ int, s *goquery.Selection) bool {
		matches := re.FindStringSubmatch(s.Text())
		// if there are exactly 2 matches
		// (one for the whole match, one for the capture group)
		if len(matches) == 2 {
			buf = matches[1]
			return false // break the loop
		}
		return true
	})

	if buf == "" {
		return nil, errors.New("INNERTUBE_CONTEXT not found in HTML")
	}

	var itc innerTubeContext
	if err := json.Unmarshal([]byte(buf), &itc); err != nil {
		return nil, fmt.Errorf("malformed InnerTube context: %w", err)
	}
	return &itc, nil
}

// fetchPlayerResponse sends a POST request to YouTube's player API and returns the response,
// which contains video metadata and caption tracks.
func fetchPlayerResponse(ctx context.Context, client *http.Client, cfg *playerConfig) (*playerResponse, error) {
	res, err := reqx.WithClientAs[playerResponse](client)(ctx,
		http.MethodPost,
		youtubePlayerAPI,
		map[string][]string{
			"User-Agent": {"Mozilla/5.0 (ChromiumStylePlatform) Cobalt/Version,gzip(gfe)"},
		},
		cfg,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// innerTubeContext represents YouTube's internal API context containing client and session information
type innerTubeContext struct {
	Client struct {
		Hl               string `json:"hl"`
		Gl               string `json:"gl"`
		RemoteHost       string `json:"remoteHost"`
		DeviceMake       string `json:"deviceMake"`
		DeviceModel      string `json:"deviceModel"`
		VisitorData      string `json:"visitorData"`
		UserAgent        string `json:"userAgent"`
		ClientName       string `json:"clientName"`
		ClientVersion    string `json:"clientVersion"`
		OsVersion        string `json:"osVersion"`
		OriginalURL      string `json:"originalUrl"`
		Theme            string `json:"theme"`
		Platform         string `json:"platform"`
		ClientFormFactor string `json:"clientFormFactor"`
		WebpSupport      bool   `json:"webpSupport"`
		ConfigInfo       struct {
			AppInstallData string `json:"appInstallData"`
		} `json:"configInfo"`
		TvAppInfo struct {
			AppQuality string `json:"appQuality"`
		} `json:"tvAppInfo"`
		AcceptHeader       string `json:"acceptHeader"`
		DeviceExperimentID string `json:"deviceExperimentId"`
		RolloutToken       string `json:"rolloutToken"`
	} `json:"client"`
	User struct {
		LockedSafetyMode bool `json:"lockedSafetyMode"`
	} `json:"user"`
	Request struct {
		UseSsl bool `json:"useSsl"`
	} `json:"request"`
	ClickTracking struct {
		ClickTrackingParams string `json:"clickTrackingParams"`
	} `json:"clickTracking"`
}

func (itc *innerTubeContext) toPlayerConfig(videoID string) *playerConfig {
	return &playerConfig{
		Context: itc,
		VideoID: videoID,
		PlaybackContext: playbackContext{
			ContentPlaybackContext: struct {
				HTML5Preference    string `json:"html5Preference"`
				SignatureTimestamp string `json:"signatureTimestamp"`
			}{
				HTML5Preference:    "HTML5_PREF_WANTS",
				SignatureTimestamp: "20249",
			},
		},
		ContentCheckOk: true,
		RacyCheckOk:    true,
	}
}

// playerConfig represents the configuration payload sent to YouTube's player API
// to retrieve video metadata and caption information.
type playerConfig struct {
	Context         *innerTubeContext `json:"context"`
	VideoID         string            `json:"videoId"`
	PlaybackContext playbackContext   `json:"playbackContext"`
	ContentCheckOk  bool              `json:"contentCheckOk"`
	RacyCheckOk     bool              `json:"racyCheckOk"`
}

// playbackContext contains playback-specific settings and preferences required by YouTube's player API
// Only used in playerConfig.
type playbackContext struct {
	ContentPlaybackContext struct {
		HTML5Preference    string `json:"html5Preference"`
		SignatureTimestamp string `json:"signatureTimestamp"`
	} `json:"contentPlaybackContext"`
}

// playerResponse represents the JSON response from YouTube's player API,
// containing caption track information and metadata.
type playerResponse struct {
	Captions struct {
		PlayerCaptionsTracklistRenderer struct {
			CaptionTracks []captionTrack `json:"captionTracks"`
		} `json:"playerCaptionsTracklistRenderer"`
	} `json:"captions"`
}

// captionTrack represents metadata for a single caption track.
type captionTrack struct {
	BaseURL        *baseURL `json:"baseUrl"`      // URL to fetch the caption track.
	VssID          string   `json:"vssId"`        // type and language (e.g., "a.en" for English automatic captions, ".en" for English manual captions).
	LanguageCode   string   `json:"languageCode"` // language code in ISO 639-1 format (e.g., "ar" for Arabic, "en" for English).
	Kind           string   `json:"kind"`         // type of the caption track (e.g., "asr" for auto-generated captions).
	IsTranslatable bool     `json:"isTranslatable"`
}

// asJSON3 adds "fmt=json3" to the base URL's query parameters.
func (ct *captionTrack) asJSON3() {
	ct.BaseURL.setQuery("fmt", "json3")
}

// toEnglish adds "tlang=en" to the base URL's query parameters
// if the captionTrack is translatable.
func (ct *captionTrack) toEnglish() error {
	if !ct.IsTranslatable {
		return errors.New("caption track is not translatable")
	}
	ct.BaseURL.setQuery("tlang", "en")
	return nil
}
