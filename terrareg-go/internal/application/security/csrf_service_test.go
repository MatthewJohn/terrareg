package security

import (
	"context"
	"errors"
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/security/csrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mocks defined inline to avoid import cycles

// MockSessionManager mocks the SessionManager interface
type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) GetCSRFToken(ctx context.Context, sessionID string) (csrf.CSRFToken, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(csrf.CSRFToken), args.Error(1)
}

func (m *MockSessionManager) CreateSession(ctx context.Context) (string, csrf.CSRFToken, error) {
	args := m.Called(ctx)
	return args.String(0), args.Get(1).(csrf.CSRFToken), args.Error(2)
}

func (m *MockSessionManager) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// MockTokenGenerator mocks the TokenGenerator interface
type MockTokenGenerator struct {
	mock.Mock
}

func (m *MockTokenGenerator) GenerateToken() (csrf.CSRFToken, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.Get(0).(csrf.CSRFToken), args.Error(1)
}

// MockTokenValidator mocks the TokenValidator interface
type MockTokenValidator struct {
	mock.Mock
}

func (m *MockTokenValidator) ValidateToken(expected, provided csrf.CSRFToken, required bool) error {
	args := m.Called(expected, provided, required)
	return args.Error(0)
}

func TestNewCSRFService(t *testing.T) {
	tests := []struct {
		name           string
		tokenGenerator csrf.TokenGenerator
		tokenValidator csrf.TokenValidator
		sessionManager SessionManager
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "all valid dependencies",
			tokenGenerator: &MockTokenGenerator{},
			tokenValidator: &MockTokenValidator{},
			sessionManager: &MockSessionManager{},
			wantErr:        false,
		},
		{
			name:           "nil token generator",
			tokenGenerator: nil,
			tokenValidator: &MockTokenValidator{},
			sessionManager: &MockSessionManager{},
			wantErr:        true,
			errMsg:         "tokenGenerator cannot be nil",
		},
		{
			name:           "nil token validator",
			tokenGenerator: &MockTokenGenerator{},
			tokenValidator: nil,
			sessionManager: &MockSessionManager{},
			wantErr:        true,
			errMsg:         "tokenValidator cannot be nil",
		},
		{
			name:           "nil session manager",
			tokenGenerator: &MockTokenGenerator{},
			tokenValidator: &MockTokenValidator{},
			sessionManager: nil,
			wantErr:        true,
			errMsg:         "sessionManager cannot be nil",
		},
		{
			name:           "all nil",
			tokenGenerator: nil,
			tokenValidator: nil,
			sessionManager: nil,
			wantErr:        true,
			errMsg:         "tokenGenerator cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, err := NewCSRFService(
				tt.tokenGenerator,
				tt.tokenValidator,
				tt.sessionManager,
			)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, service)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service)
			}
		})
	}
}

func TestCSRFService_GetOrCreateSessionToken(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		setupMock func(*MockSessionManager)
		wantErr   bool
		wantEmpty bool
	}{
		{
			name:      "existing session with valid token",
			sessionID: "existing-session-id",
			setupMock: func(m *MockSessionManager) {
				m.On("GetCSRFToken", mock.Anything, "existing-session-id").
					Return(csrf.CSRFToken("existing-token"), nil).Once()
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name:      "existing session but no token",
			sessionID: "empty-token-session",
			setupMock: func(m *MockSessionManager) {
				m.On("GetCSRFToken", mock.Anything, "empty-token-session").
					Return(csrf.CSRFToken(""), nil).Once()
				m.On("CreateSession", mock.Anything).
					Return("new-session-id", csrf.CSRFToken("new-token"), nil).Once()
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name:      "existing session with error",
			sessionID: "error-session",
			setupMock: func(m *MockSessionManager) {
				m.On("GetCSRFToken", mock.Anything, "error-session").
					Return(csrf.CSRFToken(""), errors.New("session not found")).Once()
				m.On("CreateSession", mock.Anything).
					Return("new-session-id", csrf.CSRFToken("new-token"), nil).Once()
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name:      "no session ID",
			sessionID: "",
			setupMock: func(m *MockSessionManager) {
				m.On("CreateSession", mock.Anything).
					Return("new-session-id", csrf.CSRFToken("new-token"), nil).Once()
			},
			wantErr:   false,
			wantEmpty: false,
		},
		{
			name:      "create session fails",
			sessionID: "",
			setupMock: func(m *MockSessionManager) {
				m.On("CreateSession", mock.Anything).
					Return("", csrf.CSRFToken(""), errors.New("failed to create session")).Once()
			},
			wantErr:   true,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &MockSessionManager{}
			mockGenerator := &MockTokenGenerator{}
			mockValidator := &MockTokenValidator{}
			tt.setupMock(mockSession)

			service, err := NewCSRFService(mockGenerator, mockValidator, mockSession)
			require.NoError(t, err)

			token, err := service.GetOrCreateSessionToken(context.Background(), tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, token.IsEmpty())
			} else {
				assert.NoError(t, err)
				assert.False(t, token.IsEmpty())
			}

			mockSession.AssertExpectations(t)
		})
	}
}

func TestCSRFService_ValidateRequestToken(t *testing.T) {
	tests := []struct {
		name          string
		sessionID     string
		providedToken csrf.CSRFToken
		required      bool
		setupMock     func(*MockSessionManager, *MockTokenValidator)
		wantErr       error
	}{
		{
			name:          "valid token",
			sessionID:     "session-123",
			providedToken: "abc123def456",
			required:      true,
			setupMock: func(m *MockSessionManager, v *MockTokenValidator) {
				m.On("GetCSRFToken", mock.Anything, "session-123").
					Return(csrf.CSRFToken("abc123def456"), nil).Once()
				v.On("ValidateToken", csrf.CSRFToken("abc123def456"), csrf.CSRFToken("abc123def456"), true).
					Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:          "invalid token",
			sessionID:     "session-123",
			providedToken: "xyz789",
			required:      true,
			setupMock: func(m *MockSessionManager, v *MockTokenValidator) {
				m.On("GetCSRFToken", mock.Anything, "session-123").
					Return(csrf.CSRFToken("abc123"), nil).Once()
				v.On("ValidateToken", csrf.CSRFToken("abc123"), csrf.CSRFToken("xyz789"), true).
					Return(csrf.ErrInvalidToken).Once()
			},
			wantErr: csrf.ErrInvalidToken,
		},
		{
			name:          "not required",
			sessionID:     "session-123",
			providedToken: "",
			required:      false,
			setupMock: func(m *MockSessionManager, v *MockTokenValidator) {
				// Should not call anything when not required
			},
			wantErr: nil,
		},
		{
			name:          "session error",
			sessionID:     "session-123",
			providedToken: "abc123",
			required:      true,
			setupMock: func(m *MockSessionManager, v *MockTokenValidator) {
				m.On("GetCSRFToken", mock.Anything, "session-123").
					Return(csrf.CSRFToken(""), errors.New("session expired")).Once()
			},
			wantErr: errors.New("failed to get session CSRF token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &MockSessionManager{}
			mockGenerator := &MockTokenGenerator{}
			mockValidator := &MockTokenValidator{}
			tt.setupMock(mockSession, mockValidator)

			service, err := NewCSRFService(mockGenerator, mockValidator, mockSession)
			require.NoError(t, err)

			err = service.ValidateRequestToken(context.Background(), tt.sessionID, tt.providedToken, tt.required)

			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			mockSession.AssertExpectations(t)
			mockValidator.AssertExpectations(t)
		})
	}
}
