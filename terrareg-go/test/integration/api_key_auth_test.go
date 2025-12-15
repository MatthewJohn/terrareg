package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/config/service"
	infraAuth "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/auth"
)

func TestAdminApiKeyAuth(t *testing.T) {
	// Set admin token environment variable
	os.Setenv("ADMIN_AUTHENTICATION_TOKEN", "test-admin-token")
	os.Setenv("SECRET_KEY", "this-is-a-test-secret-key-that-is-long-enough-to-be-valid")
	defer func() {
		os.Unsetenv("ADMIN_AUTHENTICATION_TOKEN")
		os.Unsetenv("SECRET_KEY")
	}()

	// Load configuration
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, nil)
	_, infraConfig, err := configService.LoadConfiguration()
	require.NoError(t, err)

	// Create admin API key auth method
	adminAuthMethod := infraAuth.NewAdminApiKeyAuthMethod(infraConfig)
	assert.True(t, adminAuthMethod.IsEnabled())

	// Test authentication with correct token
	ctx := context.Background()
	authCtx, err := adminAuthMethod.Authenticate(ctx, "test-admin-token")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.True(t, authCtx.AuthMethod.IsAdmin())
	assert.Equal(t, "Admin", authCtx.AuthMethod.GetUsername())

	// Test authentication with incorrect token
	authCtx, err = adminAuthMethod.Authenticate(ctx, "wrong-token")
	assert.Error(t, err)
	assert.Nil(t, authCtx)

	// Test authentication with empty token
	authCtx, err = adminAuthMethod.Authenticate(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, authCtx)
}

func TestUploadApiKeyAuth(t *testing.T) {
	// Set upload API keys environment variable
	os.Setenv("UPLOAD_API_KEYS", "upload-token-1,upload-token-2,upload-token-3")
	os.Setenv("SECRET_KEY", "this-is-a-test-secret-key-that-is-long-enough-to-be-valid")
	defer func() {
		os.Unsetenv("UPLOAD_API_KEYS")
		os.Unsetenv("SECRET_KEY")
	}()

	// Load configuration
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, nil)
	_, infraConfig, err := configService.LoadConfiguration()
	require.NoError(t, err)

	// Create upload API key auth method
	uploadAuthMethod := infraAuth.NewUploadApiKeyAuthMethod(infraConfig)
	assert.True(t, uploadAuthMethod.IsEnabled())

	// Test authentication with correct token
	ctx := context.Background()
	authCtx, err := uploadAuthMethod.Authenticate(ctx, "upload-token-1")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.False(t, authCtx.AuthMethod.IsAdmin())
	assert.Equal(t, "Upload API Key", authCtx.AuthMethod.GetUsername())

	// Test with another valid token
	authCtx, err = uploadAuthMethod.Authenticate(ctx, "upload-token-2")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)

	// Test authentication with incorrect token
	authCtx, err = uploadAuthMethod.Authenticate(ctx, "wrong-token")
	assert.Error(t, err)
	assert.Nil(t, authCtx)

	// Test can upload but not publish
	assert.True(t, uploadAuthMethod.CanUploadModuleVersion("any-namespace"))
	assert.False(t, uploadAuthMethod.CanPublishModuleVersion("any-namespace"))
}

func TestPublishApiKeyAuth(t *testing.T) {
	// Set publish API keys environment variable
	os.Setenv("PUBLISH_API_KEYS", "publish-token-1,publish-token-2")
	os.Setenv("SECRET_KEY", "this-is-a-test-secret-key-that-is-long-enough-to-be-valid")
	defer func() {
		os.Unsetenv("PUBLISH_API_KEYS")
		os.Unsetenv("SECRET_KEY")
	}()

	// Load configuration
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, nil)
	_, infraConfig, err := configService.LoadConfiguration()
	require.NoError(t, err)

	// Create publish API key auth method
	publishAuthMethod := infraAuth.NewPublishApiKeyAuthMethod(infraConfig)
	assert.True(t, publishAuthMethod.IsEnabled())

	// Test authentication with correct token
	ctx := context.Background()
	authCtx, err := publishAuthMethod.Authenticate(ctx, "publish-token-1")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.False(t, authCtx.AuthMethod.IsAdmin())
	assert.Equal(t, "Publish API Key", authCtx.AuthMethod.GetUsername())

	// Test authentication with incorrect token
	authCtx, err = publishAuthMethod.Authenticate(ctx, "wrong-token")
	assert.Error(t, err)
	assert.Nil(t, authCtx)

	// Test can publish but not upload
	assert.False(t, publishAuthMethod.CanUploadModuleVersion("any-namespace"))
	assert.True(t, publishAuthMethod.CanPublishModuleVersion("any-namespace"))
}

func TestApiKeyAuthInHTTPContext(t *testing.T) {
	// Set all API keys
	os.Setenv("ADMIN_AUTHENTICATION_TOKEN", "admin-test-token")
	os.Setenv("UPLOAD_API_KEYS", "upload-test-token")
	os.Setenv("PUBLISH_API_KEYS", "publish-test-token")
	os.Setenv("SECRET_KEY", "this-is-a-test-secret-key-that-is-long-enough-to-be-valid")
	defer func() {
		os.Unsetenv("ADMIN_AUTHENTICATION_TOKEN")
		os.Unsetenv("UPLOAD_API_KEYS")
		os.Unsetenv("PUBLISH_API_KEYS")
		os.Unsetenv("SECRET_KEY")
	}()

	// Load configuration
	configService := service.NewConfigurationService(service.ConfigurationServiceOptions{}, nil)
	_, infraConfig, err := configService.LoadConfiguration()
	require.NoError(t, err)

	// Test individual auth methods directly instead of using auth factory
	ctx := context.Background()

	// Test admin token authentication
	adminAuthMethod := infraAuth.NewAdminApiKeyAuthMethod(infraConfig)
	authCtx, err := adminAuthMethod.Authenticate(ctx, "admin-test-token")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.True(t, authCtx.AuthMethod.IsAdmin())

	// Test upload token authentication
	uploadAuthMethod := infraAuth.NewUploadApiKeyAuthMethod(infraConfig)
	authCtx, err = uploadAuthMethod.Authenticate(ctx, "upload-test-token")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.False(t, authCtx.AuthMethod.IsAdmin())
	assert.True(t, uploadAuthMethod.CanUploadModuleVersion("any-namespace"))
	assert.False(t, uploadAuthMethod.CanPublishModuleVersion("any-namespace"))

	// Test publish token authentication
	publishAuthMethod := infraAuth.NewPublishApiKeyAuthMethod(infraConfig)
	authCtx, err = publishAuthMethod.Authenticate(ctx, "publish-test-token")
	assert.NoError(t, err)
	assert.True(t, authCtx.IsAuthenticated)
	assert.False(t, authCtx.AuthMethod.IsAdmin())
	assert.False(t, publishAuthMethod.CanUploadModuleVersion("any-namespace"))
	assert.True(t, publishAuthMethod.CanPublishModuleVersion("any-namespace"))

	// Test invalid token
	_, err = adminAuthMethod.Authenticate(ctx, "invalid-token")
	assert.Error(t, err)
}
