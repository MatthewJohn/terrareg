
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import GitProvider


class ApiTerraregGitProviders(ErrorCatchingResource):
    """Interface to obtain git provider configurations."""

    def _get(self):
        """Return list of git providers"""
        return [
            {
                'id': git_provider.pk,
                'name': git_provider.name
            }
            for git_provider in GitProvider.get_all()
        ]
