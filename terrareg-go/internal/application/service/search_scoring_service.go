package service

import (
	"strings"
)

// SearchScoringService handles calculating relevance scores for module search results
type SearchScoringService struct{}

// NewSearchScoringService creates a new SearchScoringService
func NewSearchScoringService() *SearchScoringService {
	return &SearchScoringService{}
}

// CalculateRelevanceScore calculates the relevance score based on Python scoring algorithm
// Scoring weights:
// - Exact module name match: 20 points
// - Exact namespace match: 18 points
// - Exact provider match: 14 points
// - Exact description match: 13 points
// - Exact owner match: 12 points
// - Partial module name match: 5 points (wildcarded)
// - Partial description match: 4 points (wildcarded)
// - Partial owner match: 3 points (wildcarded)
// - Partial namespace match: 2 points (wildcarded)
func (s *SearchScoringService) CalculateRelevanceScore(query, moduleName, namespace, provider, description, owner string) int {
	query = strings.ToLower(query)
	moduleName = strings.ToLower(moduleName)
	namespace = strings.ToLower(namespace)
	provider = strings.ToLower(provider)
	description = strings.ToLower(description)
	owner = strings.ToLower(owner)

	queryParts := strings.Fields(query)
	totalScore := 0

	for _, part := range queryParts {
		// Exact matches (higher points)
		if moduleName == part {
			totalScore += 20 // Exact module name
		}
		if namespace == part {
			totalScore += 18 // Exact namespace
		}
		if provider == part {
			totalScore += 14 // Exact provider
		}
		if description == part {
			totalScore += 13 // Exact description
		}
		if owner == part {
			totalScore += 12 // Exact owner
		}

		// Partial matches (lower points)
		if strings.Contains(moduleName, part) && moduleName != part {
			totalScore += 5 // Partial module name
		}
		if strings.Contains(description, part) && description != part {
			totalScore += 4 // Partial description
		}
		if strings.Contains(owner, part) && owner != part {
			totalScore += 3 // Partial owner
		}
		if strings.Contains(namespace, part) && namespace != part {
			totalScore += 2 // Partial namespace
		}
	}

	return totalScore
}
