
import unittest.mock

from test.unit.terrareg import client, TerraregUnitTest


class TestApiTerraregIsAuthenticated(TerraregUnitTest):

    def test_authenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            mock_admin_authentication.return_value = True
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 200
            assert res.json == {'authenticated': True}

    def test_unauthenticated(self, client):
        """Test endpoint when user is authenticated."""
        with unittest.mock.patch('terrareg.server.check_admin_authentication') as mock_admin_authentication:
            mock_admin_authentication.return_value = False
            res = client.get('/v1/terrareg/auth/admin/is_authenticated')
            assert res.status_code == 401