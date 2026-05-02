package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// VersionReader handles version detection from multiple sources
type VersionReader struct {
	versionFile string
}

// NewVersionReader creates a new VersionReader
func NewVersionReader() *VersionReader {
	return &VersionReader{
		versionFile: "version.txt",
	}
}

// ReadVersion attempts to read the version from multiple sources
// Priority: 1) version.txt file, 2) build-time injection, 3) default
func (vr *VersionReader) ReadVersion() (string, error) {
	// Try to read from version.txt file first
	if version, err := vr.readFromFile(); err == nil && version != "" {
		return version, nil
	}

	// Try to read from build-time info (Go module version)
	if version, err := vr.readFromBuildInfo(); err == nil && version != "" {
		return version, nil
	}

	// Return default version
	return "unknown", nil
}

// readFromFile attempts to read version from version.txt file
func (vr *VersionReader) readFromFile() (string, error) {
	// Look for version.txt in current directory and parent directories
	paths := []string{
		vr.versionFile,
		filepath.Join("..", vr.versionFile),
		filepath.Join("..", "..", vr.versionFile),
		filepath.Join("..", "..", "..", vr.versionFile),
	}

	for _, path := range paths {
		if data, err := os.ReadFile(path); err == nil {
			version := strings.TrimSpace(string(data))
			if version != "" {
				return version, nil
			}
		}
	}

	return "", fmt.Errorf("version.txt not found or empty")
}

// readFromBuildInfo attempts to read version from Go build info
func (vr *VersionReader) readFromBuildInfo() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("no build info available")
	}

	// Try to get version from module path
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version, nil
	}

	// Try to find version from build settings
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" && setting.Value != "" {
			// Use git commit hash as version
			return setting.Value[:min(8, len(setting.Value))], nil
		}
	}

	return "", fmt.Errorf("no version found in build info")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
