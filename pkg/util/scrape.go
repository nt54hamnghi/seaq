package util

import (
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/microcosm-cc/bluemonday"
)

func HTMLToMarkdown(safeHTML []byte) ([]byte, error) {
	converter := md.NewConverter("", true, nil)
	// converter.Use(plugin.Table())

	return converter.ConvertBytes(safeHTML)
}

func SanitizeHTML(html []byte) []byte {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("href").OnElements("a")
	policy.RequireParseableURLs(true)
	policy.RequireNoFollowOnFullyQualifiedLinks(true)

	return policy.SanitizeBytes(html)
}
