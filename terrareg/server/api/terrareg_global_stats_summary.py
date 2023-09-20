
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.analytics
import terrareg.auth_wrapper


class ApiTerraregGlobalStatsSummary(ErrorCatchingResource):
    """Provide global download stats for homepage."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return number of namespaces, modules, module versions and downloads"""
        return {
            'namespaces': terrareg.models.Namespace.get_total_count(),
            'modules': terrareg.models.ModuleProvider.get_total_count(),
            'module_versions': terrareg.models.ModuleVersion.get_total_count(),
            'downloads': terrareg.analytics.AnalyticsEngine.get_total_downloads()
        }
