package git

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	provider_source_service "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
)

// GitConfig holds configuration for git operations
// Python reference: Config class with UPSTREAM_GIT_CREDENTIALS_USERNAME/PASSWORD
type GitConfig struct {
	upstreamGitCredentialsUsername string
	upstreamGitCredentialsPassword string
}

// NewDefaultGitConfig creates a new GitConfig with basic credentials
// Python reference: Config.UPSTREAM_GIT_CREDENTIALS_USERNAME/PASSWORD
func NewDefaultGitConfig(username, password string) *GitConfig {
	return &GitConfig{
		upstreamGitCredentialsUsername: username,
		upstreamGitCredentialsPassword: password,
	}
}

// GetUsername returns the configured git username
func (c *GitConfig) GetUsername() string {
	return c.upstreamGitCredentialsUsername
}

// GetPassword returns the configured git password
func (c *GitConfig) GetPassword() string {
	return c.upstreamGitCredentialsPassword
}

// AuthenticatedURLBuilder handles building authenticated git URLs
// Python reference: module_extractor.py::GitModuleExtractor::_get_authenticated_git_url
type AuthenticatedURLBuilder struct {
	providerSourceFactory model.ProviderSourceFactory
	gitConfig             *GitConfig
}

// NewAuthenticatedURLBuilder creates a new authenticated URL builder
// Python reference: module_extractor.py::GitModuleExtractor::__init__
func NewAuthenticatedURLBuilder(
	factory model.ProviderSourceFactory,
	config *GitConfig,
) *AuthenticatedURLBuilder {
	return &AuthenticatedURLBuilder{
		providerSourceFactory: factory,
		gitConfig:             config,
	}
}

// GitHubAppAuthenticator defines the interface for GitHub App authentication
// This allows us to use type assertions to access GitHub-specific methods
type GitHubAppAuthenticator interface {
	GetGithubAppInstallationID(ctx context.Context, namespace string) (string, error)
	GenerateAppInstallationToken(ctx context.Context, installationID string) (string, error)
}

// GetAuthenticatedGitURL builds an authenticated git URL for cloning
// Python reference: module_extractor.py::GitModuleExtractor::_get_authenticated_git_url (lines 837-877)
func (b *AuthenticatedURLBuilder) GetAuthenticatedGitURL(
	ctx context.Context,
	moduleProvider *model.ModuleProvider,
	gitURL string,
) (string, error) {
	// Parse URL to check scheme
	parsedURL, err := url.Parse(gitURL)
	if err != nil {
		return gitURL, fmt.Errorf("invalid git URL: %w", err)
	}

	// Only modify http/https URLs
	// Python reference: module_extractor.py line 845
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return gitURL, nil // Return SSH URLs unchanged
	}

	authenticatedNetloc := ""

	// 1. Try provider source GitHub App authentication
	// Python reference: module_extractor.py lines 850-863
	effectiveProviderSource, err := moduleProvider.GetEffectiveProviderSource(ctx)
	if err == nil && effectiveProviderSource != nil {
		// Try to get installation ID and token
		token, err := b.getGitHubAppToken(ctx, moduleProvider, effectiveProviderSource)
		if err == nil && token != "" {
			authenticatedNetloc = fmt.Sprintf("x-access-token:%s", token)
		}
		// If GitHub App auth fails, fall through to basic credentials
	}

	// 2. Fallback to basic credentials
	// Python reference: module_extractor.py lines 865-870
	if authenticatedNetloc == "" {
		username := b.gitConfig.GetUsername()
		password := b.gitConfig.GetPassword()
		if username != "" || password != "" {
			if password != "" {
				authenticatedNetloc = fmt.Sprintf("%s:%s", username, password)
			} else {
				authenticatedNetloc = username
			}
		}
	}

	// Build authenticated URL
	// Python reference: module_extractor.py lines 872-875
	if authenticatedNetloc != "" {
		domainAndPort := parsedURL.Host
		if strings.Contains(domainAndPort, "@") {
			// Remove existing credentials
			parts := strings.Split(domainAndPort, "@")
			domainAndPort = parts[len(parts)-1]
		}
		// Build URL manually to avoid URL encoding of credentials
		// Git URLs should not URL-encode the username/password part
		return fmt.Sprintf("%s://%s@%s%s", parsedURL.Scheme, authenticatedNetloc, domainAndPort, parsedURL.RequestURI()), nil
	}

	return gitURL, nil
}

// getGitHubAppToken attempts to get a GitHub App installation token
// Python reference: module_extractor.py lines 856-859
func (b *AuthenticatedURLBuilder) getGitHubAppToken(
	ctx context.Context,
	moduleProvider *model.ModuleProvider,
	providerSource provider_source_service.ProviderSourceInstance,
) (string, error) {
	// Try type assertion to GitHub App authenticator
	// This allows GitHub-specific provider sources to expose their authentication methods
	githubAuth, ok := providerSource.(GitHubAppAuthenticator)
	if !ok {
		return "", fmt.Errorf("provider source does not support GitHub App authentication")
	}

	// Get namespace for installation ID lookup
	namespace := moduleProvider.Namespace()

	// Get installation ID for the namespace
	// Python reference: provider_source.get_github_app_installation_id(namespace)
	installationID, err := githubAuth.GetGithubAppInstallationID(ctx, string(namespace.Name()))
	if err != nil || installationID == "" {
		return "", err
	}

	// Generate installation token
	// Python reference: provider_source.generate_app_installation_token(installation_id)
	token, err := githubAuth.GenerateAppInstallationToken(ctx, installationID)
	if err != nil || token == "" {
		return "", err
	}

	return token, nil
}
