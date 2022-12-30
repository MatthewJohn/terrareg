
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
from terrareg.models import Namespace


class ApiTerraregNamespaces(ErrorCatchingResource):
    """Provide interface to obtain namespaces."""

    method_decorators = {
        "post": [auth_wrapper('is_admin')]
    }

    def _get(self):
        """Return list of namespaces."""
        namespaces = Namespace.get_all(only_published=False)

        return [
            {
                "name": namespace.name,
                "view_href": namespace.get_view_url()
            }
            for namespace in namespaces
        ]

    def _post(self):
        """Create namespace."""
        namespace_name = request.json.get('name')
        namespace = Namespace.create(name=namespace_name)

        return {
            "name": namespace.name,
            "view_href": namespace.get_view_url()
        }
