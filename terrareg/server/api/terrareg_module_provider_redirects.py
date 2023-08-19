
from flask_restful import reqparse
from terrareg.models import ModuleProviderRedirect

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf


class ApiTerraregModuleProviderRedirects(ErrorCatchingResource):
    """Provide interface to delete module provider redirect."""

    def _get(self, namespace, name, provider):
        """Delete module provider."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        return [
            {
                "id": module_provider_redirect.pk,
                "namespace": module_provider_redirect.namespace.name,
                "module": module_provider_redirect.module_name,
                "provider": module_provider_redirect.provider_name
            }
            for module_provider_redirect in ModuleProviderRedirect.get_by_module_provider(module_provider)
        ]
