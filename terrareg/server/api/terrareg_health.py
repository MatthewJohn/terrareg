
from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiTerraregHealth(ErrorCatchingResource):
    """Endpoint to return 200 when healthy."""

    def _get(self):
        """Return static 200"""
        return {
            "message": "Ok"
        }
