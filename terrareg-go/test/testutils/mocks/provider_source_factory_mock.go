package mocks

import (
	"context"

	moduleModel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_source/service"
	"github.com/stretchr/testify/mock"
)

// MockProviderSourceFactory is a mock for ProviderSourceFactory
// It provides thread-safe mocking for audit service methods
type MockProviderSourceFactory struct {
	mock.Mock
}

// Ensure MockProviderAuditService implements the interface at compile time
var _ moduleModel.ProviderSourceFactory = (*MockProviderSourceFactory)(nil)

// GetProviderSourceByName mocks the method
func (m *MockProviderSourceFactory) GetProviderSourceByName(
	ctx context.Context,
	providerSourceName string,
) (service.ProviderSourceInstance, error) {

	args := m.Called(ctx, providerSourceName)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(service.ProviderSourceInstance), args.Error(1)
}
