
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace


class ApiTerraregNamespaceDetails(ErrorCatchingResource):
    """Interface to obtain custom terrareg namespace details."""

    def _get(self, namespace):
        """Return custom terrareg config for namespace."""
        namespace = Namespace.get(namespace)
        if namespace is None:
            return self._get_404_response()
        return namespace.get_details()