

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search
import terrareg.auth_wrapper


class ApiTerraregMostDownloadedModuleProviderThisWeek(ErrorCatchingResource):
    """Return data for most downloaded module provider this week."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return most downloaded module this week"""
        module_provider = terrareg.module_search.ModuleSearch.get_most_downloaded_module_provider_this_Week()
        if not module_provider:
            return {}, 404

        return module_provider.get_latest_version().get_api_outline()