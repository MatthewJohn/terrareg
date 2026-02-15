package service

import (
	"encoding/hex"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
)

// TestExtractPlatformFromFilename tests platform extraction from filenames
// Python reference: test_provider_extractor.py::TestProviderExtractor::test__get_os_and_arch
func TestExtractPlatformFromFilename(t *testing.T) {
	service := NewProviderBinaryProcessingService(nil)

	tests := []struct {
		name           string
		filename       string
		expectedOS     string
		expectedArch   string
		shouldFail     bool
	}{
		{
			name:         "Linux amd64",
			filename:     "terraform-provider-test_1.0.0_linux_amd64.zip",
			expectedOS:   "linux",
			expectedArch: "amd64",
			shouldFail:   false,
		},
		{
			name:         "Windows arm64",
			filename:     "terraform-provider-test_1.0.0_windows_arm64.zip",
			expectedOS:   "windows",
			expectedArch: "arm64",
			shouldFail:   false,
		},
		{
			name:         "Darwin amd64",
			filename:     "terraform-provider-test_1.0.0_darwin_amd64.zip",
			expectedOS:   "darwin",
			expectedArch: "amd64",
			shouldFail:   false,
		},
		{
			name:         "Linux armv7l",
			filename:     "terraform-provider-test_1.0.0_linux_armv7l.zip",
			expectedOS:   "linux",
			expectedArch: "arm",
			shouldFail:   false,
		},
		{
			name:         "FreeBSD amd64",
			filename:     "terraform-provider-test_1.0.0_freebsd_amd64.zip",
			expectedOS:   "freebsd",
			expectedArch: "amd64",
			shouldFail:   false,
		},
		{
			name:         "Invalid filename",
			filename:     "terraform-provider-test_1.0.0.zip",
			expectedOS:   "",
			expectedArch: "",
			shouldFail:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, err := service.extractPlatformFromFilename(tt.filename)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if platform.OS != tt.expectedOS {
				t.Errorf("Expected OS %s, got %s", tt.expectedOS, platform.OS)
			}

			if platform.Architecture != tt.expectedArch {
				t.Errorf("Expected Architecture %s, got %s", tt.expectedArch, platform.Architecture)
			}
		})
	}
}

// TestIsMetadataFile tests metadata file detection
func TestIsMetadataFile(t *testing.T) {
	service := NewProviderBinaryProcessingService(nil)

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "SHA256SUMS file",
			filename: "terraform-provider-test_1.0.0_SHA256SUMS",
			expected: true,
		},
		{
			name:     "SHA256SUMS.sig file",
			filename: "terraform-provider-test_1.0.0_SHA256SUMS.sig",
			expected: true,
		},
		{
			name:     "SHA256SUMS.asc file",
			filename: "terraform-provider-test_1.0.0_SHA256SUMS.asc",
			expected: true,
		},
		{
			name:     "manifest.json file",
			filename: "terraform-provider-test_1.0.0_manifest.json",
			expected: true,
		},
		{
			name:     "Binary file",
			filename: "terraform-provider-test_1.0.0_linux_amd64.zip",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isMetadataFile(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestParseBinaryPlatformString tests platform string parsing
func TestParseBinaryPlatformString(t *testing.T) {
	service := NewProviderBinaryProcessingService(nil)

	tests := []struct {
		name        string
		platform    string
		expectedOS  string
		expectedArch string
		shouldFail  bool
	}{
		{
			name:         "Valid platform",
			platform:     "linux_amd64",
			expectedOS:   "linux",
			expectedArch: "amd64",
			shouldFail:   false,
		},
		{
			name:         "Windows arm64",
			platform:     "windows_arm64",
			expectedOS:   "windows",
			expectedArch: "arm64",
			shouldFail:   false,
		},
		{
			name:       "Invalid format",
			platform:   "linux",
			shouldFail: true,
		},
		{
			name:       "Too many parts",
			platform:   "linux_amd64_extra",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os, arch, err := service.ParseBinaryPlatformString(tt.platform)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if os != tt.expectedOS {
				t.Errorf("Expected OS %s, got %s", tt.expectedOS, os)
			}

			if arch != tt.expectedArch {
				t.Errorf("Expected Architecture %s, got %s", tt.expectedArch, arch)
			}
		})
	}
}

// TestGetBinaryPlatformString tests platform string generation
func TestGetBinaryPlatformString(t *testing.T) {
	service := NewProviderBinaryProcessingService(nil)

	tests := []struct {
		os      string
		arch    string
		expected string
	}{
		{
			os:       "linux",
			arch:     "amd64",
			expected: "linux_amd64",
		},
		{
			os:       "windows",
			arch:     "arm64",
			expected: "windows_arm64",
		},
		{
			os:       "darwin",
			arch:     "amd64",
			expected: "darwin_amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.os+"_"+tt.arch, func(t *testing.T) {
			result := service.GetBinaryPlatformString(tt.os, tt.arch)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestChecksumFileValidation tests checksum file parsing and validation
// Python reference: test_provider_extractor.py::TestProviderExtractor::test__process_release_file_invalid_checksum
func TestChecksumFileValidation(t *testing.T) {
	gpgService := NewProviderExtractionGPGService()

	validChecksums := `4e13e517a3b6f474b734559c96f4fc01678ea5299b5c61844a2747727a52e80f  ./terraform-provider-test_linux_amd64.zip
aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  ./terraform-provider-test_windows_arm64.zip
`

	invalidChecksums := `invalid checksum line`

	tests := []struct {
		name       string
		content    string
		shouldFail bool
	}{
		{
			name:       "Valid checksums",
			content:    validChecksums,
			shouldFail: false,
		},
		{
			name:       "Invalid checksums",
			content:    invalidChecksums,
			shouldFail: true,
		},
		{
			name:       "Empty content",
			content:    "",
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksums, err := gpgService.ParseChecksumFile([]byte(tt.content))

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(checksums) == 0 {
				t.Errorf("Expected non-empty checksums map")
			}

			// Verify specific checksum
			expectedChecksum := "4e13e517a3b6f474b734559c96f4fc01678ea5299b5c61844a2747727a52e80f"
			actualChecksum, ok := checksums["./terraform-provider-test_linux_amd64.zip"]
			if !ok {
				t.Errorf("Expected checksum for linux_amd64 binary not found")
			}
			if actualChecksum != expectedChecksum {
				t.Errorf("Expected checksum %s, got %s", expectedChecksum, actualChecksum)
			}
		})
	}
}

// TestBinaryChecksumVerification tests binary checksum verification
func TestBinaryChecksumVerification(t *testing.T) {
	gpgService := NewProviderExtractionGPGService()

	// Create test data and checksums
	testData := []byte("test binary content")
	hasher := provider.NewSHA256Hash()
	hasher.Write(testData)
	checksumBytes := hasher.Sum(nil)
	expectedChecksum := hex.EncodeToString(checksumBytes)

	checksums := map[string]string{
		"terraform-provider-test_linux_amd64.zip": expectedChecksum,
	}

	tests := []struct {
		name       string
		filename   string
		content    []byte
		shouldFail bool
	}{
		{
			name:       "Valid checksum",
			filename:   "terraform-provider-test_linux_amd64.zip",
			content:    testData,
			shouldFail: false,
		},
		{
			name:       "Invalid checksum",
			filename:   "terraform-provider-test_linux_amd64.zip",
			content:    []byte("different content"),
			shouldFail: true,
		},
		{
			name:       "Missing checksum",
			filename:   "terraform-provider-test_windows_arm64.zip",
			content:    testData,
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gpgService.VerifyBinaryChecksum(tt.filename, tt.content, checksums)

			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
