package analytics

import (
	"context"

	moduleRepo "github.com/terrareg/terrareg/internal/domain/module/repository"
)

// GlobalStatsQuery handles getting global statistics
type GlobalStatsQuery struct {
	namespaceRepo      moduleRepo.NamespaceRepository
	moduleProviderRepo moduleRepo.ModuleProviderRepository
}

// NewGlobalStatsQuery creates a new global stats query
func NewGlobalStatsQuery(
	namespaceRepo moduleRepo.NamespaceRepository,
	moduleProviderRepo moduleRepo.ModuleProviderRepository,
) *GlobalStatsQuery {
	return &GlobalStatsQuery{
		namespaceRepo:      namespaceRepo,
		moduleProviderRepo: moduleProviderRepo,
	}
}

// GlobalStats represents global statistics
type GlobalStats struct {
	Namespaces     int `json:"namespaces"`
	Modules        int `json:"modules"`
	ModuleVersions int `json:"module_versions"`
	Downloads      int `json:"downloads"`
}

// Execute executes the query
func (q *GlobalStatsQuery) Execute(ctx context.Context) (*GlobalStats, error) {
	// Get namespace count
	namespaces, err := q.namespaceRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// For now, return simple counts
	// TODO: Implement actual counts from database aggregations
	stats := &GlobalStats{
		Namespaces:     len(namespaces),
		Modules:        0, // TODO: Count distinct module providers
		ModuleVersions: 0, // TODO: Count module versions
		Downloads:      0, // TODO: Count from analytics table
	}

	return stats, nil
}
