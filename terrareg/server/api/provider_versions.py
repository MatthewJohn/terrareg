
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model


class ApiProviderVersions(ErrorCatchingResource):

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_terraform_api')]

    def _get(self, namespace, provider):
        """Return provider version details."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        provider = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider is None:
            return self._get_404_response()

        return provider.get_versions_api_details()

