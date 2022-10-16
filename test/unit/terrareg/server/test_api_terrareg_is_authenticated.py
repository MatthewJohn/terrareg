
import unittest.mock

from test.unit.terrareg import TerraregUnitTest
from test import client


class TestApiTerraregIsAuthenticated(TerraregUnitTest):

    def test_authenticated(self, client):
        """Test endpoint when user is authenticated."""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_authenticated = unittest.mock.MagicMock(return_value=True)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 200
            assert res.json == {'authenticated': True}

    def test_unauthenticated(self, client):
        """Test endpoint when user is authenticated."""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_authenticated = unittest.mock.MagicMock(return_value=False)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 401