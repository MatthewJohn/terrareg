
from flask import request

import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregNamespaceDetails(ErrorCatchingResource):
    """Interface to obtain custom terrareg namespace details."""

    method_decorators = {
        # Limit post methods to users with FULL namespace permissions
        "post": [terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
            terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
            request_kwarg_map={'namespace': 'namespace'})]
    }

    def _get(self, namespace):
        """Return custom terrareg config for namespace."""
        namespace = terrareg.models.Namespace.get(namespace)
        if namespace is None:
            return self._get_404_response()
        return namespace.get_details()

    def _post(self, namespace):
        """Edit name/display name of a namespace"""
        namespace = terrareg.models.Namespace.get(namespace)
        if namespace is None:
            return self._get_404_response()
        
        namespace_name = request.json.get('name')
        display_name = request.json.get('display_name')
        csrf_token = request.json.get('csrf_token')

        terrareg.csrf.check_csrf_token(csrf_token)

        if namespace_name != namespace.name:
            namespace.update_name(namespace_name)
        if display_name != namespace.display_name:
            namespace.update_display_name(display_name)

        return {
            "name": namespace.name,
            "view_href": namespace.get_view_url(),
            "display_name": namespace.display_name
        }
