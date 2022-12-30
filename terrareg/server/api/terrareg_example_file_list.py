
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Example


class ApiTerraregExampleFileList(ErrorCatchingResource):
    """Interface to obtain list of example files."""

    def _get(self, namespace, name, provider, version, example):
        """Return list of files available in example."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        example_obj = Example(module_version=module_version, module_path=example)

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
