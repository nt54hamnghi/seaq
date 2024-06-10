package dumplingai

import (
	"os"

	"github.com/nt54hamnghi/hoc/util"
)

const (
	DUMPLINGAI_API = "http://localhost:3000/api/v1/get-youtube-transcript"
)

type request struct {
	// make fields public to use with json package
	VideoUrl          string `json:"videoUrl"`
	IncludeTimestamps bool   `json:"includeTimestamps"`
}

type response struct {
	Transcript string `json:"transcript"`
}

func GetTranscript(videoUrl string) (string, error) {
	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer " + os.Getenv("DUMPLINGAI_API_KEY")},
	}
	ts, err := util.Post[response](DUMPLINGAI_API, request{VideoUrl: videoUrl}, headers)
	if err != nil {
		return "", err

	}
	return ts.Transcript, nil
}
