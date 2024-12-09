package reqx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HTTPSuite struct {
	suite.Suite
	server *httptest.Server
	client *http.Client
}

func TestHTTPSuite(t *testing.T) {
	suite.Run(t, &HTTPSuite{})
}

func (s *HTTPSuite) SetupSuite() {
	mux := http.NewServeMux()

	// endpoint for GET and POST
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "%s called"}`, r.Method)
	})

	// endpoint for testing headers
	mux.HandleFunc("/with-headers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "%s"}`, r.Header.Get("X-Test"))
	})

	// endpoint for testing body
	mux.HandleFunc("/with-body", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			body = []byte(`"empty"`)
		}
		fmt.Fprintf(w, `{"message": %s}`, body)
	})

	s.server = httptest.NewServer(mux)
	s.client = s.server.Client()
}

func (s *HTTPSuite) TearDownSuite() {
	s.server.Close()
}

func (s *HTTPSuite) TestWithClient() {
	tests := []struct {
		name     string
		method   string
		endpoint string
		body     any
		headers  map[string][]string
		want     []byte
	}{
		{
			name:   "get",
			method: http.MethodGet,
			want:   []byte(`{"message": "GET called"}`),
		},
		{
			name:   "post",
			method: http.MethodPost,
			want:   []byte(`{"message": "POST called"}`),
		},
		{
			name:     "getWithHeaders",
			method:   http.MethodGet,
			endpoint: "/with-headers",
			headers:  map[string][]string{"X-Test": {"with headers"}},
			want:     []byte(`{"message": "with headers"}`),
		},
		{
			name:     "postWithBody",
			method:   http.MethodPost,
			endpoint: "/with-body",
			body:     "body",
			want:     []byte(`{"message": "body"}`),
		},
		{
			name:     "postWithNilBody",
			method:   http.MethodPost,
			endpoint: "/with-body",
			body:     nil,
			want:     []byte(`{"message": "empty"}`),
		},
	}

	ctx := context.Background()
	url := s.server.URL
	makeRequest := WithClient(s.client)

	for _, tt := range tests {
		s.Run(tt.name, func() {
			res, err := makeRequest(ctx, tt.method, url+tt.endpoint, tt.headers, tt.body)
			if s.NoError(err) {
				body, _ := io.ReadAll(res.Body)
				s.Equal(tt.want, body)
			}
			defer res.Body.Close()
		})
	}
}

func (s *HTTPSuite) TestWithClient__InputValidation() {
	tests := []struct {
		name    string
		client  *http.Client
		ctx     context.Context
		method  string
		url     string
		wantErr string
	}{
		{
			name:    "nil client",
			client:  nil,
			ctx:     context.Background(),
			method:  http.MethodGet,
			url:     "http://example.com",
			wantErr: "client cannot be nil",
		},
		{
			name:    "nil context",
			client:  s.client,
			ctx:     nil,
			method:  http.MethodGet,
			url:     "http://example.com",
			wantErr: "context cannot be nil",
		},
		{
			name:    "empty url",
			client:  s.client,
			ctx:     context.Background(),
			method:  http.MethodGet,
			url:     "",
			wantErr: "url cannot be empty",
		},
		{
			name:    "unsupported method",
			client:  s.client,
			ctx:     context.Background(),
			method:  http.MethodPatch,
			url:     "http://example.com",
			wantErr: "unsupported method: PATCH",
		},
	}

	r := s.Require()

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := WithClient(tt.client)
			_, err := req(tt.ctx, tt.method, tt.url, nil, nil)
			r.Error(err)
			r.Contains(err.Error(), tt.wantErr)
		})
	}
}

func TestInto(t *testing.T) {
	type message struct {
		Message string `json:"message"`
	}

	tests := []struct {
		name       string
		raw        *http.Response
		want       message
		wantErr    bool
		errMessage string
	}{
		{
			name: "valid JSON",
			raw: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": "hello"}`)),
			},
			want:    message{Message: "hello"},
			wantErr: false,
		},
		{
			name: "invalid JSON",
			raw: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": invalid}`)),
			},
			wantErr:    true,
			errMessage: "failed to unmarshal response body",
		},
		{
			name: "unexpected content type",
			raw: &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": "hello"}`)),
			},
			wantErr:    true,
			errMessage: "unexpected content type",
		},
		{
			name: "non-success status code",
			raw: &http.Response{
				StatusCode: http.StatusInternalServerError,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": "error"}`)),
			},
			wantErr:    true,
			errMessage: "unexpected status code",
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := Into[message](&Response{Response: tt.raw})
			if tt.wantErr {
				r.Error(err)
				r.Contains(err.Error(), tt.errMessage)
				return
			}

			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}

func TestResponse_ExpectContentType(t *testing.T) {
	testCases := []struct {
		name        string
		actual      *http.Response
		contentType string
		wantErr     bool
	}{
		{
			name: "match",
			actual: &http.Response{
				Header: http.Header{"Content-Type": []string{"application/json"}},
			},
			contentType: "application/json",
			wantErr:     false,
		},
		{
			name: "not match",
			actual: &http.Response{
				Header: http.Header{"Content-Type": []string{"text/plain"}},
			},
			contentType: "application/json",
			wantErr:     true,
		},
		{
			name:    "nil response",
			actual:  nil,
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		res := &Response{Response: tt.actual}

		t.Run(tt.name, func(*testing.T) {
			err := res.ExpectContentType(tt.contentType)
			if tt.actual == nil {
				r.Equal(ErrNilResponse, err)
				return
			}

			if tt.wantErr {
				r.Error(err)
				return
			}

			r.NoError(err)
			r.True(res.HasContentType(tt.contentType))
		})
	}
}

func TestResponse_ExpectStatusCode(t *testing.T) {
	testCases := []struct {
		name       string
		actual     *http.Response
		statusCode int
		wantErr    bool
	}{
		{
			name:       "match",
			actual:     &http.Response{StatusCode: http.StatusOK},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "not match",
			actual:     &http.Response{StatusCode: http.StatusNotFound},
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:    "nil response",
			actual:  nil,
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		res := &Response{Response: tt.actual}

		t.Run(tt.name, func(*testing.T) {
			err := res.ExpectStatusCode(tt.statusCode)
			if tt.actual == nil {
				r.Equal(ErrNilResponse, err)
				return
			}

			if tt.wantErr {
				r.Error(err)
				return
			}

			r.NoError(err)
			r.True(res.HasStatusCode(tt.statusCode))
		})
	}
}

func TestResponse_ExpectSuccess(t *testing.T) {
	tests := []struct {
		name    string
		raw     *http.Response
		wantErr bool
	}{
		{
			name: "OK",
			raw:  &http.Response{StatusCode: http.StatusOK},
		},
		{
			name: "NoContent",
			raw:  &http.Response{StatusCode: http.StatusNoContent},
		},
		{
			name:    "NotFound",
			raw:     &http.Response{StatusCode: http.StatusNotFound},
			wantErr: true,
		},
		{
			name:    "InternalServerError",
			raw:     &http.Response{StatusCode: http.StatusInternalServerError},
			wantErr: true,
		},
		{
			name:    "nil response",
			raw:     nil,
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		res := &Response{Response: tt.raw}

		t.Run(tt.name, func(*testing.T) {
			err := res.ExpectSuccess()

			if tt.raw == nil {
				r.Equal(ErrNilResponse, err)
				return
			}

			if tt.wantErr {
				r.Error(err)
				return
			}

			r.NoError(err)
			r.True(res.IsSuccess())
		})
	}
}

func TestResponse_Bytes(t *testing.T) {
	testCases := []struct {
		name string
		raw  *http.Response
		want []byte
	}{
		{
			name: "empty",
			raw: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(nil)),
			},
			want: []byte{},
		},
		{
			name: "normal",
			raw: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"message": "hello"}`)),
			},
			want: []byte(`{"message": "hello"}`),
		},
	}

	r := require.New(t)

	for _, tt := range testCases {
		t.Run(tt.name, func(*testing.T) {
			got, err := (&Response{Response: tt.raw}).Bytes()
			r.NoError(err)
			r.Equal(tt.want, got)
		})
	}
}
