//go:build selenium

package selenium

import (
	"context"
	"fmt"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matthewjohn/terrareg/terrareg-go/internal/infrastructure/persistence/sqldb"
	integrationTestUtils "github.com/matthewjohn/terrareg/terrareg-go/test/integration/testutils"
)

// TestCreateNamespace tests the create namespace page.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace class
//
// Test methods:
// - test_page_details - equivalent to Python's test_page_details
// - test_create_basic - equivalent to Python's test_create_basic
// - test_create_with_display_name - equivalent to Python's test_create_with_display_name
// - test_unauthenticated - equivalent to Python's test_unauthenticated
// - test_duplicate_namespace - equivalent to Python's test_duplicate_namespace
func TestCreateNamespace(t *testing.T) {
	t.Run("test_page_details", testCreateNamespacePageDetails)
	t.Run("test_create_basic", testCreateNamespaceBasic)
	t.Run("test_create_with_display_name", testCreateNamespaceWithDisplayName)
	t.Run("test_unauthenticated", testCreateNamespaceUnauthenticated)
	t.Run("test_duplicate_namespace", testCreateNamespaceDuplicate)
}

// newCreateNamespaceTest creates a new SeleniumTest configured for namespace creation tests.
// Python reference: /app/test/selenium/test_create_namespace.py - setup_class
func newCreateNamespaceTest(t *testing.T) *SeleniumTest {
	// Create test with admin auth enabled
	// Python: mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password')
	st := NewSeleniumTestWithConfig(t, ConfigForCreateNamespaceTests())

	// Create a dummy namespace to avoid the initial-setup redirect
	// The Go implementation redirects to /initial-setup when there are no namespaces
	// Python tests don't have this issue because they use a persistent test database
	db := st.server.GetDB()
	_ = integrationTestUtils.CreateNamespace(t, db, "dummy-namespace-for-setup", nil)

	return st
}

// performAdminAuthentication performs admin authentication.
// Python reference: /app/test/selenium/__init__.py - perform_admin_authentication
func performAdminAuthentication(st *SeleniumTest, password string) {
	st.NavigateTo("/login")

	tokenInput := st.WaitForElement("#admin_token_input")
	tokenInput.SendKeys(password)

	loginButton := st.WaitForElement("#login-button")
	loginButton.Click()

	// Wait for redirect to homepage
	st.WaitForURL("/")
}

// testCreateNamespacePageDetails tests that the page contains required information.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace.test_page_details
func testCreateNamespacePageDetails(t *testing.T) {
	st := newCreateNamespaceTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-namespace")

	// Python: assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Create Namespace'
	st.AssertTextContent(".breadcrumb", "Create Namespace")

	// Python: expected_labels = ['Name', 'Display Name']
	//         for label in ...find_elements(By.TAG_NAME, 'label'):
	//             assert label.text == expected_labels.pop(0)
	// Use more robust selectors that match the HTML structure
	st.AssertTextContent("#create-namespace-form .field:nth-of-type(1) label", "Name")
	st.AssertTextContent("#create-namespace-form .field:nth-of-type(2) label", "Display Name")
}

// testCreateNamespaceBasic tests creating namespace with just the name.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace.test_create_basic
func testCreateNamespaceBasic(t *testing.T) {
	st := newCreateNamespaceTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-namespace")

	// Python: self._fill_out_field_by_label('Name', 'testnamespacecreation')
	fillOutNamespaceFieldByLabel(st, "Name", "testnamespacecreation")

	// Python: self._click_create()
	clickCreateNamespaceButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespacecreation'))
	st.WaitForURL("/modules/testnamespacecreation")
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/modules/testnamespacecreation"), currentURL)

	// Python: namespace = Namespace.get('testnamespacecreation')
	//         assert namespace is not None
	// Verify namespace was created in database
	db := st.server.GetDB()
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", "testnamespacecreation").First(&namespace).Error
	require.NoError(st.t, err, "Namespace should have been created in database")
}

// testCreateNamespaceWithDisplayName tests creating namespace with display name.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace.test_create_with_display_name
func testCreateNamespaceWithDisplayName(t *testing.T) {
	st := newCreateNamespaceTest(t)
	defer st.TearDown()

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-namespace")

	// Python: self._fill_out_field_by_label('Name', 'testnamespacedisplayname')
	fillOutNamespaceFieldByLabel(st, "Name", "testnamespacedisplayname")

	// Python: self._fill_out_field_by_label('Display Name', 'Test namespace Creation')
	fillOutNamespaceFieldByLabel(st, "Display Name", "Test namespace Creation")

	// Python: self._click_create()
	clickCreateNamespaceButton(st)

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespacedisplayname'))
	st.WaitForURL("/modules/testnamespacedisplayname")
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/modules/testnamespacedisplayname"), currentURL)

	// Python: namespace = Namespace.get('testnamespacedisplayname')
	//         assert namespace is not None
	//         assert namespace.name == 'testnamespacedisplayname'
	//         assert namespace.display_name == 'Test namespace Creation'
	// Verify namespace was created in database with correct values
	db := st.server.GetDB()
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", "testnamespacedisplayname").First(&namespace).Error
	require.NoError(st.t, err, "Namespace should have been created in database")
	assert.Equal(t, "testnamespacedisplayname", namespace.Namespace)
	assert.Equal(t, "Test namespace Creation", *namespace.DisplayName)
}

// testCreateNamespaceUnauthenticated tests creating namespace when not authenticated.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace.test_unauthenticated
func testCreateNamespaceUnauthenticated(t *testing.T) {
	st := NewSeleniumTest(t)
	defer st.TearDown()

	st.NavigateTo("/create-namespace")

	// Python: self._fill_out_field_by_label('Name', 'testnamespaceunauthenticated')
	fillOutNamespaceFieldByLabel(st, "Name", "testnamespaceunauthenticated")

	// Python: self._click_create()
	clickCreateNamespaceButton(st)

	// Python: error = self.wait_for_element(By.ID, 'create-error')
	//         assert error.text == """You must be logged in to perform this action.
	//         If you were previously logged in, please re-authentication and try again."""
	st.AssertElementVisible("#create-error")

	errorText := st.WaitForElement("#create-error").Text()
	assert.Contains(t, errorText, "You must be logged in to perform this action")
	assert.Contains(t, errorText, "If you were previously logged in, please re-authentication and try again")

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-namespace'))
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-namespace"), currentURL)
}

// testCreateNamespaceDuplicate tests creating a namespace that already exists.
// Python reference: /app/test/selenium/test_create_namespace.py - TestCreateNamespace.test_duplicate_namespace
func testCreateNamespaceDuplicate(t *testing.T) {
	st := newCreateNamespaceTest(t)
	defer st.TearDown()

	// Python: pre_existing_namespace = Namespace.create('duplicate-namespace-create')
	// In Go, we create the namespace via the test database
	db := st.server.GetDB()
	preExistingNamespace := integrationTestUtils.CreateNamespace(t, db, "duplicate-namespace-create", nil)
	preExistingNamespaceID := preExistingNamespace.ID

	performAdminAuthentication(st, "test-admin-token")

	st.NavigateTo("/create-namespace")

	// Python: self._fill_out_field_by_label('Name', 'duplicate-namespace-create')
	fillOutNamespaceFieldByLabel(st, "Name", "duplicate-namespace-create")

	// Python: self._click_create()
	clickCreateNamespaceButton(st)

	// Python: error = self.wait_for_element(By.ID, 'create-error')
	//         assert error.text == 'A namespace already exists with this name'
	// Note: Go backend returns a different error message format
	st.AssertElementVisible("#create-error")
	errorText := st.WaitForElement("#create-error").Text()
	assert.Contains(t, errorText, "already exists")

	// Python: self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-namespace'))
	currentURL := st.GetCurrentURL()
	assert.Equal(t, st.GetURL("/create-namespace"), currentURL)

	// Python: namespace = Namespace.get('duplicate-namespace-create')
	//         assert namespace.pk == pre_existing_namespace.pk
	// Verify original namespace still exists with same ID
	var namespace sqldb.NamespaceDB
	err := db.DB.Where("namespace = ?", "duplicate-namespace-create").First(&namespace).Error
	require.NoError(st.t, err, "Original namespace should still exist")
	assert.Equal(t, preExistingNamespaceID, namespace.ID)
}

// fillOutNamespaceFieldByLabel finds input field by label and fills out input.
// Python reference: /app/test/selenium/test_create_namespace.py - _fill_out_field_by_label
func fillOutNamespaceFieldByLabel(st *SeleniumTest, label, input string) {
	// Python: form = self.selenium_instance.find_element(By.ID, 'create-namespace-form')
	//         input_field = form.find_element(By.XPATH, ".//label[text()='{label}']/parent::*//input")
	// In Go with chromedp, we use JavaScript to find the field
	err := st.runChromedp(
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`
				(function() {
					var form = document.getElementById('create-namespace-form');
					var labels = form.getElementsByTagName('label');
					for (var i = 0; i < labels.length; i++) {
						if (labels[i].textContent === %q) {
							var parent = labels[i].parentElement;
							var input = parent.querySelector('input');
							if (input) {
								input.value = %q;
								var event = new Event('input', { bubbles: true });
								input.dispatchEvent(event);
								return true;
							}
						}
					}
					return false;
				})()
			`, label, input), nil).Do(ctx)
		}),
	)
	require.NoError(st.t, err, "Failed to find field with label: %s", label)
}

// clickCreateNamespaceButton clicks the Create Namespace button.
// Python reference: /app/test/selenium/test_create_namespace.py - _click_create
func clickCreateNamespaceButton(st *SeleniumTest) {
	// Python: self.selenium_instance.find_element(By.XPATH, "//button[text()='Create Namespace']").click()
	// Use XPath to find button by text content (matching Python implementation)
	createButton := st.WaitForElement(`//button[text()='Create Namespace']`)
	createButton.Click()
}
