package reddit

import (
	"encoding/json"
	"errors"
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

func Test_comment_UnmarshalJSON(t *testing.T) {
	testCases := []struct {
		name string
		json string
	}{
		{
			name: "nested_replies",
			json: `{
				"id": "123",
				"body": "parent comment",
				"replies": {
					"kind": "Listing",
					"data": {
						"children": [
							{
								"kind": "t1",
								"data": {
									"id": "456",
									"body": "child comment",
									"replies": {
										"kind": "Listing",
										"data": {
											"children": [
												{"kind": "t1", "data": {"id": "789", "body": "grandchild comment", "replies": ""}}
											]
										}
									}
								}
							}
						]
					}
				}
			}`,
		},
	}

	r := require.New(t)

	for _, tc := range testCases {
		t.Run(tc.name, func(_ *testing.T) {
			var c comment
			err := json.Unmarshal([]byte(tc.json), &c)
			r.NoError(err)

			// Navigate to the most deeply nested comment
			grandchildComment := c.Replies.Data.Children[0].Data.(*comment).Replies.Data.Children[0]
			r.Equal("t1", grandchildComment.Kind)

			deepestComment, ok := grandchildComment.Data.(*comment)
			r.True(ok)
			r.Equal("grandchild comment", deepestComment.Body)
		})
	}
}
