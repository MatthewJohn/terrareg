
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_search
import terrareg.auth_wrapper
import terrareg.models


class ApiNamespaceProviders(ErrorCatchingResource):
    """Interface to obtain list of providers in namespace."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace):
        """Return list of providers in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int, location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int, location='args',
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        # Check if namespace exists
        if not terrareg.models.Namespace.get(name=namespace):
            return self._get_404_response()

        search_results = terrareg.provider_search.ProviderSearch.search_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace]
        )

        return {
            "meta": search_results.meta,
            "providers": [
                provider.get_latest_version().get_api_outline()
                for provider in search_results.rows
            ]
        }
