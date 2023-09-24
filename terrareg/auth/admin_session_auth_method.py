
from .base_admin_auth_method import BaseAdminAuthMethod
from .base_session_auth_method import BaseSessionAuthMethod
from .authentication_type import AuthenticationType


class AdminSessionAuthMethod(BaseAdminAuthMethod, BaseSessionAuthMethod):
    """Auth method for admin session"""

    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_PASSWORD

    @classmethod
    def check_session(cls):
        """Check admin session"""
        # There are no additional attributes to check
        return True
