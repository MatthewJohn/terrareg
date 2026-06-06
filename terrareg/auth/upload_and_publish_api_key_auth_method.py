import terrareg.api_key
from .base_api_key_auth_method import BaseApiKeyAuthMethod


class UploadAndPublishApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for upload-and-publish API keys."""

    key_type = 'upload_and_publish'

    @classmethod
    def check_auth_state(cls):
        """Check if upload-and-publish API key is provided."""
        return cls._check_api_key([])

    @classmethod
    def is_enabled(cls):
        return terrareg.api_key.ApiKey.has_active_keys(terrareg.api_key.ApiKeyType.UPLOAD_AND_PUBLISH)

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
        """Get username of current user."""
        return 'Upload and Publish API Key'