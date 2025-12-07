package service

import (
	"context"
	"testing"
	"time"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/auth/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTerraformIdpAuthorizationCodeRepository is a mock for the authorization code repository
type MockTerraformIdpAuthorizationCodeRepository struct {
	mock.Mock
}

func (m *MockTerraformIdpAuthorizationCodeRepository) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	args := m.Called(ctx, key, data, expiry)
	return args.Error(0)
}

func (m *MockTerraformIdpAuthorizationCodeRepository) FindByKey(ctx context.Context, key string) (*repository.AuthCodeRecord, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*repository.AuthCodeRecord), args.Error(1)
}

func (m *MockTerraformIdpAuthorizationCodeRepository) DeleteByKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockTerraformIdpAuthorizationCodeRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTerraformIdpAuthorizationCodeRepository) WithTransaction(tx interface{}) repository.TerraformIdpAuthorizationCodeRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.TerraformIdpAuthorizationCodeRepository)
}

// MockTerraformIdpAccessTokenRepository is a mock for the access token repository
type MockTerraformIdpAccessTokenRepository struct {
	mock.Mock
}

func (m *MockTerraformIdpAccessTokenRepository) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	args := m.Called(ctx, key, data, expiry)
	return args.Error(0)
}

func (m *MockTerraformIdpAccessTokenRepository) FindByKey(ctx context.Context, key string) (*repository.AccessTokenRecord, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*repository.AccessTokenRecord), args.Error(1)
}

func (m *MockTerraformIdpAccessTokenRepository) DeleteByKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockTerraformIdpAccessTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTerraformIdpAccessTokenRepository) WithTransaction(tx interface{}) repository.TerraformIdpAccessTokenRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.TerraformIdpAccessTokenRepository)
}

// MockTerraformIdpSubjectIdentifierRepository is a mock for the subject identifier repository
type MockTerraformIdpSubjectIdentifierRepository struct {
	mock.Mock
}

func (m *MockTerraformIdpSubjectIdentifierRepository) Create(ctx context.Context, key string, data []byte, expiry time.Time) error {
	args := m.Called(ctx, key, data, expiry)
	return args.Error(0)
}

func (m *MockTerraformIdpSubjectIdentifierRepository) FindByKey(ctx context.Context, key string) (*repository.SubjectIdentifierRecord, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*repository.SubjectIdentifierRecord), args.Error(1)
}

func (m *MockTerraformIdpSubjectIdentifierRepository) DeleteByKey(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockTerraformIdpSubjectIdentifierRepository) DeleteExpired(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTerraformIdpSubjectIdentifierRepository) WithTransaction(tx interface{}) repository.TerraformIdpSubjectIdentifierRepository {
	args := m.Called(tx)
	return args.Get(0).(repository.TerraformIdpSubjectIdentifierRepository)
}

func TestNewTerraformIdpService(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	assert.NotNil(t, service)
}

func TestCreateAuthorizationCode(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	req := AuthorizationCodeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://localhost:3000/callback",
		Scope:        "openid profile",
		State:        "test-state",
		ResponseType: "code",
	}

	authCodeRepo.On("Create", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)

	resp, err := service.CreateAuthorizationCode(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Code)
	assert.Equal(t, "test-state", resp.State)
	assert.Regexp(t, `^[A-Za-z0-9+/=]+$`, resp.Code) // Base64 pattern

	authCodeRepo.AssertExpectations(t)
}

func TestCreateAuthorizationCode_UnsupportedResponseType(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	req := AuthorizationCodeRequest{
		ClientID:     "test-client",
		RedirectURI:  "http://localhost:3000/callback",
		Scope:        "openid profile",
		State:        "test-state",
		ResponseType: "token", // Unsupported
	}

	resp, err := service.CreateAuthorizationCode(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unsupported response_type")
}

func TestExchangeCodeForToken_Success(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	req := AccessTokenRequest{
		GrantType:   "authorization_code",
		Code:        "test-auth-code",
		RedirectURI: "http://localhost:3000/callback",
		ClientID:    "test-client",
	}

	// Mock authorization code lookup
	authData := map[string]interface{}{
		"client_id":     "test-client",
		"redirect_uri":  "http://localhost:3000/callback",
		"scope":         "openid profile",
		"state":         "test-state",
		"response_type": "code",
	}
	authCodeRecord := &repository.AuthCodeRecord{
		Key:  "test-auth-code",
		Data: []byte(`{"client_id":"test-client","redirect_uri":"http://localhost:3000/callback"}`),
	}

	authCodeRepo.On("FindByKey", mock.Anything, "test-auth-code").Return(authCodeRecord, nil)
	authCodeRepo.On("DeleteByKey", mock.Anything, "test-auth-code").Return(nil)
	accessTokenRepo.On("Create", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)

	resp, err := service.ExchangeCodeForToken(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.AccessToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, 3600, resp.ExpiresIn)
	assert.Equal(t, "openid profile", resp.Scope)

	authCodeRepo.AssertExpectations(t)
	accessTokenRepo.AssertExpectations(t)
}

func TestExchangeCodeForToken_UnsupportedGrantType(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	req := AccessTokenRequest{
		GrantType:   "client_credentials", // Unsupported
		Code:        "test-auth-code",
		RedirectURI: "http://localhost:3000/callback",
		ClientID:    "test-client",
	}

	resp, err := service.ExchangeCodeForToken(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unsupported grant_type")
}

func TestExchangeCodeForToken_InvalidCode(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	req := AccessTokenRequest{
		GrantType:   "authorization_code",
		Code:        "invalid-code",
		RedirectURI: "http://localhost:3000/callback",
		ClientID:    "test-client",
	}

	authCodeRepo.On("FindByKey", mock.Anything, "invalid-code").Return(nil, assert.AnError)

	resp, err := service.ExchangeCodeForToken(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid authorization code")

	authCodeRepo.AssertExpectations(t)
}

func TestValidateToken_Success(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	token := "test-access-token"

	// Mock access token lookup
	userInfoData := map[string]interface{}{
		"sub":   "terraform-user",
		"name":  "Terraform User",
		"email": "terraform@example.com",
		"iss":   "terraform-idp",
		"aud":   "test-client",
	}

	tokenRecord := &repository.AccessTokenRecord{
		Key:  "test-access-token",
		Data: []byte(`{"sub":"terraform-user","name":"Terraform User"}`),
	}

	accessTokenRepo.On("FindByKey", mock.Anything, token).Return(tokenRecord, nil)

	resp, err := service.ValidateToken(context.Background(), token)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "terraform-user", resp.Sub)
	assert.Equal(t, "Terraform User", resp.Name)
	assert.Equal(t, "terraform@example.com", resp.Email)

	accessTokenRepo.AssertExpectations(t)
}

func TestValidateToken_InvalidToken(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	token := "invalid-token"

	accessTokenRepo.On("FindByKey", mock.Anything, token).Return(nil, assert.AnError)

	resp, err := service.ValidateToken(context.Background(), token)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid or expired token")

	accessTokenRepo.AssertExpectations(t)
}

func TestRevokeToken_Success(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	token := "test-access-token"

	accessTokenRepo.On("DeleteByKey", mock.Anything, token).Return(nil)

	err := service.RevokeToken(context.Background(), token)

	assert.NoError(t, err)

	accessTokenRepo.AssertExpectations(t)
}

func TestRevokeToken_Error(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	token := "test-access-token"

	accessTokenRepo.On("DeleteByKey", mock.Anything, token).Return(assert.AnError)

	err := service.RevokeToken(context.Background(), token)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke token")

	accessTokenRepo.AssertExpectations(t)
}

func TestCleanupExpired(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	authCodeRepo.On("DeleteExpired", mock.Anything).Return(int64(5), nil)
	accessTokenRepo.On("DeleteExpired", mock.Anything).Return(int64(3), nil)
	subjectIdentifierRepo.On("DeleteExpired", mock.Anything).Return(int64(1), nil)

	err := service.CleanupExpired(context.Background())

	assert.NoError(t, err)

	authCodeRepo.AssertExpectations(t)
	accessTokenRepo.AssertExpectations(t)
	subjectIdentifierRepo.AssertExpectations(t)
}

func TestStoreSubjectIdentifier(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	subject := "terraform-user"
	clientID := "test-client"
	data := map[string]interface{}{
		"custom_field": "custom_value",
	}

	subjectIdentifierRepo.On("Create", mock.Anything, mock.AnythingOfType("string"), mock.Anything, mock.Anything).Return(nil)

	err := service.StoreSubjectIdentifier(context.Background(), subject, clientID, data)

	assert.NoError(t, err)

	subjectIdentifierRepo.AssertExpectations(t)
}

func TestGetSubjectIdentifier(t *testing.T) {
	authCodeRepo := &MockTerraformIdpAuthorizationCodeRepository{}
	accessTokenRepo := &MockTerraformIdpAccessTokenRepository{}
	subjectIdentifierRepo := &MockTerraformIdpSubjectIdentifierRepository{}

	service := NewTerraformIdpService(authCodeRepo, accessTokenRepo, subjectIdentifierRepo)

	subject := "terraform-user"
	clientID := "test-client"

	record := &repository.SubjectIdentifierRecord{
		Key:  "terraform-user:test-client",
		Data: []byte(`{"custom_field":"custom_value"}`),
	}

	subjectIdentifierRepo.On("FindByKey", mock.Anything, "terraform-user:test-client").Return(record, nil)

	resp, err := service.GetSubjectIdentifier(context.Background(), subject, clientID)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "custom_value", resp["custom_field"])

	subjectIdentifierRepo.AssertExpectations(t)
}
