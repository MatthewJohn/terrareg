
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model


class ApiProvider(ErrorCatchingResource):

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, provider, version=None):
        """Return provider details."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)

        provider = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider is None:
            return self._get_404_response()

        provider_version = None
        if version is not None:
            provider_version = terrareg.provider_version_model.ProviderVersion.get(
                provider=provider,
                version=version
            )
        else:
            provider_version = provider.get_latest_version()

        if provider_version is None:
            return self._get_404_response()

        return provider_version.get_api_details()

