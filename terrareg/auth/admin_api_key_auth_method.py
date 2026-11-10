
from .base_admin_auth_method import BaseAdminAuthMethod
from .base_api_key_auth_method import BaseApiKeyAuthMethod
import terrareg.config
import terrareg.models


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    key_type = 'admin'

    @classmethod
    def get_valid_keys(cls):
        return [terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN]

    @classmethod
    def check_auth_state(cls):
        """Check if admin API key is provided"""
        return cls._check_api_key(cls.get_valid_keys())

    @classmethod
    def is_enabled(cls):
        """Whether admin API key auth is configured."""
        return bool(terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN or terrareg.models.ApiKey.has_active_keys(terrareg.models.ApiKeyType.ADMIN))
