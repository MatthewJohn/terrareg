
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models


class ApiTerraregNamespaces(ErrorCatchingResource):
    """Provide interface to obtain namespaces."""

    method_decorators = {
        "post": [terrareg.auth_wrapper.auth_wrapper('is_admin')]
    }

    def _get(self):
        """Return list of namespaces."""
        namespaces = terrareg.models.Namespace.get_all(only_published=False)

        return [
            {
                "name": namespace.name,
                "view_href": namespace.get_view_url(),
                "display_name": namespace.display_name
            }
            for namespace in namespaces
        ]

    def _post(self):
        """Create namespace."""
        namespace_name = request.json.get('name')
        namespace = terrareg.models.Namespace.create(name=namespace_name)

        return {
            "name": namespace.name,
            "view_href": namespace.get_view_url()
        }
