
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import Namespace
from terrareg.module_search import ModuleSearch


class ApiModuleDetails(ErrorCatchingResource):
    def _get(self, namespace, name):
        """Return latest version for each module provider."""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        namespace, _ = Namespace.extract_analytics_token(namespace)

        search_results = ModuleSearch.search_module_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace],
            modules=[name]
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

