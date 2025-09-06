package source

import "github.com/Yunsang-Jeong/terraform-config-parser/pkg/filesystem"

// Source represents different sources of Terraform configurations
type Source interface {
	// Fetch retrieves the Terraform files and returns a filesystem reader
	Fetch() (filesystem.FileReader, string, error) // fs, rootPath, error
	// Cleanup removes any temporary resources
	Cleanup() error
}

// SourceConfig holds common configuration for all sources
type SourceConfig struct {
	// Ref specifies the git reference to use (branch, tag, or commit hash)
	Ref string
	// Subdirectory within the source
	SubDir string
}
