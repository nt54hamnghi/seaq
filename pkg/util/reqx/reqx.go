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

var ErrNilResponse = errors.New("response is nil")

// GetAs makes a GET request and unmarshals the response into a struct of type T
func GetAs[T any](ctx context.Context, url string, headers map[string][]string) (T, error) {
	return WithClientAs[T](http.DefaultClient)(ctx, http.MethodGet, url, headers, nil)
}

// PostAs makes a POST request and unmarshals the response into a struct of type T
func PostAs[T any](ctx context.Context, url string, headers map[string][]string, body any) (T, error) {
	return WithClientAs[T](http.DefaultClient)(ctx, http.MethodPost, url, headers, body)
}

// Get is a convenience function for making a GET request
func Get(ctx context.Context, url string, headers map[string][]string) (*Response, error) {
	return WithClient(http.DefaultClient)(ctx, http.MethodGet, url, headers, nil)
}

// Post is a convenience function for making a POST request
func Post(ctx context.Context, url string, headers map[string][]string, body any) (*Response, error) {
	return WithClient(http.DefaultClient)(ctx, http.MethodPost, url, headers, body)
}

type (
	RequestFunc          func(ctx context.Context, method string, url string, headers map[string][]string, body any) (*Response, error)
	RequestAsFunc[T any] func(ctx context.Context, method string, url string, headers map[string][]string, body any) (T, error)
)

// WithClientAs returns a RequestAsFunc that uses the provided http.Client
// to make requests and unmarshal the response into a struct of type T
func WithClientAs[T any](client *http.Client) RequestAsFunc[T] {
	return func(ctx context.Context, method, url string, headers map[string][]string, body any) (T, error) {
		res, err := WithClient(client)(ctx, method, url, headers, body)
		if err != nil {
			var zero T
			return zero, err
		}
		return Into[T](res)
	}
}

// WithClient returns a RequestFunc that uses the provided http.Client to make requests
func WithClient(client *http.Client) RequestFunc {
	return func(ctx context.Context, method, url string, headers map[string][]string, body any) (*Response, error) {
		// Validate inputs
		if client == nil {
			return nil, errors.New("client cannot be nil")
		}
		if ctx == nil {
			return nil, errors.New("context cannot be nil")
		}
		if method != http.MethodGet && method != http.MethodPost {
			return nil, fmt.Errorf("unsupported method: %s", method)
		}
		if url == "" {
			return nil, errors.New("url cannot be empty")
		}

		buf := bytes.NewBuffer(nil)
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("marshal body: %w", err)
			}
			buf = bytes.NewBuffer(b)
		}

		// Prepare request
		req, err := http.NewRequestWithContext(ctx, method, url, buf)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		for k, vs := range headers {
			req.Header[k] = append(req.Header[k], vs...)
		}
		req.Header.Set("User-Agent", "go-http-client/1.1")

		// Send request
		// Caller is responsible for closing the response body
		res, err := client.Do(req) // nolint: bodyclose
		if err != nil {
			return nil, fmt.Errorf("send request: %w", err)
		}

		return &Response{
			Response: res,
			Request:  req,
		}, nil
	}
}

// Response wraps an http.Response and its corresponding http.Request
type Response struct {
	*http.Response
	Request *http.Request
}

// Into unmarshals the response body into a struct of type T
func Into[T any](res *Response) (T, error) {
	var t T

	raw, err := res.Bytes()
	if err != nil {
		return t, err
	}

	if err := res.ExpectContentType("application/json"); err != nil {
		return t, err
	}

	if err = json.Unmarshal(raw, &t); err != nil {
		return t, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return t, nil
}

// String returns the response body as a string.
// Returns an error if the response has an unsuccessful status code.
func (r *Response) String() (string, error) {
	b, err := r.Bytes()
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Bytes returns the response body as a byte slice.
// Returns an error if the response has an unsuccessful status code.
func (r *Response) Bytes() ([]byte, error) {
	if err := r.ExpectSuccess(); err != nil {
		return nil, err
	}

	defer r.Body.Close()
	return io.ReadAll(r.Body)
}

// ExpectContentType checks if the response has the expected content type
// and returns an error if it does not
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

// HasContentType returns true if the response has the expected content type
func (r *Response) HasContentType(contentType string) bool {
	return r.ExpectContentType(contentType) == nil
}

// ExpectStatusCode checks if the response has the expected status code
// and returns an error if it does not
func (r *Response) ExpectStatusCode(statusCode int) error {
	if r.Response == nil {
		return ErrNilResponse
	}
	if r.StatusCode != statusCode {
		return fmt.Errorf("unexpected status code: %d", r.StatusCode)
	}
	return nil
}

// HasStatusCode returns true if the response has the expected status code
func (r *Response) HasStatusCode(statusCode int) bool {
	return r.ExpectStatusCode(statusCode) == nil
}

// ExpectSuccess checks if the response has a successful status code
// and returns an error if it does not
func (r *Response) ExpectSuccess() error {
	if r.Response == nil {
		return ErrNilResponse
	}

	if code := r.StatusCode; code <= 199 || code >= 300 {
		return fmt.Errorf("unexpected status code: %s", r.Status)
	}

	return nil
}

// IsSuccess returns true if the response has a successful status code
func (r *Response) IsSuccess() bool {
	return r.ExpectSuccess() == nil
}
