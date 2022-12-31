
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiModuleVersionDetails(ErrorCatchingResource):

    def _get(self, namespace, name, provider, version):
        """Return list of version."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return self._get_404_response()

        return module_version.get_api_details()

