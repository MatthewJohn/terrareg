package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// MockGetModuleProviderQuery is a mock for GetModuleProviderQuery
type MockGetModuleProviderQuery struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *MockGetModuleProviderQuery) Execute(ctx context.Context, namespace, module, provider string) (*model.ModuleProvider, error) {
	args := m.Called(ctx, namespace, module, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ModuleProvider), args.Error(1)
}
