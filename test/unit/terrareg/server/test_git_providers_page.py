import unittest.mock

from test.unit.terrareg import TerraregUnitTest, mock_models
from test import client, app_context, test_request_context


class TestGitProvidersPage(TerraregUnitTest):
    """Test git providers settings page access."""

    def _mock_auth_method(self, is_admin):
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin.return_value = is_admin
        mock_auth_method.can_access_read_api.return_value = True
        mock_auth_method.is_authenticated.return_value = True
        return mock_auth_method

    def test_non_admin_is_rejected(self, client, app_context, test_request_context, mock_models):
        """Non-admin users should not access the git providers settings page."""
        with app_context, test_request_context, client, unittest.mock.patch(
            'terrareg.auth.AuthFactory.get_current_auth_method',
            return_value=self._mock_auth_method(False)
        ):
            res = client.get('/git-providers')

        assert res.status_code == 403
        assert b'Permission denied' in res.data

    def test_admin_can_access(self, client, app_context, test_request_context, mock_models):
        """Admins should access the git providers settings page."""
        with app_context, test_request_context, client, unittest.mock.patch(
            'terrareg.auth.AuthFactory.get_current_auth_method',
            return_value=self._mock_auth_method(True)
        ):
            res = client.get('/git-providers')

        assert res.status_code == 200
        assert b'Create Git Provider' in res.data