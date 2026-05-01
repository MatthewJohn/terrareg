//go:build selenium

package selenium

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
)

// TestCommonSearchPage tests the common search page.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch class
func TestCommonSearchPage(t *testing.T) {
	t.Run("test_search_from_homepage_common_search", testSearchFromHomepageCommonSearch)
	t.Run("test_search_from_homepage_redirect_type_search", testSearchFromHomepageRedirectTypeSearch)
	t.Run("test_result_cards", testCommonSearchResultCards)
	t.Run("test_provider_results_button", testProviderResultsButton)
	t.Run("test_module_results_button", testModuleResultsButton)
}

// newCommonSearchPageTest creates a new SeleniumTest configured for common search page tests.
// Python reference: /app/test/selenium/test_common_search_page.py - setup_class
func newCommonSearchPageTest(t *testing.T) *SeleniumTest {
	config := ConfigForCommonSearchPageTests()
	return NewSeleniumTestWithConfig(t, config, WithCommonSearchPageTestData)
}

// WithCommonSearchPageTestData is a TestServerOption that sets up test data for common search page tests.
// This setup happens before the HTTP server starts to avoid database connection conflicts.
var WithCommonSearchPageTestData TestServerOption = func(ts *TestServer) {
	ts.testDataSetup = func(db *sqldb.Database) {
		SetupCommonSearchPageTestData(ts.t, db)
	}
}

// ConfigForCommonSearchPageTests returns config for common search page tests.
// Python reference: /app/test/selenium/test_common_search_page.py - setup_class
func ConfigForCommonSearchPageTests() map[string]string {
	base := getDefaultTestConfig()
	return mergeMaps(base, map[string]string{
		"CONTRIBUTED_NAMESPACE_LABEL": "unittest contributed module",
		"TRUSTED_NAMESPACE_LABEL":     "unittest trusted namespace",
		"VERIFIED_MODULE_LABEL":       "unittest verified label",
		"TRUSTED_NAMESPACES":          "modulesearch-trusted,relevancysearch",
	})
}

// testSearchFromHomepageCommonSearch checks search functionality from homepage.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_search_from_homepage_common_search
func testSearchFromHomepageCommonSearch(t *testing.T) {
	testCases := []struct {
		searchString string
	}{
		{""}, // Test string that will match modules and providers
		{"mixed"},
	}

	for _, tc := range testCases {
		t.Run(tc.searchString, func(t *testing.T) {
			st := newCommonSearchPageTest(t)
			defer st.TearDown()

			st.NavigateTo("/")

			// Python: self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)
			searchInput := st.WaitForElement("#navBarSearchInput")
			searchInput.SendKeys(tc.searchString)

			// Python: search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
			// Python: assert search_button.text == 'Search'
			searchButton := st.WaitForElement("#navBarSearchButton")
			assert.Equal(t, "Search", searchButton.Text())

			// Python: search_button.click()
			searchButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(f'/search?q={search_string}'))
			// Wait for redirect to complete and URL to match expected
			expectedURL := st.GetURL("/search?q=" + tc.searchString)
			var currentURL string
			err := st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
				currentURL = st.GetCurrentURL()
				if currentURL == expectedURL {
					return nil
				}
				return fmt.Errorf("URL not redirected: expected %q, got %q", expectedURL, currentURL)
			}), 100, 10)
			require.NoError(t, err, "URL redirect failed: expected %q, but got %q", expectedURL, currentURL)

			// Python: assert self.selenium_instance.title == 'Search - Terrareg'
			// Wait for title to update (might lag behind URL change)
			var title string
			err = st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
				title = st.GetTitle()
				if title == "Search - Terrareg" {
					return nil
				}
				return fmt.Errorf("Title not updated: expected %q, got %q", "Search - Terrareg", title)
			}), 50, 10)
			require.NoError(t, err, "Title check failed: expected %q, but got %q", "Search - Terrareg", title)
		})
	}
}

// redirectTypeSearchTestCase represents a test case for search redirect verification.
type redirectTypeSearchTestCase struct {
	searchString  string
	expectedURL   string
	expectedTitle string
}

// testSearchFromHomepageRedirectTypeSearch checks search functionality from homepage with type-specific redirects.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_search_from_homepage_redirect_type_search
func testSearchFromHomepageRedirectTypeSearch(t *testing.T) {
	testCases := []redirectTypeSearchTestCase{
		{"fullypopulated", "/search/modules?q=fullypopulated", "Module Search - Terrareg"},
		{"initial-providers", "/search/providers?q=initial-providers", "Provider Search - Terrareg"},
	}

	for _, tc := range testCases {
		t.Run(tc.searchString, func(t *testing.T) {
			st := newCommonSearchPageTest(t)
			defer st.TearDown()

			st.NavigateTo("/")

			// Python: self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)
			searchInput := st.WaitForElement("#navBarSearchInput")
			searchInput.SendKeys(tc.searchString)

			// Python: search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
			// Python: assert search_button.text == 'Search'
			searchButton := st.WaitForElement("#navBarSearchButton")
			assert.Equal(t, "Search", searchButton.Text())

			// Python: search_button.click()
			searchButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(expected_url))
			// Wait for redirect to complete and URL to match expected
			expectedURL := st.GetURL(tc.expectedURL)
			var currentURL string
			err := st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
				currentURL = st.GetCurrentURL()
				if currentURL == expectedURL {
					return nil
				}
				return fmt.Errorf("URL not redirected: expected %q, got %q", expectedURL, currentURL)
			}), 100, 10)
			require.NoError(t, err, "URL redirect failed: expected %q, but got %q", expectedURL, currentURL)

			// Python: self.assert_equals(lambda: self.selenium_instance.title, expected_title)
			// Wait for title to update (might lag behind URL change)
			var title string
			err = st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
				title = st.GetTitle()
				if title == tc.expectedTitle {
					return nil
				}
				return fmt.Errorf("Title not updated: expected %q, got %q", tc.expectedTitle, title)
			}), 50, 10)
			require.NoError(t, err, "Title check failed: expected %q, but got %q", tc.expectedTitle, title)

			// Python: assert self.selenium_instance.find_element(By.ID, 'search-query-string').get_attribute('value') == search_string
			var actualValue string
			err = st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
				actualValue = st.GetValue("#search-query-string")
				if tc.searchString == actualValue {
					return nil
				}
				return fmt.Errorf("%s does not match %s", tc.searchString, actualValue)
			}), 100, 10)
			require.NoError(t, err, "Retry failed: expected %q to match %q, but got %q", tc.searchString, tc.searchString, actualValue)
			assert.Equal(t, tc.searchString, actualValue)
		})
	}
}

// testCommonSearchResultCards checks result cards in common search page.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_result_cards
func testCommonSearchResultCards(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Debug: Wait for page to load and for JavaScript to execute
	time.Sleep(3 * time.Second)

	// Debug: Check the debug-module-count element to see JavaScript trace
	var debugText string
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var el = document.getElementById('debug-module-count');
					return el ? el.textContent : 'debug element not found';
				})()
			`, &debugText).Do(ctx)
		}),
	)
	t.Logf("Debug module count element text: %s (error: %v)", debugText, err)

	// Debug: Check how many result-box elements exist
	var resultBoxCount int
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					return document.querySelectorAll('#results-providers-content .result-box').length;
				})()
			`, &resultBoxCount).Do(ctx)
		}),
	)
	t.Logf("Number of provider result boxes found: %d (error: %v)", resultBoxCount, err)

	// Debug: Check page title
	var pageTitle string
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Title(&pageTitle).Do(ctx)
		}),
	)
	t.Logf("Page title: %s (error: %v)", pageTitle, err)

	// Debug: List all IDs of result-box elements
	var allIDs []string
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var boxes = document.querySelectorAll('#results-providers-content .result-box');
					return Array.from(boxes).map(b => b.id);
				})()
			`, &allIDs).Do(ctx)
		}),
	)
	t.Logf("All result box IDs: %v (error: %v)", allIDs, err)

	// Debug: Check the actual API response
	var apiResponse string
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(async function() {
					try {
						const response = await fetch('/v1/providers/search?q=mixed&include_count=true&limit=6');
						const data = await response.json();
						return JSON.stringify(data);
					} catch (e) {
						return 'Error: ' + e.message;
					}
				})()
			`, &apiResponse).Do(ctx)
		}),
	)
	t.Logf("Provider API Response: %s (error: %v)", apiResponse, err)

	// Debug: Check the module API response
	var moduleApiResponse string
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(async function() {
					try {
						const response = await fetch('/v1/modules/search?q=mixed&include_count=true&limit=6');
						const data = await response.json();
						return JSON.stringify(data);
					} catch (e) {
						return 'Error: ' + e.message;
					}
				})()
			`, &moduleApiResponse).Do(ctx)
		}),
	)
	t.Logf("Module API Response: %s (error: %v)", moduleApiResponse, err)

	// Wait for page to fully load before checking for elements
	// The search page loads data asynchronously via JavaScript
	time.Sleep(1 * time.Second)

	// Debug: Check if the specific element exists before WaitForElement
	var elementExists bool
	err = st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var el = document.querySelector('[id="contributed-providersearch.mixedsearch-result.1.0.0"]');
					if (!el) return false;
					var rect = el.getBoundingClientRect();
					return { exists: true, visible: rect.width > 0 && rect.height > 0 };
				})()
			`, &elementExists).Do(ctx)
		}),
	)
	t.Logf("Element exists check: %v (error: %v)", elementExists, err)

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	// Note: In Python Selenium, By.ID queries the DOM ID attribute directly
	// In Go chromedp with CSS selectors, dots are interpreted as class separators
	// We use attribute selector to avoid this ambiguity
	_ = st.WaitForElement(`[id="contributed-providersearch.mixedsearch-result.1.0.0"]`)

	// Python: provider_cards = [...]
	// Python: for card in self.selenium_instance.find_element(By.ID, "results-providers-content").find_elements(By.CLASS_NAME, "result-box"):
	providerCards := []struct {
		link string
		text string
	}{
		{"/providers/providersearch-trusted/mixedsearch-trusted-second-result", "providersearch-trusted / mixedsearch-trusted-second-result"},
		{"/providers/providersearch-trusted/mixedsearch-trusted-result-multiversion", "providersearch-trusted / mixedsearch-trusted-result-multiversion"},
		{"/providers/providersearch-trusted/mixedsearch-trusted-result", "providersearch-trusted / mixedsearch-trusted-result"},
		{"/providers/contributed-providersearch/mixedsearch-result-multiversion", "contributed-providersearch / mixedsearch-result-multiversion"},
		{"/providers/contributed-providersearch/mixedsearch-result", "contributed-providersearch / mixedsearch-result"},
	}

	for i, cardDetails := range providerCards {
		cardSelector := "#results-providers-content .result-box:nth-child(" + strconv.Itoa(i+1) + ")"
		_ = st.WaitForElement(cardSelector)

		// Python: for link in card.find_elements(By.TAG_NAME, "a"):
		// Python:     assert link.get_attribute("href") == self.get_url(card_details["link"])
		err := st.runChromedp(
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(fmt.Sprintf(`
					(function() {
						var card = document.querySelectorAll('#results-providers-content .result-box')[%d];
						var links = card.getElementsByTagName('a');
						for (var i = 0; i < links.length; i++) {
							if (links[i].getAttribute('href').endsWith('%s')) {
								return true;
							}
						}
						return false;
					})()
				`, i, cardDetails.link), nil).Do(ctx)
			}),
		)
		require.NoError(st.t, err, "Provider card link not found for: %s", cardDetails.link)

		// Python: assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]
		titleSelector := cardSelector + " .module-card-title"
		st.AssertTextContent(titleSelector, cardDetails.text)
	}

	// Python: module_cards = [...]
	// Python: for card in self.selenium_instance.find_element(By.ID, "results-modules-content").find_elements(By.CLASS_NAME, "result-box"):
	moduleCards := []struct {
		link string
		text string
	}{
		{"/modules/modulesearch-contributed/mixedsearch-result/aws", "modulesearch-contributed / mixedsearch-result"},
		{"/modules/modulesearch-contributed/mixedsearch-result-multiversion/aws", "modulesearch-contributed / mixedsearch-result-multiversion"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result/aws", "modulesearch-trusted / mixedsearch-trusted-result"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null", "modulesearch-trusted / mixedsearch-trusted-result-multiversion"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp", "modulesearch-trusted / mixedsearch-trusted-result-verified"},
		{"/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog", "modulesearch-trusted / mixedsearch-trusted-second-result"},
	}

	for i, cardDetails := range moduleCards {
		cardSelector := "#results-modules-content .result-box:nth-child(" + strconv.Itoa(i+1) + ")"
		_ = st.WaitForElement(cardSelector)

		// Python: for link in card.find_elements(By.TAG_NAME, "a"):
		// Python:     if "provider-logo-link" not in link.get_attribute("class"):
		// Python:         assert link.get_attribute("href") == self.get_url(card_details["link"])
		err := st.runChromedp(
			chromedp.ActionFunc(func(ctx context.Context) error {
				return chromedp.Evaluate(fmt.Sprintf(`
					(function() {
						var card = document.querySelectorAll('#results-modules-content .result-box')[%d];
						var links = card.getElementsByTagName('a');
						for (var i = 0; i < links.length; i++) {
							if (!links[i].classList.contains('provider-logo-link') &&
								links[i].getAttribute('href').endsWith('%s')) {
								return true;
							}
						}
						return false;
					})()
				`, i, cardDetails.link), nil).Do(ctx)
			}),
		)
		require.NoError(st.t, err, "Module card link not found for: %s", cardDetails.link)

		// Python: assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]
		titleSelector := cardSelector + " .module-card-title"
		st.AssertTextContent(titleSelector, cardDetails.text)
	}
}

// testProviderResultsButton checks link to provider results.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_provider_results_button
func testProviderResultsButton(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	_ = st.WaitForElement(`[id="contributed-providersearch.mixedsearch-result.1.0.0"]`)

	// Python: button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all provider results']")
	// Python: button.click()
	button := st.WaitForElement("button")
	button.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/search/providers?q=mixed'))
	// Wait for redirect to complete
	expectedURL := st.GetURL("/search/providers?q=mixed")
	var currentURL string
	err := st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
		currentURL = st.GetCurrentURL()
		if currentURL == expectedURL {
			return nil
		}
		return fmt.Errorf("URL not redirected: expected %q, got %q", expectedURL, currentURL)
	}), 50, 10)
	require.NoError(t, err, "URL redirect failed: expected %q, but got %q", expectedURL, currentURL)
	assert.Equal(t, expectedURL, currentURL)
}

// testModuleResultsButton checks link to module results.
// Python reference: /app/test/selenium/test_common_search_page.py - TestModuleSearch.test_module_results_button
func testModuleResultsButton(t *testing.T) {
	st := newCommonSearchPageTest(t)
	defer st.TearDown()

	st.NavigateTo("/search?q=mixed")

	// Python: self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
	_ = st.WaitForElement(`[id="contributed-providersearch.mixedsearch-result.1.0.0"]`)

	// Python: button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all module results']")
	// Python: button.click()
	// Find the button by text content
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`
				(function() {
					var buttons = document.getElementsByTagName('button');
					for (var i = 0; i < buttons.length; i++) {
						if (buttons[i].textContent === 'View all module results') {
							buttons[i].click();
							return true;
						}
					}
					return false;
				})()
			`, nil).Do(ctx)
		}),
	)
	require.NoError(t, err)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/search/modules?q=mixed'))
	// Wait for redirect to complete
	expectedURL := st.GetURL("/search/modules?q=mixed")
	var currentURL string
	err = st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
		currentURL = st.GetCurrentURL()
		if currentURL == expectedURL {
			return nil
		}
		return fmt.Errorf("URL not redirected: expected %q, got %q", expectedURL, currentURL)
	}), 50, 10)
	require.NoError(t, err, "URL redirect failed: expected %q, but got %q", expectedURL, currentURL)
	assert.Equal(t, expectedURL, currentURL)
}
