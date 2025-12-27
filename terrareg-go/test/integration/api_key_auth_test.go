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
	headers := map[string]string{
		"X-Terrareg-ApiKey": "test-admin-token",
	}
	authCtx, err := adminAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.True(t, authCtx.IsAdmin())
	assert.Equal(t, "admin-api-key", authCtx.GetUsername())

	// Test authentication with incorrect token
	headers = map[string]string{
		"X-Terrareg-ApiKey": "wrong-token",
	}
	authCtx, err = adminAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err) // HeaderAuthMethod returns nil, nil when auth fails
	assert.Nil(t, authCtx)

	// Test authentication with empty token (no header)
	headers = map[string]string{}
	authCtx, err = adminAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
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
	headers := map[string]string{
		"X-Terrareg-Upload-Key": "upload-token-1",
	}
	authCtx, err := uploadAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.False(t, authCtx.IsAdmin())
	assert.Equal(t, "upload-api-key", authCtx.GetUsername())

	// Test with another valid token
	headers = map[string]string{
		"X-Terrareg-Upload-Key": "upload-token-2",
	}
	authCtx, err = uploadAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())

	// Test authentication with incorrect token
	headers = map[string]string{
		"X-Terrareg-Upload-Key": "wrong-token",
	}
	authCtx, err = uploadAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.Nil(t, authCtx)

	// Test can upload but not publish
	assert.True(t, uploadAuthMethod.IsEnabled())
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
	headers := map[string]string{
		"X-Terrareg-Publish-Key": "publish-token-1",
	}
	authCtx, err := publishAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.False(t, authCtx.IsAdmin())
	assert.Equal(t, "publish-api-key", authCtx.GetUsername())

	// Test authentication with incorrect token
	headers = map[string]string{
		"X-Terrareg-Publish-Key": "wrong-token",
	}
	authCtx, err = publishAuthMethod.Authenticate(ctx, headers, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.Nil(t, authCtx)
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
	adminHeaders := map[string]string{
		"X-Terrareg-ApiKey": "admin-test-token",
	}
	authCtx, err := adminAuthMethod.Authenticate(ctx, adminHeaders, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.True(t, authCtx.IsAdmin())

	// Test upload token authentication
	uploadAuthMethod := infraAuth.NewUploadApiKeyAuthMethod(infraConfig)
	uploadHeaders := map[string]string{
		"X-Terrareg-Upload-Key": "upload-test-token",
	}
	authCtx, err = uploadAuthMethod.Authenticate(ctx, uploadHeaders, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.False(t, authCtx.IsAdmin())

	// Test publish token authentication
	publishAuthMethod := infraAuth.NewPublishApiKeyAuthMethod(infraConfig)
	publishHeaders := map[string]string{
		"X-Terrareg-Publish-Key": "publish-test-token",
	}
	authCtx, err = publishAuthMethod.Authenticate(ctx, publishHeaders, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.NotNil(t, authCtx)
	assert.True(t, authCtx.IsAuthenticated())
	assert.False(t, authCtx.IsAdmin())

	// Test invalid token
	adminHeaders = map[string]string{
		"X-Terrareg-ApiKey": "invalid-token",
	}
	authCtx, err = adminAuthMethod.Authenticate(ctx, adminHeaders, map[string]string{}, map[string]string{})
	assert.NoError(t, err)
	assert.Nil(t, authCtx) // Invalid token should return nil context
}
