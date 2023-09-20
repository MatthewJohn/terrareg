
from flask_restful import reqparse, inputs

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.module_search
import terrareg.auth_wrapper


class ApiModuleList(ErrorCatchingResource):

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
            default=None, help='Limits modules to a specific provider.',
            action='append', dest='providers'
        )
        parser.add_argument(
            'verified', type=inputs.boolean, location='args',
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        search_results = terrareg.module_search.ModuleSearch.search_module_providers(
            providers=args.providers,
            verified=args.verified,
            offset=args.offset,
            limit=args.limit
        )

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }
