
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth
import terrareg.provider_source.factory


class GithubAuthStatus(ErrorCatchingResource):
    """Interface to provide details about current authentication status with Github"""

    def _get(self, provider_source: str):
        """Provide authentication status."""

        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return self._get_404_response()

        github_authenticated = False
        username = None
        if auth_method := terrareg.auth.GithubAuthMethod.get_current_instance():
            github_authenticated = True
            username = auth_method.get_username()

        return {
            "auth": github_authenticated,
            "username": username
        }
