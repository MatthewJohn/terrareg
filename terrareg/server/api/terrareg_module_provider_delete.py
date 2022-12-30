
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.csrf import check_csrf_token


class ApiTerraregModuleProviderDelete(ErrorCatchingResource):
    """Provide interface to delete module provider."""

    method_decorators = [auth_wrapper('check_namespace_access',
                                      UserGroupNamespacePermissionType.FULL,
                                      request_kwarg_map={'namespace': 'namespace'})]

    def _delete(self, namespace, name, provider):
        """Delete module provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )

        args = parser.parse_args()

        check_csrf_token(args.csrf_token)

        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        module_provider_id = module_provider.id

        module_provider.delete()
