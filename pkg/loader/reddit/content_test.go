package reddit

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_replies_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name      string
		json      string
		wantError error
		wantNil   bool
	}{
		{
			name:      "emptyString",
			json:      `""`,
			wantError: nil,
			wantNil:   true,
		},
		{
			name:      "nonEmptyString",
			json:      `"test"`,
			wantError: errors.New("unexpected non-empty string replies"),
		},
		{
			name:      "validListing",
			json:      `{"kind":"Listing","data":{"children":[]}}`,
			wantError: nil,
			wantNil:   false,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			var rp replies
			err := rp.UnmarshalJSON([]byte(tc.json))

			if tc.wantError != nil {
				r.Equal(tc.wantError, err)
				return
			}

			if tc.wantNil {
				r.Nil(rp.listing)
			} else {
				r.NotNil(rp.listing)
				r.Equal("Listing", rp.Kind)
				r.Empty(rp.Data.Children)
			}
		})
	}
}

func Test_listing_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name              string
		json              string
		wantChildrenLen   int
		wantChildrenKinds []string
	}{
		{
			name:              "empty",
			json:              `{"kind":"Listing","data":{"children":[]}}`,
			wantChildrenLen:   0,
			wantChildrenKinds: []string{},
		},
		{
			name:              "unsupported",
			json:              `{"kind":"Listing","data":{"children":[{"kind":"t42","data":{}}]}}`,
			wantChildrenLen:   1,
			wantChildrenKinds: []string{"t42"},
		},
		{
			name: "comment",
			json: `{
				"kind": "Listing",
				"data": {
					"children": [
						{"kind": "t1", "data": { "body": "this is a comment" }}
					]
				}
			}`,
			wantChildrenLen:   1,
			wantChildrenKinds: []string{"t1"},
		},
		{
			name: "link",
			json: `{
				"kind": "Listing",
				"data": {
					"children": [
						{"kind": "t3", "data": { "title": "click me", "url": "https://www.google.com" }}
					]
				}
			}`,
			wantChildrenLen:   1,
			wantChildrenKinds: []string{"t3"},
		},
		{
			name: "mixed",
			json: `{
				"kind": "Listing",
				"data": {
					"children": [
						{"kind": "t3", "data": {}},
						{"kind": "t1", "data": {}},
						{"kind": "t42", "data": {}}
					]
				}
			}`,
			wantChildrenLen:   3,
			wantChildrenKinds: []string{"t3", "t1", "t42"},
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			var l listing
			err := l.UnmarshalJSON([]byte(tc.json))

			r.NoError(err)
			r.Equal("Listing", l.Kind)
			r.Len(l.Data.Children, tc.wantChildrenLen)

			for i, wantKind := range tc.wantChildrenKinds {
				child := l.Data.Children[i]

				r.Equal(wantKind, child.Kind)

				obj := child.Data
				switch child.Kind {
				case "t1":
					_, ok := obj.(*comment)
					r.True(ok, "object is not a comment")
				case "t3":
					_, ok := obj.(*link)
					r.True(ok, "object is not a link")
				default:
					_, ok := obj.(*unsupported)
					r.True(ok, "object is not an unsupported")
				}
			}
		})
	}
}

func Test_parseRedditUrl(t *testing.T) {
	testCases := []struct {
		name    string
		rawURL  string
		wantURL string
	}{
		{
			name:    "validPostUrl",
			rawURL:  "https://www.reddit.com/r/golang/comments/123/hello_world/",
			wantURL: "https://www.reddit.com/r/golang/comments/123/hello_world/.json",
		},
		{
			name:    "validCommentUrl",
			rawURL:  "https://www.reddit.com/r/golang/comments/123/comment/456/",
			wantURL: "https://www.reddit.com/r/golang/comments/123/comment/456/.json",
		},
		{
			name:    "urlAlreadyHasJson",
			rawURL:  "https://www.reddit.com/r/golang/comments/123/.json",
			wantURL: "https://www.reddit.com/r/golang/comments/123/.json",
		},
		{
			name:    "redditHomepage",
			rawURL:  "https://www.reddit.com/",
			wantURL: "https://www.reddit.com/.json",
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			result, err := parseRedditURL(tc.rawURL)

			r.NoError(err)
			r.NotNil(result)
			r.Equal(tc.wantURL, result.String())
		})
	}
}

func Test_parseRedditPath(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		wantMap map[string]string
		wantErr bool
	}{
		{
			name: "untitledPost",
			url:  "https://www.reddit.com/r/golang/comments/123abc",
			wantMap: map[string]string{
				"subreddit": "golang",
				"post-id":   "123abc",
			},
		},
		{
			name: "untitledPostWithJson",
			url:  "https://www.reddit.com/r/golang/comments/123abc/.json",
			wantMap: map[string]string{
				"subreddit": "golang",
				"post-id":   "123abc",
			},
		},
		{
			name: "titledPost",
			url:  "https://www.reddit.com/r/golang/comments/123abc/hello_world",
			wantMap: map[string]string{
				"subreddit": "golang",
				"post-id":   "123abc",
				"title":     "hello_world",
			},
		},
		{
			name: "titledPostWithJson",
			url:  "https://www.reddit.com/r/golang/comments/123abc/hello_world/.json",
			wantMap: map[string]string{
				"subreddit": "golang",
				"post-id":   "123abc",
				"title":     "hello_world",
			},
		},
		{
			name: "comment",
			url:  "https://www.reddit.com/r/golang/comments/123abc/comment/456def",
			wantMap: map[string]string{
				"subreddit":  "golang",
				"post-id":    "123abc",
				"comment-id": "456def",
			},
		},
		{
			name: "commentWithJson",
			url:  "https://www.reddit.com/r/golang/comments/123abc/comment/456def/.json",
			wantMap: map[string]string{
				"subreddit":  "golang",
				"post-id":    "123abc",
				"comment-id": "456def",
			},
		},
		{
			name:    "invalidPath",
			url:     "https://www.reddit.com/r/golang/",
			wantErr: true,
		},
		{
			name:    "invalidFormat",
			url:     "https://www.reddit.com/r/golang/posts/123",
			wantErr: true,
		},
		{
			name:    "emptyPath",
			url:     "https://www.reddit.com/",
			wantErr: true,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			u, err := url.Parse(tc.url)
			r.NoError(err)

			result, err := parseRedditPath(u)

			if tc.wantErr {
				r.Error(err)
				r.Nil(result)
				return
			}

			r.NoError(err)
			r.Equal(tc.wantMap, result)
		})
	}
}

func Test_GetRedditContentAsDocuments(t *testing.T) {
	// Mock Reddit API responses
	postResponse := `[
		{
			"kind": "Listing",
			"data": {
				"children": [
					{
						"kind": "t3",
						"data": {
							"id": "123abc",
							"name": "t3_123abc",
							"author": "testuser",
							"title": "Test Post Title",
							"url": "https://example.com",
							"domain": "example.com",
							"selftext": "This is the post content",
							"subreddit": "golang",
							"subreddit_id": "t5_2rc7j",
							"created": 0.0,
							"over_18": false
						}
					}
				]
			}
		},
		{
			"kind": "Listing",
			"data": {
				"children": [{"kind": "t3","data": {}}]
			}
		}
	]`

	commentResponse := `[
		{
			"kind": "Listing",
			"data": {
				"children": [{"kind": "t3","data": {}}]
			}
		},
		{
			"kind": "Listing",
			"data": {
				"children": [
					{
						"kind": "t1",
						"data": {
							"id": "456def",
							"name": "t1_456def",
							"author": "commenter",
							"body": "This is a test comment",
							"parent_id": "t3_123abc",
							"subreddit": "golang",
							"subreddit_id": "t5_2rc7j",
							"created": 0.0
						}
					}
				]
			}
		}
	]`

	testCases := []struct {
		name         string
		inputPath    string
		mockResponse string
		statusCode   int
		wantErr      bool
		wantContent  string
		wantMetadata map[string]any
	}{
		{
			name:         "validPost",
			inputPath:    "/r/golang/comments/123abc/test_post/.json",
			mockResponse: postResponse,
			statusCode:   200,
			wantErr:      false,
			wantContent:  "Test Post Title\n\nhttps://example.com\n\nThis is the post content",
			wantMetadata: map[string]any{
				"id":           "123abc",
				"name":         "t3_123abc",
				"author":       "testuser",
				"domain":       "example.com",
				"subreddit":    "golang",
				"subreddit_id": "t5_2rc7j",
				"created":      float32(0.0),
				"over_18":      false,
			},
		},
		{
			name:         "validComment",
			inputPath:    "/r/golang/comments/123abc/comment/456def/.json",
			mockResponse: commentResponse,
			statusCode:   200,
			wantErr:      false,
			wantContent:  "This is a test comment",
			wantMetadata: map[string]any{
				"id":           "456def",
				"name":         "t1_456def",
				"author":       "commenter",
				"parent_id":    "t3_123abc",
				"subreddit":    "golang",
				"subreddit_id": "t5_2rc7j",
				"created":      float32(0.0),
			},
		},
		{
			name:         "unexpectedListingLength",
			inputPath:    "/r/golang/comments/123abc/.json",
			mockResponse: `[{"kind":"Listing","data":{"children":[]}}]`,
			statusCode:   200,
			wantErr:      true,
		},
		{
			name:      "emptyChildren",
			inputPath: "/r/golang/comments/123abc/.json",
			mockResponse: `[
				{"kind":"Listing","data":{"children":[]}},
				{"kind":"Listing","data":{"children":[]}}
			]`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				// Verify the request path ends with .json
				if !strings.HasSuffix(req.URL.Path, ".json") {
					t.Errorf("Expected request path to end with .json, got: %s", req.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.statusCode)
				if tc.statusCode == 200 && tc.mockResponse != "" {
					if _, err := w.Write([]byte(tc.mockResponse)); err != nil {
						t.Errorf("Failed to write response: %v", err)
					}
				}
			}))
			defer server.Close()

			// Build full URL with test server
			fullURL, err := url.ParseRequestURI(server.URL + tc.inputPath)
			r.NoError(err)

			// Execute the function
			ctx := context.Background()
			docs, err := getRedditContentAsDocuments(ctx, fullURL)

			// Verify results
			if tc.wantErr {
				r.Error(err)
				r.Nil(docs)
				return
			}

			r.NoError(err)
			r.Len(docs, 1)
			r.Equal(tc.wantContent, docs[0].PageContent)
			r.Equal(tc.wantMetadata, docs[0].Metadata)
		})
	}
}
