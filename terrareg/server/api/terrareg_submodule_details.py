
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregSubmoduleDetails(ErrorCatchingResource):
    """Interface to obtain submodule details."""

    def _get(self, namespace, name, provider, version, submodule):
        """Return details of submodule."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        submodule_obj = terrareg.models.Submodule.get(module_version=module_version, module_path=submodule)

        if not submodule_obj:
            return self._get_404_response()

        return submodule_obj.get_terrareg_api_details(
            request_domain=urllib.parse.urlparse(request.base_url).hostname)
