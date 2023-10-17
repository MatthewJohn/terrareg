
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.repository_model


class GithubRepositories(ErrorCatchingResource):
    """Interface to provide details about current Github repositories for the logged in user"""

    def _get(self, provider_source):
        """Provide organisation details."""
        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return self._get_404_response()

        if not (auth_method := terrareg.auth.GithubAuthMethod.get_current_instance()):
            return []

        return [
            {
                "kind": repository.kind.value,
                "id": repository.provider_id,
                "full_name": repository.id,
                "owner_login": repository.owner,
                "owner_type": "owner",
                "published_id": None
            }
            for repository in terrareg.repository_model.Repository.get_repositories_by_owner_list(owners=auth_method.get_github_organisations())
            if repository.kind is not None
        ]
