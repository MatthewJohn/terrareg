import datetime

from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource, api_error
import terrareg.auth
import terrareg.auth_wrapper
import terrareg.csrf
import terrareg.models
from terrareg.errors import InvalidApiKeyTypeError


class ApiTerraregApiKeys(ErrorCatchingResource):
    """List and create API keys."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _post_arg_parser(self):
        """Return parser for API key creation."""
        parser = reqparse.RequestParser()
        parser.add_argument('name', type=str, required=True, location='json')
        parser.add_argument('key_type', type=str, required=True, location='json')
        parser.add_argument('namespace', type=str, required=False, default=None, location='json')
        parser.add_argument('expires_at', type=str, required=False, default=None, location='json')
        parser.add_argument('csrf_token', type=str, required=False, default=None, location='json')
        return parser

    def _serialize_api_key(self, api_key):
        """Serialize API key details without exposing the secret."""
        return {
            'id': api_key.pk,
            'name': api_key.name,
            'key_type': api_key.key_type,
            'key_prefix': api_key.key_prefix,
            'created_at': api_key.created_at.isoformat() if api_key.created_at else None,
            'created_by': api_key.created_by,
            'last_used_at': api_key.last_used_at.isoformat() if api_key.last_used_at else None,
            'expires_at': api_key.expires_at.isoformat() if api_key.expires_at else None,
            'is_active': api_key.is_active,
            'namespace': api_key.namespace,
        }

    def _get(self):
        """Return all API keys."""
        return [
            self._serialize_api_key(api_key)
            for api_key in terrareg.models.ApiKey.get_all()
        ]

    def _post(self):
        """Create a new API key."""
        args = self._post_arg_parser().parse_args()
        terrareg.csrf.check_csrf_token(args.csrf_token)

        expires_at = None
        if args.expires_at:
            try:
                expires_at = datetime.datetime.fromisoformat(args.expires_at)
            except ValueError:
                return api_error('Invalid expires_at value, expected ISO-8601 datetime'), 400

        try:
            api_key, plaintext_key = terrareg.models.ApiKey.create(
                name=args.name,
                key_type=args.key_type,
                created_by=terrareg.auth.AuthFactory().get_current_auth_method().get_username(),
                expires_at=expires_at,
                namespace=args.namespace or None,
            )
        except InvalidApiKeyTypeError as exc:
            return api_error(str(exc)), 400

        return {
            'api_key': self._serialize_api_key(api_key),
            'plaintext_key': plaintext_key,
        }, 201


class ApiTerraregApiKey(ErrorCatchingResource):
    """Revoke an API key."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _delete(self, api_key_id):
        """Revoke an API key."""
        api_key = terrareg.models.ApiKey.get(api_key_id)
        if api_key is None:
            return {'message': 'API key does not exist.'}, 400

        api_key.revoke()
        return {}, 204