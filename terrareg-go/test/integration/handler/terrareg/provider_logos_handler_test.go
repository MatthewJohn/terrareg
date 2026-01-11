package terrareg_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/application/query/provider_logo"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/provider_logo/model"
	"github.com/matthewjohn/terrareg/terrareg-go/internal/interfaces/http/handler/terrareg"
	"github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// FakeProviderLogoRepository is a fake implementation of ProviderLogoRepository for testing
type FakeProviderLogoRepository struct {
	logos map[string]model.ProviderLogoInfo
}

func (f *FakeProviderLogoRepository) GetProviderLogo(providerName string) (*model.ProviderLogoInfo, bool) {
	info, exists := f.logos[providerName]
	return &info, exists
}

func (f *FakeProviderLogoRepository) GetAllProviderLogos() map[string]model.ProviderLogoInfo {
	return f.logos
}

// TestProviderLogosHandler_HandleGetProviderLogos_Success tests successful provider logos retrieval
func TestProviderLogosHandler_HandleGetProviderLogos_Success(t *testing.T) {
	// Create fake repository with test data
	fakeRepo := &FakeProviderLogoRepository{
		logos: map[string]model.ProviderLogoInfo{
			"aws": {
				Source: "/static/logos/aws.png",
				Alt:    "AWS Logo",
				Tos:    "AWS Terms of Service",
				Link:   "https://aws.amazon.com",
			},
			"google": {
				Source: "/static/logos/google.png",
				Alt:    "Google Cloud Logo",
				Tos:    "Google Terms of Service",
				Link:   "https://cloud.google.com",
			},
			"azurerm": {
				Source: "/static/logos/azure.png",
				Alt:    "Azure Logo",
				Tos:    "Microsoft Terms of Service",
				Link:   "https://azure.microsoft.com",
			},
		},
	}

	getAllProviderLogosQuery := provider_logo.NewGetAllProviderLogosQuery(fakeRepo)
	handler := terrareg.NewProviderLogosHandler(getAllProviderLogosQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/provider_logos", nil)
	w := httptest.NewRecorder()

	handler.HandleGetProviderLogos(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	response := testutils.GetJSONBody(t, w)

	// Should have all three providers
	assert.Contains(t, response, "aws")
	assert.Contains(t, response, "google")
	assert.Contains(t, response, "azurerm")

	// Check aws logo structure
	awsLogo := response["aws"].(map[string]interface{})
	assert.Equal(t, "/static/logos/aws.png", awsLogo["source"])
	assert.Equal(t, "AWS Logo", awsLogo["alt"])
	assert.Equal(t, "AWS Terms of Service", awsLogo["tos"])
	assert.Equal(t, "https://aws.amazon.com", awsLogo["link"])

	// Check google logo structure
	googleLogo := response["google"].(map[string]interface{})
	assert.Equal(t, "/static/logos/google.png", googleLogo["source"])
	assert.Equal(t, "Google Cloud Logo", googleLogo["alt"])
}

// TestProviderLogosHandler_HandleGetProviderLogos_Empty tests with no provider logos
func TestProviderLogosHandler_HandleGetProviderLogos_Empty(t *testing.T) {
	fakeRepo := &FakeProviderLogoRepository{
		logos: map[string]model.ProviderLogoInfo{},
	}

	getAllProviderLogosQuery := provider_logo.NewGetAllProviderLogosQuery(fakeRepo)
	handler := terrareg.NewProviderLogosHandler(getAllProviderLogosQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/provider_logos", nil)
	w := httptest.NewRecorder()

	handler.HandleGetProviderLogos(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should be empty object
	assert.Equal(t, map[string]interface{}{}, response)
}

// TestProviderLogosHandler_HandleGetProviderLogos_PartialData tests with some logos having incomplete data
func TestProviderLogosHandler_HandleGetProviderLogos_PartialData(t *testing.T) {
	fakeRepo := &FakeProviderLogoRepository{
		logos: map[string]model.ProviderLogoInfo{
			"aws": {
				Source: "/static/logos/aws.png",
				Alt:    "AWS Logo",
				Tos:    "",
				Link:   "",
			},
		},
	}

	getAllProviderLogosQuery := provider_logo.NewGetAllProviderLogosQuery(fakeRepo)
	handler := terrareg.NewProviderLogosHandler(getAllProviderLogosQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/provider_logos", nil)
	w := httptest.NewRecorder()

	handler.HandleGetProviderLogos(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should have aws provider
	assert.Contains(t, response, "aws")
	awsLogo := response["aws"].(map[string]interface{})
	assert.Equal(t, "/static/logos/aws.png", awsLogo["source"])
	assert.Equal(t, "AWS Logo", awsLogo["alt"])
	// Empty fields should still be present as empty strings
	assert.Equal(t, "", awsLogo["tos"])
	assert.Equal(t, "", awsLogo["link"])
}

// TestProviderLogosHandler_HandleGetProviderLogos_SingleProvider tests with single provider
func TestProviderLogosHandler_HandleGetProviderLogos_SingleProvider(t *testing.T) {
	fakeRepo := &FakeProviderLogoRepository{
		logos: map[string]model.ProviderLogoInfo{
			"hashicorp": {
				Source: "/static/logos/hashicorp.png",
				Alt:    "HashiCorp Logo",
				Tos:    "HashiCorp Terms",
				Link:   "https://www.hashicorp.com",
			},
		},
	}

	getAllProviderLogosQuery := provider_logo.NewGetAllProviderLogosQuery(fakeRepo)
	handler := terrareg.NewProviderLogosHandler(getAllProviderLogosQuery)

	req := httptest.NewRequest("GET", "/v1/terrareg/provider_logos", nil)
	w := httptest.NewRecorder()

	handler.HandleGetProviderLogos(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	response := testutils.GetJSONBody(t, w)

	// Should have exactly one provider
	assert.Len(t, response, 1)
	assert.Contains(t, response, "hashicorp")
}
