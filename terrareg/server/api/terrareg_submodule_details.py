
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Submodule


class ApiTerraregSubmoduleDetails(ErrorCatchingResource):
    """Interface to obtain submodule details."""

    def _get(self, namespace, name, provider, version, submodule):
        """Return details of submodule."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        submodule_obj = Submodule.get(module_version=module_version, module_path=submodule)

        return submodule_obj.get_terrareg_api_details()
