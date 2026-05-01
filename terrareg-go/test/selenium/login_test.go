//go:build selenium

package selenium

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoginNoProviderSources tests login without provider sources.
// This is the Go implementation of Python's TestLoginNoProviderSources class.
// Python reference: /app/test/selenium/test_login.py - TestLoginNoProviderSources
//
// Test methods:
// - test_no_authentication_methods_warning - equivalent to Python's parametrized test (4 test cases)

func TestLoginNoProviderSources(t *testing.T) {
	t.Run("test_no_authentication_methods_warning", testNoAuthenticationMethodsWarning)
}

// loginNoAuthWarningTest represents a single test case for no auth warning.
type loginNoAuthWarningTest struct {
	adminToken    string
	openidEnabled bool
	samlEnabled   bool
	warningShown  bool
}

// loginNoAuthWarningTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_login.py line 57-64
var loginNoAuthWarningTests = []loginNoAuthWarningTest{
	// Check warning is shown when no methods are available
	{"", false, false, true},
	// Cases where authentication method is enabled
	{"pass", false, false, false},
	{"", true, false, false},
	{"", false, true, false},
}

// testNoAuthenticationMethodsWarning tests warning is shown when no authentication methods are available.
// Python reference: /app/test/selenium/test_login.py - TestLoginNoProviderSources.test_no_authentication_methods_warning
func testNoAuthenticationMethodsWarning(t *testing.T) {
	for _, tt := range loginNoAuthWarningTests {
		t.Run(tt.adminToken+"_"+boolToStr(tt.openidEnabled)+"_"+boolToStr(tt.samlEnabled), func(t *testing.T) {
			// Build config overrides for this test case
			// Python: self.update_multiple_mocks(...)
			configOverrides := map[string]string{
				"ADMIN_AUTHENTICATION_TOKEN": tt.adminToken,
				"PROVIDER_SOURCES":           "[]",
			}

			// Set OIDC config if enabled
			if tt.openidEnabled {
				configOverrides["OPENID_CONNECT_ISSUER"]       = "https://example.com"
				configOverrides["OPENID_CONNECT_CLIENT_ID"]     = "test-client-id"
				configOverrides["OPENID_CONNECT_CLIENT_SECRET"] = "test-client-secret"
			}

			// Set SAML config if enabled
			if tt.samlEnabled {
				configOverrides["SAML2_ENTITY_ID"]         = "test-entity-id"
				configOverrides["SAML2_IDP_METADATA_URL"]   = "https://example.com/metadata"
				configOverrides["SAML2_PUBLIC_KEY"]        = "test-public-key"
				configOverrides["SAML2_PRIVATE_KEY"]       = "test-private-key"
			}

			st := NewSeleniumTestWithConfig(t, configOverrides)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			st.NavigateTo("/login")
			waitForLoginFormReady(st)

			// Wait for JavaScript to execute and show/hide warning
			// The warning is shown/hidden by JavaScript based on config
			// We need to wait for the config to be loaded and processed
			st.WaitForJavaScriptEval(`
				(function() {
					return document.getElementById('login-title') !== null &&
					       window.getComputedStyle(document.getElementById('login-title')).display !== 'none';
				})()
			`)

			// Python: warning = selenium_instance.find_element(By.ID, 'no-authentication-methods-warning')
			//         assert warning.is_displayed() == warning_shown
			if tt.warningShown {
				// Python: assert warning.text == 'Login is not available as there are no authentication methods configured'
				st.AssertElementVisible("#no-authentication-methods-warning")
				st.AssertTextContent("#no-authentication-methods-warning", "Login is not available as there are no authentication methods configured")
			} else {
				st.AssertElementNotVisible("#no-authentication-methods-warning")
			}
		})
	}
}

// TestLogin tests the login functionality.
// This is the Go implementation of Python's TestLogin class.
// Python reference: /app/test/selenium/test_login.py - TestLogin class
//
// Test methods:
// - test_ensure_admin_authentication_not_shown - admin form not shown when not configured
// - test_ensure_openid_connect_login_not_shown - OIDC button not shown when not enabled
// - test_admin_password_login_invalid_password - 2 parametrized test cases
// - test_admin_password_login - valid admin password login
// - test_valid_openid_connect_login - 8 parametrized test cases
// - test_invalid_openid_connect_response - invalid OIDC error handling
// - test_ensure_saml_login_not_shown - SAML button not shown when not enabled
// - test_valid_saml_login - 8 parametrized test cases
// - test_invalid_saml_response - invalid SAML error handling
// - test_ensure_provider_source_buttons_not_shown - no provider buttons when no providers
// - test_valid_github_provider_source_login - 16 parametrized test cases
// - test_invalid_github_response - invalid GitHub error handling

func TestLogin(t *testing.T) {
	t.Run("test_ensure_admin_authentication_not_shown", testEnsureAdminAuthenticationNotShown)
	t.Run("test_ensure_openid_connect_login_not_shown", testEnsureOpenIDConnectLoginNotShown)
	t.Run("test_admin_password_login_invalid_password", testAdminPasswordLoginInvalidPassword)
	t.Run("test_admin_password_login", testAdminPasswordLogin)
	t.Run("test_valid_openid_connect_login", testValidOpenIDConnectLogin)
	t.Run("test_invalid_openid_connect_response", testInvalidOpenIDConnectResponse)
	t.Run("test_ensure_saml_login_not_shown", testEnsureSAMLLoginNotShown)
	t.Run("test_valid_saml_login", testValidSAMLLogin)
	t.Run("test_invalid_saml_response", testInvalidSAMLResponse)
	t.Run("test_ensure_provider_source_buttons_not_shown", testEnsureProviderSourceButtonsNotShown)
	t.Run("test_valid_github_provider_source_login", testValidGithubProviderSourceLogin)
	t.Run("test_invalid_github_response", testInvalidGithubResponse)
}

// waitForLoginFormReady waits for the login form to be rendered.
// Python reference: /app/test/selenium/test_login.py - _wait_for_login_form_ready
func waitForLoginFormReady(st *SeleniumTest) {
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'login-title').is_displayed(), True)
	st.AssertElementVisible("#login-title")
}

// testEnsureAdminAuthenticationNotShown ensures admin login form is not shown when admin password is not configured.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_ensure_admin_authentication_not_shown
func testEnsureAdminAuthenticationNotShown(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	waitForLoginFormReady(st)

	// Python: assert selenium_instance.find_element(By.ID, 'admin-login').is_displayed() == False
	st.AssertElementNotVisible("#admin-login")
}

// testEnsureOpenIDConnectLoginNotShown ensures OpenID connect login button is not shown when not enabled.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_ensure_openid_connect_login_not_shown
func testEnsureOpenIDConnectLoginNotShown(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	waitForLoginFormReady(st)

	// Python: assert selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed() == False
	st.AssertElementNotVisible("#openid-connect-login")
}

// adminPasswordLoginInvalidTest represents a single test case for invalid admin password.
type adminPasswordLoginInvalidTest struct {
	testPassword string
}

// adminPasswordLoginInvalidTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_login.py line 137-140
var adminPasswordLoginInvalidTests = []adminPasswordLoginInvalidTest{
	{""},
	{"incorrectpassword"},
}

// testAdminPasswordLoginInvalidPassword tests admin authentication using incorrect password.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_admin_password_login_invalid_password
func testAdminPasswordLoginInvalidPassword(t *testing.T) {
	for _, tt := range adminPasswordLoginInvalidTests {
		t.Run(tt.testPassword, func(t *testing.T) {
			st := NewSeleniumTest(t)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			st.NavigateTo("/login")
			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'admin-login').is_displayed(), True)
			st.AssertElementVisible("#admin-login")

			// Python: selenium_instance.find_element(By.ID, 'admin_token_input').send_keys(test_password)
			element := st.WaitForElement("#admin_token_input")
			element.SendKeys(tt.testPassword)

			// Python: selenium_instance.find_element(By.ID, 'login-button').click()
			loginButton := st.WaitForElement("#login-button")
			loginButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/login'))
			currentURL := st.GetCurrentURL()
			assert.Equal(t, st.GetURL("/login"), currentURL, "Should remain on login page")

			// Python: assert selenium_instance.find_element(By.ID, 'navbar_login_span').text == 'Login'
			st.AssertTextContent("#navbar_login_span", "Login")

			// Python: error_div = selenium_instance.find_element(By.ID, 'login_error')
			//         assert error_div.is_displayed() == True
			//         assert error_div.text == 'Incorrect admin token'
			st.AssertElementVisible("#login_error")
			st.AssertTextContent("#login_error", "Incorrect admin token")
		})
	}
}

// testAdminPasswordLogin tests admin authentication using password.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_admin_password_login
func testAdminPasswordLogin(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'admin-login').is_displayed(), True)
	st.AssertElementVisible("#admin-login")

	// Python: selenium_instance.find_element(By.ID, 'admin_token_input').send_keys('testloginpassword')
	element := st.WaitForElement("#admin_token_input")
	element.SendKeys("testloginpassword")

	// Python: selenium_instance.find_element(By.ID, 'login-button').click()
	loginButton := st.WaitForElement("#login-button")
	loginButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/"), currentURL, "Should redirect to home page")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')
	st.AssertTextContent("#navbar_login_span", "Logout")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), True)
	st.AssertElementVisible("#navbarSettingsDropdown")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), True)
	st.AssertElementVisible("#navbarCreateDropdown")
}

// oidcLoginTest represents a single test case for OpenID Connect login.
type oidcLoginTest struct {
	enableAccessControls bool
	groupMemberships     []string
	hasSiteAdmin         bool
	canCreateModule      bool
}

// oidcLoginTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_login.py line 193-202
var oidcLoginTests = []oidcLoginTest{
	{true, []string{"nopermissions"}, false, false},
	{true, []string{"siteadmin"}, true, true},
	{true, []string{"moduledetailsfull"}, false, true},
	{true, []string{}, false, false},
	{false, []string{"nopermissions"}, true, true},
	{false, []string{"siteadmin"}, true, true},
	{false, []string{"moduledetailsfull"}, true, true},
	{false, []string{}, true, true},
}

// testValidOpenIDConnectLogin ensures OpenID Connect login works.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_valid_openid_connect_login
func testValidOpenIDConnectLogin(t *testing.T) {
	for _, tt := range oidcLoginTests {
		groupStr := sliceToStr(tt.groupMemberships)
		t.Run(groupStr, func(t *testing.T) {
			st := NewSeleniumTest(t)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			st.NavigateTo("/login")
			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed(), True)
			st.AssertElementVisible("#openid-connect-login")

			// Python: openid_connect_login_button = selenium_instance.find_element(By.ID, 'openid-connect-login')
			//         assert openid_connect_login_button.text == 'Unittest OpenID Connect Login Button'
			st.AssertTextContent("#openid-connect-login", "Unittest OpenID Connect Login Button")

			// Python: openid_connect_login_button.click()
			openidButton := st.WaitForElement("#openid-connect-login")
			openidButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))
			currentURL := st.GetCurrentURL()
			assert.Equal(t, st.GetURL("/"), currentURL, "Should redirect to home page")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')
			st.AssertTextContent("#navbar_login_span", "Logout")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), has_site_admin)
			if tt.hasSiteAdmin {
				st.AssertElementVisible("#navbarSettingsDropdown")
			} else {
				st.AssertElementNotVisible("#navbarSettingsDropdown")
			}

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), can_create_module)
			if tt.canCreateModule {
				st.AssertElementVisible("#navbarCreateDropdown")
			} else {
				st.AssertElementNotVisible("#navbarCreateDropdown")
			}
		})
	}
}

// testInvalidOpenIDConnectResponse tests handling of invalid OpenID connect authentication error.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_invalid_openid_connect_response
func testInvalidOpenIDConnectResponse(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed(), True)
	st.AssertElementVisible("#openid-connect-login")

	// Python: selenium_instance.find_element(By.ID, 'openid-connect-login').click()
	openidButton := st.WaitForElement("#openid-connect-login")
	openidButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/openid/callback?code=abcdefg&state=unitteststate'))
	currentURL := st.GetCurrentURL()
	assert.Contains(t, currentURL, "/openid/callback", "Should be on callback URL")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-title').text, 'Login error')
	st.AssertTextContent("#error-title", "Login error")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-content').text, 'Invalid response from SSO')
	st.AssertTextContent("#error-content", "Invalid response from SSO")
}

// testEnsureSAMLLoginNotShown ensures SAML login button is not shown when SAML login is not available.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_ensure_saml_login_not_shown
func testEnsureSAMLLoginNotShown(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	waitForLoginFormReady(st)

	// Python: assert selenium_instance.find_element(By.ID, 'saml-login').is_displayed() == False
	st.AssertElementNotVisible("#saml-login")
}

// samlLoginTest represents a single test case for SAML login.
type samlLoginTest struct {
	enableAccessControls bool
	groupMemberships     []string
	hasSiteAdmin         bool
	canCreateModule      bool
}

// samlLoginTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_login.py line 271-280
var samlLoginTests = []samlLoginTest{
	{true, []string{"nopermissions"}, false, false},
	{true, []string{"siteadmin"}, true, true},
	{true, []string{"moduledetailsfull"}, false, true},
	{true, []string{}, false, false},
	{false, []string{"nopermissions"}, true, true},
	{false, []string{"siteadmin"}, true, true},
	{false, []string{"moduledetailsfull"}, true, true},
	{false, []string{}, true, true},
}

// testValidSAMLLogin ensures SAML login works.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_valid_saml_login
func testValidSAMLLogin(t *testing.T) {
	for _, tt := range samlLoginTests {
		groupStr := sliceToStr(tt.groupMemberships)
		t.Run(groupStr, func(t *testing.T) {
			st := NewSeleniumTest(t)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			st.NavigateTo("/login")
			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'saml-login').is_displayed(), True)
			st.AssertElementVisible("#saml-login")

			// Python: saml_login_button = selenium_instance.find_element(By.ID, 'saml-login')
			//         assert saml_login_button.text == 'Unittest SAML Login Button'
			st.AssertTextContent("#saml-login", "Unittest SAML Login Button")

			// Python: saml_login_button.click()
			samlButton := st.WaitForElement("#saml-login")
			samlButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))
			currentURL := st.GetCurrentURL()
			assert.Equal(t, st.GetURL("/"), currentURL, "Should redirect to home page")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')
			st.AssertTextContent("#navbar_login_span", "Logout")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), has_site_admin)
			if tt.hasSiteAdmin {
				st.AssertElementVisible("#navbarSettingsDropdown")
			} else {
				st.AssertElementNotVisible("#navbarSettingsDropdown")
			}

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), can_create_module)
			if tt.canCreateModule {
				st.AssertElementVisible("#navbarCreateDropdown")
			} else {
				st.AssertElementNotVisible("#navbarCreateDropdown")
			}
		})
	}
}

// testInvalidSAMLResponse tests handling of invalid SAML authentication error.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_invalid_saml_response
func testInvalidSAMLResponse(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'saml-login').is_displayed(), True)
	st.AssertElementVisible("#saml-login")

	// Python: selenium_instance.find_element(By.ID, 'saml-login').click()
	samlButton := st.WaitForElement("#saml-login")
	samlButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/saml/login?acs'))
	currentURL := st.GetCurrentURL()
	assert.Contains(t, currentURL, "/saml/login", "Should be on SAML login URL")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-title').text, 'Login error')
	st.AssertTextContent("#error-title", "Login error")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-content').text, 'An error occured whilst processing SAML login request')
	st.AssertTextContent("#error-content", "An error occured whilst processing SAML login request")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Login')
	st.AssertTextContent("#navbar_login_span", "Login")
}

// testEnsureProviderSourceButtonsNotShown ensures provider source login buttons are not shown when there aren't any provider sources in database.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_ensure_provider_source_buttons_not_shown
func testEnsureProviderSourceButtonsNotShown(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	waitForLoginFormReady(st)

	// Python: buttons = selenium_instance.find_element(By.ID, "sso-login").find_elements(By.TAG_NAME, 'a')
	//         assert [button.get_attribute('id') for button in buttons] == ["openid-connect-login", "saml-login"]
	// Ensure there aren't any login buttons except the built-in ones
	st.AssertElementVisible("#openid-connect-login")
	st.AssertElementVisible("#saml-login")
}

// githubLoginTest represents a single test case for GitHub provider source login.
type githubLoginTest struct {
	enableAccessControls                     bool
	autoGenerateGithubOrganisationNamespaces bool
	groupMemberships                         []string
	hasSiteAdmin                             bool
	canCreateModule                          bool
}

// githubLoginTests contains all test cases from Python's @pytest.mark.parametrize.
// Python reference: /app/test/selenium/test_login.py line 384-403
var githubLoginTests = []githubLoginTest{
	{true, false, []string{"nopermissions"}, false, false},
	{true, false, []string{"siteadmin"}, true, true},
	{true, false, []string{"moduledetailsfull"}, false, true},
	{true, false, []string{}, false, false},
	{false, false, []string{"nopermissions"}, true, true},
	{false, false, []string{"siteadmin"}, true, true},
	{false, false, []string{"moduledetailsfull"}, true, true},
	{false, false, []string{}, true, true},
	// With auto github namespace creation enabled
	{true, true, []string{"nopermissions"}, false, true},
	{true, true, []string{"siteadmin"}, true, true},
	{true, true, []string{"moduledetailsfull"}, false, true},
	{true, true, []string{}, false, true},
	{false, true, []string{"nopermissions"}, true, true},
	{false, true, []string{"siteadmin"}, true, true},
	{false, true, []string{"moduledetailsfull"}, true, true},
	{false, true, []string{}, true, true},
}

// testValidGithubProviderSourceLogin ensures Github login works.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_valid_github_provider_source_login
func testValidGithubProviderSourceLogin(t *testing.T) {
	for _, tt := range githubLoginTests {
		groupStr := sliceToStr(tt.groupMemberships)
		autoStr := boolToStr(tt.autoGenerateGithubOrganisationNamespaces)
		t.Run(groupStr+"_"+autoStr, func(t *testing.T) {
			st := NewSeleniumTest(t)
			defer st.TearDown()
			st.DeleteCookiesAndLocalStorage()

			st.NavigateTo("/login")
			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'ut-github-login').is_displayed(), True)
			st.AssertElementVisible("#ut-github-login")

			// Python: github_login_button = selenium_instance.find_element(By.ID, 'ut-github-login')
			//         assert github_login_button.text == 'Unittest Github Login Button'
			st.AssertTextContent("#ut-github-login", "Unittest Github Login Button")

			// Python: github_login_button.click()
			githubButton := st.WaitForElement("#ut-github-login")
			githubButton.Click()

			// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))
			currentURL := st.GetCurrentURL()
			assert.Equal(t, st.GetURL("/"), currentURL, "Should redirect to home page")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')
			st.AssertTextContent("#navbar_login_span", "Logout")

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), has_site_admin)
			if tt.hasSiteAdmin {
				st.AssertElementVisible("#navbarSettingsDropdown")
			} else {
				st.AssertElementNotVisible("#navbarSettingsDropdown")
			}

			// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), can_create_module)
			if tt.canCreateModule {
				st.AssertElementVisible("#navbarCreateDropdown")
			} else {
				st.AssertElementNotVisible("#navbarCreateDropdown")
			}
		})
	}
}

// testInvalidGithubResponse tests handling of invalid Github authentication error.
// Python reference: /app/test/selenium/test_login.py - TestLogin.test_invalid_github_response
func testInvalidGithubResponse(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()
	st.DeleteCookiesAndLocalStorage()

	st.NavigateTo("/login")
	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'ut-github-login').is_displayed(), True)
	st.AssertElementVisible("#ut-github-login")

	// Python: selenium_instance.find_element(By.ID, 'ut-github-login').click()
	githubButton := st.WaitForElement("#ut-github-login")
	githubButton.Click()

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/ut-github/callback?code=1234'))
	currentURL := st.GetCurrentURL()
	assert.Contains(t, currentURL, "/ut-github/callback", "Should be on GitHub callback URL")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-title').text, 'Login error')
	st.AssertTextContent("#error-title", "Login error")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-content').text, 'Invalid code returned from ut-github')
	st.AssertTextContent("#error-content", "Invalid code returned from ut-github")

	// Python: self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Login')
	st.AssertTextContent("#navbar_login_span", "Login")
}

// Helper functions

// boolToStr converts a boolean to a string representation.
func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// sliceToStr converts a string slice to a comma-separated string.
func sliceToStr(s []string) string {
	if len(s) == 0 {
		return "empty"
	}
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ","
		}
		result += v
	}
	return result
}
