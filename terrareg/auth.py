
import datetime

from flask import g, request, session

import terrareg.config
from terrareg.models import Session
import terrareg.openid_connect
import terrareg.saml


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
                    NotAuthenticated]:
            if not cls.is_enabled():
                continue

            if auth_method := cls.get_current_instance():
                setattr(g, AuthFactory.FLASK_GLOBALS_AUTH_KEY, auth_method)
                return auth_method

        raise Exception('Unable to determine current auth type - not caught by NotAuthenticated')


class BaseAuthMethod:
    """Base auth method"""

    def is_built_in_admin(self):
        """Whether user is the built-in admin"""
        return False

    def is_admin(self):
        """Whether user is an admin"""
        return False

    def is_authenticated(self):
        """Whether user is authenticated"""
        return True

    @classmethod
    def is_enabled(cls):
        """Whether authentication method is enabled"""
        raise NotImplementedError

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        raise NotImplementedError

    @classmethod
    def get_current_instance(cls):
        """Get instance of auth method, if user is authenticated"""
        return cls() if cls.check_auth_state() else None

    @classmethod
    def check_auth_state(cls):
        """Check whether user is logged in using this method and return instance of object"""
        raise NotImplementedError

    def check_namespace_access(self, namespace):
        """Check level of access to namespace"""
        raise NotImplementedError


class NotAuthenticated(BaseAuthMethod):
    """Base auth method for unauthenticated users"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return True

    def is_authenticated(self):
        """Whether user is authenticated"""
        return False

    def check_namespace_access(self, namespace):
        """Unauthenticated users have no namespace access."""
        return None

    @classmethod
    def check_auth_state(cls):
        """Always return True as a last-catch auth mechanism"""
        return True

    @classmethod
    def is_enabled(cls):
        return True


class BaseAdminAuthMethod(BaseAuthMethod):
    """Base auth method admin authentication"""

    def is_admin(self):
        """Return whether user is an admin"""
        return True

    def is_built_in_admin(self):
        """Whether user is the built-in admin"""
        return True

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN)


class BaseApiKeyAuthMethod(BaseAuthMethod):
    """Base auth method for API key-based authentication"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    @classmethod
    def _check_api_key(cls, valid_keys):
        """Whether whether API key is valid"""
        if not isinstance(valid_keys, list):
            return False

        # Obtain API key from request, ensuring that it is
        # not empty
        actual_key = request.headers.get('X-Terrareg-ApiKey', '')
        if not actual_key:
            return False
    
        # Iterate through API keys, ensuring 
        for valid_key in valid_keys:
            # Ensure the valid key is not empty:
            if actual_key and actual_key == valid_key:
                return True
        return False


class BaseSessionAuthMethod(BaseAuthMethod):
    """Base auth method for session-based authentication"""

    SESSION_AUTH_TYPE_KEY = 'authentication_type'
    SESSION_AUTH_TYPE_VALUE = None

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return True

    @classmethod
    def check_session_auth_type(cls):
        """Check if the current type of authenticate is set in session."""
        # Check if auth type value has been overriden
        if cls.SESSION_AUTH_TYPE_VALUE is None:
            raise NotImplementedError

        return session.get(cls.SESSION_AUTH_TYPE_KEY, None) == cls.SESSION_AUTH_TYPE_VALUE

    @classmethod
    def check_session(cls):
        """Check if auth-specific session is valid."""
        raise NotImplementedError

    @classmethod
    def check_auth_state(cls):
        """Check if session is valid."""
        # Ensure session secret key is set,
        # session ID is present and valid and
        # is_admin_authenticated session is set
        if (not terrareg.config.Config().SECRET_KEY or
                not Session.check_session(session.get('session_id', None)) or
                not session.get('is_admin_authenticated', False)):
            return False

        # Ensure session type is set to the current session and session is valid
        return cls.check_session_auth_type() and cls.check_session()


class UploadApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for upload API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if upload API key is provided"""
        return cls._check_api_key(terrareg.config.Config().UPLOAD_API_KEYS)

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().UPLOAD_API_KEYS)


class PublishApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for publish API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if upload API key is provided"""
        return cls._check_api_key(terrareg.config.Config().PUBLISH_API_KEYS)

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().PUBLISH_API_KEYS)


class SamlAuthMethod(BaseSessionAuthMethod):
    """Auth method for SAML authentication"""

    SESSION_AUTH_TYPE_VALUE = 5

    @classmethod
    def check_session(cls):
        """Check SAML session is valid"""
        auth = terrareg.saml.Saml2.initialise_request_auth_object(request)
        return auth.is_authenticated()

    @classmethod
    def is_enabled(cls):
        return terrareg.saml.Saml2.is_enabled()

class OpenidConnectAuthMethod(BaseSessionAuthMethod):
    """Auth method for OpenID authentication"""
    
    SESSION_AUTH_TYPE_VALUE = 4
    
    @classmethod
    def check_session(cls):
        """Check OpenID session"""
        # Check OpenID connect expiry time
        if datetime.datetime.now() >= datetime.datetime.fromtimestamp(
                session.get('openid_connect_expires_at', 0)):
            return False

        try:
            terrareg.openid_connect.OpenidConnect.validate_session_token(session.get('openid_connect_id_token'))
        except:
            # Catch any exceptions when validating session and return False
            return False
        # If validate did not raise errors, return True
        return True

    @classmethod
    def is_enabled(cls):
        return terrareg.openid_connect.OpenidConnect.is_enabled()


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if admin API key is provided"""
        return cls._check_api_key([terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN])


class AdminSessionAuthMethod(BaseAdminAuthMethod, BaseSessionAuthMethod):
    """Auth method for admin session"""

    SESSION_AUTH_TYPE_VALUE = 3

    @classmethod
    def check_session(cls):
        """Check admin session"""
        # There are no additional attributes to check
        return True
