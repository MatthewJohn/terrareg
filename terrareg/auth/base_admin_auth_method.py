

import terrareg.config
from .base_auth_method import BaseAuthMethod


class BaseAdminAuthMethod(BaseAuthMethod):
    """Base auth method admin authentication"""

    def is_admin(self):
        """Return whether user is an admin"""
        return True

    def is_built_in_admin(self):
        """Whether user is the built-in admin"""
        return True

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        return True

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        # Allow full access to all namespaces
        return True

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN)

    def get_username(self):
        """Get username of current user"""
        return 'Built-in admin'

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        return True
