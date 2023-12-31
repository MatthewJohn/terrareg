
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory


class GithubOrganisations(ErrorCatchingResource):
    """Interface to provide details about current Github organisations for the logged in user"""

    def _get(self, provider_source):
        """Provide organisation details."""
        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return self._get_404_response()

        organisations = []
        if auth_method := terrareg.auth.GithubAuthMethod.get_current_instance():

            for namespace_name in auth_method.get_github_organisations():
                if (namespace := terrareg.models.Namespace.get(name=namespace_name, include_redirect=False)):
                    organisations.append({
                        "name": namespace.name,
                        "type": "user" if namespace.namespace_type is terrareg.namespace_type.NamespaceType.GITHUB_USER else "organization",
                        "admin": True,
                        "can_publish_providers": namespace.can_publish_providers
                    })

        return organisations
