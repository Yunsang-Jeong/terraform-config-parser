package version

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Version information - can be overridden at build time with -ldflags
var (
	Version = "" // Will be set by build or read from file
)

// init runs at package initialization and sets the version if not provided at build time
func init() {
	if Version == "" {
		Version = getVersionFromFile()
	}
}

// getVersionFromFile reads version from .version file
func getVersionFromFile() string {
	// Try to read version from .version file in current directory
	if content, err := os.ReadFile(".version"); err == nil {
		version := strings.TrimSpace(string(content))
		if version != "" {
			return version
		}
	}
	
	// Try embedded version file (for Go 1.16+ if we add it later)  
	// This would be: //go:embed .version
	// var embeddedVersion string
	
	// Fallback to dev if file doesn't exist or is empty
	return "dev"
}

// GetVersion returns the full version string
func GetVersion() string {
	return Version
}

// GetFullVersion returns detailed version information  
func GetFullVersion() string {
	return fmt.Sprintf("%s (go: %s)", Version, runtime.Version())
}
