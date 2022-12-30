
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.module_search import ModuleSearch


class ApiNamespaceModules(ErrorCatchingResource):
    """Interface to obtain list of modules in namespace."""

    def _get(self, namespace):
        """Return list of modules in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        search_results = ModuleSearch.search_module_providers(
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
                for module_provider in search_results.module_providers
            ]
        }
