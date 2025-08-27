package source

import (
	"os"
	"path/filepath"
	"terraform-config-parser/pkg/filesystem"

	"github.com/spf13/afero"
)

// LocalSource represents a local filesystem source
type LocalSource struct {
	Path string
	Config SourceConfig
}

func NewLocalSource(path string, config SourceConfig) *LocalSource {
	return &LocalSource{
		Path:   path,
		Config: config,
	}
}

func (s *LocalSource) Fetch() (filesystem.FileReader, string, error) {
	rootPath := s.Path
	if s.Config.SubDir != "" {
		rootPath = filepath.Join(s.Path, s.Config.SubDir)
	}

	// Check if path exists
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		return nil, "", err
	}

	// Create Afero adapter for OS filesystem
	aferoAdapter := filesystem.NewAferoAdapter(afero.NewOsFs())
	return aferoAdapter, rootPath, nil
}

func (s *LocalSource) Cleanup() error {
	// No cleanup needed for local files
	return nil
}