
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth


class GithubAuthStatus(ErrorCatchingResource):
    """Interface to provide details about current authentication status with Github"""

    def _get(self):
        """Provide authentication status."""
        github_authenticated = False
        username = None
        if terrareg.auth.GithubAuthMethod.is_enabled() and (auth_method := terrareg.auth.GithubAuthMethod.get_current_instance()):
            github_authenticated = True
            username = auth_method.get_username()

        return {
            "auth": github_authenticated,
            "username": username
        }
