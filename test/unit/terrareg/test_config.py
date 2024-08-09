
import unittest.mock
import re

import pytest

import terrareg.config


class TestConfig:
    """Test Config class."""

    @classmethod
    def setup_class(cls):
        """Setup class"""
        # Create list to hold all tested variables
        cls.tested_variables = []

    @classmethod
    def teardown_class(cls):
        """Ensure that all configs from Config class have been tested."""
        valid_config_re = re.compile(r'^[A-Z]')

        untested_properties = []

        for prop in dir(terrareg.config.Config):

            # Check if attribute looks like a config variable
            if not valid_config_re.match(prop):
                continue

            # If property has not been tested, add to list of untested configs
            if prop not in cls.tested_variables:
                untested_properties.append(prop)

        # Assert that there were no untested configs
        if untested_properties:
            print('Untested configs: ', untested_properties)
        assert untested_properties == []

    @classmethod
    def register_checked_config(cls, config):
        """Register a config item as having been tested."""
        if config not in cls.tested_variables:
            cls.tested_variables.append(config)

    @pytest.mark.parametrize('config_name,override_expected_value', [
        ('ADMIN_AUTHENTICATION_TOKEN', None),
        ('ANALYTICS_TOKEN_DESCRIPTION', None),
        ('ANALYTICS_TOKEN_PHRASE', None),
        ('APPLICATION_NAME', None),
        ('CONTRIBUTED_NAMESPACE_LABEL', None),
        ('DATABASE_URL', None),
        ('DATA_DIRECTORY', 'unittest-value/data'),
        ('UPLOAD_DIRECTORY', None),
        ('EXAMPLES_DIRECTORY', None),
        ('EXAMPLE_ANALYTICS_TOKEN', None),
        ('GIT_PROVIDER_CONFIG', None),
        ('LOGO_URL', None),
        ('MODULES_DIRECTORY', None),
        ('SECRET_KEY', None),
        ('SSL_CERT_PRIVATE_KEY', None),
        ('SSL_CERT_PUBLIC_KEY', None),
        ('TERRAFORM_EXAMPLE_VERSION_TEMPLATE', None),
        ('TERRAFORM_EXAMPLE_VERSION_TEMPLATE_PRE_MAJOR', None),
        ('TRUSTED_NAMESPACE_LABEL', None),
        ('VERIFIED_MODULE_LABEL', None),
        ('INFRACOST_API_KEY', None),
        ('INFRACOST_PRICING_API_ENDPOINT', None),
        ('DOMAIN_NAME', None),
        ('PUBLIC_URL', None),
        ('ADDITIONAL_MODULE_TABS', None),
        ('OPENID_CONNECT_LOGIN_TEXT', None),
        ('OPENID_CONNECT_CLIENT_ID', None),
        ('OPENID_CONNECT_CLIENT_SECRET', None),
        ('OPENID_CONNECT_ISSUER', None),
        ('SAML2_LOGIN_TEXT', None),
        ('SAML2_ENTITY_ID', None),
        ('SAML2_IDP_METADATA_URL', None),
        ('SAML2_ISSUER_ENTITY_ID', None),
        ('SAML2_LOGIN_TEXT', None),
        ('SAML2_PRIVATE_KEY', None),
        ('SAML2_PUBLIC_KEY', None),
        ('SAML2_GROUP_ATTRIBUTE', None),
        ('INTERNAL_EXTRACTION_ANALYTICS_TOKEN', None),
        ('MODULE_LINKS', None),
        ('DEFAULT_TERRAFORM_VERSION', None),
        ('TERRAFORM_ARCHIVE_MIRROR', None),
        ('SENTRY_DSN', None),
        ('TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH', None),
        ('TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT', None),
        ('TERRAFORM_PRESIGNED_URL_SECRET', None),
        ('GITHUB_URL', None),
        ('GITHUB_API_URL', None),
        ('GITHUB_APP_CLIENT_ID', None),
        ('GITHUB_APP_CLIENT_SECRET', None),
        ('GITHUB_LOGIN_TEXT', None),
        ('GO_PACKAGE_CACHE_DIRECTORY', None),
        ('PROVIDER_CATEGORIES', None),
        ('PROVIDER_SOURCES', None),
        ('SITE_WARNING', None),
        ('UPSTREAM_GIT_CREDENTIALS_USERNAME', None),
        ('UPSTREAM_GIT_CREDENTIALS_PASSWORD', None),
    ])
    def test_string_configs(self, config_name, override_expected_value):
        """Test string configs to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        with unittest.mock.patch('os.environ', {config_name: 'unittest-value'}):
            assert getattr(terrareg.config.Config(), config_name) == (override_expected_value if override_expected_value is not None else 'unittest-value')

    @pytest.mark.parametrize('config_name, test_value, test_expected', [
        ('SENTRY_TRACES_SAMPLE_RATE', '1.523', 1.523)
    ])
    def test_custom_string_configs(self, config_name, test_value, test_expected):
        """Test string configs with custom values to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        with unittest.mock.patch('os.environ', {config_name: test_value}):
            assert getattr(terrareg.config.Config(), config_name) == test_expected

    @pytest.mark.parametrize('config_name', [
        'ADMIN_SESSION_EXPIRY_MINS',
        'LISTEN_PORT',
        'GIT_CLONE_TIMEOUT',
        'REDIRECT_DELETION_LOOKBACK_DAYS',
        'TERRAFORM_OIDC_IDP_SESSION_EXPIRY',
        'TERRAFORM_PRESIGNED_URL_EXPIRY_SECONDS',
    ])
    def test_integer_configs(self, config_name):
        """Test integer configs to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        with unittest.mock.patch('os.environ', {config_name: '582612'}):
            assert getattr(terrareg.config.Config(), config_name) == 582612

    @pytest.mark.parametrize('test_value,expected_value', [
        # Check empty value produces an empty array
        ('', []),
        # Ensure a single value produces an array with a single value
        ('unittest-value', ['unittest-value']),
        # Ensure a single value with a space produces an array with a single value
        ('unittest value', ['unittest value']),
        # Ensure a single value with a leading/trailing comma produces an array with a single value
        (',unittest-value,', ['unittest-value']),
        # Ensure multiple values produce an array with a both values
        ('unittest-value1,test-value2', ['unittest-value1', 'test-value2'])
    ])
    @pytest.mark.parametrize('config_name', [
        ('ALLOWED_PROVIDERS'),
        ('ANALYTICS_AUTH_KEYS'),
        ('PUBLISH_API_KEYS'),
        ('REQUIRED_MODULE_METADATA_ATTRIBUTES'),
        ('TRUSTED_NAMESPACES'),
        ('UPLOAD_API_KEYS'),
        ('VERIFIED_MODULE_NAMESPACES'),
        ('IGNORE_ANALYTICS_TOKEN_AUTH_KEYS'),
        ('OPENID_CONNECT_SCOPES'),
        ('EXAMPLE_FILE_EXTENSIONS'),
    ])
    def test_list_configs(self, config_name, test_value, expected_value):
        """Test list configs to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        # Check that input value produces expected list value
        with unittest.mock.patch('os.environ', {config_name: test_value}):
            assert getattr(terrareg.config.Config(), config_name) == expected_value

    @pytest.mark.parametrize('config_name,enum,expected_default', [
        ('MODULE_VERSION_REINDEX_MODE', terrareg.config.ModuleVersionReindexMode, terrareg.config.ModuleVersionReindexMode.LEGACY),
        ('SERVER', terrareg.config.ServerType, terrareg.config.ServerType.BUILTIN),
        ('ALLOW_MODULE_HOSTING', terrareg.config.ModuleHostingMode, terrareg.config.ModuleHostingMode.ALLOW),
        ('DEFAULT_UI_DETAILS_VIEW', terrareg.config.DefaultUiInputOutputView, terrareg.config.DefaultUiInputOutputView.TABLE),
    ])
    def test_enum_configs(self, config_name, enum, expected_default):
        """Test enum configs to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        check_dict = {None: expected_default}
        check_dict.update({
            i.value: i
            for i in enum
        })
        # Check that input value produces expected list value
        for test_env_value, expected_enum in check_dict.items():
            os_env = {} if test_env_value is None else {config_name: test_env_value}
            with unittest.mock.patch('os.environ', os_env):
                assert getattr(terrareg.config.Config(), config_name) == expected_enum

    @pytest.mark.parametrize('test_value,expected_value', [
        ('true', True),
        ('True', True),
        ('TRUE', True),
        ('false', False),
        ('False', False),
        ('FALSE', False),
        ('0', False),
        ('1', True),
        ('yes', True),
        ('no', False)
    ])
    @pytest.mark.parametrize('config_name', [
        'ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER',
        'ALLOW_CUSTOM_GIT_URL_MODULE_VERSION',
        'ALLOW_UNIDENTIFIED_DOWNLOADS',
        'AUTO_CREATE_MODULE_PROVIDER',
        'AUTO_CREATE_NAMESPACE',
        'AUTO_PUBLISH_MODULE_VERSIONS',
        'DEBUG',
        'DELETE_EXTERNALLY_HOSTED_ARTIFACTS',
        'DISABLE_TERRAREG_EXCLUSIVE_LABELS',
        'AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION',
        'ENABLE_SECURITY_SCANNING',
        'AUTOGENERATE_USAGE_BUILDER_VARIABLES',
        'THREADED',
        'INFRACOST_TLS_INSECURE_SKIP_VERIFY',
        'ENABLE_ACCESS_CONTROLS',
        'SAML2_DEBUG',
        'OPENID_CONNECT_DEBUG',
        "MANAGE_TERRAFORM_RC_FILE",
        'DISABLE_ANALYTICS',
        'ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION',
        'ALLOW_UNAUTHENTICATED_ACCESS',
        'AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES',
        'MODULE_VERSION_USE_GIT_COMMIT',
    ])
    def test_boolean_configs(self, config_name, test_value, expected_value):
        """Test boolean configs to ensure they are overridden with environment variables."""
        self.register_checked_config(config_name)
        # Check that input value generates the expected boolean value
        with unittest.mock.patch('os.environ', {config_name: test_value}):
            assert getattr(terrareg.config.Config(), config_name) is expected_value

