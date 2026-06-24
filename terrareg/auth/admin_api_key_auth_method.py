
from typing import List

from .base_admin_auth_method import BaseAdminAuthMethod
from .base_api_key_auth_method import BaseApiKeyAuthMethod
import terrareg.config
import terrareg.api_key
import terrareg.api_key_type


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    key_type = terrareg.api_key_type.ApiKeyType.ADMIN

    @classmethod
    def get_static_keys(cls) -> List[str]:
        if admin_token := terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN:
            return [admin_token]
        return []
