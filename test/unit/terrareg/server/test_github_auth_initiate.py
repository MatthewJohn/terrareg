
import datetime
import unittest.mock

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest, mock_models
from test import client, app_context, test_request_context
import terrareg.auth


class TestGithubAuthInitiate(TerraregUnitTest):
    """Test Github Auth initiate API."""

    def test_without_sessions_enabled(self, client, mock_models):
        """Test endpoint without session secret key being set."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', None), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock:
            # Update real app secret key
            self.SERVER._app.secret_key = None

            res = client.get("/github/login")
            assert res.status_code == 200
            assert "Sessions are not available" in res.data.decode('utf-8')

    def test_without_github_configurations(self, client, mock_models):
        """Test endpoint with github configuration not being set."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.github.Github.is_enabled', unittest.mock.MagicMock(return_value=False)) as mock_is_enabled:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            res = client.get("/github/login")

            assert res.status_code == 200
            assert "Github authentication is not enabled" in res.data.decode('utf-8')

            mock_is_enabled.assert_called_once_with()

    def test_call(self, client, mock_models):
        """Test valid call to endpoint."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.github.Github.is_enabled', unittest.mock.MagicMock(return_value=True)) as mock_is_enabled, \
                unittest.mock.patch('terrareg.github.Github.get_login_redirect_url',
                                    unittest.mock.MagicMock(return_value='https://examplegithub.com/login?client_id=abcdefg123')) as mock_get_login_redirect_url:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            res = client.get("/github/login")

            assert res.status_code == 302
            assert res.headers.get("Location") == "https://examplegithub.com/login?client_id=abcdefg123"

            mock_is_enabled.assert_called_once_with()
            mock_get_login_redirect_url.assert_called_once_with()
