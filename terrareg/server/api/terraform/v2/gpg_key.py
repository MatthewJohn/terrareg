
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf
import terrareg.errors
import terrareg.models
import terrareg.database


class ApiGpgKey(ErrorCatchingResource):
    """Provide interface to create GPG Keys."""

    method_decorators = {
        "get": [terrareg.auth_wrapper.auth_wrapper("can_access_read_api")],
        "delete": [
            terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
                terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
                request_kwarg_map={'namespace': 'namespace'})
        ],
    }

    def _get(self, namespace, key_id):
        """Get details for a given GPG key for a namespace"""
        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if not namespace_obj:
            return self._get_404_response()

        gpg_key = terrareg.models.GpgKey.get_by_id_and_namespace(namespace=namespace_obj, id_=key_id)
        if not gpg_key:
            return self._get_404_response()

        return {
            "data": gpg_key.get_api_data()
        }

    def _delete_arg_parser(self):
        """Return arg parser for delete method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )
        return parser

    def _delete(self, namespace, key_id):
        """
        Perform deletion of GPG key
        """
        args = self._delete_arg_parser().parse_args()
        terrareg.csrf.check_csrf_token(args.csrf_token)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if not namespace_obj:
            return self._get_404_response()

        gpg_key = terrareg.models.GpgKey.get_by_id_and_namespace(namespace=namespace_obj, id_=key_id)
        gpg_key.delete()

        return {}, 201
