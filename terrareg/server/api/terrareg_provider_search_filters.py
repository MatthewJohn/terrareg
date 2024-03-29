
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_search
import terrareg.auth_wrapper


class ApiTerraregProviderSearchFilters(ErrorCatchingResource):
    """
    Return list of filters available for provider search.
    """

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return list of available filters and filter counts for search query."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'q', type=str,
            required=True,
            location='args',
            help='The search string.'
        )
        args = parser.parse_args()

        return terrareg.provider_search.ProviderSearch.get_search_filters(query=args.q)