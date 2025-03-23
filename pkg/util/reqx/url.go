package reqx

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidPlaceholder = errors.New("invalid placeholder format")

// ParseURL returns a function that parses URLs and validates the hostname for a specific host.
// The returned function takes a raw URL string and returns a parsed *url.URL if:
// - The URL is valid and can be parsed
// - The URL's hostname matches the specified host
// - Any URL fragments (#) are removed before parsing
//
// Example:
//
//	parser := ParseURL("example.com")
//	url, err := parser("https://example.com/path?q=value#fragment")
//	// url.Host == "example.com"
//	// url.Path == "/path"
//	// url.RawQuery == "q=value"
//	// Fragment is removed
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

// ParsePath extracts named parameters from a URL path based on a template pattern.
// It matches path segments against a template where placeholders are denoted by {name}.
// Returns a map of placeholder names to their corresponding values from the path.
//
// Parameters:
//   - path: The actual URL path to parse (e.g., "/users/123/posts/456")
//   - tmpl: The template pattern with placeholders (e.g., "/users/{id}/posts/{postId}")
//
// Returns:
//   - map[string]string: Key-value pairs of placeholder names and their values
//   - error: If path is empty, template is empty, segments don't match, or invalid placeholder format
//
// Example:
//
//	matches, err := ParsePath("/users/123/posts/456", "/users/{userId}/posts/{postId}")
//	// matches = map[string]string{
//	//   "userId": "123",
//	//   "postId": "456",
//	// }
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

		if p == "" {
			return nil, errors.New("empty path part")
		}

		key, err := extractKey(t)
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

func extractKey(input string) (string, error) {
	if len(input) < 3 {
		return "", ErrInvalidPlaceholder
	}

	if input[0] != '{' || input[len(input)-1] != '}' {
		return "", ErrInvalidPlaceholder
	}

	return input[1 : len(input)-1], nil
}
