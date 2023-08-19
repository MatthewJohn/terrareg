
from flask_restful import reqparse
from terrareg.models import ModuleProviderRedirect

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf


class ApiTerraregModuleProviderRedirectDelete(ErrorCatchingResource):
    """Provide interface to delete module provider redirect."""

    method_decorators = [
        terrareg.auth_wrapper.auth_wrapper(
            'check_namespace_access',
            terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
            request_kwarg_map={'namespace': 'namespace'}
        )
    ]

    def _delete(self, namespace, name, provider, module_provider_redirect_id):
        """Delete module provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )
        parser.add_argument(
            'force',
            type=bool,
            required=False,
            default=False,
            location='json',
            help='Whether to force deletion of provider, ignoring check for whether the redirect is in use'
        )

        args = parser.parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        if not (module_provider_redirect := ModuleProviderRedirect(module_provider_redirect_id)) or module_provider_redirect.module_provider_id != module_provider.pk:
            return {
                'status': 'Error',
                'message': 'Module provider redirect does not exist'
            }
        
        module_provider_redirect.delete(force=args.force)
