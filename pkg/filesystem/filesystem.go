package filesystem

import "os"

// FileReader defines the interface for reading files and directories
type FileReader interface {
	// DirExists checks if a directory exists
	DirExists(dirname string) (bool, error)
	
	// ReadDir reads the directory and returns file info
	ReadDir(dirname string) ([]os.FileInfo, error)
	
	// ReadFile reads the entire file content
	ReadFile(filename string) ([]byte, error)
}