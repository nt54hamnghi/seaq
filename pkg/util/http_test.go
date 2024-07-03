package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type HttpSuite struct {
	suite.Suite
	server *httptest.Server
}

func TestHttpSuite(t *testing.T) {
	suite.Run(t, &HttpSuite{})
}

func (s *HttpSuite) SetupSuite() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "%s called"}`, r.Method)
	})

	mux.HandleFunc("/get-with-headers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "%s"}`, r.Header.Get("Test"))
	})

	mux.HandleFunc("/post-with-body", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fmt.Fprintf(w, `{"message": %s}`, body)
	})

	mux.HandleFunc("/not-found", http.NotFound)

	s.server = httptest.NewServer(mux)
}

func (s *HttpSuite) TearDownSuite() {
	s.server.Close()
}

func (s *HttpSuite) TestDo() {

	url := s.server.URL

	type response struct {
		Message string `json:"message"`
	}

	testCases := []struct {
		name     string
		method   string
		endpoint string
		body     any
		headers  map[string][]string
		expected response
	}{
		{
			name:     "get",
			method:   http.MethodGet,
			expected: response{Message: "GET called"},
		},
		{
			name:     "post",
			method:   http.MethodPost,
			expected: response{Message: "POST called"},
		},
	}

	ctx := context.Background()
	t := s.T()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := do[response](ctx, tc.method, url, nil, nil)
			s.Equal(nil, err)
			s.Equal(tc.expected, res)

		})
	}

}

func (s *HttpSuite) TestDoRaw_Fail() {
	url := s.server.URL
	ctx := context.Background()
	expectedErr := errors.New("unexpected status code: 404")
	_, err := doRaw(ctx, http.MethodGet, url+"/not-found", nil, nil)

	s.Equal(expectedErr, err)
}

func (s *HttpSuite) TestDoRaw() {
	// replace the real URL with the mock server's URL
	url := s.server.URL

	testCases := []struct {
		name     string
		method   string
		endpoint string
		body     any
		headers  map[string][]string
		expected []byte
	}{
		{
			name:     "get",
			method:   http.MethodGet,
			expected: []byte(`{"message": "GET called"}`),
		},
		{
			name:     "post",
			method:   http.MethodPost,
			expected: []byte(`{"message": "POST called"}`),
		},
		{
			name:     "getWithHeaders",
			method:   http.MethodGet,
			endpoint: "/get-with-headers",
			headers: map[string][]string{
				"Test": {"with headers"},
			},
			expected: []byte(`{"message": "with headers"}`),
		},
		{
			name:     "postWithBody",
			method:   http.MethodPost,
			endpoint: "/post-with-body",
			body:     "body",
			expected: []byte(`{"message": "body"}`),
		},
	}

	ctx := context.Background()
	t := s.T()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := doRaw(ctx, tc.method, url+tc.endpoint, tc.body, tc.headers)

			s.Equal(nil, err)
			s.Equal(tc.expected, res)
		})
	}

}
