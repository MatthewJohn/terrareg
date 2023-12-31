
import datetime
import unittest.mock

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest, mock_models
from test import client, app_context, test_request_context
import terrareg.auth
from test.integration.terrareg.fixtures import test_github_provider_source


class TestGithubAuthInitiate(TerraregUnitTest):
    """Test Github Auth initiate API."""

    def test_without_sessions_enabled(self, client, mock_models, test_github_provider_source):
        """Test endpoint without session secret key being set."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', None), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock:
            # Update real app secret key
            self.SERVER._app.secret_key = None

            res = client.get("/test-github-provider/login")
            assert res.status_code == 200
            assert "Sessions are not available" in res.data.decode('utf-8')

    def test_without_github_configurations(self, client, mock_models):
        """Test endpoint with github configuration not being set."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            res = client.get("/github/login")

            assert res.status_code == 200
            assert "github authentication is not enabled" in res.data.decode('utf-8')

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert 'csrf_token' not in session
                assert 'session_id' not in session

    def test_call(self, client, mock_models, test_github_provider_source):
        """Test valid call to endpoint."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            res = client.get("/test-github-provider/login")

            assert res.status_code == 302
            assert res.headers.get("Location") == "https://github.example.com/login/oauth/authorize?client_id=unittest-client-id"

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert 'csrf_token' in session
                assert 'session_id' in session
