
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
import terrareg.auth


class ApiTerraregIsAuthenticated(ErrorCatchingResource):
    """Interface to teturn whether user is authenticated as an admin."""

    method_decorators = [auth_wrapper('is_authenticated')]

    def _get(self):
        """Return information about current user."""
        auth_method = terrareg.auth.AuthFactory().get_current_auth_method()
        return {
            'authenticated': True,
            'site_admin': auth_method.is_admin(),
            'namespace_permissions': {
                namespace.name: permission.value
                for namespace, permission in auth_method.get_all_namespace_permissions().items()
            }
        }
