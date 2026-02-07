package module

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/url/service"
	infraConfig "github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestURLService creates a URL service for testing
func createTestURLService(t *testing.T) *service.URLService {
	cfg := &infraConfig.InfrastructureConfig{
		PublicURL: "http://localhost:5000",
	}
	urlService, err := service.NewURLService(cfg)
	require.NoError(t, err, "Failed to create URL service")
	return urlService
}

// TestGetExampleDetailsQuery_getCostAnalysis_NoDetails tests getCostAnalysis when example has no details
// Python reference: /app/terrareg/models.py Example.get_terrareg_api_details() - cost analysis section
func TestGetExampleDetailsQuery_getCostAnalysis_NoDetails(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example without details
	exampleName := "test-example"
	example := model.NewExample("examples/test-example", &exampleName, nil)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when example has no details")
}

// TestGetExampleDetailsQuery_getCostAnalysis_NoInfracost tests getCostAnalysis when details have no infracost data
func TestGetExampleDetailsQuery_getCostAnalysis_NoInfracost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with details but no infracost data
	exampleName := "test-example"
	details := model.NewModuleDetails([]byte("# Test README"))
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when details have no infracost data")
}

// TestGetExampleDetailsQuery_getCostAnalysis_InvalidJSON tests getCostAnalysis with invalid JSON in infracost
func TestGetExampleDetailsQuery_getCostAnalysis_InvalidJSON(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with invalid JSON infracost data
	exampleName := "test-example"
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost([]byte("invalid json {"))
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when infracost data is invalid JSON")
}

// TestGetExampleDetailsQuery_getCostAnalysis_MissingTotalMonthlyCost tests getCostAnalysis when infracost JSON is valid but missing total_monthly_cost
func TestGetExampleDetailsQuery_getCostAnalysis_MissingTotalMonthlyCost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with valid infracost JSON but missing total_monthly_cost
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_hourly_cost": "0.001",
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when total_monthly_cost is missing")
}

// TestGetExampleDetailsQuery_getCostAnalysis_ValidInfracost tests getCostAnalysis with valid infracost data
func TestGetExampleDetailsQuery_getCostAnalysis_ValidInfracost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with valid infracost data
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": 12.34,
		"total_hourly_cost": 0.017,
		"projects": [
			{
				"name": "test-project",
				"breakdown": []
			}
		]
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis when infracost data is valid")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// Calculate expected yearly cost: 12.34 * 12 = 148.08
	expectedYearlyCost := "148.08"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should be monthly cost * 12")
}

// TestGetExampleDetailsQuery_getCostAnalysis_ZeroCost tests getCostAnalysis with zero cost
func TestGetExampleDetailsQuery_getCostAnalysis_ZeroCost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with zero cost infracost data
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": 0.00,
		"total_hourly_cost": 0.0,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis even with zero cost")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// 0.00 * 12 = 0.00
	expectedYearlyCost := "0.00"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should be 0.00 for zero monthly cost")
}

// TestGetExampleDetailsQuery_getCostAnalysis_LargeCost tests getCostAnalysis with large cost values
func TestGetExampleDetailsQuery_getCostAnalysis_LargeCost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with large cost infracost data
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": 1234.56,
		"total_hourly_cost": 1.702,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis for large costs")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// 1234.56 * 12 = 14814.72
	expectedYearlyCost := "14814.72"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should correctly calculate large values")
}

// TestGetExampleDetailsQuery_getCostAnalysis_ScientificNotation tests getCostAnalysis with scientific notation in cost
func TestGetExampleDetailsQuery_getCostAnalysis_ScientificNotation(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with scientific notation cost
	exampleName := "test-example"
	// Using scientific notation: 1.23e-4 = 0.000123
	infracostJSON := []byte(`{
		"total_monthly_cost": 1.23e-4,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis with scientific notation")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// 0.000123 * 12 = 0.001476, rounded to 2 decimal places = 0.00
	expectedYearlyCost := "0.00"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should handle scientific notation")
}

// TestGetExampleDetailsQuery_getCostAnalysis_TotalMonthlyCostIsString tests getCostAnalysis when total_monthly_cost is a string instead of float
func TestGetExampleDetailsQuery_getCostAnalysis_TotalMonthlyCostIsString(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with string total_monthly_cost (invalid type)
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": "12.34",
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when total_monthly_cost is not a number")
}

// TestGetExampleDetailsQuery_getCostAnalysis_NegativeCost tests getCostAnalysis with negative cost
func TestGetExampleDetailsQuery_getCostAnalysis_NegativeCost(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with negative cost (edge case, but should handle gracefully)
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": -10.50,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis even with negative cost")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// -10.50 * 12 = -126.00
	expectedYearlyCost := "-126.00"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should handle negative values")
}

// TestGetExampleDetailsQuery_getCostAnalysis_FractionalPennies tests getCostAnalysis with fractional pennies
func TestGetExampleDetailsQuery_getCostAnalysis_FractionalPennies(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with fractional pennies
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": 10.123456,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis with fractional pennies")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// 10.123456 * 12 = 121.481472, formatted to 2 decimal places = 121.48
	expectedYearlyCost := "121.48"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should round to 2 decimal places")
}

// TestGetExampleDetailsQuery_getCostAnalysis_TotalMonthlyCostIsInteger tests getCostAnalysis when total_monthly_cost is an integer
func TestGetExampleDetailsQuery_getCostAnalysis_TotalMonthlyCostIsInteger(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with integer total_monthly_cost (valid JSON number)
	exampleName := "test-example"
	infracostJSON := []byte(`{
		"total_monthly_cost": 100,
		"projects": []
	}`)
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost(infracostJSON)
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.NotNil(t, result, "getCostAnalysis should return CostAnalysis with integer monthly cost")
	assert.NotNil(t, result.YearlyCost, "YearlyCost should not be nil")

	// 100 * 12 = 1200.00
	expectedYearlyCost := "1200.00"
	assert.Equal(t, expectedYearlyCost, *result.YearlyCost, "YearlyCost should format integer as 2 decimal places")
}

// TestGetExampleDetailsQuery_getCostAnalysis_EmptyInfracostJSON tests getCostAnalysis with empty infracost JSON
func TestGetExampleDetailsQuery_getCostAnalysis_EmptyInfracostJSON(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with empty infracost JSON
	exampleName := "test-example"
	details := model.NewModuleDetails([]byte("# Test README")).
		WithInfracost([]byte("{}"))
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when infracost JSON is empty")
}

// TestGetExampleDetailsQuery_getCostAnalysis_NilInfracostBytes tests getCostAnalysis when infracost bytes are nil
func TestGetExampleDetailsQuery_getCostAnalysis_NilInfracostBytes(t *testing.T) {
	urlService := createTestURLService(t)
	query := NewGetExampleDetailsQuery(nil, nil, urlService)

	// Create example with nil infracost bytes (using NewModuleDetails without WithInfracost)
	exampleName := "test-example"
	details := model.NewModuleDetails([]byte("# Test README"))
	example := model.NewExample("examples/test-example", &exampleName, details)

	result := query.getCostAnalysis(example)

	assert.Nil(t, result, "getCostAnalysis should return nil when infracost bytes are nil")
}
