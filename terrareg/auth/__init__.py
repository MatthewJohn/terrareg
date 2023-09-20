
from flask import g

from .admin_api_key_auth_method import AdminApiKeyAuthMethod
from .admin_session_auth_method import AdminSessionAuthMethod
from .upload_api_key_auth_method import UploadApiKeyAuthMethod
from .publish_api_key_auth_method import PublishApiKeyAuthMethod
from .saml_auth_method import SamlAuthMethod
from .openid_auth_method import OpenidConnectAuthMethod
from .terraform_oidc_auth_method import TerraformOidcAuthMethod
from .terraform_analytics_auth_key_auth_method import TerraformAnalyticsAuthKeyAuthMethod
from .terraform_ignore_analytics_auth_method import TerraformIgnoreAnalyticsAuthMethod
from .terraform_internal_extraction import TerraformInternalExtractionAuthMethod
from .not_authenticated import NotAuthenticated
from .authentication_type import AuthenticationType


class AuthFactory:
    """
    Factory for obtaining current user authentication method
    and authentication providing wrappers
    """

    FLASK_GLOBALS_AUTH_KEY = 'user_auth'

    def get_current_auth_method(self):
        """Obtain user's current login state"""
        # Check if current authenticate type has been determined
        if (current_auth_method := g.get(AuthFactory.FLASK_GLOBALS_AUTH_KEY, None)) is not None:
            return current_auth_method

        # Iterate through auth methods, checking if user is authenticated
        for cls in [AdminApiKeyAuthMethod,
                    AdminSessionAuthMethod,
                    UploadApiKeyAuthMethod,
                    PublishApiKeyAuthMethod,
                    SamlAuthMethod,
                    OpenidConnectAuthMethod,
                    TerraformOidcAuthMethod,
                    TerraformAnalyticsAuthKeyAuthMethod,
                    TerraformIgnoreAnalyticsAuthMethod,
                    TerraformInternalExtractionAuthMethod,
                    NotAuthenticated]:
            if not cls.is_enabled():
                continue

            if auth_method := cls.get_current_instance():
                setattr(g, AuthFactory.FLASK_GLOBALS_AUTH_KEY, auth_method)
                return auth_method

        raise Exception('Unable to determine current auth type - not caught by NotAuthenticated')
