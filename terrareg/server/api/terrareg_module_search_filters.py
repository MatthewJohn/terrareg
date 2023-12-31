
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search
import terrareg.auth_wrapper


class ApiTerraregModuleSearchFilters(ErrorCatchingResource):
    """
    Return list of filters available for search.
    
    *Deprecation*: The `/v1/terrareg/search_filters` endpoint has been deprecated in favour of `/v1/terrareg/modules/search/filters`

    The previous endpoint will be removed in a future major release.
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

        return terrareg.module_search.ModuleSearch.get_search_filters(query=args.q)