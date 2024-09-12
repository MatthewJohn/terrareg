
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.version import VERSION


class ApiTerraregVersion(ErrorCatchingResource):
    """Interface to obtain version of Terrareg."""

    def _get(self):
        """Return version"""

        return {
            "version": VERSION
        }
