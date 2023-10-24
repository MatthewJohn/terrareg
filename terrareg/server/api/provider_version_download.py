
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.provider_binary_types
import terrareg.provider_version_binary_model


class ApiProviderVersionDownload(ErrorCatchingResource):

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_terraform_api')]

    def _get(self, namespace, provider, version, os, arch):
        """Return provider details."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)

        if namespace_obj is None:
            return self._get_404_response()

        provider = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider is None:
            return self._get_404_response()

        provider_version = terrareg.provider_version_model.ProviderVersion.get(
            provider=provider,
            version=version
        )

        if provider_version is None:
            return self._get_404_response()

        try:
            operating_system_type = terrareg.provider_binary_types.ProviderBinaryOperatingSystemType(os)
        except ValueError:
            return self._get_404_response()

        try:
            architecture_type = terrareg.provider_binary_types.ProviderBinaryArchitectureType(arch)
        except ValueError:
            return self._get_404_response()

        binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=operating_system_type,
            architecture_type=architecture_type
        )
        if not binary:
            return self._get_404_response()

        return binary.get_api_outline()

