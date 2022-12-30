
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.csrf
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type


class ApiTerraregModuleVersionDelete(ErrorCatchingResource):
    """Provide interface to delete module version."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper(
        'check_namespace_access',
        terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
        request_kwarg_map={'namespace': 'namespace'})
    ]

    def _delete(self, namespace, name, provider, version):
        """Delete module version."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        module_version_id = module_version.id

        module_version.delete()

        return {
            'status': 'Success'
        }