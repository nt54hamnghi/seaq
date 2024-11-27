package reqx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrMismatchedPathTemplate = errors.New("path and template don't match")

type ErrInvalidURL struct {
	inner error
}

func (e ErrInvalidURL) Error() string {
	return "invalid URL: " + e.inner.Error()
}

func (e ErrInvalidURL) Unwrap() error {
	return e.inner
}

func ParseURL(host string) func(rawUrl string) (*url.URL, error) {
	return func(rawUrl string) (*url.URL, error) {
		// remove fragment
		if fidx := strings.Index(rawUrl, "#"); fidx != -1 {
			rawUrl = rawUrl[:fidx]
		}

		// parse url
		parsed, err := url.ParseRequestURI(rawUrl)
		if err != nil {
			return nil, ErrInvalidURL{inner: err}
		}

		// validate hostname
		if parsed.Hostname() != host {
			return nil, ErrInvalidURL{
				inner: fmt.Errorf("invalid hostname, expected %q, got %q", host, parsed.Hostname()),
			}
		}

		return parsed, nil
	}
}

func ParsePath(path, tmpl string) (map[string]string, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	if tmpl == "" {
		return nil, errors.New("empty template")
	}

	pathParts := strings.Split(path, "/")
	tmplParts := strings.Split(tmpl, "/")

	if len(pathParts) != len(tmplParts) {
		return nil, ErrMismatchedPathTemplate
	}

	matches := make(map[string]string)

	for i := 0; i < len(pathParts); i++ {
		p, t := pathParts[i], tmplParts[i]
		if p == t {
			continue
		}

		if !strings.HasPrefix(t, "{") || !strings.HasSuffix(t, "}") {
			return nil, ErrMismatchedPathTemplate
		}

		key := t[1 : len(t)-1]
		if key == "" {
			return nil, fmt.Errorf("invalid template part %q at %d", t, i)
		}
		matches[key] = p
	}

	if len(matches) == 0 {
		return nil, errors.New("template doesn't contain any placeholders")
	}

	return matches, nil
}
