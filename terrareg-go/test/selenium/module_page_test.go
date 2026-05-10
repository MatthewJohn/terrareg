//go:build selenium

package selenium

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModulePage tests the module page functionality.
// This is the Go implementation of the module details page tests.
// Python reference: /app/test/selenium/test_edit_namespace.py - navigation to /modules/moduledetails
//
// Test methods:
// - test_module_page_with_providers - tests module page with providers displays correctly
// - test_module_page_without_providers - tests module page with no providers
// - test_module_page_non_existent - tests non-existent module page
// - test_module_page_navigation - tests navigation to provider details

func TestModulePage(t *testing.T) {
	t.Run("test_module_page_with_providers", testModulePageWithProviders)
	t.Run("test_module_page_without_providers", testModulePageWithoutProviders)
	t.Run("test_module_page_non_existent", testModulePageNonExistent)
	t.Run("test_module_page_navigation", testModulePageNavigation)
}

// testModulePageWithProviders tests the module page with providers displays correctly.
// Python reference: /app/test/selenium/test_edit_namespace.py - test_navigation_from_namespace
func testModulePageWithProviders(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Navigate to module page with multiple providers
	// Python: self.selenium_instance.get(self.get_url("/modules/moduledetails"))
	st.NavigateTo("/modules/moduledetails")

	// Verify page title
	// Python: assert self.selenium_instance.title == 'moduledetails - Terrareg'
	title := st.GetTitle()
	assert.Equal(t, "moduledetails - Terrareg", title, "Module page title should match")

	// Verify breadcrumb is displayed correctly
	// Python reference: breadcrumb shows 'Modules > moduledetails > testmodule'
	breadcrumbElement := st.WaitForElement("#breadcrumb-ul")
	breadcrumbText := breadcrumbElement.Text()
	assert.Contains(t, breadcrumbText, "Modules", "Breadcrumb should contain 'Modules'")
	assert.Contains(t, breadcrumbText, "moduledetails", "Breadcrumb should contain namespace name")

	// Verify provider cards are displayed
	// The page calls /v1/terrareg/modules/moduledetails/testmodule API
	// and uses createSearchResultCard to display results
	// Python: createSearchResultCard('module-list-table', 'module', module_data)
	moduleListTable := st.WaitForElement("#module-list-table")
	require.NotNil(t, moduleListTable, "Module list table should be displayed")

	// Verify "module does not exist" message is NOT shown
	// Python: $('#module-does-not-exist') should be hidden
	st.AssertElementNotVisible("#module-does-not-exist")

	// Verify "no results" message is NOT shown (there are providers)
	st.AssertElementNotVisible("#no-results")
}

// testModulePageWithoutProviders tests the module page with no providers.
// Python reference: /app/test/selenium/test_edit_namespace.py - similar pattern for empty results
func testModulePageWithoutProviders(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Navigate to module page that has no providers
	// This would require setting up test data with a module that has no providers
	// For now, we'll use a URL that would return no providers
	st.NavigateTo("/modules/emptynamespace/emptymodule")

	// Verify "no results" message is shown
	// Python: $('#no-results').removeClass('default-hidden')
	st.AssertElementVisible("#no-results")

	// Verify "module does not exist" message is NOT shown
	// (it's different from "no results" - "module does not exist" means the module itself doesn't exist)
	st.AssertElementNotVisible("#module-does-not-exist")

	// Verify module list table is hidden or empty
	st.AssertElementNotVisible("#module-list-table")
}

// testModulePageNonExistent tests non-existent module page.
// Python reference: /app/test/selenium/test_edit_namespace.py - similar pattern for 404
func testModulePageNonExistent(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Navigate to non-existent module
	st.NavigateTo("/modules/doesnotexist/nonexistent")

	// Verify "module does not exist" message is shown
	// Python: $('#module-does-not-exist').removeClass('default-hidden')
	st.AssertElementVisible("#module-does-not-exist")

	noResultsElement := st.WaitForElement("#module-does-not-exist")
	noResultsText := noResultsElement.Text()
	assert.Contains(t, noResultsText, "does not exist", "Error message should indicate module does not exist")

	// Verify "no results" message is NOT shown
	st.AssertElementNotVisible("#no-results")

	// Verify module list table is hidden
	st.AssertElementNotVisible("#module-list-table")
}

// testModulePageNavigation tests navigation from module page to provider details.
// Python reference: /app/test/selenium/test_edit_namespace.py - test_navigation_from_namespace
func testModulePageNavigation(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Navigate to module page with providers
	st.NavigateTo("/modules/moduledetails")

	// Wait for module list to load
	st.WaitForElement("#module-list-table")

	// Click on a provider card and verify navigation
	// The provider cards are created by createSearchResultCard JavaScript function
	// Python: Click on provider card, verify navigation to /modules/{namespace}/{name}/{provider}

	// Find the first provider card link
	// The link is generated as: /modules/${details.namespace}/${details.name}/${details.provider}
	linkElement := st.WaitForElement("#module-list-table a")
	firstProviderLink := linkElement.GetAttribute("href")
	require.NotEmpty(t, firstProviderLink, "Should find provider link")

	// Click on the provider link
	linkElement.Click()

	// Verify navigation to provider details page
	// Python: assert self.selenium_instance.current_url == self.get_url("/modules/...")
	currentURL := st.GetCurrentURL()
	assert.Contains(t, currentURL, "/modules/", "URL should contain /modules/")
	// Provider details URL has three path segments (namespace/module/provider)
	// while module page has only two (namespace/module)
	assert.Contains(t, currentURL, "moduledetails/", "URL should contain namespace name")
}
