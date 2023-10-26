
from flask_restful import reqparse, inputs

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.provider_search
import terrareg.module_search
import terrareg.auth_wrapper


class ApiProviderSearch(ErrorCatchingResource):

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Search for modules, given query string, namespace or provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'q', type=str,
            required=True,
            location='args',
            help='The search string.'
        )
        parser.add_argument(
            'offset', type=int, location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int, location='args',
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'namespace', type=str, location='args',
            default=None, help='Limits modules to a specific namespace.',
            action='append', dest='namespaces'
        )
        parser.add_argument(
            'trusted_namespaces', type=inputs.boolean, location='args',
            default=None, help='Limits modules to include trusted namespaces.'
        )
        parser.add_argument(
            'contributed', type=inputs.boolean, location='args',
            default=None, help='Limits modules to include contributed modules.'
        )
        parser.add_argument(
            'include_count', type=inputs.boolean, location='args', default=False,
            help='Whether to include total result count. This is not part of the Terraform API spec.'
        )

        args = parser.parse_args()

        namespace_trust_filters = terrareg.module_search.NamespaceTrustFilter.UNSPECIFIED
        # If either trusted namespaces or contributed have been provided
        # (irrelevant of whether they are set to true or false),
        # setup the filter to no longer be unspecified.
        if args.trusted_namespaces is not None or args.contributed is not None:
            namespace_trust_filters = []

        if args.trusted_namespaces:
            namespace_trust_filters.append(terrareg.module_search.NamespaceTrustFilter.TRUSTED_NAMESPACES)
        if args.contributed:
            namespace_trust_filters.append(terrareg.module_search.NamespaceTrustFilter.CONTRIBUTED)

        search_results = terrareg.provider_search.ProviderSearch.search_providers(
            query=args.q,
            namespaces=args.namespaces,
            namespace_trust_filters=namespace_trust_filters,
            offset=args.offset,
            limit=args.limit
        )

        res = {
            "meta": search_results.meta,
            "providers": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.rows
                if module_provider.get_latest_version()
            ]
        }
        if args.include_count:
            res['count'] = search_results.count
        return res
