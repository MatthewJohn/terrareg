
import terrareg.config
import terrareg.auth
from .base_auth_method import BaseAuthMethod


class NotAuthenticated(BaseAuthMethod):
    """Base auth method for unauthenticated users"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    def is_authenticated(self):
        """Whether user is authenticated"""
        return False

    def check_namespace_access(self, permission_type, namespace):
        """Unauthenticated users have no namespace access."""
        return False

    @classmethod
    def check_auth_state(cls):
        """Always return True as a last-catch auth mechanism"""
        return True

    @classmethod
    def is_enabled(cls):
        return True

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        # If API key authentication is not configured for publishing modules,
        # RBAC is not enabled and unauthenticated access is enabled,
        # allow unauthenticated access
        if ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and
                (not terrareg.auth.PublishApiKeyAuthMethod.is_enabled()) and
                terrareg.config.Config().ALLOW_UNAUTHENTICATED_ACCESS):
            return True
        return False

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        # If API key authentication is not configured for uploading modules,
        # RBAC is not enabled and unauthenticated access is enabled,
        # allow unauthenticated access
        if ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and
                (not terrareg.auth.UploadApiKeyAuthMethod.is_enabled()) and
                terrareg.config.Config().ALLOW_UNAUTHENTICATED_ACCESS):
            return True
        return False

    def get_username(self):
        """Get username of current user"""
        return 'Unauthenticated User'

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # Unauthenticated users can only access 'read' APIs
        # if global anonymous access is allowed
        return terrareg.config.Config().ALLOW_UNAUTHENTICATED_ACCESS
