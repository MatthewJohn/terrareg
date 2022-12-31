
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search


class ApiTerraregMostRecentlyPublishedModuleVersion(ErrorCatchingResource):
    """Return data for most recently published module version."""

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        module_version = terrareg.module_search.ModuleSearch.get_most_recently_published()
        if not module_version:
            return {}, 404
        return module_version.get_api_outline()
