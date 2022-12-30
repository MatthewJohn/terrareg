
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace, ModuleProvider, ModuleVersion
from terrareg.analytics import AnalyticsEngine


class ApiTerraregGlobalStatsSummary(ErrorCatchingResource):
    """Provide global download stats for homepage."""

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        return {
            'namespaces': Namespace.get_total_count(),
            'modules': ModuleProvider.get_total_count(),
            'module_versions': ModuleVersion.get_total_count(),
            'downloads': AnalyticsEngine.get_total_downloads()
        }
