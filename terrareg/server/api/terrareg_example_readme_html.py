
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregExampleReadmeHtml(ErrorCatchingResource):
    """Interface to obtain example REAMDE in HTML format."""

    def _get(self, namespace, name, provider, version, example):
        """Return HTML formatted README of example."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_obj = terrareg.models.Example.get(module_version=module_version, module_path=example)

        return example_obj.get_readme_html(server_hostname=urllib.parse.urlparse(request.base_url).hostname)
