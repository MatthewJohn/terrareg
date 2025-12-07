package service

import (
	"github.com/matthewjohn/terrareg/terrareg-go/internal/config"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
)

// NamespaceService provides domain services for namespaces
type NamespaceService struct {
	config *config.Config
}

// NewNamespaceService creates a new namespace service
func NewNamespaceService(config *config.Config) *NamespaceService {
	return &NamespaceService{
		config: config,
	}
}

// IsTrusted checks if a namespace is in the trusted list
func (s *NamespaceService) IsTrusted(namespace *model.Namespace) bool {
	if namespace == nil {
		return false
	}

	for _, ns := range s.config.TrustedNamespaces {
		if ns == namespace.Name() {
			return true
		}
	}

	return false
}

// IsAutoVerified checks if a namespace is in the auto-verified list
func (s *NamespaceService) IsAutoVerified(namespace *model.Namespace) bool {
	if namespace == nil {
		return false
	}

	for _, ns := range s.config.VerifiedModuleNamespaces {
		if ns == namespace.Name() {
			return true
		}
	}

	return false
}
