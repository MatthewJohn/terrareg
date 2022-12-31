
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.analytics


class ApiModuleProviderDownloadsSummary(ErrorCatchingResource):
    """Provide download summary for module provider."""

    def _get(self, namespace, name, provider):
        """Return list of download counts for module provider."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        return {
            "data": {
                "type": "module-downloads-summary",
                "id": module_provider.id,
                "attributes": terrareg.analytics.AnalyticsEngine.get_module_provider_download_stats(module_provider)
            }
        }
