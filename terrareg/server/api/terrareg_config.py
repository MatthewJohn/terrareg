
import terrareg.config
import terrareg.openid_connect
import terrareg.saml
import terrareg.auth
from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiTerraregConfig(ErrorCatchingResource):
    """Endpoint to return config used by UI."""

    def _get(self):
        """Return config."""
        config = terrareg.config.Config()
        return {
            'TRUSTED_NAMESPACE_LABEL': config.TRUSTED_NAMESPACE_LABEL,
            'CONTRIBUTED_NAMESPACE_LABEL': config.CONTRIBUTED_NAMESPACE_LABEL,
            'VERIFIED_MODULE_LABEL': config.VERIFIED_MODULE_LABEL,
            'ANALYTICS_TOKEN_PHRASE': config.ANALYTICS_TOKEN_PHRASE,
            'ANALYTICS_TOKEN_DESCRIPTION': config.ANALYTICS_TOKEN_DESCRIPTION,
            'EXAMPLE_ANALYTICS_TOKEN': config.EXAMPLE_ANALYTICS_TOKEN,
            'ALLOW_MODULE_HOSTING': config.ALLOW_MODULE_HOSTING,
            'UPLOAD_API_KEYS_ENABLED': bool(config.UPLOAD_API_KEYS),
            'PUBLISH_API_KEYS_ENABLED': bool(config.PUBLISH_API_KEYS),
            'DISABLE_TERRAREG_EXCLUSIVE_LABELS': config.DISABLE_TERRAREG_EXCLUSIVE_LABELS,
            'ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER': config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER,
            'ALLOW_CUSTOM_GIT_URL_MODULE_VERSION': config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION,
            'SECRET_KEY_SET': bool(config.SECRET_KEY),
            'ADDITIONAL_MODULE_TABS': config.ADDITIONAL_MODULE_TABS,
            'OPENID_CONNECT_ENABLED': terrareg.openid_connect.OpenidConnect.is_enabled(),
            'OPENID_CONNECT_LOGIN_TEXT': config.OPENID_CONNECT_LOGIN_TEXT,
            'SAML_ENABLED': terrareg.saml.Saml2.is_enabled(),
            'SAML_LOGIN_TEXT': config.SAML2_LOGIN_TEXT,
            'ADMIN_LOGIN_ENABLED': terrareg.auth.AdminApiKeyAuthMethod.is_enabled(),
            'AUTO_CREATE_NAMESPACE': config.AUTO_CREATE_NAMESPACE,
            'AUTO_CREATE_MODULE_PROVIDER': config.AUTO_CREATE_MODULE_PROVIDER
        }