package reqx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidPlaceholder = errors.New("invalid placeholder format")

func ParseURL(host string) func(rawUrl string) (*url.URL, error) {
	return func(rawUrl string) (*url.URL, error) {
		// remove fragment
		if fidx := strings.Index(rawUrl, "#"); fidx != -1 {
			rawUrl = rawUrl[:fidx]
		}

		// parse url
		parsed, err := url.ParseRequestURI(rawUrl)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		// validate hostname
		if parsed.Hostname() != host {
			return nil, fmt.Errorf("invalid hostname, expected %q, got %q", host, parsed.Hostname())
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
		return nil, errors.New("path and template mismatch")
	}

	matches := make(map[string]string)

	for i := 0; i < len(pathParts); i++ {
		p, t := pathParts[i], tmplParts[i]
		if p == t {
			continue
		}

		key, err := extract(t)
		if err != nil {
			return nil, err
		}

		matches[key] = p
	}

	if len(matches) == 0 {
		return nil, errors.New("template doesn't contain any placeholders")
	}

	return matches, nil
}

func extract(input string) (string, error) {
	if len(input) < 3 {
		return "", ErrInvalidPlaceholder
	}

	if input[0] != '{' || input[len(input)-1] != '}' {
		return "", ErrInvalidPlaceholder
	}

	return input[1 : len(input)-1], nil
}
