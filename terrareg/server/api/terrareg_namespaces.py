
from flask import request
from flask_restful import reqparse, inputs

from terrareg.server.error_catching_resource import ErrorCatchingResource, api_error
from terrareg.errors import (
    DuplicateNamespaceDisplayNameError, NamespaceAlreadyExistsError,
    InvalidNamespaceNameError, InvalidNamespaceDisplayNameError
)
import terrareg.auth_wrapper
import terrareg.models
import terrareg.csrf
import terrareg.auth_wrapper


class ApiTerraregNamespaces(ErrorCatchingResource):
    """Provide interface to obtain namespaces."""

    method_decorators = {
        "get": [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')],
        "post": [terrareg.auth_wrapper.auth_wrapper('is_admin')]
    }

    def _get_arg_parser(self):
        """Return arg parser for get method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'only_published', type=inputs.boolean,
            location='args',
            default=False,
            help='Whether to only show namespaces with published modules'
        )
        parser.add_argument(
            'offset', type=int,
            location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            location='args',
            default=None, help='Pagination limit'
        )
        return parser

    def _get(self):
        """
        Return list of namespaces.

        The offset/limit arguments are currently optional.
        Without them, all namespaces will be returned in a list (legacy response format).
        Providing these values will return an object with a meta object and a list of namespaces.
        """
        parser = self._get_arg_parser()
        args = parser.parse_args()

        namespace_results = terrareg.models.Namespace.get_all(
            only_published=args.only_published, limit=args.limit, offset=args.offset
        )

        namespace_list = [
            {
                "name": namespace.name,
                "view_href": namespace.get_view_url(),
                "display_name": namespace.display_name
            }
            for namespace in namespace_results.rows
        ]

        if args.limit is not None:
            return {
                "meta": namespace_results.meta,
                "namespaces": namespace_list
            }
        else:
            return namespace_list

    def _post(self):
        """Create namespace."""
        namespace_name = request.json.get('name')
        display_name = request.json.get('display_name')
        csrf_token = request.json.get('csrf_token')

        terrareg.csrf.check_csrf_token(csrf_token)

        try:
            namespace = terrareg.models.Namespace.create(
                name=namespace_name,
                display_name=display_name)
        except (InvalidNamespaceNameError, NamespaceAlreadyExistsError,
                InvalidNamespaceDisplayNameError, DuplicateNamespaceDisplayNameError) as exc:
            return api_error(str(exc)), 400

        return {
            "name": namespace.name,
            "view_href": namespace.get_view_url(),
            "display_name": namespace.display_name
        }
