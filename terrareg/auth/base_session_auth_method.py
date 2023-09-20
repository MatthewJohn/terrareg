
import flask

import terrareg.config
import terrareg.models
from .base_auth_method import BaseAuthMethod


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

        return flask.session.get(cls.SESSION_AUTH_TYPE_KEY, None) == cls.SESSION_AUTH_TYPE_VALUE.value

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
                not terrareg.models.Session.check_session(flask.session.get('session_id', None)) or
                not flask.session.get('is_admin_authenticated', False)):
            return False

        # Ensure session type is set to the current session and session is valid
        return cls.check_session_auth_type() and cls.check_session()

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # Logged in SSO users can access any 'read' APIs
        return True
