
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregModuleVersionDetails(ErrorCatchingResource):
    """Interface to obtain module verison details."""

    def _get(self, namespace, name, provider, version=None):
        """Return details about module version."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        if version is not None:
            module_version = terrareg.models.ModuleVersion.get(module_provider=module_provider, version=version)
        else:
            module_version = module_provider.get_latest_version()

        if module_version is None:
            return self._get_404_response()

        return module_version.get_terrareg_api_details(
            request_domain=urllib.parse.urlparse(request.base_url).hostname)
