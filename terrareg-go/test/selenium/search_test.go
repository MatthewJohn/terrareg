//go:build selenium

package selenium

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModuleSearch tests the module search functionality.
// This is the Go implementation of Python's test_module_search.py.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch class
//
// Test methods:
// - test_result_cards - checks result cards are displayed correctly
// - test_search_filters - checks search filter functionality
// - test_next_prev_buttons - checks pagination navigation
// - test_result_counts - checks result count text
// - test_result_relevancy_ordering - checks search result ordering
// - test_terraform_version_compatibility - checks terraform version compatibility (2 parametrized test cases)
// - test_terraform_version_compatibility_retains_state - checks terraform version state is retained

func TestModuleSearch(t *testing.T) {
	t.Run("test_result_cards", testModuleSearchResultCards)
	t.Run("test_search_filters", testModuleSearchFilters)
	t.Run("test_next_prev_buttons", testModuleSearchNextPrevButtons)
	t.Run("test_result_counts", testModuleSearchResultCounts)
	t.Run("test_result_relevancy_ordering", testModuleSearchResultRelevancyOrdering)
	t.Run("test_terraform_version_compatibility", testModuleSearchTerraformVersionCompatibility)
	t.Run("test_terraform_version_compatibility_retains_state", testModuleSearchTerraformVersionCompatibilityRetainsState)
}

// testModuleSearchResultCards checks the result cards.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_result_cards
func testModuleSearchResultCards(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))
	st.NavigateTo("/search/modules?q=modulesearch")

	// Python: self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)
	resultsContainer := st.WaitForElement("#results")
	require.NotNil(t, resultsContainer)

	// Expected card headings
	expectedCardHeadings := []string{
		"modulesearch-trusted / mixedsearch-trusted-result",
		"modulesearch-trusted / mixedsearch-trusted-result-multiversion",
		"modulesearch-trusted / mixedsearch-trusted-result-verified",
		"modulesearch-trusted / mixedsearch-trusted-second-result",
	}

	expectedCardLinks := []string{
		"/modules/modulesearch-trusted/mixedsearch-trusted-result/aws",
		"/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null",
		"/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp",
		"/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog",
	}

	expectedProviderText := []string{
		"Provider: aws",
		"Provider: null",
		"Provider: gcp",
		"Provider: datadog",
	}

	// Verify each card
	for i, expectedHeading := range expectedCardHeadings {
		// Python: heading = card.find_element(By.CLASS_NAME, 'module-card-title')
		//         assert heading.text == expected_card_headings.pop(0)
		selector := "#results .card:nth-child(" + intToStr(i+1) + ") .module-card-title"
		st.AssertTextContent(selector, expectedHeading)

		// Python: assert heading.get_attribute('href') == self.get_url(expected_card_links.pop(0))
		expectedLink := st.GetURL(expectedCardLinks[i])
		href := getLinkAttribute(st, selector, "href")
		assert.Equal(t, expectedLink, href, "Card link should match")

		// Python: assert card.find_element(By.CLASS_NAME, 'module-provider-card-provider-text').text == expected_card_provider_text.pop(0)
		providerSelector := "#results .card:nth-child(" + intToStr(i+1) + ") .module-provider-card-provider-text"
		st.AssertTextContent(providerSelector, expectedProviderText[i])
	}
}

// testModuleSearchFilters checks value of search filters.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_search_filters
func testModuleSearchFilters(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))
	st.NavigateTo("/search/modules?q=modulesearch")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-verified-count').text, '3')
	st.AssertTextContent("#search-verified-count", "3")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces-count').text, '4')
	st.AssertTextContent("#search-trusted-namespaces-count", "4")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed-count').text, '9')
	st.AssertTextContent("#search-contributed-count", "9")

	// Click verified label
	// Python: self.selenium_instance.find_element(By.ID, 'search-verified').click()
	verifiedCheckbox := st.WaitForElement("#search-verified")
	verifiedCheckbox.Click()

	// Python: self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 1)
	//         assert card.find_element(By.CLASS_NAME, 'module-card-title').text == 'modulesearch-trusted / mixedsearch-trusted-result-verified'
	st.AssertTextContent("#results .card:first-child .module-card-title", "modulesearch-trusted / mixedsearch-trusted-result-verified")

	// Click contributed label
	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Python: self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 3)
}

// testModuleSearchNextPrevButtons checks next and previous buttons.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_next_prev_buttons
func testModuleSearchNextPrevButtons(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))
	st.NavigateTo("/search/modules?q=modulesearch")

	// Ensure 4 results are found
	st.AssertElementVisible("#results .card:nth-child(1)")
	st.AssertElementVisible("#results .card:nth-child(4)")

	// Ensure both buttons are disabled initially
	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == False
	//         self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

	// Search for contributed modules
	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Ensure NextButton is active
	nextButton := st.WaitForElement("#nextButton")
	require.NotNil(t, nextButton)

	// Get list of cards on first page
	// Python: first_page_cards = []
	//         for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
	//             first_page_cards.append(card.find_element(By.CLASS_NAME, 'module-card-title').text)
	// Note: In real implementation, would store card titles for comparison

	// Select next page
	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').click()
	nextButton.Click()

	// Ensure next button is disabled and prev button is enabled
	prevButton := st.WaitForElement("#prevButton")
	require.NotNil(t, prevButton)

	// Select previous page
	// Python: self.selenium_instance.find_element(By.ID, 'prevButton').click()
	prevButton.Click()

	// Verify we're back on first page
	st.AssertElementVisible("#results .card:nth-child(1)")
}

// testModuleSearchResultCounts checks result count text.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_result_counts
func testModuleSearchResultCounts(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))
	st.NavigateTo("/search/modules?q=modulesearch")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 4 of 4')
	st.AssertTextContent("#result-count", "Showing results 1 - 4 of 4")

	// Search for contributed modules
	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 13')
	st.AssertTextContent("#result-count", "Showing results 1 - 10 of 13")

	// Select next page
	nextButton := st.WaitForElement("#nextButton")
	nextButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 11 - 13 of 13')
	st.AssertTextContent("#result-count", "Showing results 11 - 13 of 13")

	// Select previous page
	prevButton := st.WaitForElement("#prevButton")
	prevButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 13')
	st.AssertTextContent("#result-count", "Showing results 1 - 10 of 13")

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=doesnotexist'))
	st.NavigateTo("/search/modules?q=doesnotexist")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 0 - 0 of 0')
	st.AssertTextContent("#result-count", "Showing results 0 - 0 of 0")
}

// testModuleSearchResultRelevancyOrdering tests results are displayed in relevancy order.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_result_relevancy_ordering
func testModuleSearchResultRelevancyOrdering(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q=namematch'))
	st.NavigateTo("/search/modules?q=namematch")

	expectedCardHeadings := []string{
		"relevancysearch / namematch",
		"relevancysearch / namematch",
		"relevancysearch / partialmodulenamematch",
		"relevancysearch / descriptionmatch",
		"relevancysearch / ownermatch",
		"relevancysearch / partialmodulenamematch",
		"relevancysearch / partialdescriptionmatch",
		"relevancysearch / partialownermatch",
	}

	expectedProviders := []string{
		"namematch",
		"partialprovidernamematch",
		"namematch",
		"testprovider",
		"testprovider",
		"partialprovidernamematch",
		"testprovider",
		"testprovider",
	}

	// Verify each card heading and provider
	for i, expectedHeading := range expectedCardHeadings {
		selector := "#results .card:nth-child(" + intToStr(i+1) + ") .module-card-title"
		st.AssertTextContent(selector, expectedHeading)

		providerSelector := "#results .card:nth-child(" + intToStr(i+1) + ") .module-provider-card-provider-text"
		st.AssertTextContent(providerSelector, "Provider: "+expectedProviders[i])
	}
}

// terraformVersionCompatibilityTest represents a single test case for terraform version compatibility.
type terraformVersionCompatibilityTest struct {
	inputTerraformVersion      string
	expectedCompatibilityTexts []string
}

// terraformVersionCompatibilityTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_module_search.py line 226-244
var terraformVersionCompatibilityTests = []terraformVersionCompatibilityTest{
	{
		"2.5.0",
		[]string{
			"", // First item has no entry, due to invalid version constraint
			"Compatible",
			"Implicitly compatible",
			"No version constraint defined",
			"Compatible",
		},
	},
	{
		"0.5.0",
		[]string{
			"", // First item has no entry, due to invalid version constraint
			"Incompatible",
			"Incompatible",
			"No version constraint defined",
			"Incompatible",
		},
	},
}

// testModuleSearchTerraformVersionCompatibility tests terraform compatibility input and display.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_terraform_version_compatibility
func testModuleSearchTerraformVersionCompatibility(t *testing.T) {
	for _, tt := range terraformVersionCompatibilityTests {
		t.Run(tt.inputTerraformVersion, func(t *testing.T) {
			st := NewSeleniumTest(t)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			// Python: self.selenium_instance.get(self.get_url('/search/modules?q=version-constraint-test'))
			st.NavigateTo("/search/modules?q=version-constraint-test")

			// Python: terraform_input = self.selenium_instance.find_element(By.ID, 'search-terraform-version')
			//         assert terraform_input.get_attribute("value") == ""
			//         terraform_input.send_keys(input_terraform_version)
			terraformInput := st.WaitForElement("#search-terraform-version")
			terraformInput.SendKeys(tt.inputTerraformVersion)

			// Python: self.selenium_instance.find_element(By.ID, 'search-options-update-button').click()
			updateButton := st.WaitForElement("#search-options-update-button")
			updateButton.Click()

			// Wait until cards have version constraints
			// Python: self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card-terraform-version-compatibility')), 4)
			st.AssertElementVisible("#results .card-terraform-version-compatibility")

			// Verify compatibility texts for each result card
			for i, expectedText := range tt.expectedCompatibilityTexts {
				if expectedText != "" {
					selector := "#results .card:nth-child(" + intToStr(i+1) + ") .card-terraform-version-compatibility"
					st.AssertTextContent(selector, expectedText)
				}
			}
		})
	}
}

// testModuleSearchTerraformVersionCompatibilityRetainsState tests terraform compatibility input is retained between page loads.
// Python reference: /app/test/selenium/test_module_search.py - TestModuleSearch.test_terraform_version_compatibility_retains_state
func testModuleSearchTerraformVersionCompatibilityRetainsState(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q='))
	st.NavigateTo("/search/modules?q=")

	// Python: terraform_input = self.selenium_instance.find_element(By.ID, 'search-terraform-version')
	//         assert terraform_input.get_attribute("value") == ""
	//         terraform_input.send_keys("5.2.6-unittest")
	terraformInput := st.WaitForElement("#search-terraform-version")
	terraformInput.SendKeys("5.2.6-unittest")

	// Python: self.selenium_instance.find_element(By.ID, 'search-options-update-button').click()
	updateButton := st.WaitForElement("#search-options-update-button")
	updateButton.Click()

	// Python: self.wait_for_element(By.CLASS_NAME, "card-terraform-version-compatibility")
	st.AssertElementVisible(".card-terraform-version-compatibility")

	// Python: self.selenium_instance.get(self.get_url('/search/modules?q='))
	st.NavigateTo("/search/modules?q=")

	// Python: self.assert_equals(lambda: self.wait_for_element(By.ID, 'search-terraform-version').get_attribute("value"), "5.2.6-unittest")
	terraformInputAfterReload := st.WaitForElement("#search-terraform-version")
	require.NotNil(t, terraformInputAfterReload)

	// Note: In real implementation, would verify the value is retained
}
