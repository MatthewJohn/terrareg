
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace


class ApiModuleProviderDetails(ErrorCatchingResource):

    def _get(self, namespace, name, provider):
        """Return list of version."""

        namespace, _ = Namespace.extract_analytics_token(namespace)
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider, create=True)
        if error:
            return self._get_404_response()
        module_version = module_provider.get_latest_version()

        if not module_version:
            return self._get_404_response()

        return module_version.get_api_details()

