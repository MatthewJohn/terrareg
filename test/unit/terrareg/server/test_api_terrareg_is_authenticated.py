
import unittest.mock

from terrareg.auth import AdminApiKeyAuthMethod, NotAuthenticated, UserGroupNamespacePermissionType
from test.unit.terrareg import (
    TerraregUnitTest, mock_models, setup_test_data
)
import terrareg.models
from test import client

test_data = {
    'testnamespace': {
        'id': 1
    }
}

user_group_data = {
    'usergroup': {
        'id': 1,
        'site_admin': False,
        'namespace_permissions': {
            'namespace1': UserGroupNamespacePermissionType.FULL
        }
    }
}

class TestApiTerraregIsAuthenticated(TerraregUnitTest):

    def test_authenticated_site_admin(self, client, mock_models):
        """Test endpoint when user is authenticated."""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminApiKeyAuthMethod())
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 200
            assert res.json == {'authenticated': True, 'namespace_permissions': {}, 'site_admin': True}

    @setup_test_data(test_data, user_group_data=user_group_data)
    def test_authenticated_with_permissions(self, client, mock_models):
        """Test authenticated with permissions"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_authenticated.return_value = True
        mock_auth_method.is_admin.return_value = False
        mock_auth_method.get_all_namespace_permissions.return_value = {
            terrareg.models.Namespace('testnamespace'): UserGroupNamespacePermissionType.FULL
        }
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 200
            assert res.json == {'authenticated': True, 'namespace_permissions': {'testnamespace': 'FULL'}, 'site_admin': False}

    def test_unauthenticated(self, client, mock_models):
        """Test endpoint when user is authenticated."""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=NotAuthenticated())
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 401
