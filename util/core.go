package util

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func Get[T any](url string, headers map[string][]string) (T, error) {
	return do[T]("GET", url, nil, headers)
}

func Post[T any](url string, body any, headers map[string][]string) (T, error) {
	return do[T]("POST", url, body, headers)
}

func do[T any](method string, url string, body any, headers map[string][]string) (T, error) {
	var res T

	// convert struct to into a JSON-encoded byte slice
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return res, err
	}

	// wrap the JSON byte slice in a `*bytes.Buffer` so it can be read by the HTTP request body.
	req, err := http.NewRequest(method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return res, err
	}
	req.Header = headers

	// create a new HTTP client and send the request
	client := &http.Client{}
	rawResp, err := client.Do(req)
	if err != nil {
		return res, err
	}
	defer rawResp.Body.Close()

	// decode the JSON response into a struct
	err = json.NewDecoder(rawResp.Body).Decode(&res)
	if err != nil {
		return res, err
	}

	return res, err
}
