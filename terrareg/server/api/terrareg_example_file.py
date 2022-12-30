
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import ExampleFile


class ApiTerraregExampleFile(ErrorCatchingResource):
    """Interface to obtain content of example file."""

    def _get(self, namespace, name, provider, version, example_file):
        """Return conent of example file in example module."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_file_obj = ExampleFile.get_by_path(module_version=module_version, file_path=example_file)

        if example_file_obj is None:
            return {'message': 'Example file object does not exist.'}

        return example_file_obj.get_content(server_hostname=request.host)
