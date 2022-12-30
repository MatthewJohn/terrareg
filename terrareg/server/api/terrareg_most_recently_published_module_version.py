
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.module_search import ModuleSearch


class ApiTerraregMostRecentlyPublishedModuleVersion(ErrorCatchingResource):
    """Return data for most recently published module version."""

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        module_version = ModuleSearch.get_most_recently_published()
        if not module_version:
            return {}, 404
        return module_version.get_api_outline()
