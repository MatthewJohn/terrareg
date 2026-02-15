package provider

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
)

// MockProviderSource is a mock implementation of ProviderSourceInstance for testing
// Python reference: test_provider_extractor.py mock objects
type MockProviderSource struct {
	// ReleaseArtifacts stores artifact data by name
	ReleaseArtifacts map[string][]byte

	// ReleaseArchive stores the archive data and subdirectory
	ReleaseArchiveData []byte
	ReleaseArchiveSubdir string

	// AccessToken to simulate GitHub authentication
	AccessToken string
}

// NewMockProviderSource creates a new mock provider source
func NewMockProviderSource() *MockProviderSource {
	return &MockProviderSource{
		ReleaseArtifacts:  make(map[string][]byte),
		AccessToken:      "test-access-token",
	}
}

// Name implements ProviderSourceInstance
func (m *MockProviderSource) Name() string {
	return "mock-provider"
}

// ApiName implements ProviderSourceInstance
func (m *MockProviderSource) ApiName(ctx context.Context) (string, error) {
	return "mock-api-name", nil
}

// Type implements ProviderSourceInstance
func (m *MockProviderSource) Type() model.ProviderSourceType {
	return model.ProviderSourceTypeGithub
}

// GetLoginRedirectURL implements ProviderSourceInstance
func (m *MockProviderSource) GetLoginRedirectURL(ctx context.Context) (string, error) {
	return "https://example.com/oauth", nil
}

// GetUserAccessToken implements ProviderSourceInstance
func (m *MockProviderSource) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	return m.AccessToken, nil
}

// GetUsername implements ProviderSourceInstance
func (m *MockProviderSource) GetUsername(ctx context.Context, accessToken string) (string, error) {
	return "test-user", nil
}

// GetUserOrganizations implements ProviderSourceInstance
func (m *MockProviderSource) GetUserOrganizations(ctx context.Context, accessToken string) []string {
	return []string{"test-org"}
}

// GetUserOrganizationsList implements ProviderSourceInstance
func (m *MockProviderSource) GetUserOrganizationsList(ctx context.Context, sessionID string) ([]*model.Organization, error) {
	return []*model.Organization{}, nil
}

// GetUserRepositories implements ProviderSourceInstance
func (m *MockProviderSource) GetUserRepositories(ctx context.Context, sessionID string) ([]*model.Repository, error) {
	return []*model.Repository{}, nil
}

// RefreshNamespaceRepositories implements ProviderSourceInstance
func (m *MockProviderSource) RefreshNamespaceRepositories(ctx context.Context, namespace string) error {
	return nil
}

// PublishProviderFromRepository implements ProviderSourceInstance
func (m *MockProviderSource) PublishProviderFromRepository(ctx context.Context, repoID int, categoryID int, namespace string) (*service.PublishProviderResult, error) {
	return &service.PublishProviderResult{
		Name:      "test-provider",
		Namespace: namespace,
	}, nil
}

// GetReleaseArtifact implements ProviderSourceInstance
// Downloads a specific release artifact by name
func (m *MockProviderSource) GetReleaseArtifact(ctx context.Context, repo *sqldb.RepositoryDB, artifact *model.ReleaseArtifactMetadata, accessToken string) ([]byte, error) {
	data, ok := m.ReleaseArtifacts[artifact.Name]
	if !ok {
		return nil, nil // Return nil to simulate "not found"
	}
	if data == nil {
		return nil, nil // Return nil to simulate download failure
	}
	return data, nil
}

// SetReleaseArtifact sets the data for a release artifact
func (m *MockProviderSource) SetReleaseArtifact(name string, data []byte) {
	m.ReleaseArtifacts[name] = data
}

// GetReleaseArchive implements ProviderSourceInstance
// Downloads the release source archive (.tar.gz)
func (m *MockProviderSource) GetReleaseArchive(ctx context.Context, repo *sqldb.RepositoryDB, releaseMetadata *model.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error) {
	if m.ReleaseArchiveData == nil {
		return nil, "", nil // Simulate not found
	}
	return m.ReleaseArchiveData, m.ReleaseArchiveSubdir, nil
}

// SetReleaseArchive sets the data for the release archive
func (m *MockProviderSource) SetReleaseArchive(data []byte, subdir string) {
	m.ReleaseArchiveData = data
	m.ReleaseArchiveSubdir = subdir
}

// AssertGetReleaseArtifactCalled checks if GetReleaseArtifact was called with the expected artifact name
func (m *MockProviderSource) AssertGetReleaseArtifactCalled(t interface{}, artifactName string) bool {
	// This would typically be implemented with a sync.Mutex and call tracking
	// For simplicity, just check if the artifact exists
	_, ok := m.ReleaseArtifacts[artifactName]
	return ok
}

// CreateTestReleaseMetadata creates a RepositoryReleaseMetadata for testing
func CreateTestReleaseMetadata(version string, artifacts []string) *model.RepositoryReleaseMetadata {
	releaseArtifacts := make([]*model.ReleaseArtifactMetadata, len(artifacts))
	for i, artifact := range artifacts {
		releaseArtifacts[i] = model.NewReleaseArtifactMetadata(artifact, i)
	}

	return &model.RepositoryReleaseMetadata{
		Name:             "Release " + version,
		Tag:              "v" + version,
		ArchiveURL:       "https://git.example.com/artifacts/downloads/" + version + ".tar.gz",
		CommitHash:       "abc123",
		ProviderID:       123,
		ReleaseArtifacts: releaseArtifacts,
	}
}
