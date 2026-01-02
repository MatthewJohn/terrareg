package provider_source

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
	provider_source_model "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/repository"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/model"
	repository_repository "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/repository/repository"
)

// Repository represents a source code repository
// This is a simplified interface for GitHub provider source operations
// Python reference: repository_model.py::Repository
type Repository interface {
	Owner() string
	Name() string
}

// ProviderVersion represents a provider version
// This is a simplified interface for GitHub provider source operations
// Python reference: provider_version_model.py::ProviderVersion
type ProviderVersion interface {
	Repository() (Repository, error)
	GitTag() (string, error)
}

// GithubProviderSource implements GitHub-specific provider source functionality
// Python reference: provider_source/github.py::GithubProviderSource
type GithubProviderSource struct {
	*BaseProviderSource
	httpClient             *http.Client
	privateKeyContent      []byte
	privateKeyContentMutex  sync.RWMutex
}

// NewGithubProviderSource creates a new GitHub provider source instance
// Python reference: github.py::__init__
func NewGithubProviderSource(
	name string,
	repo repository.ProviderSourceRepository,
	sourceClass *service.GithubProviderSourceClass,
) *GithubProviderSource {
	return &GithubProviderSource{
		BaseProviderSource: NewBaseProviderSource(name, repo, sourceClass),
		httpClient:         &http.Client{Timeout: 30 * time.Second},
	}
}

// TEMPORARY CACHE FOR INSTALLATION ACCESS TOKENS
// @TODO REMOVE
var installationIdTokens = make(map[string]string)
var installationIdTokensMutex sync.RWMutex

// Config properties (from database)

// config loads the configuration from the database
// Python reference: base.py::_config property
func (g *GithubProviderSource) config(ctx context.Context) (*provider_source_model.ProviderSourceConfig, error) {
	return g.Config(ctx)
}

// clientID returns the GitHub OAuth client ID
// Python reference: github.py::_client_id property
func (g *GithubProviderSource) clientID(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.ClientID, nil
}

// clientSecret returns the GitHub OAuth client secret
// Python reference: github.py::_client_secret property
func (g *GithubProviderSource) clientSecret(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.ClientSecret, nil
}

// baseURL returns the GitHub base URL
// Python reference: github.py::_base_url property
func (g *GithubProviderSource) baseURL(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.BaseURL, nil
}

// apiURL returns the GitHub API URL
// Python reference: github.py::_api_url property
func (g *GithubProviderSource) apiURL(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.ApiURL, nil
}

// loginButtonText returns the text for the login button
// Python reference: github.py::login_button_text property
func (g *GithubProviderSource) LoginButtonText(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.LoginButtonText, nil
}

// privateKeyPath returns the path to the private key
// Python reference: github.py::_private_key_path property
func (g *GithubProviderSource) privateKeyPath(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.PrivateKeyPath, nil
}

// privateKey returns the content of the private key
// Python reference: github.py::_private_key property
func (g *GithubProviderSource) privateKey(ctx context.Context) ([]byte, error) {
	g.privateKeyContentMutex.RLock()
	if g.privateKeyContent != nil {
		g.privateKeyContentMutex.RUnlock()
		return g.privateKeyContent, nil
	}
	g.privateKeyContentMutex.RUnlock()

	// Load private key from file
	privateKeyPath, err := g.privateKeyPath(ctx)
	if err != nil {
		return nil, err
	}

	if privateKeyPath == "" {
		return nil, nil
	}

	// Check if file exists
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		return nil, nil
	}

	content, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	g.privateKeyContentMutex.Lock()
	g.privateKeyContent = content
	g.privateKeyContentMutex.Unlock()

	return content, nil
}

// githubAppID returns the GitHub App ID
// Python reference: github.py::github_app_id property
func (g *GithubProviderSource) githubAppID(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}
	return config.AppID, nil
}

// autoGenerateGithubOrganisationNamespaces returns whether to auto-generate namespaces
// Python reference: github.py::auto_generate_github_organisation_namespaces property
func (g *GithubProviderSource) AutoGenerateGithubOrganisationNamespaces(ctx context.Context) (bool, error) {
	config, err := g.config(ctx)
	if err != nil {
		return false, err
	}
	return config.AutoGenerateNamespaces, nil
}

// IsEnabled returns whether GitHub authentication is enabled
// Python reference: github.py::is_enabled property
func (g *GithubProviderSource) IsEnabled(ctx context.Context) (bool, error) {
	clientID, err := g.clientID(ctx)
	if err != nil {
		return false, err
	}
	clientSecret, err := g.clientSecret(ctx)
	if err != nil {
		return false, err
	}
	baseURL, err := g.baseURL(ctx)
	if err != nil {
		return false, err
	}
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return false, err
	}

	return clientID != "" && clientSecret != "" && baseURL != "" && apiURL != "", nil
}

// GetLoginRedirectURL generates the OAuth login redirect URL
// Python reference: github.py::get_login_redirect_url
func (g *GithubProviderSource) GetLoginRedirectURL(ctx context.Context) (string, error) {
	clientID, err := g.clientID(ctx)
	if err != nil {
		return "", err
	}
	baseURL, err := g.baseURL(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/login/oauth/authorize?client_id=%s", baseURL, clientID), nil
}

// getDefaultAccessToken returns the default access token
// Python reference: github.py::_get_default_access_token
func (g *GithubProviderSource) getDefaultAccessToken(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}

	// Prefer default_installation_id
	if config.DefaultInstallationID != "" {
		return g.GenerateAppInstallationToken(ctx, config.DefaultInstallationID)
	}

	// Fallback to default_access_token
	return config.DefaultAccessToken, nil
}

// GetUserAccessToken obtains an access token from an OAuth code
// Python reference: github.py::get_user_access_token
func (g *GithubProviderSource) GetUserAccessToken(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", nil
	}

	clientID, err := g.clientID(ctx)
	if err != nil {
		return "", err
	}
	clientSecret, err := g.clientSecret(ctx)
	if err != nil {
		return "", err
	}
	baseURL, err := g.baseURL(ctx)
	if err != nil {
		return "", err
	}

	// Build form data
	formData := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/login/oauth/access_token", baseURL), strings.NewReader(formData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		// Parse response as query string format
		body := make([]byte, resp.ContentLength)
		_, err = resp.Body.Read(body)
		if err != nil {
			return "", err
		}

		// Parse query string format: "access_token=xxx&..."
		values, err := url.ParseQuery(string(body))
		if err != nil {
			return "", err
		}

		if accessTokens := values["access_token"]; len(accessTokens) == 1 {
			return accessTokens[0], nil
		}
	}

	return "", nil
}

// GetUsername gets the username of an authenticated user
// Python reference: github.py::get_username
func (g *GithubProviderSource) GetUsername(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", nil
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/user", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		if login, ok := result["login"].(string); ok {
			return login, nil
		}
	}

	return "", nil
}

// GetUserOrganisations gets the organisations the user is a member of
// Python reference: github.py::get_user_organisations
func (g *GithubProviderSource) GetUserOrganisations(ctx context.Context, accessToken string) ([]string, error) {
	if accessToken == "" {
		return []string{}, nil
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL+"/user/memberships/orgs", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var response []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return nil, err
		}

		var organisations []string
		for _, orgMembership := range response {
			// Check if state is active and role is admin
			state, _ := orgMembership["state"].(string)
			role, _ := orgMembership["role"].(string)

			if state != "active" || role != "admin" {
				continue
			}

			// Get organization login name
			if org, ok := orgMembership["organization"].(map[string]interface{}); ok {
				if login, ok := org["login"].(string); ok && login != "" {
					organisations = append(organisations, login)
				}
			}
		}

		return organisations, nil
	}

	return []string{}, nil
}

// GetPublicSourceURL returns the public URL for a repository
// Python reference: github.py::get_public_source_url
func (g *GithubProviderSource) GetPublicSourceURL(ctx context.Context, repo Repository) (string, error) {
	baseURL, err := g.baseURL(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", baseURL, repo.Owner(), repo.Name()), nil
}

// GetPublicArtifactDownloadURL returns the public download URL for a release artifact
// Python reference: github.py::get_public_artifact_download_url
func (g *GithubProviderSource) GetPublicArtifactDownloadURL(ctx context.Context, providerVersion ProviderVersion, artifactName string) (string, error) {
	repo, err := providerVersion.Repository()
	if err != nil {
		return "", err
	}

	sourceURL, err := g.GetPublicSourceURL(ctx, repo)
	if err != nil {
		return "", err
	}

	gitTag, err := providerVersion.GitTag()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/releases/download/%s/%s", sourceURL, gitTag, artifactName), nil
}

// generateJWT generates a JWT for GitHub App authentication
// Python reference: github.py::_generate_jwt
func (g *GithubProviderSource) generateJWT(ctx context.Context) (string, error) {
	privateKey, err := g.privateKey(ctx)
	if err != nil {
		return "", err
	}
	if privateKey == nil {
		return "", nil
	}

	appIDStr, err := g.githubAppID(ctx)
	if err != nil {
		return "", err
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return "", err
	}

	// Parse PEM block
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create JWT token
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(600 * time.Second)), // 10 minutes
		Issuer:    strconv.FormatInt(appID, 10),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(key)
}

// GenerateAppInstallationToken generates an installation access token
// Python reference: github.py::generate_app_installation_token
func (g *GithubProviderSource) GenerateAppInstallationToken(ctx context.Context, installationID string) (string, error) {
	if installationID == "" {
		return "", nil
	}

	// Check cache
	installationIdTokensMutex.RLock()
	if token, ok := installationIdTokens[installationID]; ok {
		installationIdTokensMutex.RUnlock()
		return token, nil
	}
	installationIdTokensMutex.RUnlock()

	// Generate JWT
	jwt, err := g.generateJWT(ctx)
	if err != nil {
		return "", err
	}
	if jwt == "" {
		return "", nil
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return "", err
	}

	// Make request to GitHub API
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/app/installations/%s/access_tokens", apiURL, installationID), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("unable to generate installation token: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, _ := result["token"].(string)

	// Cache token
	installationIdTokensMutex.Lock()
	installationIdTokens[installationID] = token
	installationIdTokensMutex.Unlock()

	return token, nil
}

// GetAppMetadata returns GitHub App metadata
// Python reference: github.py::_get_app_metadata
func (g *GithubProviderSource) GetAppMetadata(ctx context.Context) (map[string]interface{}, error) {
	jwt, err := g.generateJWT(ctx)
	if err != nil {
		return nil, err
	}
	if jwt == "" {
		return nil, fmt.Errorf("no private key configured for GitHub App")
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/app", apiURL), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("could not obtain app metadata: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetAppInstallationURL generates the GitHub App installation URL
// Python reference: github.py::get_app_installation_url
func (g *GithubProviderSource) GetAppInstallationURL(ctx context.Context) (string, error) {
	metadata, err := g.GetAppMetadata(ctx)
	if err != nil {
		return "", err
	}

	if htmlURL, ok := metadata["html_url"].(string); ok {
		return fmt.Sprintf("%s/installations/new", htmlURL), nil
	}

	return "", fmt.Errorf("app metadata missing html_url")
}

// GetDefaultAccessToken returns the default access token for GitHub API calls
// Python reference: github.py::_get_default_access_token
func (g *GithubProviderSource) GetDefaultAccessToken(ctx context.Context) (string, error) {
	config, err := g.config(ctx)
	if err != nil {
		return "", err
	}

	// Prefer default_installation_id
	if config.DefaultInstallationID != "" {
		return g.GenerateAppInstallationToken(ctx, config.DefaultInstallationID)
	}

	// Fallback to default_access_token
	return config.DefaultAccessToken, nil
}

// IsEntityOrgOrUser determines if a GitHub entity is a user or organization
// Python reference: github.py::_is_entity_org_or_user
// Returns: "GITHUB_USER", "GITHUB_ORGANISATION", or "" if not found
func (g *GithubProviderSource) IsEntityOrgOrUser(ctx context.Context, identity string, accessToken string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("access token is required")
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s", apiURL, identity), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return "", fmt.Errorf("unable to authenticate to GitHub API")
	}
	if resp.StatusCode != 200 {
		return "", nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if type_, ok := result["type"].(string); ok {
		if type_ == "User" {
			return "GITHUB_USER", nil
		} else if type_ == "Organization" {
			return "GITHUB_ORGANISATION", nil
		}
	}

	return "", nil
}

// GetGithubAppInstallationID gets the GitHub App installation ID for a namespace
// Python reference: github.py::get_github_app_installation_id
func (g *GithubProviderSource) GetGithubAppInstallationID(ctx context.Context, namespace string) (string, error) {
	jwt, err := g.generateJWT(ctx)
	if err != nil {
		return "", err
	}
	if jwt == "" {
		return "", fmt.Errorf("no private key configured for GitHub App")
	}

	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return "", err
	}

	// Try organization first
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/orgs/%s/installation", apiURL, namespace), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check if organization exists
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		if id, ok := result["id"].(float64); ok {
			return fmt.Sprintf("%.0f", id), nil
		}
	}

	// Try user
	req, err = http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/users/%s/installation", apiURL, namespace), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwt)

	resp, err = g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}
		if id, ok := result["id"].(float64); ok {
			return fmt.Sprintf("%.0f", id), nil
		}
	}

	return "", fmt.Errorf("no installation found for namespace: %s", namespace)
}

// GetAccessTokenForProvider gets an access token for a specific provider
// Python reference: github.py::_get_access_token_for_provider
func (g *GithubProviderSource) GetAccessTokenForProvider(ctx context.Context, namespace string) (string, error) {
	// First, try to get installation ID for the namespace
	installationID, err := g.GetGithubAppInstallationID(ctx, namespace)
	if err == nil && installationID != "" {
		// Generate installation token
		return g.GenerateAppInstallationToken(ctx, installationID)
	}

	// Fallback to default access token
	return g.GetDefaultAccessToken(ctx)
}

// GetCommitHashByRelease returns the commit hash for a given tag
// Python reference: github.py::_get_commit_hash_by_release
func (g *GithubProviderSource) GetCommitHashByRelease(ctx context.Context, repo *model.Repository, tagName string, accessToken string) (string, error) {
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/git/ref/tags/%s", apiURL, repo.Owner(), repo.Name(), tagName)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", nil
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if object, ok := result["object"].(map[string]interface{}); ok {
		if sha, ok := object["sha"].(string); ok {
			return sha, nil
		}
	}

	return "", nil
}

// GetReleaseArtifactsMetadata gets the artifacts metadata for a release
// Python reference: github.py::_get_release_artifacts_metadata
func (g *GithubProviderSource) GetReleaseArtifactsMetadata(ctx context.Context, repo *model.Repository, releaseID int, accessToken string) ([]*provider_source_model.ReleaseArtifactMetadata, error) {
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/%d/assets", apiURL, repo.Owner(), repo.Name(), releaseID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return []*provider_source_model.ReleaseArtifactMetadata{}, nil
	}

	var assets []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, err
	}

	var artifacts []*provider_source_model.ReleaseArtifactMetadata
	for _, asset := range assets {
		name, _ := asset["name"].(string)
		providerIDFloat, ok := asset["id"].(float64)
		if !ok || name == "" {
			continue
		}
		providerID := int(providerIDFloat)
		artifacts = append(artifacts, provider_source_model.NewReleaseArtifactMetadata(name, providerID))
	}

	return artifacts, nil
}

// GetReleaseArtifact downloads a release artifact
// Python reference: github.py::get_release_artifact
func (g *GithubProviderSource) GetReleaseArtifact(ctx context.Context, repo *model.Repository, artifact *provider_source_model.ReleaseArtifactMetadata, accessToken string) ([]byte, error) {
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/repos/%s/%s/releases/assets/%d", apiURL, repo.Owner(), repo.Name(), artifact.ProviderID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, nil
	}

	return io.ReadAll(resp.Body)
}

// GetReleaseArchive downloads the release archive tarball
// Python reference: github.py::get_release_archive
func (g *GithubProviderSource) GetReleaseArchive(ctx context.Context, repo *model.Repository, releaseMetadata *provider_source_model.RepositoryReleaseMetadata, accessToken string) ([]byte, string, error) {
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return nil, "", err
	}

	archiveID := fmt.Sprintf("%s-%s-%s", repo.Owner(), repo.Name(), releaseMetadata.CommitHash[:7])

	url := fmt.Sprintf("%s/repos/%s/%s/tarball/%s", apiURL, repo.Owner(), repo.Name(), releaseMetadata.Tag)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, archiveID, err
	}
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, archiveID, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, archiveID, nil
	}

	content, err := io.ReadAll(resp.Body)
	return content, archiveID, err
}

// AddRepository creates a repository from GitHub API metadata
// Python reference: github.py::_add_repository
func (g *GithubProviderSource) AddRepository(ctx context.Context, repoRepo repository_repository.RepositoryRepository, repositoryMetadata map[string]interface{}) error {
	// Extract required fields
	repoIDFloat, ok := repositoryMetadata["id"].(float64)
	if !ok {
		return nil
	}
	repoID := strconv.FormatFloat(repoIDFloat, 'f', -1, 64)

	repoName, ok := repositoryMetadata["name"].(string)
	if !ok {
		return nil
	}

	ownerObj, ok := repositoryMetadata["owner"].(map[string]interface{})
	if !ok {
		return nil
	}
	ownerName, _ := ownerObj["login"].(string)
	if ownerName == "" {
		return nil
	}

	cloneURL, _ := repositoryMetadata["clone_url"].(string)

	// Extract optional fields
	var description *string
	if desc, ok := repositoryMetadata["description"].(string); ok && desc != "" {
		description = &desc
	}

	var logoURL *string
	if avatar, ok := ownerObj["avatar_url"].(string); ok && avatar != "" {
		logoURL = &avatar
	}

	// Check if repository already exists
	exists, err := repoRepo.Exists(ctx, g.Name(), repoID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// Create new repository
	newRepo, err := model.NewRepository(
		&repoID,
		ownerName,
		repoName,
		description,
		&cloneURL,
		logoURL,
		g.Name(),
	)
	if err != nil {
		return err
	}

	_, err = repoRepo.Create(ctx, newRepo)
	return err
}

// UpdateRepositories refreshes the list of repositories from GitHub
// Python reference: github.py::update_repositories
func (g *GithubProviderSource) UpdateRepositories(ctx context.Context, repoRepo repository_repository.RepositoryRepository, accessToken string) error {
	apiURL, err := g.apiURL(ctx)
	if err != nil {
		return err
	}

	page := 1
	for {
		url := fmt.Sprintf("%s/user/repos?visibility=public&affiliation=owner,organization_member&sort=created&direction=desc&per_page=100&page=%d", apiURL, page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		resp, err := g.httpClient.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			return fmt.Errorf("invalid response code from github: %d", resp.StatusCode)
		}

		var results []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		for _, repository := range results {
			g.AddRepository(ctx, repoRepo, repository)
		}

		if len(results) < 100 {
			break
		}

		page++
	}

	return nil
}
