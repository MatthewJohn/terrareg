
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.auth


class ApiTerraregIsAuthenticated(ErrorCatchingResource):
    """Interface to return whether user is authenticated as an admin."""

    def _get(self):
        """Return information about current user."""
        auth_method = terrareg.auth.AuthFactory().get_current_auth_method()
        return {
            "read_access": auth_method.can_access_read_api(),
            "authenticated": auth_method.is_authenticated(),
            "site_admin": auth_method.is_admin(),
            "namespace_permissions": {
                namespace.name: permission.value
                for namespace, permission in auth_method.get_all_namespace_permissions().items()
            }
        }
