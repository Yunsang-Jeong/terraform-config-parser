package source

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/filesystem"
	"github.com/Yunsang-Jeong/terraform-config-parser/pkg/logger"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"go.uber.org/zap"
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
	logger.Info("Starting git repository clone", zap.String("url", s.URL), zap.String("ref", s.Config.Ref), zap.String("subdir", s.Config.SubDir))

	// Create in-memory filesystem for Git operations
	billyFs := memfs.New()

	// Clone options
	cloneOptions := &git.CloneOptions{
		URL:   s.URL,
		Depth: 1,
	}

	// Set authentication if available
	if auth := s.getAuthentication(); auth != nil {
		logger.Debug("Using authentication for git clone", zap.String("username", auth.Username))
		cloneOptions.Auth = auth
	} else {
		logger.Debug("No authentication configured for git clone")
	}

	// Set reference (branch, tag, or commit) if specified
	if s.Config.Ref != "" {
		refType := detectRefType(s.Config.Ref)
		logger.Debug("Cloning specific reference", zap.String("ref", s.Config.Ref), zap.String("type", getRefTypeName(refType)))

		switch refType {
		case RefTypeBranch:
			cloneOptions.ReferenceName = plumbing.ReferenceName("refs/heads/" + s.Config.Ref)
			cloneOptions.SingleBranch = true
		case RefTypeTag:
			cloneOptions.ReferenceName = plumbing.ReferenceName("refs/tags/" + s.Config.Ref)
			cloneOptions.SingleBranch = true
		case RefTypeCommit:
			// For commits, we'll clone and then checkout the specific commit
			logger.Debug("Will checkout commit after clone", zap.String("commit", s.Config.Ref))
		}
	} else {
		logger.Debug("Cloning default branch")
	}

	// Clone repository directly to in-memory storage
	_, err := git.Clone(memory.NewStorage(), billyFs, cloneOptions)
	if err != nil {
		ref := "default"
		if s.Config.Ref != "" {
			ref = s.Config.Ref
		}
		logger.Error("Failed to clone git repository", zap.String("url", s.URL), zap.String("ref", ref), zap.Error(err))
		return nil, "", fmt.Errorf("failed to clone repository %s (ref: %s): %w", s.URL, ref, err)
	}

	// Create Billy adapter
	billyAdapter := filesystem.NewBillyAdapter(billyFs)

	// Return root path based on subdirectory config
	rootPath := "."
	if s.Config.SubDir != "" {
		rootPath = s.Config.SubDir
		logger.Debug("Using subdirectory", zap.String("subdir", s.Config.SubDir))
	}

	logger.Info("Successfully cloned git repository", zap.String("url", s.URL), zap.String("root_path", rootPath))
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
			// Fine-grained tokens start with "github_pat_"
			username := "token"
			if strings.HasPrefix(token, "github_pat_") {
				username = "x-access-token"
			}
			return &http.BasicAuth{
				Username: username,
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
		// Fine-grained GitHub tokens start with "github_pat_"
		username := "token"
		if strings.HasPrefix(token, "github_pat_") {
			username = "x-access-token"
		}
		return &http.BasicAuth{
			Username: username,
			Password: token,
		}
	}

	return nil
}

func (s *GitSource) Cleanup() error {
	return nil
}

// RefType represents the type of git reference
type RefType int

const (
	RefTypeBranch RefType = iota
	RefTypeTag
	RefTypeCommit
)

// detectRefType determines if the ref is a branch, tag, or commit hash
func detectRefType(ref string) RefType {
	if ref == "" {
		return RefTypeBranch // default branch
	}

	// Check if it's a commit hash (SHA-1: 40 hex chars, or SHA-256: 64 hex chars)
	commitRegex := regexp.MustCompile(`^[a-f0-9]{7,64}$`)
	if commitRegex.MatchString(ref) {
		return RefTypeCommit
	}

	// Check common tag patterns (v1.0.0, 1.0.0, release-1.0, etc.)
	tagRegex := regexp.MustCompile(`^(v?\d+\.\d+(\.\d+)?|.*-\d+\.\d+(\.\d+)?|release-.+)$`)
	if tagRegex.MatchString(ref) {
		return RefTypeTag
	}

	// Default to branch for everything else
	return RefTypeBranch
}

// getRefTypeName returns the string representation of RefType
func getRefTypeName(refType RefType) string {
	switch refType {
	case RefTypeBranch:
		return "branch"
	case RefTypeTag:
		return "tag"
	case RefTypeCommit:
		return "commit"
	default:
		return "unknown"
	}
}
