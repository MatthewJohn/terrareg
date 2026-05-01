package user_group

import (
	modulemodel "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// Helper to create a mock namespace for testing
func createMockNamespace(id int, name types.NamespaceName) *modulemodel.Namespace {
	return modulemodel.ReconstructNamespace(id, name, nil, modulemodel.NamespaceTypeNone)
}
