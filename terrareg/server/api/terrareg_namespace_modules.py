
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models


class ApiTerraregNamespaceModules(ErrorCatchingResource):
    """Interface to obtain list of modules in namespace."""

    def _get(self, namespace):
        """Return list of modules in namespace"""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            location='args',
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        module_providers = [
            module_provider
            for module in namespace_obj.get_all_modules()
            for module_provider in module.get_providers()
        ]
        if not module_providers:
            return self._get_404_response()

        meta = {
            'limit': args.limit,
            'current_offset': args.offset
        }
        if len(module_providers) > (args.offset + args.limit):
            meta['next_offset'] = (args.offset + args.limit)
        if args.offset > 0:
            meta['prev_offset'] = max(args.offset - args.limit, 0)

        return {
            "meta": meta,
            "modules": [
                module_provider.get_api_outline()
                if module_provider.get_latest_version() is None else
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers[args.offset:args.offset + args.limit]
            ]
        }
