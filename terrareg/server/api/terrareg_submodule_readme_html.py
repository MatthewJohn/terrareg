
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregSubmoduleReadmeHtml(ErrorCatchingResource):
    """Interface to obtain submodule REAMDE in HTML format."""

    def _get(self, namespace, name, provider, version, submodule):
        """Return HTML formatted README of submodule."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        submodule_obj = terrareg.models.Submodule.get(module_version=module_version, module_path=submodule)

        return submodule_obj.get_readme_html(server_hostname=urllib.parse.urlparse(request.base_url).hostname)
