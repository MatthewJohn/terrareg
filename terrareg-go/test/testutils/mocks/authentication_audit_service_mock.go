package mocks

import (
	"context"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/audit/service"
	"github.com/stretchr/testify/mock"
)

// MockAuthenticationAuditService is a mock for AuthenticationAuditServiceInterface
type MockAuthenticationAuditService struct {
	mock.Mock
}

// Ensure MockAuthenticationAuditService implements the interface at compile time
var _ service.AuthenticationAuditServiceInterface = (*MockAuthenticationAuditService)(nil)

// LogUserLogin mocks the method
func (m *MockAuthenticationAuditService) LogUserLogin(ctx context.Context, username, authMethod string) error {
	args := m.Called(ctx, username, authMethod)
	return args.Error(0)
}
