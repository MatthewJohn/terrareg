
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.analytics


class ApiV2Provider(ErrorCatchingResource):
    """Interface for providing provider details"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get_arg_parser(self):
        """Get arg parser for get endpoint"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            "include",
            type=str,
            help="List of linked resources to include in response. Currently supports: provider-versions, categories",
            default="",
            required=False,
            location="args",
        )
        return parser

    def _get(self, namespace: str, provider: str):
        """Return provider details."""

        args = self._get_arg_parser().parse_args()

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider_obj is None:
            return self._get_404_response()

        downloads = terrareg.analytics.ProviderAnalytics.get_provider_total_downloads(provider=provider_obj)

        includes = []
        include_names = args.include.split(",")
        if "provider-versions" in include_names:
            for provider_version in provider_obj.get_all_versions():
                includes.append(provider_version.get_v2_include())
        if "categories" in include_names:
            includes.append(provider_obj.category.get_v2_include())

        data = {
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

        if args.include:
            data["included"] = includes

        return data
