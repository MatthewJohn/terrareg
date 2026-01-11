package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockListModulesQuery is a mock for ListModulesQuery
type MockListModulesQuery struct {
	mock.Mock
}

// Execute mocks the Execute method
// Returns: (interface{}, error) - caller should type assert to []*model.ModuleProvider
func (m *MockListModulesQuery) Execute(ctx context.Context, namespace ...string) (interface{}, error) {
	args := m.Called(ctx, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0), args.Error(1)
}
