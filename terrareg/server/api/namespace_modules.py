
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search
import terrareg.auth_wrapper


class ApiNamespaceModules(ErrorCatchingResource):
    """Interface to obtain list of modules in namespace."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace):
        """Return list of modules in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int, location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int, location='args',
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        search_results = terrareg.module_search.ModuleSearch.search_module_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace],
            include_internal=True
        )

        if not search_results.module_providers:
            return self._get_404_response()

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.rows
            ]
        }
