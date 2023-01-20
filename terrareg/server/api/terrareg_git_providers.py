
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregGitProviders(ErrorCatchingResource):
    """Interface to obtain git provider configurations."""

    def _get(self):
        """Return list of git providers"""
        return [
            {
                'id': git_provider.pk,
                'name': git_provider.name
            }
            for git_provider in terrareg.models.GitProvider.get_all()
        ]
