
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf
import terrareg.errors
import terrareg.models
import terrareg.database


class ApiGpgKey(ErrorCatchingResource):
    """Provide interface to create GPG Keys."""

    method_decorators = [
        terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
            terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
            # Obtain namespace for check from post data
            kwarg_values={'namespace': lambda: request.get_json().get("data", {}).get("attributes").get("namespace")})
    ]


    def _post(self):
        """Handle update to settings."""
        data = request.get_json().get("data", {})
        attributes = data.get("attributes", {})
        namespace_name = attributes.get("namespace")
        ascii_armor = attributes.get("ascii-armor")
        csrf = data.get("csrf_token")


        terrareg.csrf.check_csrf_token(csrf)

        if not (namespace := terrareg.models.Namespace.get(name=namespace_name)):
            return {'message': 'Namespace does not exist'}, 400

        gpg_key = terrareg.models.GpgKey.create(namespace=namespace, ascii_armor=ascii_armor)

        return {
            "data": gpg_key.get_api_data()
        }
