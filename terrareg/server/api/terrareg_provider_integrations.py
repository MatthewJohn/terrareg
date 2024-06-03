
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper


class ApiTerraregProviderIntegrations(ErrorCatchingResource):
    """Interface to provide list of integration URLs"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, provider):
        """Return list of integration URLs"""
        _, provider_obj, error = self.get_provider_by_names(namespace, provider)
        if error:
            return error

        integrations = provider_obj.get_integrations()

        return [
            integrations[integration]
            for integration in ['import']
            if integration in integrations
        ]
