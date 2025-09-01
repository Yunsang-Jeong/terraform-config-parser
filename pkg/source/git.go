package source

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/filesystem"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
)

// GitSource represents a Git repository source
type GitSource struct {
	URL    string
	Config SourceConfig
}

func NewGitSource(url string, config SourceConfig) *GitSource {
	return &GitSource{
		URL:    url,
		Config: config,
	}
}

func (s *GitSource) Fetch() (filesystem.FileReader, string, error) {
	// Create in-memory filesystem for Git operations
	billyFs := memfs.New()

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL:   s.URL,
		Depth: 1,
	}

	// Set authentication if available
	if auth := s.getAuthentication(); auth != nil {
		cloneOptions.Auth = auth
	}

	// Set branch if specified
	if s.Config.Branch != "" {
		cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + s.Config.Branch)
		cloneOptions.SingleBranch = true
	}

	// Clone repository directly to in-memory storage
	_, err := git.Clone(memory.NewStorage(), billyFs, cloneOptions)
	if err != nil {
		return nil, "", fmt.Errorf("failed to clone repository: %w", err)
	}

	// Create Billy adapter
	billyAdapter := filesystem.NewBillyAdapter(billyFs)

	// Return root path based on subdirectory config
	rootPath := "."
	if s.Config.SubDir != "" {
		rootPath = s.Config.SubDir
	}

	return billyAdapter, rootPath, nil
}

func (s *GitSource) getAuthentication() *http.BasicAuth {
	// Parse URL to determine provider
	parsedURL, err := url.Parse(s.URL)
	if err != nil {
		return nil
	}

	hostname := strings.ToLower(parsedURL.Hostname())

	// GitHub
	if strings.Contains(hostname, "github.com") {
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			return &http.BasicAuth{
				Username: "token",
				Password: token,
			}
		}
	}

	// GitLab
	if strings.Contains(hostname, "gitlab.com") || strings.Contains(hostname, "gitlab") {
		if token := os.Getenv("GITLAB_TOKEN"); token != "" {
			return &http.BasicAuth{
				Username: "gitlab-ci-token",
				Password: token,
			}
		}
	}

	// GIT_TOKEN (generic)
	if token := os.Getenv("GIT_TOKEN"); token != "" {
		return &http.BasicAuth{
			Username: "token",
			Password: token,
		}
	}

	return nil
}

func (s *GitSource) Cleanup() error {
	return nil
}
