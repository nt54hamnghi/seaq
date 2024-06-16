package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Get[T any](ctx context.Context, url string, headers map[string][]string) (T, error) {
	return do[T](ctx, http.MethodGet, url, nil, headers)
}

func GetRaw(ctx context.Context, url string, headers map[string][]string) ([]byte, error) {
	return doRaw(ctx, http.MethodGet, url, nil, headers)
}

func Post[T any](ctx context.Context, url string, body any, headers map[string][]string) (T, error) {
	return do[T](ctx, http.MethodPost, url, body, headers)
}

func PostRaw(ctx context.Context, url string, body any, headers map[string][]string) ([]byte, error) {
	return doRaw(ctx, http.MethodPost, url, body, headers)
}

func do[T any](
	ctx context.Context,
	method string,
	url string,
	body any,
	headers map[string][]string,
) (T, error) {
	var res T

	raw, err := doRaw(ctx, method, url, body, headers)
	if err != nil {
		return res, err
	}

	// decode the JSON response into a struct
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func doRaw(
	ctx context.Context,
	method string,
	url string,
	body any,
	headers map[string][]string,
) ([]byte, error) {

	// convert struct to into a JSON-encoded byte slice
	var buf io.Reader

	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer([]byte{})
	}

	// wrap the JSON byte slice in a `*bytes.Buffer`
	// so it can be read by the HTTP request body.
	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = headers

	// create a new HTTP client and send the request
	client := &http.Client{}
	rawResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer rawResp.Body.Close()

	// read the response body
	return io.ReadAll(rawResp.Body)

}
