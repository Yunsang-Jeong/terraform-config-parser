package source

import (
	"fmt"
	"terraform-config-parser/pkg/filesystem"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

func (s *GitSource) Cleanup() error {
	return nil
}
