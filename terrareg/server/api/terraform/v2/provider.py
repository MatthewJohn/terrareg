
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.analytics


class ApiV2Provider(ErrorCatchingResource):
    """Interface for providing provider details"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace: str, provider: str):
        """Return provider details."""

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider_obj is None:
            return self._get_404_response()

        downloads = terrareg.analytics.ProviderAnalytics.get_provider_total_downloads(provider=provider_obj)

        return {
            "data": {
                "type": "providers",
                "id": provider_obj.pk,
                "attributes": {
                    "alias": provider_obj.alias,
                    "description": provider_obj.repository.description,
                    "downloads": downloads,
                    "featured": provider_obj.featured,
                    "full-name": provider_obj.full_name,
                    "logo-url": provider_obj.logo_url,
                    "name": provider_obj.name,
                    "namespace": provider_obj.namespace.name,
                    "owner-name": provider_obj.owner_name,
                    "repository-id": provider_obj.repository_id,
                    "robots-noindex": provider_obj.robots_noindex,
                    "source": provider_obj.source_url,
                    "tier": provider_obj.tier.value,
                    "unlisted": provider_obj.unlisted,
                    "warning": provider_obj.warning
                },
                "links": {
                    "self": f"/v2/providers/{provider_obj.pk}"
                }
            }
        }
