
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiTerraregModuleProviderDetails(ErrorCatchingResource):
    """Interface to obtain module provider details."""

    def _get(self, namespace, name, provider):
        """Return details about module version."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        # If a version exists, obtain the details for that
        latest_version = module_provider.get_latest_version()
        if latest_version is not None:
            return latest_version.get_terrareg_api_details(request_domain=request.host)

        # Otherwise, return module provider details
        return module_provider.get_terrareg_api_details()
