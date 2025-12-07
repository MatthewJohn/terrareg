package presenter

import (
	"testing"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/domain/module/model"
	"github.com/stretchr/testify/assert"
)

func createMockNamespaceForTest(name string, displayName *string, namespaceType string) *model.Namespace {
	namespace, _ := model.NewNamespace(name, displayName, model.NamespaceType(namespaceType))
	return namespace
}

func TestNamespacePresenter_ToDTO(t *testing.T) {
	// Arrange
	displayName := "Example Namespace"
	ns := createMockNamespaceForTest("example", &displayName, "GITHUB_ORGANISATION")
	presenter := NewNamespacePresenter()

	// Act
	result := presenter.ToDTO(ns)

	// Assert
	assert.Equal(t, "example", result.Name)
	assert.Equal(t, "Example Namespace", *result.DisplayName)
	assert.Equal(t, "GITHUB_ORGANISATION", result.Type)
}

func TestNamespacePresenter_ToDTO_WithoutDisplayName(t *testing.T) {
	// Arrange
	ns := createMockNamespaceForTest("example", nil, "NONE")
	presenter := NewNamespacePresenter()

	// Act
	result := presenter.ToDTO(ns)

	// Assert
	assert.Equal(t, "example", result.Name)
	assert.Nil(t, result.DisplayName)
	assert.Equal(t, "NONE", result.Type)
}

func TestNamespacePresenter_ToListDTO(t *testing.T) {
	// Arrange
	displayName1 := "Organization A"
	displayName2 := "User Namespace"

	ns1 := createMockNamespaceForTest("org-a", &displayName1, "GITHUB_ORGANISATION")
	ns2 := createMockNamespaceForTest("user-a", &displayName2, "GITHUB_USER")
	ns3 := createMockNamespaceForTest("system", nil, "NONE")

	namespaces := []*model.Namespace{ns1, ns2, ns3}
	presenter := NewNamespacePresenter()

	// Act
	result := presenter.ToListDTO(namespaces)

	// Assert
	assert.Len(t, result.Namespaces, 3)

	// First namespace
	assert.Equal(t, "org-a", result.Namespaces[0].Name)
	assert.Equal(t, "Organization A", *result.Namespaces[0].DisplayName)
	assert.Equal(t, "GITHUB_ORGANISATION", result.Namespaces[0].Type)

	// Second namespace
	assert.Equal(t, "user-a", result.Namespaces[1].Name)
	assert.Equal(t, "User Namespace", *result.Namespaces[1].DisplayName)
	assert.Equal(t, "GITHUB_USER", result.Namespaces[1].Type)

	// Third namespace
	assert.Equal(t, "system", result.Namespaces[2].Name)
	assert.Nil(t, result.Namespaces[2].DisplayName)
	assert.Equal(t, "NONE", result.Namespaces[2].Type)
}

func TestNamespacePresenter_ToListArray(t *testing.T) {
	// Arrange
	displayName := "Test Namespace"
	ns := createMockNamespaceForTest("test", &displayName, "NONE")
	namespaces := []*model.Namespace{ns}
	presenter := NewNamespacePresenter()

	// Act
	result := presenter.ToListArray(namespaces)

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "test", result[0].Name)
	assert.Equal(t, "Test Namespace", *result[0].DisplayName)
	assert.Equal(t, "NONE", result[0].Type)
}

func TestNamespacePresenter_ToListArray_Empty(t *testing.T) {
	// Arrange
	var namespaces []*model.Namespace
	presenter := NewNamespacePresenter()

	// Act
	result := presenter.ToListArray(namespaces)

	// Assert
	assert.Empty(t, result)
}