// Request + Extra
package reqx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
)

// region: --- errors

var (
	ErrNilResponse = errors.New("response is nil")
)

// endregion: --- errors

// GetAs is a convenience function for making a GET request and unmarshaling the response body to a type.
func GetAs[T any](ctx context.Context, url string, headers map[string][]string) (T, error) {
	resp, err := Do(ctx, http.MethodGet, url, nil, headers)
	if err != nil {
		var zero T
		return zero, err
	}
	return Into[T](resp)
}

// PostAs is a convenience function for making a POST request and unmarshaling the response body to a type.
func PostAs[T any](ctx context.Context, url string, body any, headers map[string][]string) (T, error) {
	resp, err := Do(ctx, http.MethodPost, url, body, headers)
	if err != nil {
		var zero T
		return zero, err
	}
	return Into[T](resp)
}

// Get is a convenience function for making a GET request
func Get(ctx context.Context, url string, headers map[string][]string) (*Response, error) {
	return Do(ctx, http.MethodGet, url, nil, headers)
}

// Post is a convenience function for making a POST request
func Post(ctx context.Context, url string, body any, headers map[string][]string) (*Response, error) {
	return Do(ctx, http.MethodPost, url, body, headers)
}

func Do(
	ctx context.Context,
	method string,
	url string,
	body any,
	headers map[string][]string,
) (*Response, error) {

	return DoWith(&http.Client{}, ctx, method, url, body, headers)
}

func DoWith(
	client *http.Client,
	ctx context.Context,
	method string,
	url string,
	body any,
	headers map[string][]string,
) (*Response, error) {

	// Prepare request body
	var buf io.Reader
	if body == nil {
		buf = bytes.NewBuffer(nil)
	} else {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		buf = bytes.NewBuffer(b)
	}

	// Prepare request
	req, err := http.NewRequestWithContext(ctx, method, url, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("User-Agent", "go-http-client/1.1")

	// Send request
	// create a new HTTP client and send the request
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// read the response body
	return &Response{
		Response: res,
		Request:  req,
	}, nil
}

type Response struct {
	*http.Response
	Request *http.Request
}

func Into[T any](resp *Response) (T, error) {
	var res T

	if err := resp.ExpectContentType("application/json"); err != nil {
		return res, err
	}

	raw, err := resp.Bytes()
	if err != nil {
		return res, err
	}

	// decode the JSON response into a struct
	if err = json.Unmarshal(raw, &res); err != nil {
		return res, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return res, nil
}

func (r *Response) String() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (r *Response) Bytes() ([]byte, error) {
	if err := r.ExpectSuccess(); err != nil {
		return nil, err
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

func (r *Response) ExpectContentType(contentType string) error {
	if r.Response == nil {
		return ErrNilResponse
	}

	actual := r.Header.Get("Content-Type")
	parsed, _, err := mime.ParseMediaType(actual)
	if err != nil || parsed != contentType {
		return fmt.Errorf("unexpected content type: %s", actual)
	}

	return nil
}

func (r *Response) HasContentType(contentType string) bool {
	return r.ExpectContentType(contentType) == nil
}

func (r *Response) ExpectStatusCode(statusCode int) error {
	if r.Response == nil {
		return ErrNilResponse
	}
	if r.StatusCode != statusCode {
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return nil
}

func (r *Response) HasStatusCode(statusCode int) bool {
	return r.ExpectStatusCode(statusCode) == nil
}

func (r *Response) ExpectSuccess() error {
	if r.Response == nil {
		return ErrNilResponse
	}

	if code := r.StatusCode; code <= 199 || code >= 300 {
		return fmt.Errorf("unexpected status code: %s", r.Status)
	}

	return nil
}

func (r *Response) IsSuccess() bool {
	return r.ExpectSuccess() == nil
}
