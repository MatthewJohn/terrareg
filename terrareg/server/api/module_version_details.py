
import terrareg.analytics
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiModuleVersionDetails(ErrorCatchingResource):

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider, version):
        """Return list of version."""

        namespace, _ = terrareg.analytics.AnalyticsEngine.extract_analytics_token(namespace)
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return self._get_404_response()

        return module_version.get_api_details()

