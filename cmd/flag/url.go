package flag

import (
	"net/url"
)

type URL url.URL

// String implements the pflag.Value interface
// It returns the string representation of the URL
func (u *URL) String() string {
	return (*url.URL)(u).String()
}

// Set implements the pflag.Value interface
// It parses the input string and sets the URL
func (u *URL) Set(s string) error {
	res, err := url.ParseRequestURI(s)
	if err != nil {
		return err
	}

	*u = URL(*res)
	return nil
}

// Type implements the pflag.Value interface
// It returns the type of the URL flag in help message
func (u *URL) Type() string {
	return "url"
}
