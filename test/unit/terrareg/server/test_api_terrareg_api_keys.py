import unittest.mock

from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context
import terrareg.models


class TestApiTerraregApiKeys(TerraregUnitTest):
    """Test API key management endpoints."""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method."""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_auth_method.can_access_read_api = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    def test_revoke_api_key(self, client, app_context, test_request_context):
        """Test API key revoke uses the existing delete endpoint."""
        api_key = unittest.mock.MagicMock()

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.models.ApiKey.get', return_value=api_key):
            res = client.delete('/v1/terrareg/api-keys/7')

        assert res.status_code == 204
        api_key.revoke.assert_called_once_with()

    def test_create_api_key_returns_plaintext_key(self, client, app_context, test_request_context):
        """Test API key creation returns the one-time plaintext key."""
        created_api_key = unittest.mock.MagicMock(
            pk=11,
            name='upload-key',
            key_type='upload',
            key_prefix='prefix',
            created_at=None,
            created_by='admin',
            last_used_at=None,
            expires_at=None,
            is_active=True,
            namespace=None,
        )

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True), \
                unittest.mock.patch('terrareg.models.ApiKey.create', return_value=(created_api_key, 'secret-value')):
            res = client.post('/v1/terrareg/api-keys', json={
                'name': 'upload-key',
                'key_type': 'upload',
                'namespace': None,
                'expires_at': None,
                'csrf_token': 'csrf-token'
            })

        assert res.status_code == 201
        assert res.json == {
            'api_key': {
                'id': 11,
                'name': 'upload-key',
                'key_type': 'upload',
                'key_prefix': 'prefix',
                'created_at': None,
                'created_by': 'admin',
                'last_used_at': None,
                'expires_at': None,
                'is_active': True,
                'namespace': None,
            },
            'plaintext_key': 'secret-value',
        }

    def test_delete_api_key(self, client, app_context, test_request_context):
        """Test API key hard delete uses the dedicated delete endpoint."""
        api_key = unittest.mock.MagicMock()

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.models.ApiKey.get', return_value=api_key):
            res = client.delete('/v1/terrareg/api-keys/7/delete')

        assert res.status_code == 204
        api_key.delete.assert_called_once_with()
