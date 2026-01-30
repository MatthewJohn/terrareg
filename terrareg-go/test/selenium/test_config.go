package selenium

// ConfigForHomepageTests returns environment variables for homepage tests.
// This matches Python's test_homepage.py - TestHomepage.setup_class exactly.
// Python reference: /app/test/selenium/test_homepage.py - TestHomepage.setup_class
//
// Python patches applied:
// - mock.patch('terrareg.config.Config.APPLICATION_NAME', 'unittest application name')
// - mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', return_value=2005)
// - mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module')
// - mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace')
// - mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label')
// - mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['trustednamespace'])
func ConfigForHomepageTests() map[string]string {
	// Start with default test config
	defaults := getDefaultTestConfig()

	// Add homepage-specific overrides
	homepageOverrides := map[string]string{
		"APPLICATION_NAME":            "unittest application name",
		"CONTRIBUTED_NAMESPACE_LABEL": "unittest contributed module",
		"TRUSTED_NAMESPACE_LABEL":     "unittest trusted namespace",
		"VERIFIED_MODULE_LABEL":       "unittest verified label",
		"TRUSTED_NAMESPACES":          "trustednamespace",
	}

	return mergeMaps(defaults, homepageOverrides)
}

// getDefaultTestConfig returns the default configuration for all tests.
// These are the standard config values needed for the test server to start.
// Note: LISTEN_PORT will be overridden to a random port by the test server.
// Python reference: /app/test/selenium/__init__.py - _get_database_path() returns 'temp-selenium.db'
func getDefaultTestConfig() map[string]string {
	return map[string]string{
		"LISTEN_PORT":                          "5000", // Valid port (will be overridden to random port by test server)
		"PUBLIC_URL":                           "http://127.0.0.1:5000",
		"DATABASE_URL":                         "sqlite://./temp-selenium.db", // File-based DB in current directory (./ makes it relative)
		"SECRET_KEY":                           "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"ADMIN_AUTHENTICATION_TOKEN":           "test-admin-token",
		"ALLOW_MODULE_HOSTING":                 "true",
		"DEBUG":                                "true",
		"SESSION_COOKIE_NAME":                  "terrareg_session",
		"SESSION_EXPIRY_MINS":                  "60",
		"ADMIN_SESSION_EXPIRY_MINS":            "60",
		"SESSION_REFRESH_MINS":                 "5",
		"TRUSTED_NAMESPACE_LABEL":              "Trusted",
		"CONTRIBUTED_NAMESPACE_LABEL":          "Contributed",
		"VERIFIED_MODULE_LABEL":                "Verified",
		"ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER": "true",
		"ALLOW_CUSTOM_GIT_URL_MODULE_VERSION":  "true",
		"AUTO_CREATE_NAMESPACE":                "true",
		"AUTO_CREATE_MODULE_PROVIDER":          "true",
		"DISABLE_ANALYTICS":                    "false",
		"AUTO_PUBLISH_MODULE_VERSIONS":         "true",
		"MODULE_VERSION_REINDEX_MODE":          "legacy",
		"PRODUCT":                              "terraform",
		"DEFAULT_TERRAFORM_VERSION":            "1.5.7",
		"MANAGE_TERRAFORM_RC_FILE":             "false",
		"MODULES_DIRECTORY":                    "modules",
		"EXAMPLES_DIRECTORY":                   "examples",
		"PROVIDER_SOURCES":                     "[]",
		"PROVIDER_CATEGORIES":                  `[{"id": 1, "name": "Example Category", "slug": "example-category", "user-selectable": true}]`,
		"GITHUB_URL":                           "https://github.com",
		"GITHUB_API_URL":                       "https://api.github.com",
		"GITHUB_LOGIN_TEXT":                    "Login with Github",
		"OPENID_CONNECT_LOGIN_TEXT":            "Login using OpenID Connect",
		"SAML2_LOGIN_TEXT":                     "Login using SAML",
		"INFRACOST_TLS_INSECURE_SKIP_VERIFY":   "false",
		"ALLOW_UNIDENTIFIED_DOWNLOADS":         "false",
		// Terraform OIDC settings - use test values to avoid file I/O
		"TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH":     "",
		"TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"TERRAFORM_OIDC_IDP_SESSION_EXPIRY":       "3600",
	}
}

// ConfigForNoAuthTests returns environment variables for tests with no authentication methods enabled.
// This is the Go equivalent of Python's test login tests with no auth providers.
// Python reference: /app/test/selenium/test_login.py - TestLoginNoProviderSources
func ConfigForNoAuthTests() map[string]string {
	return getDefaultTestConfig()
}

// ConfigForAdminTokenTests returns environment variables for admin token authentication tests.
// This is the Go equivalent of Python's test login tests with admin token.
// Python reference: /app/test/selenium/test_login.py - TestLogin
func ConfigForAdminTokenTests() map[string]string {
	return getDefaultTestConfig()
}

// ConfigForOIDCTests returns environment variables for OpenID Connect authentication tests.
// This is the Go equivalent of Python's test_valid_openid_connect_login.
// Python reference: /app/test/selenium/test_login.py - test_valid_openid_connect_login
func ConfigForOIDCTests() map[string]string {
	return mergeMaps(getDefaultTestConfig(), map[string]string{
		"OPENID_CONNECT_ISSUER":        "http://127.0.0.1:5000",
		"OPENID_CONNECT_CLIENT_ID":     "test-client-id",
		"OPENID_CONNECT_CLIENT_SECRET": "test-client-secret",
	})
}

// ConfigForSAMLTests returns environment variables for SAML authentication tests.
// This is the Go equivalent of Python's test_valid_saml_login.
// Python reference: /app/test/selenium/test_login.py - test_valid_saml_login
func ConfigForSAMLTests() map[string]string {
	return mergeMaps(getDefaultTestConfig(), map[string]string{
		"SAML2_IDP_METADATA_URL": "http://127.0.0.1:5000/saml-metadata",
		"SAML2_ENTITY_ID":        "test-entity-id",
	})
}

// ConfigForGitHubOAuthTests returns environment variables for GitHub OAuth authentication tests.
// This is the Go equivalent of Python's test_valid_github_provider_source_login.
// Python reference: /app/test/selenium/test_login.py - test_valid_github_provider_source_login
func ConfigForGitHubOAuthTests() map[string]string {
	return mergeMaps(getDefaultTestConfig(), map[string]string{
		"GITHUB_APP_CLIENT_ID":     "test-github-client-id",
		"GITHUB_APP_CLIENT_SECRET": "test-github-client-secret",
	})
}

// mergeMaps merges multiple maps into one, with later maps overriding earlier ones.
// This is equivalent to Python's dictionary unpacking: {**dict1, **dict2}
func mergeMaps(maps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// ConfigForInitialSetupTests returns config for initial setup tests.
// This creates an empty database scenario, like Python's _TEST_DATA = {}
// Python reference: /app/test/selenium/test_initial_setup.py - TestInitialSetup
func ConfigForInitialSetupTests() map[string]string {
	baseConfig := getDefaultTestConfig()

	// Override for initial setup - empty database triggers setup wizard
	// Set SECRET_KEY and auth tokens to empty to simulate unconfigured state
	return mergeMaps(baseConfig, map[string]string{
		"SECRET_KEY":                 "", // Unconfigured - no secret key set
		"UPLOAD_API_KEYS":            "",
		"ADMIN_AUTHENTICATION_TOKEN": "",
		// Keep ALLOW_MODULE_HOSTING and other settings
		"ALLOW_MODULE_HOSTING":        "true",
		"AUTO_CREATE_NAMESPACE":       "true",
		"AUTO_CREATE_MODULE_PROVIDER": "true",
	})
}

// ConfigForCreateNamespaceTests returns config for namespace creation tests.
// This is the Go equivalent of Python's test_create_namespace with admin token.
// Python reference: /app/test/selenium/test_create_namespace.py
func ConfigForCreateNamespaceTests() map[string]string {
	// Namespace creation tests use admin auth
	return ConfigForAdminTokenTests()
}

// ConfigForCreateModuleProviderTests returns config for module provider tests.
// Python reference: /app/test/selenium/test_create_module_provider.py
func ConfigForCreateModuleProviderTests() map[string]string {
	// Module provider creation tests use admin auth
	return ConfigForAdminTokenTests()
}
