
import terrareg.config
import terrareg.models
from .base_api_key_auth_method import BaseApiKeyAuthMethod


class UploadApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for upload API key"""

    key_type = 'upload'

    @classmethod
    def get_valid_keys(cls):
        return terrareg.config.Config().UPLOAD_API_KEYS

    @classmethod
    def check_auth_state(cls):
        """Check if upload API key is provided"""
        return cls._check_api_key(cls.get_valid_keys())

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().UPLOAD_API_KEYS or terrareg.models.ApiKey.has_active_keys(terrareg.models.ApiKeyType.UPLOAD))

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        key = self.matched_api_key
        if key is not None and key.namespace is not None:
            return key.namespace == namespace
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        return False

    def get_username(self):
        """Get username of current user"""
        return 'Upload API Key'

