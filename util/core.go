package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// region: --- file operations

func WriteFile(filename string, msg string) error {
	content := []byte(msg)

	// write to file with 0644 permission
	// 0644 means owner can read and write, group can read, and others can read
	// writes data to the named file, creating it if necessary, truncate it if it already exists
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// endregion: --- file operations

// region: --- http operations

func Get[T any](ctx context.Context, url string, headers map[string][]string) (T, error) {
	return do[T](ctx, http.MethodGet, url, nil, headers)
}

func Post[T any](ctx context.Context, url string, body any, headers map[string][]string) (T, error) {
	return do[T](ctx, http.MethodPost, url, body, headers)
}

func do[T any](ctx context.Context, method string, url string, body any, headers map[string][]string) (T, error) {
	var res T

	// convert struct to into a JSON-encoded byte slice
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return res, fmt.Errorf("failed to marshal body: %w", err)
	}

	// wrap the JSON byte slice in a `*bytes.Buffer` so it can be read by the HTTP request body.
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return res, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = headers

	// create a new HTTP client and send the request
	client := &http.Client{}
	rawResp, err := client.Do(req)
	if err != nil {
		return res, fmt.Errorf("failed to send request: %w", err)
	}
	defer rawResp.Body.Close()

	// decode the JSON response into a struct
	err = json.NewDecoder(rawResp.Body).Decode(&res)
	if err != nil {
		return res, fmt.Errorf("failed to decode response: %w", err)
	}

	return res, nil
}

// endregion: --- http operations
