package github

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/nt54hamnghi/seaq/pkg/util/reqx"
)

// Repository represents a GitHub repository.
type Repository struct {
	Owner      string   // GitHub username or organization name that owns the repository
	Repo       string   // name of the repository
	OriginURL  *url.URL // original GitHub URL of the repository
	ContentURL *url.URL // GitHub API URL for accessing repository contents
}

// ParseRepositoryURL takes a GitHub repository URL and returns a Repository struct.
func ParseRepositoryURL(v string) (Repository, error) {
	repoURL, err := reqx.ParseURL("github.com")(v)
	if err != nil {
		return Repository{}, err
	}

	matches, err := reqx.ParsePath(repoURL.Path, "/{owner}/{repo}")
	if err != nil {
		return Repository{}, err
	}

	// originURL: https://github.com/:owner/:repo
	// contentURL: https://api.github.com/repos/:owner/:repo/contents
	contentURL := &url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("/repos/%s/%s/contents", matches["owner"], matches["repo"]),
	}

	return Repository{
		Owner:      matches["owner"],
		Repo:       matches["repo"],
		OriginURL:  repoURL,
		ContentURL: contentURL,
	}, nil
}

func (r Repository) DownloadPattern(ctx context.Context, patternName string) (string, error) {
	downloadURL := r.ContentURL.JoinPath("patterns", patternName, "system.md")
	res, err := reqx.Get(ctx, downloadURL.String(), map[string][]string{
		// for downloading the raw file content
		"Accept": {"application/vnd.github.raw+json"},
	})
	if err != nil {
		return "", err
	}

	return res.String()
}

// GetPatternNames retrieves all system pattern files from the repository's patterns directory.
// Returns a list of patterns found, or an error if the patterns
// directory doesn't exist or can't be accessed.
func (r Repository) GetPatternNames(ctx context.Context) ([]string, error) {
	content, err := reqx.GetAs[getContentResponse](ctx,
		r.ContentURL.String()+"/patterns",
		map[string][]string{
			// to get the contents in a consistent object format, with a git tree URL
			"Accept": {"application/vnd.github.object+json"},
		},
	)
	if err != nil {
		return nil, err
	}

	if content.Type != "dir" {
		return nil, fmt.Errorf("expected directory, got %s", content.Type)
	}

	if content.GitURL == nil {
		return nil, fmt.Errorf("unexpected: no git tree URL found")
	}

	// recursive=1: get the tree recursively
	tree, err := reqx.GetAs[getGitTree](ctx, *content.GitURL+"?recursive=1", nil)
	if err != nil {
		return nil, err
	}

	patterns := make([]string, 0, len(tree.Tree))
	for _, l := range tree.Tree {
		if l.Type == nil || l.Size == nil || l.Path == nil {
			continue
		}

		// nested patterns are not supported yet!
		if strings.HasSuffix(*l.Path, "system.md") && *l.Type == "blob" && *l.Size > 0 {
			if strings.Count(*l.Path, "/") == 1 {
				patterns = append(patterns, strings.TrimSuffix(*l.Path, "/system.md"))
			}
		}
	}

	return patterns, nil
}

// getContentResponse represents the GitHub API response for getting repository contents.
// It is the list of directory items
//
// See: https://docs.github.com/rest/repos/contents#get-repository-getContentResponse
type getContentResponse struct {
	Type   string  `json:"type"`    // type (e.g. "blob", "tree")
	GitURL *string `json:"git_url"` // Git tree URL
}

// getGitTree represents the GitHub API response for getting a flat tree of a path.
// It is the hierarchy between files in a Git repository.
//
// See: https://docs.github.com/rest/git/trees#get-a-tree
type getGitTree struct {
	Tree      []leaf `json:"tree"`
	Truncated bool   `json:"truncated"` // true if exceeds the limit of 100,000 entries with a maximum size of 7 MB
}

// leaf represents a single node in a Git tree
type leaf struct {
	Type *string `json:"type"` // object type (e.g. "blob", "tree")
	Size *int64  `json:"size"`
	Path *string `json:"path"`
	URL  *string `json:"url"`
}
