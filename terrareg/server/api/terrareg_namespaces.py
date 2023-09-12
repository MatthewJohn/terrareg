
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource, api_error
from terrareg.errors import (
    DuplicateNamespaceDisplayNameError, NamespaceAlreadyExistsError,
    InvalidNamespaceNameError, InvalidNamespaceDisplayNameError
)
import terrareg.auth_wrapper
import terrareg.models
import terrareg.csrf
import terrareg.auth_wrapper


class ApiTerraregNamespaces(ErrorCatchingResource):
    """Provide interface to obtain namespaces."""

    method_decorators = {
        "get": [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')],
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
        display_name = request.json.get('display_name')
        csrf_token = request.json.get('csrf_token')

        terrareg.csrf.check_csrf_token(csrf_token)

        try:
            namespace = terrareg.models.Namespace.create(
                name=namespace_name,
                display_name=display_name)
        except (InvalidNamespaceNameError, NamespaceAlreadyExistsError,
                InvalidNamespaceDisplayNameError, DuplicateNamespaceDisplayNameError) as exc:
            return api_error(str(exc)), 400

        return {
            "name": namespace.name,
            "view_href": namespace.get_view_url(),
            "display_name": namespace.display_name
        }
