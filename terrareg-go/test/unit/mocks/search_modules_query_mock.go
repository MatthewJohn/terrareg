package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
	modulequery "github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/module"
)

// MockSearchModulesQuery is a mock for SearchModulesQuery
type MockSearchModulesQuery struct {
	mock.Mock
}

// Execute mocks the Execute method
func (m *MockSearchModulesQuery) Execute(ctx context.Context, params modulequery.SearchParams) (*modulequery.SearchResult, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*modulequery.SearchResult), args.Error(1)
}
