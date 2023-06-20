
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregExampleDetails(ErrorCatchingResource):
    """Interface to obtain example details."""

    def _get(self, namespace, name, provider, version, example):
        """Return details of example."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_obj = terrareg.models.Example.get(module_version=module_version, module_path=example)
        if example_obj is None:
            return self._get_404_response()

        return example_obj.get_terrareg_api_details(
            request_domain=urllib.parse.urlparse(request.base_url).hostname)
