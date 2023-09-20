
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregExampleFileList(ErrorCatchingResource):
    """Interface to obtain list of example files."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider, version, example):
        """Return list of files available in example."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_obj = terrareg.models.Example(module_version=module_version, module_path=example)

        if example_obj is None:
            return self._get_404_response()

        return [
            {
                'filename': example_file.file_name,
                'path': example_file.path,
                'content_href': '/v1/terrareg/modules/{module_version_id}/example/files/{file_path}'.format(
                    module_version_id=module_version.id,
                    file_path=example_file.path)
            }
            for example_file in sorted(example_obj.get_files())
        ]
