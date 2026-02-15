package testutils

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider"
	providermodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// CreateProviderVersionWithExtractor creates a provider version with extraction wrapper
// This mimics the Python test fixture setup for provider extraction tests
func CreateProviderVersionWithExtractor(
	t *testing.T,
	db *sqldb.Database,
	namespaceID int,
	providerName string,
	version string,
) (*provider.Provider, *provider.ProviderVersion, *providermodel.RepositoryReleaseMetadata, func()) {
	// Create provider
	providerDB := &sqldb.ProviderDB{
		NamespaceID: namespaceID,
		Name:        providerName,
	}
	err := db.DB.Create(providerDB).Error
	require.NoError(t, err)

	prov := provider.ReconstructProvider(
		int(providerDB.ID),
		providerDB.NamespaceID,
		providerDB.Name,
		providerDB.Description,
		string(providerDB.Tier),
		(*int)(providerDB.ProviderCategoryID),
		(*int)(providerDB.RepositoryID),
		(*int)(providerDB.LatestVersionID),
		providerDB.DefaultProviderSourceAuth,
	)

	// Create provider version
	versionDB := &sqldb.ProviderVersionDB{
		ProviderID: prov.ID(),
		Version:    version,
		Beta:       false,
	}
	err = db.DB.Create(versionDB).Error
	require.NoError(t, err)

	provVersion := provider.ReconstructProviderVersion(
		int(versionDB.ID),
		versionDB.ProviderID,
		versionDB.Version,
		versionDB.GitTag,
		versionDB.Beta,
		nil, // PublishedAt - will be set by repository
		0,   // GPGKeyID
		nil, // ProtocolVersions
	)

	// Create release metadata
	releaseArtifacts := []*providermodel.ReleaseArtifactMetadata{
		providermodel.NewReleaseArtifactMetadata(fmt.Sprintf("terraform-provider-%s_%s_windows_arm64.zip", providerName, version), 1),
		providermodel.NewReleaseArtifactMetadata(fmt.Sprintf("terraform-provider-%s_%s_manifest.json", providerName, version), 2),
		providermodel.NewReleaseArtifactMetadata(fmt.Sprintf("terraform-provider-%s_%s_linux_amd64.zip", providerName, version), 3),
	}
	releaseMetadata := &providermodel.RepositoryReleaseMetadata{
		Name:             fmt.Sprintf("Release %s", version),
		Tag:              fmt.Sprintf("v%s", version),
		ArchiveURL:       fmt.Sprintf("https://git.example.com/artifacts/downloads/%s.tar.gz", version),
		CommitHash:       "abcdefg123455",
		ProviderID:       123,
		ReleaseArtifacts: releaseArtifacts,
	}

	// Return cleanup function
	cleanup := func() {
		db.DB.Where("provider_version_id = ?", versionDB.ID).Delete(&sqldb.ProviderVersionDocumentationDB{})
		db.DB.Where("provider_version_id = ?", versionDB.ID).Delete(&sqldb.ProviderVersionBinaryDB{})
		db.DB.Delete(&sqldb.ProviderVersionDB{}, versionDB.ID)
		db.DB.Delete(&sqldb.ProviderDB{}, prov.ID())
	}

	return prov, provVersion, releaseMetadata, cleanup
}

// CreateTestGPGKey creates a test GPG key in the database
// Python reference: test_data.py::get_test_gpg_key_data
func CreateTestGPGKey(t *testing.T, db *sqldb.Database, namespaceID int, fingerprint string) *provider.GPGKey {
	// This is a test GPG key for unit testing
	// In production, real GPG keys would be imported
	asciiArmor := `-----BEGIN PGP PUBLIC KEY BLOCK-----

mQINBGYrFgUBEACttrW6z7QhlYsn5HGT45EaKRcVsWzJ5Gj2yRnPbwFiWpLNdCfj
4FqYlH8I8E7vVZ0Q2vF3pN2eJ5yD0Y4yF2xN3eL7wK9yL3hQ5sR8wY1zB2xC5dV
6yF8hL2pN3eR7yK1zA5xC8dV2yF9hM2zC6eL3wP9sR1yB2zC5dV6yF8hL2pN3eR
=RIEB
-----END PGP PUBLIC KEY BLOCK-----`

	gpgKeyDB := &sqldb.GPGKeyDB{
		NamespaceID: namespaceID,
		ASCIIArmor:  []byte(asciiArmor),
		KeyID:       &fingerprint,
		Fingerprint: &fingerprint,
	}

	err := db.DB.Create(gpgKeyDB).Error
	require.NoError(t, err)

	gpgKey := provider.ReconstructGPGKey(
		int(gpgKeyDB.ID),
		string(gpgKeyDB.ASCIIArmor),
		string(gpgKeyDB.ASCIIArmor),
		fingerprint,
		nil, // TrustSignature
		*gpgKeyDB.CreatedAt,
		*gpgKeyDB.UpdatedAt,
	)

	return gpgKey
}

// CreateMockReleaseMetadata creates mock release metadata for testing
func CreateMockReleaseMetadata(version string, artifacts []string) *providermodel.RepositoryReleaseMetadata {
	providerName := "multiple-versions" // Default from Python tests

	releaseArtifacts := make([]*providermodel.ReleaseArtifactMetadata, len(artifacts))
	for i, artifact := range artifacts {
		releaseArtifacts[i] = providermodel.NewReleaseArtifactMetadata(
			fmt.Sprintf("terraform-provider-%s_%s", providerName, artifact),
			i,
		)
	}

	return &providermodel.RepositoryReleaseMetadata{
		Name:             fmt.Sprintf("Release %s", version),
		Tag:              fmt.Sprintf("v%s", version),
		ArchiveURL:       fmt.Sprintf("https://git.example.com/artifacts/downloads/%s.tar.gz", version),
		CommitHash:       "abcdefg123455",
		ProviderID:       123,
		ReleaseArtifacts: releaseArtifacts,
	}
}

// CreateTestTarArchive creates a test tar.gz archive for testing
// Python reference: test_provider_extractor.py::_obtain_source_code test setup
func CreateTestTarArchive(t *testing.T, files map[string]string, stripPrefix string) []byte {
	tempDir := t.TempDir()

	// Create all files
	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		err := os.MkdirAll(dir, 0755)
		require.NoError(t, err)

		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create tar.gz archive
	archivePath := filepath.Join(tempDir, "archive.tar.gz")
	err := createTarGz(tempDir, archivePath, stripPrefix)
	require.NoError(t, err)

	// Read and return
	data, err := os.ReadFile(archivePath)
	require.NoError(t, err)

	return data
}

// createTarGz creates a tar.gz archive from a directory
func createTarGz(srcDir, destFile, stripPrefix string) error {
	// Create destination file
	outFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Create gzip writer
	gzWriter := gzip.NewWriter(outFile)
	defer gzWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Walk through source directory
	return filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the destination file itself
		if filePath == destFile {
			return nil
		}

		// Create relative path
		relPath, err := filepath.Rel(srcDir, filePath)
		if err != nil {
			return err
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Strip prefix if specified
		if stripPrefix != "" {
			// If stripPrefix contains a path separator, we need to strip the first component
			parts := filepath.SplitList(relPath)
			if len(parts) > 1 && parts[0] == stripPrefix {
				relPath = filepath.Join(parts[1:]...)
			}
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.IsDir() {
			data, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}
			if _, err := tarWriter.Write(data); err != nil {
				return err
			}
		}

		return nil
	})
}

// GenerateTestChecksums generates SHA256 checksums for test data
func GenerateTestChecksums(t *testing.T, data map[string]string) map[string]string {
	checksums := make(map[string]string)
	for filename, content := range data {
		hash := sha256.Sum256([]byte(content))
		checksums[filename] = hex.EncodeToString(hash[:])
	}
	return checksums
}

// GenerateTestSignature generates a test GPG signature
// Python reference: test_provider_extractor.py::test_obtain_gpg_key
// Note: In real tests, we use pre-generated signatures. For now, return placeholder.
func GenerateTestSignature(t *testing.T, data []byte, keyID string) []byte {
	// This is a placeholder - in real tests we would use actual GPG signatures
	// For unit tests, we mock the verification process
	signatures := map[string]string{
		"A0FC4319ABAF9C28A16821DF4F3072E58D16FF6D": `iLMEAAEKAB0WIQSg/EMZ+6+cKKFoId9PMHLljRb/bQUCZVsR+gAKCRBPMHLljRb/bXmpA/9Ycl/a9ZKFevCamJLjMxw2K7OV12hWdR5X5pZ/Rse1gAOYQNaSbKwchM0ChDh/nrFMYzvErHsw/he8OjOKG3KtIxGITPvTgjL7Zj0OxJSQAAgQN/bmDNM/jxhYevNsJjqnHeSBHm7U6IsLHFKNiSDj1c2yom4pUnkCiCt3juqNNA==`,
	}

	sig, ok := signatures[keyID]
	require.True(t, ok, "No test signature for key: %s", keyID)

	decoded, err := hex.DecodeString(sig)
	require.NoError(t, err)
	return decoded
}

// CopyProviderFixtureFiles copies provider fixture files to a test directory
func CopyProviderFixtureFiles(t *testing.T, destDir string, fixtureName string) {
	fixtureDir := filepath.Join("test/integration/testutils/fixtures/providers", fixtureName)

	err := filepath.Walk(fixtureDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			relPath, err := filepath.Rel(fixtureDir, path)
			if err != nil {
				return err
			}
			if relPath == "." {
				return nil
			}
			destPath := filepath.Join(destDir, relPath)
			return os.MkdirAll(destPath, 0755)
		}

		relPath, err := filepath.Rel(fixtureDir, path)
		if err != nil {
			return err
		}
		destPath := filepath.Join(destDir, relPath)

		// Create parent directory if needed
		destDirPath := filepath.Dir(destPath)
		if err := os.MkdirAll(destDirPath, 0755); err != nil {
			return err
		}

		// Copy file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, content, info.Mode())
	})

	require.NoError(t, err)
}
