
from .base_admin_auth_method import BaseAdminAuthMethod
from .base_api_key_auth_method import BaseApiKeyAuthMethod
import terrareg.config


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if admin API key is provided"""
        return cls._check_api_key([terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN])
