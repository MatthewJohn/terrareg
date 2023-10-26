
from flask_restful import reqparse, inputs

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_search
import terrareg.auth_wrapper


class ApiProviderList(ErrorCatchingResource):
    """Interface to list all providers"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return list of modules."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int, location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int, location='args',
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'provider', type=str, location='args',
            default=None, help='Limits providers by specific providers.',
            action='append', dest='providers'
        )

        args = parser.parse_args()

        search_results = terrareg.provider_search.ProviderSearch.search_providers(
            providers=args.providers,
            offset=args.offset,
            limit=args.limit
        )

        return {
            "meta": search_results.meta,
            "providers": [
                provider.get_latest_version().get_api_outline()
                for provider in search_results.rows
            ]
        }
