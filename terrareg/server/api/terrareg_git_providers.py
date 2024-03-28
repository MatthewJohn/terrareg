
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregGitProviders(ErrorCatchingResource):
    """Interface to obtain git provider configurations."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return list of git providers"""
        return [
            {
                "id": git_provider.pk,
                "name": git_provider.name,
                "git_path_template": git_provider.git_path_template,
            }
            for git_provider in terrareg.models.GitProvider.get_all()
        ]
