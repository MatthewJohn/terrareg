
import terrareg.models
from .base_api_key_auth_method import BaseApiKeyAuthMethod


class ModuleFullApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for module-full API key (upload + publish)"""

    @classmethod
    def check_auth_state(cls):
        """Check if module-full API key is provided"""
        return cls._check_api_key([], key_type=terrareg.models.ApiKeyType.MODULE_FULL)

    @classmethod
    def is_enabled(cls):
        return terrareg.models.ApiKey.has_active_keys(terrareg.models.ApiKeyType.MODULE_FULL)

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        key = self.matched_api_key
        if key is not None and key.namespace is not None:
            return key.namespace == namespace
        return True

    def can_publish_module_version(self, namespace):
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
        return 'Module Full API Key'
