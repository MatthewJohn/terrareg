

from enum import Enum


class AuthenticationType(Enum):
    """Determine the method of authentication."""
    NOT_CHECKED = 0
    NOT_AUTHENTICATED = 1
    AUTHENTICATION_TOKEN = 2
    SESSION_PASSWORD = 3
    SESSION_OPENID_CONNECT = 4
    SESSION_SAML = 5
