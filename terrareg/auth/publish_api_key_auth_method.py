
from typing import Optional, List

import terrareg.config
import terrareg.api_key
import terrareg.api_key_type
from .base_api_key_auth_method import BaseApiKeyAuthMethod


class PublishApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for publish API key"""

    key_type = terrareg.api_key_type.ApiKeyType.PUBLISH

    @classmethod
    def get_static_keys(cls) -> List[str]:
        return terrareg.config.Config().PUBLISH_API_KEYS

    def can_publish_module_version(self, namespace: Optional[str]) -> bool:
        """Whether user can publish module version within a namespace."""
        key = self.matched_api_key
        if key is not None and key.namespace is not None:
            return key.namespace == namespace
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        return False

    def get_username(self):
        """Get username of current user"""
        return 'Publish API Key'
