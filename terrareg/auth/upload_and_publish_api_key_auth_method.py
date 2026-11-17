
from typing import List, Optional

import terrareg.user_group_namespace_permission_type
import terrareg.api_key
import terrareg.api_key_type
from .base_api_key_auth_method import BaseApiKeyAuthMethod


class UploadAndPublishApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for upload-and-publish API keys."""

    key_type = terrareg.api_key_type.ApiKeyType.UPLOAD_AND_PUBLISH

    @classmethod
    def get_static_keys(cls) -> List[str]:
        """Check if upload-and-publish API key is provided."""
        return []


    def can_upload_module_version(self, namespace: Optional[str]):
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

    def check_namespace_access(self, permission_type: terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType, namespace: Optional[str]):
        """Check access level to a given namespace."""
        return False

    def get_username(self):
        """Get username of current user."""
        return 'Upload and Publish API Key'