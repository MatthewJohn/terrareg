package module

import (
	types "github.com/matthewjohn/terrareg/terrareg-go/internal/domain/shared/types"
)

// ImportModuleVersionRequest represents a module import request shared between domain and application layers
type ImportModuleVersionRequest struct {
	Namespace types.NamespaceName
	Module    types.ModuleName
	Provider  types.ModuleProviderName
	Version   types.ModuleVersion
	GitTag    types.GitTag
}
