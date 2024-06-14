package dumplingai

import (
	"context"
	"os"

	"github.com/nt54hamnghi/hoc/util"
)

const (
	DumplingAiApi = "http://localhost:3000/api/v1/get-youtube-transcript"
	// DumplingAiApi = "https://app.dumplingai.com/api/v1/get-youtube-transcript"
)

type request struct {
	// make fields public to use with json package
	VideoUrl          string `json:"videoUrl"`
	IncludeTimestamps bool   `json:"includeTimestamps"`
}

type response struct {
	Transcript string `json:"transcript"`
}

func GetTranscript(ctx context.Context, videoUrl string) (string, error) {
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + os.Getenv("DUMPLINGAI_API_KEY")},
	}
	ts, err := util.Post[response](ctx, DumplingAiApi, request{VideoUrl: videoUrl}, headers)
	if err != nil {
		return "", err

	}
	return ts.Transcript, nil
}
