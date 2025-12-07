package unit

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/service"
	"github.com/stretchr/testify/assert"
)

func TestSearchScoringService(t *testing.T) {
	service := service.NewSearchScoringService()

	tests := []struct {
		name     string
		query    string
		module   string
		ns       string
		provider string
		desc     string
		owner    string
		expected int
	}{
		{
			name:     "Exact module name match",
			query:    "test",
			module:   "test",
			ns:       "other",
			provider: "other",
			desc:     "other",
			owner:    "other",
			expected: 20,
		},
		{
			name:     "Exact namespace match",
			query:    "hashicorp",
			module:   "consul",
			ns:       "hashicorp",
			provider: "aws",
			desc:     "Provider for EC2",
			owner:    "OtherCorp",
			expected: 18,
		},
		{
			name:     "Multiple exact matches",
			query:    "test",
			module:   "test",
			ns:       "test",
			provider: "test",
			desc:     "test",
			owner:    "test",
			expected: 20 + 18 + 14 + 13 + 12, // All exact matches
		},
		{
			name:     "Partial module name match",
			query:    "test",
			module:   "testing",
			ns:       "other",
			provider: "other",
			desc:     "other",
			owner:    "other",
			expected: 5,
		},
		{
			name:     "Partial description match",
			query:    "aws",
			module:   "other",
			ns:       "other",
			provider: "other",
			desc:     "AWS provider for EC2",
			owner:    "other",
			expected: 4,
		},
		{
			name:     "No matches",
			query:    "nomatch",
			module:   "test",
			ns:       "other",
			provider: "other",
			desc:     "other",
			owner:    "other",
			expected: 0,
		},
		{
			name:     "Multiple words in query",
			query:    "test aws",
			module:   "test",
			ns:       "other",
			provider: "aws",
			desc:     "other",
			owner:    "other",
			expected: 20 + 14, // test matches module (20) + aws matches provider (14)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := service.CalculateRelevanceScore(tt.query, tt.module, tt.ns, tt.provider, tt.desc, tt.owner)
			assert.Equal(t, tt.expected, score)
		})
	}
}