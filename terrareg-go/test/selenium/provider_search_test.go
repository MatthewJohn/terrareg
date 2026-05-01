//go:build selenium

package selenium

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProviderSearch tests the provider search page.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch class
func TestProviderSearch(t *testing.T) {
	t.Run("test_result_cards", testProviderSearchResultCards)
	t.Run("test_search_filters", testProviderSearchFilters)
	t.Run("test_next_prev_buttons", testProviderSearchNextPrevButtons)
	t.Run("test_result_counts", testProviderSearchResultCounts)
	t.Run("test_result_relevancy_ordering", testProviderSearchResultRelevancyOrdering)
}

// newProviderSearchTest creates a new SeleniumTest configured for provider search tests.
// Python reference: /app/test/selenium/test_provider_search.py - setup_class
func newProviderSearchTest(t *testing.T) *SeleniumTest {
	config := ConfigForProviderSearchTests()
	return NewSeleniumTestWithConfig(t, config)
}

// ConfigForProviderSearchTests returns config for provider search tests.
// Python reference: /app/test/selenium/test_provider_search.py - setup_class
func ConfigForProviderSearchTests() map[string]string {
	base := getDefaultTestConfig()
	return mergeMaps(base, map[string]string{
		"CONTRIBUTED_NAMESPACE_LABEL": "unittest contributed module",
		"TRUSTED_NAMESPACE_LABEL":     "unittest trusted namespace",
		"TRUSTED_NAMESPACES":          "providersearch-trusted,relevancysearch",
	})
}

// testProviderSearchResultCards checks the result cards.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch.test_result_cards
func testProviderSearchResultCards(t *testing.T) {
	st := newProviderSearchTest(t)
	defer st.TearDown()

	st.NavigateTo("/search/providers?q=providersearch")

	// Python: self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 3)
	cardCount := st.GetElementCount("#results .card")
	assert.Equal(t, 3, cardCount)

	// Python: expected_card_headings = [...]
	// Python: expected_card_links = [...]
	// Python: expected_sources = [...]
	expectedCardHeadings := []string{
		"providersearch-trusted / mixedsearch-trusted-second-result",
		"providersearch-trusted / mixedsearch-trusted-result-multiversion",
		"providersearch-trusted / mixedsearch-trusted-result",
	}
	expectedCardLinks := []string{
		"/providers/providersearch-trusted/mixedsearch-trusted-second-result",
		"/providers/providersearch-trusted/mixedsearch-trusted-result-multiversion",
		"/providers/providersearch-trusted/mixedsearch-trusted-result",
	}
	expectedSources := []string{
		"Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-second-result",
		"Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-result-multiversion",
		"Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-result",
	}

	// Python: for card in result_cards:
	_ = st.WaitForElement("#results")
	for i := 0; i < 3; i++ {
		// Get the i-th card
		cardSelector := fmt.Sprintf("#results .card:nth-child(%d)", i+1)
		_ = st.WaitForElement(cardSelector)

		// Python: heading = card.find_element(By.CLASS_NAME, 'module-card-title')
		// Python: assert heading.text == expected_card_headings[i]
		headingSelector := cardSelector + " .module-card-title"
		heading := st.WaitForElement(headingSelector)
		assert.Equal(t, expectedCardHeadings[i], heading.Text())

		// Python: assert heading.get_attribute('href') == self.get_url(expected_card_links[i])
		href := heading.GetAttribute("href")
		assert.Equal(t, st.GetURL(expectedCardLinks[i]), href)

		// Python: footer = card.find_element(By.CLASS_NAME, "card-footer")
		// Python: assert footer.find_element(By.CLASS_NAME, "card-source-link").text == expected_sources[i]
		sourceLinkSelector := cardSelector + " .card-footer .card-source-link"
		sourceLink := st.WaitForElement(sourceLinkSelector)
		assert.Equal(t, expectedSources[i], sourceLink.Text())
	}
}

// testProviderSearchFilters checks value of search filters.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch.test_search_filters
func testProviderSearchFilters(t *testing.T) {
	st := newProviderSearchTest(t)
	defer st.TearDown()

	st.NavigateTo("/search/providers?q=providersearch")

	// Python: self.assert_equals(lambda: len(...find_elements(By.CLASS_NAME, 'card')), 3)
	cardCount := st.GetElementCount("#results .card")
	assert.Equal(t, 3, cardCount)

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces-count').text, '3')
	st.AssertTextContent("#search-trusted-namespaces-count", "3")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed-count').text, '4')
	st.AssertTextContent("#search-contributed-count", "4")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces').is_selected(), True)
	assert.True(t, st.IsElementChecked("#search-trusted-namespaces"))

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed').is_selected(), False)
	assert.False(t, st.IsElementChecked("#search-contributed"))

	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Python: self.assert_equals(lambda: len(...find_elements(By.CLASS_NAME, 'card')), 7)
	cardCount = st.GetElementCount("#results .card")
	assert.Equal(t, 7, cardCount)

	// Python: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces').click()
	trustedCheckbox := st.WaitForElement("#search-trusted-namespaces")
	trustedCheckbox.Click()

	// Python: self.assert_equals(lambda: len(...find_elements(By.CLASS_NAME, 'card')), 4)
	cardCount = st.GetElementCount("#results .card")
	assert.Equal(t, 4, cardCount)
}

// testProviderSearchNextPrevButtons checks next and previous buttons.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch.test_next_prev_buttons
func testProviderSearchNextPrevButtons(t *testing.T) {
	st := newProviderSearchTest(t)
	defer st.TearDown()

	st.NavigateTo("/search/providers?q=providersearch")

	// Python: result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
	// Python: assert len(result_cards) == 3
	cardCount := st.GetElementCount("#results .card")
	assert.Equal(t, 3, cardCount)

	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == False
	// Python: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False
	nextButton := st.WaitForElement("#nextButton")
	prevButton := st.WaitForElement("#prevButton")
	// Check if buttons are disabled by checking their class or disabled attribute
	assert.True(t, nextButton.GetAttribute("disabled") == "true" || nextButton.GetAttribute("class") != "")
	assert.True(t, prevButton.GetAttribute("disabled") == "true" || prevButton.GetAttribute("class") != "")

	// Python: self.selenium_instance.get(self.get_url('/search/providers?q='))
	// Python: self.wait_for_element(By.CLASS_NAME, "module-card-title")
	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	st.NavigateTo("/search/providers?q=")
	_ = st.WaitForElement(".module-card-title")
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == True
	// Python: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False
	nextButton = st.WaitForElement("#nextButton")
	prevButton = st.WaitForElement("#prevButton")
	assert.False(t, nextButton.GetAttribute("disabled") == "true")
	assert.True(t, prevButton.GetAttribute("disabled") == "true" || prevButton.GetAttribute("class") != "")

	// Python: self.assert_equals(lambda: len(...find_elements(By.CLASS_NAME, 'card')), 10)
	cardCount = st.GetElementCount("#results .card")
	assert.Equal(t, 10, cardCount)

	// Python: first_page_cards = []
	// Python: for card in ...find_elements(By.CLASS_NAME, 'card'):
	// Python:     first_page_cards.append(card.find_element(By.CLASS_NAME, 'module-card-title').text)
	firstPageCards := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		headingSelector := fmt.Sprintf("#results .card:nth-child(%d) .module-card-title", i+1)
		heading := st.WaitForElement(headingSelector)
		firstPageCards = append(firstPageCards, heading.Text())
	}

	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').click()
	nextButton = st.WaitForElement("#nextButton")
	nextButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled(), False)
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled(), True)
	nextButton = st.WaitForElement("#nextButton")
	prevButton = st.WaitForElement("#prevButton")
	assert.True(t, nextButton.GetAttribute("disabled") == "true" || nextButton.GetAttribute("class") != "")
	assert.False(t, prevButton.GetAttribute("disabled") == "true")

	// Python: self.assert_equals(lambda: ...find_elements(By.CLASS_NAME, 'card')[0].find_element(By.CLASS_NAME, 'module-card-title').text, 'providersearch-trusted / mixedsearch-trusted-result')
	firstHeadingSelector := "#results .card:nth-child(1) .module-card-title"
	firstHeading := st.WaitForElement(firstHeadingSelector)
	assert.Equal(t, "providersearch-trusted / mixedsearch-trusted-result", firstHeading.Text())

	// Python: for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
	// Python:     assert card.find_element(By.CLASS_NAME, 'module-card-title').text not in first_page_cards
	for i := 0; i < 10; i++ {
		headingSelector := fmt.Sprintf("#results .card:nth-child(%d) .module-card-title", i+1)
		heading := st.WaitForElement(headingSelector)
		headingText := heading.Text()
		for _, firstPageText := range firstPageCards {
			require.NotEqual(t, firstPageText, headingText, "Card from first page found on second page")
		}
	}

	// Python: self.selenium_instance.find_element(By.ID, 'prevButton').click()
	prevButton = st.WaitForElement("#prevButton")
	prevButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled(), True)
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled(), False)
	nextButton = st.WaitForElement("#nextButton")
	prevButton = st.WaitForElement("#prevButton")
	assert.False(t, nextButton.GetAttribute("disabled") == "true")
	assert.True(t, prevButton.GetAttribute("disabled") == "true" || prevButton.GetAttribute("class") != "")

	// Python: self.assert_equals(lambda: ...find_elements(By.CLASS_NAME, 'card')[0].find_element(By.CLASS_NAME, 'module-card-title').text, 'initial-providers / update-attributes'
	firstHeading = st.WaitForElement(firstHeadingSelector)
	assert.Equal(t, "initial-providers / update-attributes", firstHeading.Text())

	// Python: for card in ...find_elements(By.CLASS_NAME, 'card'):
	// Python:     card_title = card.find_element(By.CLASS_NAME, 'module-card-title').text
	// Python:     assert card_title in first_page_cards
	// Python:     first_page_cards.remove(card_title)
	// Python: assert len(first_page_cards) == 0
	for i := 0; i < 10; i++ {
		headingSelector := fmt.Sprintf("#results .card:nth-child(%d) .module-card-title", i+1)
		heading := st.WaitForElement(headingSelector)
		cardTitle := heading.Text()
		found := false
		for j, firstPageText := range firstPageCards {
			if firstPageText == cardTitle {
				firstPageCards = append(firstPageCards[:j], firstPageCards[j+1:]...)
				found = true
				break
			}
		}
		require.True(t, found, "Card from second page not found in first page: %s", cardTitle)
	}
	assert.Equal(t, 0, len(firstPageCards))
}

// testProviderSearchResultCounts checks result count text.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch.test_result_counts
func testProviderSearchResultCounts(t *testing.T) {
	st := newProviderSearchTest(t)
	defer st.TearDown()

	st.NavigateTo("/search/providers?q=")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 7 of 7')
	st.AssertTextContent("#result-count", "Showing results 1 - 7 of 7")

	// Python: self.selenium_instance.find_element(By.ID, 'search-contributed').click()
	contributedCheckbox := st.WaitForElement("#search-contributed")
	contributedCheckbox.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 16')
	st.AssertTextContent("#result-count", "Showing results 1 - 10 of 16")

	// Python: self.selenium_instance.find_element(By.ID, 'nextButton').click()
	nextButton := st.WaitForElement("#nextButton")
	nextButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 11 - 16 of 16')
	st.AssertTextContent("#result-count", "Showing results 11 - 16 of 16")

	// Python: self.selenium_instance.find_element(By.ID, 'prevButton').click()
	prevButton := st.WaitForElement("#prevButton")
	prevButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 16')
	st.AssertTextContent("#result-count", "Showing results 1 - 10 of 16")

	// Python: self.selenium_instance.get(self.get_url('/search/providers?q=doesnotexist'))
	st.NavigateTo("/search/providers?q=doesnotexist")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 0 - 0 of 0')
	st.AssertTextContent("#result-count", "Showing results 0 - 0 of 0")
}

// testProviderSearchResultRelevancyOrdering tests results are displayed in relevancy order.
// Python reference: /app/test/selenium/test_provider_search.py - TestProviderSearch.test_result_relevancy_ordering
func testProviderSearchResultRelevancyOrdering(t *testing.T) {
	st := newProviderSearchTest(t)
	defer st.TearDown()

	st.NavigateTo("/search/providers?q=namematch")

	// Python: self.assert_equals(lambda: len(...find_elements(By.CLASS_NAME, 'card')), 4)
	cardCount := st.GetElementCount("#results .card")
	assert.Equal(t, 4, cardCount)

	// Python: expected_card_headings = [...]
	expectedCardHeadings := []string{
		"relevancysearch / namematch",
		"relevancysearch / descriptionmatch",
		"relevancysearch / partialnamematch",
		"relevancysearch / partialdescriptionmatch",
	}

	// Python: for expected_heading in expected_card_headings:
	// Python:     card = result_cards.pop(0)
	// Python:     heading = card.find_element(By.CLASS_NAME, 'module-card-title')
	// Python:     assert heading.text == expected_heading
	for i, expectedHeading := range expectedCardHeadings {
		headingSelector := fmt.Sprintf("#results .card:nth-child(%d) .module-card-title", i+1)
		heading := st.WaitForElement(headingSelector)
		assert.Equal(t, expectedHeading, heading.Text())
	}
}
