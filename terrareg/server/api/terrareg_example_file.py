
import urllib.parse

from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregExampleFile(ErrorCatchingResource):
    """Interface to obtain content of example file."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider, version, example_file):
        """Return content of example file in example module."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_file_obj = terrareg.models.ExampleFile.get_by_path(module_version=module_version, file_path=example_file)

        if example_file_obj is None:
            return {'message': 'Example file object does not exist.'}

        return example_file_obj.get_content(server_hostname=urllib.parse.urlparse(request.base_url).hostname)
