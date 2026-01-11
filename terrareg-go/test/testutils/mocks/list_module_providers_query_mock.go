package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// MockListModuleProvidersQuery is a mock for ListModuleProvidersQuery
type MockListModuleProvidersQuery struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *MockListModuleProvidersQuery) Execute(ctx context.Context, namespace, module string) ([]*model.ModuleProvider, error) {
	args := m.Called(ctx, namespace, module)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ModuleProvider), args.Error(1)
}
