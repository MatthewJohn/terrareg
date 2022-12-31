
from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiTerraregModuleProviderIntegrations(ErrorCatchingResource):
    """Interface to provide list of integration URLs"""

    def _get(self, namespace, name, provider):
        """Return list of integration URLs"""
        _, _ , module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        integrations = module_provider.get_integrations()

        return [
            integrations[integration]
            for integration in ['upload', 'import', 'hooks_bitbucket', 'hooks_github', 'hooks_gitlab', 'publish']
            if integration in integrations
        ]
