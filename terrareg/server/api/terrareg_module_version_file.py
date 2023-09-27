
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregModuleVersionFile(ErrorCatchingResource):
    """Interface to obtain content of module version file."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider, version, path):
        """Return conent of module version file."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        module_version_file = terrareg.models.ModuleVersionFile.get(module_version=module_version, path=path)

        if module_version_file is None:
            return {'message': 'Module version file does not exist.'}, 400

        return module_version_file.get_content()
