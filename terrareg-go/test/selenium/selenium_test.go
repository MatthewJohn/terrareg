//go:build selenium

package selenium

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultTimeout = 5 * time.Second
	pollInterval   = 100 * time.Millisecond
)

// SeleniumTest provides base functionality for Selenium tests.
// This is the Go equivalent of Python's test.selenium.SeleniumTest class.
//
// Python reference: /app/test/selenium/__init__.py - SeleniumTest class
type SeleniumTest struct {
	t           *testing.T
	server      *TestServer
	baseURL     string
	AllocCtx    context.Context // Exported chromedp allocator context
	allocCancel context.CancelFunc
	ctxCancel   context.CancelFunc // Context cancel function for chromedp context
}

// NewSeleniumTest creates a new Selenium test instance.
// This is the Go equivalent of Python's setup_class method.
func NewSeleniumTest(t *testing.T) *SeleniumTest {
	st := &SeleniumTest{
		t: t,
	}

	st.setupServer()
	st.setupBrowser()

	return st
}

// NewSeleniumTestWithConfig creates a new Selenium test instance with custom config.
// This allows individual test classes to override configuration like Python's setup_class.
// Python reference: /app/test/selenium/test_homepage.py - TestHomepage.setup_class
func NewSeleniumTestWithConfig(t *testing.T, configOverrides map[string]string, opts ...TestServerOption) *SeleniumTest {
	st := &SeleniumTest{
		t: t,
	}

	st.server = NewTestServer(st.t, configOverrides, opts...)
	st.baseURL = st.server.baseURL
	st.setupBrowser()

	return st
}

// setupServer starts a test Terrareg server with the actual application.
// This is the Go equivalent of Python's _setup_server method.
// Python reference: /app/test/selenium/__init__.py - SeleniumTest._setup_server()
func (st *SeleniumTest) setupServer() {
	// Create test server with default configuration
	// Individual tests can override this by calling NewTestServer directly
	configOverrides := ConfigForAdminTokenTests()
	st.server = NewTestServer(st.t, configOverrides)
	st.baseURL = st.server.baseURL
}

// setupBrowser initializes the Chrome browser for testing.
// This is the Go equivalent of Python's Selenium/Firefox setup.
func (st *SeleniumTest) setupBrowser() {
	// Create Chrome DP allocator options
	// Use headless mode unless RUN_INTERACTIVELY is set
	opts := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", os.Getenv("RUN_INTERACTIVELY") == ""),
		chromedp.Flag("disable-gpu", "true"),
		chromedp.Flag("no-sandbox", "true"),
		chromedp.Flag("disable-dev-shm-usage", "true"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	// Create the allocator context
	allocatorCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	st.allocCancel = allocCancel

	// Create the browser context from the allocator context
	// This is the standard chromedp pattern - the allocator is inherited from the parent
	ctx, cancel := chromedp.NewContext(allocatorCtx, chromedp.WithLogf(log.Printf))
	st.AllocCtx = ctx
	chromedpCancel := cancel // Save the chromedp cancel function

	// Set a 60-second timeout for all chromedp operations
	// This must be longer than the longest individual wait (e.g., WaitForURL uses 30s)
	ctx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	st.AllocCtx = ctx
	st.ctxCancel = func() {
		timeoutCancel()  // Cancel the timeout context
		chromedpCancel() // Cancel the chromedp context
	}

	// Allocate the browser by running an initial task
	// The executor will be embedded in the context after this call
	// Navigate to about:blank to initialize the browser
	if err := chromedp.Run(st.AllocCtx, chromedp.Navigate("about:blank")); err != nil {
		st.t.Fatalf("Failed to start browser: %v", err)
	}
}

// TearDown cleans up the Selenium test resources.
// This is the Go equivalent of Python's teardown_class method.
func (st *SeleniumTest) TearDown() {
	// Close browser context
	if st.ctxCancel != nil {
		st.ctxCancel()
	}

	// Close allocator
	if st.allocCancel != nil {
		st.allocCancel()
	}

	// Shutdown server
	if st.server != nil {
		st.server.Shutdown()
	}
}

// GetURL returns the full URL for a given path.
// This is the Go equivalent of Python's get_url method.
// Python reference: /app/test/selenium/__init__.py - get_url()
func (st *SeleniumTest) GetURL(path string) string {
	return st.baseURL + path
}

// runChromedp is a helper method to run chromedp actions with the proper context.
// This ensures that all chromedp operations have access to the browser executor.
func (st *SeleniumTest) runChromedp(actions ...chromedp.Action) error {
	return chromedp.Run(st.AllocCtx, actions...)
}

// NavigateTo navigates the browser to a specific path.
func (st *SeleniumTest) NavigateTo(path string) {
	url := st.GetURL(path)
	err := st.Retry(chromedp.ActionFunc(func(ctx context.Context) error {
		return st.runChromedp(chromedp.Navigate(url))
	}), 50, 3)
	require.NoError(st.t, err, "Failed to navigate to %s", url)
}

// NavigateToURL navigates the browser to a full URL (not a path).
// Use this when you have a complete URL that should not have the base URL prepended.
func (st *SeleniumTest) NavigateToURL(fullURL string) {
	err := st.runChromedp(chromedp.Navigate(fullURL))
	require.NoError(st.t, err, "Failed to navigate to %s", fullURL)
}

// WaitForElement waits for an element to be present and optionally visible.
// This is the Go equivalent of Python's wait_for_element method.
// Python reference: /app/test/selenium/__init__.py - wait_for_element()
func (st *SeleniumTest) WaitForElement(selector string, opts ...ElementOption) *Element {
	opt := &elementOptions{
		timeout:         defaultTimeout,
		ensureDisplayed: true,
	}
	for _, o := range opts {
		o(opt)
	}

	// Detect XPath selector (starts with //)
	isXPath := strings.HasPrefix(selector, "//")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), opt.timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			require.Fail(st.t, "Element not found within timeout: %s", selector)
		case <-ticker.C:
			var found bool
			var visible bool

			// Use runChromedp to properly access the browser
			err := st.runChromedp(
				chromedp.ActionFunc(func(ctx context.Context) error {
					// Check if element exists
					var textContent string
					queryOpts := chromedp.ByQuery
					if isXPath {
						queryOpts = chromedp.BySearch
					}
					err := chromedp.Text(selector, &textContent, queryOpts).Do(ctx)
					if err != nil {
						found = false
						return nil
					}
					found = true

					if !opt.ensureDisplayed {
						return nil
					}

					// Check if visible - for XPath, we need different approach
					if isXPath {
						// For XPath, use $x() helper to get element and check visibility
						return chromedp.Evaluate(fmt.Sprintf(`
							(function() {
								var els = document.evaluate(%q, document, null, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE, null);
								if (!els.snapshotLength) return false;
								var el = els.snapshotItem(0);
								if (!el) return false;
								var rect = el.getBoundingClientRect();
								return rect.width > 0 && rect.height > 0;
							})()
						`, selector), &visible).Do(ctx)
					} else {
						// For CSS selectors, use querySelector
						return chromedp.Evaluate(fmt.Sprintf(`
							(function() {
								var el = document.querySelector(%q);
								if (!el) return false;
								var rect = el.getBoundingClientRect();
								return rect.width > 0 && rect.height > 0;
							})()
						`, selector), &visible).Do(ctx)
					}
				}),
			)

			if err == nil && found {
				if !opt.ensureDisplayed || visible {
					return &Element{
						selector: selector,
						isXPath:  isXPath,
						ctx:      st.AllocCtx,
						st:       st,
					}
				}
			}
		}
	}
}

// GetTitle returns the page title.
func (st *SeleniumTest) GetTitle() string {
	var title string
	err := st.runChromedp(chromedp.Title(&title))
	require.NoError(st.t, err, "Failed to get page title")
	return title
}

// GetCurrentURL returns the current browser URL.
// Uses a 5-second internal timeout to avoid blocking if the page is waiting for async operations.
func (st *SeleniumTest) GetCurrentURL() string {
	// Create a context with a short timeout for this specific operation
	ctx, cancel := context.WithTimeout(st.AllocCtx, 5*time.Second)
	defer cancel()

	var url string
	err := chromedp.Run(ctx, chromedp.Location(&url))
	if err != nil {
		// If we get a context deadline exceeded, the page might be waiting for async operations
		// Return an empty string rather than failing the test
		if ctx.Err() == context.DeadlineExceeded {
			return ""
		}
		require.NoError(st.t, err, "Failed to get current URL")
	}
	return url
}

// DeleteCookiesAndLocalStorage clears all cookies and local storage.
// This is the Go equivalent of Python's delete_cookies_and_local_storage method.
// Python reference: /app/test/selenium/__init__.py - delete_cookies_and_local_storage()
func (st *SeleniumTest) DeleteCookiesAndLocalStorage() {
	// Navigate to a simple page first
	st.NavigateTo("/")

	// Clear cookies and local storage using chromedp.Run
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Use JavaScript to clear cookies
			return chromedp.Evaluate(`document.cookie.split(";").forEach(function(c) { document.cookie = c.replace(/^ +/, "").replace(/=.*/, "=;expires=" + new Date().toUTCString() + ";path=/"); });`, nil).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Clear local storage
			return chromedp.Evaluate("window.localStorage.clear();", nil).Do(ctx)
		}),
	)
	if err != nil {
		log.Printf("Warning: Failed to clear cookies/local storage: %v", err)
	}
}

// AssertTextContent asserts that an element contains specific text.
func (st *SeleniumTest) AssertTextContent(selector, expectedText string) {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	require.NoError(st.t, err, "Element not found: %s", selector)
	assert.Contains(st.t, text, expectedText, "Element text does not contain expected value")
}

// AssertElementVisible asserts that an element is visible.
// It waits up to the default timeout for the element to become visible.
// This uses the existing WaitForElement with ensureDisplayed=true.
func (st *SeleniumTest) AssertElementVisible(selector string) {
	st.WaitForElement(selector) // ensureDisplayed=true by default
}

// AssertElementNotVisible asserts that an element either doesn't exist or is not visible.
func (st *SeleniumTest) AssertElementNotVisible(selector string) {
	var visible bool
	err := st.runChromedp(chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			var el = document.querySelector(%q);
			if (!el) return false;
			var rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, selector), &visible))
	if err != nil {
		// Element doesn't exist or error - which is fine for "not visible"
		return
	}
	assert.False(st.t, visible, "Element should not be visible: %s", selector)
}

// AssertElementExists asserts that an element exists in the DOM.
func (st *SeleniumTest) AssertElementExists(selector string) {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	require.NoError(st.t, err, "Element not found: %s", selector)
}

// AssertElementNotExists asserts that an element does not exist in the DOM.
// Uses a short timeout to quickly fail if the element exists.
// Python reference: /app/test/selenium/test_homepage.py - pytest.raises(NoSuchElementException)
func (st *SeleniumTest) AssertElementNotExists(selector string) {
	// Create a context with a short timeout (500ms) to quickly check if element exists
	ctx, cancel := context.WithTimeout(st.AllocCtx, 500*time.Millisecond)
	defer cancel()

	var text string
	err := chromedp.Run(ctx, chromedp.Text(selector, &text, chromedp.ByQuery))
	assert.Error(st.t, err, "Element should not exist: %s", selector)
}

// AssertAttributeValue asserts that an element has the expected attribute value.
func (st *SeleniumTest) AssertAttributeValue(selector, attribute, expectedValue string) {
	var value string
	err := st.runChromedp(
		chromedp.AttributeValue(selector, attribute, &value, nil),
	)
	require.NoError(st.t, err, "Failed to get attribute '%s' for element: %s", attribute, selector)
	assert.Equal(st.t, expectedValue, value, "Element %s attribute '%s' has unexpected value", selector, attribute)
}

// ElementOption is a function that modifies element options.
type ElementOption func(*elementOptions)

// elementOptions holds options for WaitForElement.
type elementOptions struct {
	timeout         time.Duration
	ensureDisplayed bool
}

// WithTimeout sets a custom timeout for WaitForElement.
func WithTimeout(timeout time.Duration) ElementOption {
	return func(o *elementOptions) {
		o.timeout = timeout
	}
}

// WithoutVisibilityCheck disables the visibility check in WaitForElement.
func WithoutVisibilityCheck() ElementOption {
	return func(o *elementOptions) {
		o.ensureDisplayed = false
	}
}

// Element represents a DOM element.
// This is the Go equivalent of Python's WebElement.
type Element struct {
	selector string
	isXPath  bool
	ctx      context.Context
	st       *SeleniumTest
}

// Text returns the text content of the element.
func (e *Element) Text() string {
	var text string
	queryOpts := chromedp.ByQuery
	if e.isXPath {
		queryOpts = chromedp.BySearch
	}
	err := e.st.runChromedp(chromedp.Text(e.selector, &text, queryOpts))
	require.NoError(e.st.t, err, "Failed to get text for element: %s", e.selector)
	return text
}

// Click clicks the element.
func (e *Element) Click() {
	queryOpts := chromedp.ByQuery
	if e.isXPath {
		queryOpts = chromedp.BySearch
	}
	err := e.st.runChromedp(chromedp.Click(e.selector, queryOpts))
	require.NoError(e.st.t, err, "Failed to click element: %s", e.selector)
}

// IsDisplayed returns true if the element is visible.
func (e *Element) IsDisplayed() bool {
	var visible bool
	err := e.st.runChromedp(chromedp.Evaluate(fmt.Sprintf(`
		(function() {
			var el = document.querySelector(%q);
			if (!el) return false;
			var rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		})()
	`, e.selector), &visible))
	if err != nil {
		return false
	}
	return visible
}

// Exists returns true if the element exists in the DOM.
func (e *Element) Exists() bool {
	var text string
	queryOpts := chromedp.ByQuery
	if e.isXPath {
		queryOpts = chromedp.BySearch
	}
	err := e.st.runChromedp(chromedp.Text(e.selector, &text, queryOpts))
	return err == nil
}

// SendKeys sends keystrokes to the element.
func (e *Element) SendKeys(keys string) {
	queryOpts := chromedp.ByQuery
	if e.isXPath {
		queryOpts = chromedp.BySearch
	}
	err := e.st.runChromedp(chromedp.SendKeys(e.selector, keys, queryOpts))
	require.NoError(e.st.t, err, "Failed to send keys to element: %s", e.selector)
}

// WaitForURL waits for the current URL to match the expected path.
func (st *SeleniumTest) WaitForURL(expectedPath string) {
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			require.Fail(st.t, fmt.Sprintf("URL did not change to expected path: %s, Current: %s", expectedPath, st.GetCurrentURL()))
		case <-ticker.C:
			currentURL := st.GetCurrentURL()
			if strings.HasSuffix(currentURL, expectedPath) {
				return
			}
		}
	}
}

// GetAttribute retrieves an attribute value from an element.
func (st *SeleniumTest) GetAttribute(selector, attr string) string {
	var value string
	err := st.runChromedp(chromedp.AttributeValue(selector, attr, &value, nil))
	if err != nil {
		return ""
	}
	return value
}

// GetValue retrieves the current value of an input element (DOM property, not HTML attribute).
// This is different from GetAttribute which returns the initial HTML attribute value.
func (st *SeleniumTest) GetValue(selector string) string {
	var value string
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var el = document.querySelector(%q);
					return el ? el.value : null;
				})()
			`, selector), &value).Do(ctx)
		}),
	)
	if err != nil {
		return ""
	}
	return value
}

// GetElementAttribute retrieves an attribute value from an element (on Element).
func (e *Element) GetAttribute(attr string) string {
	var value string
	err := e.st.runChromedp(chromedp.AttributeValue(e.selector, attr, &value, nil))
	if err != nil {
		return ""
	}
	return value
}

// SelectOption selects an option in a select dropdown by value attribute.
func (st *SeleniumTest) SelectOption(selector, value string) {
	// Wait for dropdown to be populated with options (async AJAX)
	st.WaitForDropdownOptions(selector, 3)
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var select = document.querySelector(%q);
					select.value = %q;
					var event = new Event('change', { bubbles: true });
					select.dispatchEvent(event);
				})()
			`, selector, value), nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err)
}

func (st *SeleniumTest) Retry(callback chromedp.ActionFunc, retries int, sleepTimeMillis int) error {
	var err error
	for i := 0; i < retries; i++ {
		err = st.runChromedp(callback)
		if err == nil {
			return nil
		}
		// chromedp.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
		time.Sleep(time.Duration(sleepTimeMillis) * time.Millisecond)
	}
	return err
}

// WaitForDropdownOptions waits for a select dropdown to be populated with options.
// This is needed because dropdowns are populated asynchronously via AJAX.
func (st *SeleniumTest) WaitForDropdownOptions(selector string, minOptions int) {
	err := st.Retry(
		chromedp.ActionFunc(func(ctx context.Context) error {
			var count int
			err := chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var select = document.querySelector(%q);
					if (!select || !select.options) {
						return 0;
					}
					return select.options.length;
				})()
			`, selector), &count).Do(ctx)
			if err != nil {
				return err
			}
			if count >= minOptions {
				return nil
			}
			return fmt.Errorf("only %d options available, need at least %d", count, minOptions)
		}),
		5,
		500,
	)
	require.NoError(st.t, err, "Dropdown %s did not get populated with at least %d options", selector, minOptions)
}

// SelectOptionByVisibleText selects an option in a select dropdown by visible text.
// Matches Python: select.select_by_visible_text(text)
func (st *SeleniumTest) SelectOptionByVisibleText(selector, text string) {
	// Wait for dropdown to be populated with options (async AJAX)
	st.WaitForDropdownOptions(selector, 3)

	// Then select the option by visible text
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var select = document.querySelector(%q);
					for (var i = 0; i < select.options.length; i++) {
						if (select.options[i].text === %q) {
							select.selectedIndex = i;
							select.value = select.options[i].value;
							var event = new Event('change', { bubbles: true });
							select.dispatchEvent(event);
							return true;
						}
				 }
				 return false;
				})()
			`, selector, text), nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err, "Option with text '%s' not found in select", text)
}

// SelectOptionOnElement selects an option in a select dropdown (on Element).
func (e *Element) SelectOption(value string) {
	err := e.st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var select = document.querySelector(%q);
					select.value = %q;
					var event = new Event('change', { bubbles: true });
					select.dispatchEvent(event);
				})()
			`, e.selector, value), nil).Do(ctx)
		}),
	)
	require.NoError(e.st.t, err)
}

// IsElementChecked returns true if checkbox/radio is checked.
func (st *SeleniumTest) IsElementChecked(selector string) bool {
	var checked bool
	err := st.runChromedp(
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				var el = document.querySelector(%q);
				return el ? el.checked : false;
			})()
		`, selector), &checked),
	)
	if err != nil {
		return false
	}
	return checked
}

// IsChecked returns true if the checkbox/radio element is checked.
func (e *Element) IsChecked() bool {
	var checked bool
	err := e.st.runChromedp(
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				var el = document.querySelector(%q);
				return el ? el.checked : false;
			})()
		`, e.selector), &checked),
	)
	if err != nil {
		return false
	}
	return checked
}

// GetElementCount returns the number of elements matching selector.
func (st *SeleniumTest) GetElementCount(selector string) int {
	var count int64
	err := st.runChromedp(
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				return document.querySelectorAll(%q).length;
			})()
		`, selector), &count),
	)
	if err != nil {
		return 0
	}
	return int(count)
}

// GetProgressBarValue retrieves the value attribute of a progress bar element.
func (st *SeleniumTest) GetProgressBarValue(selector string) int {
	var value string
	err := st.runChromedp(chromedp.AttributeValue(selector, "value", &value, nil))
	if err != nil {
		return 0
	}
	// Convert string to int
	i, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return i
}

// IsStruckThrough checks if an element has strike-through styling.
func (st *SeleniumTest) IsStruckThrough(selector string) bool {
	var struckThrough bool
	err := st.runChromedp(
		chromedp.Evaluate(fmt.Sprintf(`
			(function() {
				var el = document.querySelector(%q);
				if (!el) return false;
				// Check for <strike> or <s> tag
				if (el.tagName === 'STRIKE' || el.tagName === 'S') return true;
				// Check for text-decoration: line-through
				var style = window.getComputedStyle(el);
				return style.textDecoration === 'line-through';
			})()
		`, selector), &struckThrough),
	)
	if err != nil {
		return false
	}
	return struckThrough
}

// GetInnerHTML returns the inner HTML of an element.
func (st *SeleniumTest) GetInnerHTML(selector string) string {
	var html string
	err := st.runChromedp(chromedp.InnerHTML(selector, &html, chromedp.ByQuery))
	if err != nil {
		return ""
	}
	return html
}

// GetElementText returns the full text content of an element.
// This is similar to AssertTextContent but returns the text instead of asserting on it.
func (st *SeleniumTest) GetElementText(selector string) string {
	var text string
	err := st.runChromedp(chromedp.Text(selector, &text, chromedp.ByQuery))
	if err != nil {
		return ""
	}
	return text
}

// WaitForURLContains waits for the current URL to contain the expected string.
func (st *SeleniumTest) WaitForURLContains(expectedStr string) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			require.Fail(st.t, "URL did not contain expected string")
		case <-ticker.C:
			currentURL := st.GetCurrentURL()
			if strings.Contains(currentURL, expectedStr) {
				return
			}
		}
	}
}

// WaitForTitle waits for the page title to match the expected value.
func (st *SeleniumTest) WaitForTitle(expectedTitle string) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			require.Fail(st.t, "Title did not match expected value")
		case <-ticker.C:
			currentTitle := st.GetTitle()
			if currentTitle == expectedTitle {
				return
			}
		}
	}
}

// ClearInput clears the value of an input element.
// This is the Go equivalent of Python's element.clear() method.
// Python reference: /app/test/selenium/test_create_module_provider.py - input_field.clear()
func (st *SeleniumTest) ClearInput(selector string) {
	err := st.runChromedp(
		chromedp.Focus(selector, chromedp.ByQuery),
		chromedp.Clear(selector, chromedp.ByQuery),
	)
	require.NoError(st.t, err, "Failed to clear input: %s", selector)
}

// WaitForJavaScriptEval waits for a JavaScript expression to evaluate to true.
// This is useful for waiting on JavaScript execution or async operations.
func (st *SeleniumTest) WaitForJavaScriptEval(jsExpression string) {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			require.Fail(st.t, "JavaScript expression did not evaluate to true within timeout")
		case <-ticker.C:
			var result bool
			err := st.runChromedp(
				chromedp.Evaluate(jsExpression, &result),
			)
			if err == nil && result {
				return
			}
		}
	}
}
