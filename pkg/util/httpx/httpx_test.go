package httpx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

	mux.HandleFunc("/with-headers", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"message": "%s"}`, r.Header.Get("Test"))
	})

	mux.HandleFunc("/with-body", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			body = []byte(`"empty"`)
		}
		fmt.Fprintf(w, `{"message": %s}`, body)
	})

	mux.HandleFunc("/not-found", http.NotFound)

	s.server = httptest.NewServer(mux)
}

func (s *HttpSuite) TearDownSuite() {
	s.server.Close()
}

func (s *HttpSuite) TestDo() {
	// replace the real URL with the mock server's URL
	url := s.server.URL

	testCases := []struct {
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
			headers: map[string][]string{
				"Test": {"with headers"},
			},
			want: []byte(`{"message": "with headers"}`),
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
	t := s.T()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := Do(ctx, tc.method, url+tc.endpoint, tc.body, tc.headers)
			s.Equal(nil, err)
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)
			s.Equal(tc.want, body)
		})
	}

}

func TestResponse_Expec_Nil_Response(t *testing.T) {
	resp := &Response{Response: nil}
	asserts := assert.New(t)

	err := resp.ExpectSuccess()
	asserts.Equal(ErrNilResponse, err)
	asserts.False(resp.IsSuccess())

	err = resp.ExpectStatusCode(http.StatusOK)
	asserts.Equal(ErrNilResponse, err)
	asserts.False(resp.HasStatusCode(http.StatusOK))

	err = resp.ExpectContentType("application/json")
	asserts.Equal(ErrNilResponse, err)
	asserts.False(resp.HasContentType("application/json"))
}

func TestResponse_ExpectSuccess(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "OK",
			statusCode: http.StatusOK,
		},
		{
			name:       "No Content",
			statusCode: http.StatusNoContent,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{StatusCode: tt.statusCode},
		}

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectSuccess()
			asserts.Nil(err)
			asserts.True(resp.IsSuccess())
		})
	}
}

func TestResponse_ExpectSuccess_Error(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "Not Found",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "Internal Server Error",
			statusCode: http.StatusInternalServerError,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{StatusCode: tt.statusCode},
		}

		want := fmt.Errorf("unexpected status code: %d", tt.statusCode)

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectSuccess()
			asserts.Equal(want, err)
			asserts.False(resp.IsSuccess())
		})
	}
}

func TestResponse_ExpectStatusCode(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "OK",
			statusCode: http.StatusOK,
		},
		{
			name:       "No Content",
			statusCode: http.StatusNoContent,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{StatusCode: tt.statusCode},
		}

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectStatusCode(tt.statusCode)
			asserts.Nil(err)
			asserts.True(resp.HasStatusCode(tt.statusCode))
		})
	}
}
func TestResponse_ExpectStatusCode_Error(t *testing.T) {
	testCases := []struct {
		name   string
		actual int
		want   int
	}{
		{
			name:   "Not Found",
			actual: http.StatusNotFound,
			want:   http.StatusOK,
		},
		{
			name:   "Internal Server Error",
			actual: http.StatusInternalServerError,
			want:   http.StatusOK,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{StatusCode: tt.actual},
		}

		want := fmt.Errorf("unexpected status code: %d", tt.actual)

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectStatusCode(tt.want)
			asserts.Equal(want, err)
			asserts.False(resp.HasStatusCode(tt.want))
		})
	}
}

func TestResponse_ExpectContentType(t *testing.T) {
	testCases := []struct {
		name   string
		actual string
		want   string
	}{
		{
			name:   "normal",
			actual: "text/html",
			want:   "text/html",
		},
		{
			name:   "with charset",
			actual: "text/html; charset=utf-8",
			want:   "text/html",
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{
				Header: http.Header{"Content-Type": []string{tt.actual}},
			},
		}

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectContentType(tt.want)
			asserts.Nil(err)
			asserts.True(resp.HasContentType(tt.want))
		})
	}
}

func TestResponse_ExpectContentType_Error(t *testing.T) {
	testCases := []struct {
		name   string
		actual string
		want   string
	}{
		{
			name:   "normal",
			actual: "text/plain",
			want:   "text/html",
		},
		{
			name:   "with charset",
			actual: "application/xml; charset=utf-8",
			want:   "application/json",
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{
				Header: http.Header{"Content-Type": []string{tt.actual}},
			},
		}

		want := fmt.Errorf("unexpected content type: %s", tt.actual)

		t.Run(tt.name, func(t *testing.T) {
			err := resp.ExpectContentType(tt.want)
			asserts.Equal(want, err)
			asserts.False(resp.HasContentType(tt.want))
		})
	}
}

func TestResponse_Bytes(t *testing.T) {
	testCases := []struct {
		name   string
		actual []byte
		want   []byte
	}{
		{
			name:   "empty",
			actual: []byte{},
			want:   []byte{},
		},
		{
			name:   "normal",
			actual: []byte(`{"message": "hello"}`),
			want:   []byte(`{"message": "hello"}`),
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(tt.actual)),
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			res, err := resp.Bytes()
			asserts.Nil(err)
			asserts.Equal(tt.want, res)
		})
	}
}

func TestResponse_String(t *testing.T) {
	testCases := []struct {
		name   string
		actual string
		want   string
	}{
		{
			name:   "empty",
			actual: "",
			want:   "",
		},
		{
			name:   "normal",
			actual: `{"message": "hello"}`,
			want:   `{"message": "hello"}`,
		},
	}

	asserts := assert.New(t)

	for _, tt := range testCases {
		resp := &Response{
			Response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(tt.actual)),
			},
		}
		t.Run(tt.name, func(t *testing.T) {
			res, err := resp.String()
			asserts.Nil(err)
			asserts.Equal(tt.want, res)
		})
	}
}

func TestInto(t *testing.T) {
	type message struct {
		Message string `json:"message"`
	}

	resp := &Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(
				bytes.NewBufferString(`{"message": "hello"}`),
			),
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
		},
	}

	asserts := assert.New(t)

	res, err := Into[message](resp)
	asserts.Nil(err)
	asserts.Equal(res, message{Message: "hello"})
}

func TestInto_Error(t *testing.T) {
	type message struct {
		Message string `json:"message"`
	}

	resp := &Response{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(
				bytes.NewBufferString(`{"message": "hello"}`),
			),
		},
	}

	asserts := assert.New(t)

	_, err := Into[message](resp)
	asserts.NotNil(err)
}
