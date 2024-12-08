package reqx

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	const host = "example.com"

	tests := []struct {
		name   string
		rawURL string
	}{
		{
			name:   "valid URL with matching host",
			rawURL: "http://example.com/path",
		},
		{
			name:   "valid URL with fragment",
			rawURL: "http://example.com/path#fragment",
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			u, err := ParseURL(host)(tt.rawURL)

			r.NoError(err)
			r.NotNil(u)
			r.Equal(host, u.Hostname())
		})
	}
}

func TestParseURL_Error(t *testing.T) {
	const host = "example.com"

	tests := []struct {
		name   string
		rawURL string
	}{
		{
			name:   "relative URL",
			rawURL: "example.com/path",
		},
		{
			name:   "mismatched host",
			rawURL: "http://another.com/path",
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			u, err := ParseURL(host)(tt.rawURL)

			r.Error(err)
			r.Nil(u)
		})
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		tmpl string
		want map[string]string
	}{
		{
			name: "normal",
			path: "/users/123/orders/456",
			tmpl: "/users/{userId}/orders/{orderId}",
			want: map[string]string{"userId": "123", "orderId": "456"},
		},
		{
			name: "single element",
			path: "/users",
			tmpl: "/{entity}",
			want: map[string]string{"entity": "users"},
		},
		{
			name: "no leading slash",
			path: "users",
			tmpl: "{entity}",
			want: map[string]string{"entity": "users"},
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			result, err := ParsePath(tt.path, tt.tmpl)

			r.NoError(err)
			r.Equal(tt.want, result)
		})
	}
}

func TestParsePath_Error(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		tmpl    string
		wantErr error
	}{
		{
			name:    "path and template mismatch",
			path:    "/users/123",
			tmpl:    "/users",
			wantErr: fmt.Errorf("path and template mismatch"),
		},
		{
			name:    "empty path",
			path:    "",
			tmpl:    "/{entity}",
			wantErr: errors.New("empty path"),
		},
		{
			name:    "empty template",
			path:    "/users",
			tmpl:    "",
			wantErr: errors.New("empty template"),
		},
		{
			name:    "empty placeholder",
			path:    "/users",
			tmpl:    "/",
			wantErr: ErrInvalidPlaceholder,
		},
		{
			name:    "placeholder without text",
			path:    "/users",
			tmpl:    "/{}",
			wantErr: ErrInvalidPlaceholder,
		},
		{
			name:    "placeholder without closing brace",
			path:    "/users",
			tmpl:    "/{entity",
			wantErr: ErrInvalidPlaceholder,
		},
		{
			name:    "placeholder without opening brace",
			path:    "/users",
			tmpl:    "/entity}",
			wantErr: ErrInvalidPlaceholder,
		},

		{
			name:    "missing parts",
			path:    "/users//orders/456",
			tmpl:    "/users/{userId}/orders/{orderId}",
			wantErr: errors.New("empty path part"),
		},
	}

	r := require.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			_, err := ParsePath(tt.path, tt.tmpl)

			r.Equal(tt.wantErr, err)
		})
	}
}
