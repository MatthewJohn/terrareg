
from flask import request
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf
import terrareg.errors
import terrareg.models
import terrareg.database


class ApiGpgKeys(ErrorCatchingResource):
    """Provide interface to create GPG Keys."""

    method_decorators = {
        "get": [terrareg.auth_wrapper.auth_wrapper("can_access_read_api")],
        "post": [
            terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
                terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
                # Obtain namespace for check from post data
                kwarg_values={'namespace': lambda: request.get_json().get("data", {}).get("attributes").get("namespace")})
        ],
    }

    def _get_arg_parser(self):
        """Obtain argument parser for get request"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'filter[namespace]',
            type=str,
            required=True,
            location='args',
            dest='namespaces',
            help=(
                'Comma-separated list of namespaces to obtain GPG keys for'
            )
        )
        return parser

    def _get(self):
        """Lists GPG keys for given namespaces"""
        args = self._get_arg_parser().parse_args()
        gpg_keys = []
        for namespace_name in args.namespaces.split(","):
            namespace = terrareg.models.Namespace.get(name=namespace_name)
            if namespace:
                gpg_keys += terrareg.models.GpgKey.get_by_namespace(namespace=namespace)

        return {
            "data": [
                gpg_key.get_api_data()
                for gpg_key in gpg_keys
            ]
        }

    def _post(self):
        """
        Handle creation of GPG key.

        POST Body must be JSON, in the format:
        ```
        {
            "data": {
                "type": "gpg-keys",
                "attributes": {
                    "namespace": "my-namespace",
                    "ascii-armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\n...\n-----END PGP PUBLIC KEY BLOCK-----\n"
                },
                "csrf_token": "xxxaaabbccc"
            }
        }
        ```
        """
        data = request.get_json().get("data", {})
        attributes = data.get("attributes", {})
        namespace_name = attributes.get("namespace")
        ascii_armor = attributes.get("ascii-armor")
        csrf = data.get("csrf_token")


        terrareg.csrf.check_csrf_token(csrf)

        if not namespace_name or not (namespace := terrareg.models.Namespace.get(name=namespace_name)):
            return {'message': 'Namespace does not exist'}, 400

        gpg_key = terrareg.models.GpgKey.create(namespace=namespace, ascii_armor=ascii_armor)

        return {
            "data": gpg_key.get_api_data()
        }
