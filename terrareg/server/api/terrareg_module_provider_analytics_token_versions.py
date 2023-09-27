
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.analytics
import terrareg.auth_wrapper


class ApiTerraregModuleProviderAnalyticsTokenVersions(ErrorCatchingResource):
    """Provide download summary for module provider."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider):
        """Return list of download counts for module provider."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error
        return terrareg.analytics.AnalyticsEngine.get_module_provider_token_versions(module_provider)