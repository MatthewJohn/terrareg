

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search


class ApiTerraregMostDownloadedModuleProviderThisWeek(ErrorCatchingResource):
    """Return data for most downloaded module provider this week."""

    def _get(self):
        """Return most downloaded module this week"""
        module_provider = terrareg.module_search.ModuleSearch.get_most_downloaded_module_provider_this_Week()
        if not module_provider:
            return {}, 404

        return module_provider.get_latest_version().get_api_outline()