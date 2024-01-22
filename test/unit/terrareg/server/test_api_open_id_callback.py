
import datetime
import unittest.mock

import pytest

import terrareg.errors
from terrareg.namespace_type import NamespaceType
from test.unit.terrareg import TerraregUnitTest, mock_models
from test import client, app_context, test_request_context
import terrareg.auth
import terrareg.audit_action


class TestApiOpenIdCallback(TerraregUnitTest):
    """Test OpenIDC Auth callback API."""

    def test_without_openid_connect_state(self, client, mock_models):
        """Test endpoint without openid_connect_state."""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token') as mock_fetch_access_token:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            res = client.get("/openid/callback")

            assert res.status_code == 400
            assert res.json == {}
            # assert "Invalid code returned from test-github-provider" in res.data.decode('utf-8')

            mock_fetch_access_token.assert_not_called()
            mock_create_audit_event.assert_not_called()

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session == {}

    def test_fetch_access_token_exception(self, client, mock_models):
        """Test endpoint with fetch_access_token throwing exception"""

        def fetch_access_token_side_effect(*args, **kwargs):
            raise Exception("Error in fetch_access_token")

        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example-public.example.com'), \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', unittest.mock.MagicMock(side_effect=fetch_access_token_side_effect)) as mock_fetch_access_token:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            with client.session_transaction() as session:
                # set a user id without going through the login route
                session["openid_connect_state"] = "unittest-openid-connect-state"

            res = client.get("/openid/callback?some=kwarg")

            assert res.status_code == 200
            assert "Invalid response from SSO" in res.data.decode('utf-8')

            mock_fetch_access_token.assert_called_once_with(uri='https://example-public.example.com:443/openid/callback?some=kwarg', valid_state='unittest-openid-connect-state')
            mock_create_audit_event.assert_not_called()

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session == {'openid_connect_state': 'unittest-openid-connect-state'}

    def test_fetch_access_token_empty_response(self, client, mock_models):
        """Test endpoint with fetch_access_token returning empty value"""

        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example-public.example.com'), \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', unittest.mock.MagicMock(return_value=None)) as mock_fetch_access_token:
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            with client.session_transaction() as session:
                # set a user id without going through the login route
                session["openid_connect_state"] = "unittest-openid-connect-state"

            res = client.get("/openid/callback?some=kwarg")

            assert res.status_code == 200
            assert "Invalid response from SSO" in res.data.decode('utf-8')

            mock_fetch_access_token.assert_called_once_with(uri='https://example-public.example.com:443/openid/callback?some=kwarg', valid_state='unittest-openid-connect-state')
            mock_create_audit_event.assert_not_called()

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session == {'openid_connect_state': 'unittest-openid-connect-state'}

    def test_empty_user_info(self, client, mock_models):
        """Test endpoint with empty user info"""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example-public.example.com'), \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', unittest.mock.MagicMock(return_value={
                    "id_token": "unittest-id-token", "access_token": "unittest-access-token", "expires_in": 5000})) as mock_fetch_access_token, \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.get_user_info', unittest.mock.MagicMock(return_value={})):
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            with client.session_transaction() as session:
                # set a user id without going through the login route
                session["openid_connect_state"] = "unittest-openid-connect-state"

            before_call = datetime.datetime.now()
            res = client.get("/openid/callback?some=kwarg")
            after_call = datetime.datetime.now()

            assert res.status_code == 302
            assert res.location == '/'

            mock_fetch_access_token.assert_called_once_with(uri='https://example-public.example.com:443/openid/callback?some=kwarg', valid_state='unittest-openid-connect-state')
            mock_create_audit_event.assert_called_once_with(action=terrareg.audit_action.AuditAction.USER_LOGIN, object_type=None, object_id=None, old_value=None, new_value=None)

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session.get('openid_connect_state') == 'unittest-openid-connect-state'
                assert session.get('authentication_type') == 4
                assert session.get('is_admin_authenticated') is True
                assert session.get('openid_connect_access_token') == 'unittest-access-token'
                assert session.get('openid_connect_id_token') == 'unittest-id-token'
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) >= (before_call + datetime.timedelta(seconds=5000))
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) <= (after_call + datetime.timedelta(seconds=5000))
                assert session.get('openid_groups') == []
                assert session.get('openid_username') is None
                assert session.get('session_id')
                assert session.get('csrf_token')

    def test_endpoint(self, client, mock_models):
        """Test endpoint"""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example-public.example.com'), \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', unittest.mock.MagicMock(return_value={
                    "id_token": "unittest-id-token", "access_token": "unittest-access-token", "expires_in": 5000})) as mock_fetch_access_token, \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.get_user_info', unittest.mock.MagicMock(return_value={
                    "groups": ["a-group-1", "second-group"], "preferred_username": "unittest-username"})):
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            with client.session_transaction() as session:
                # set a user id without going through the login route
                session["openid_connect_state"] = "unittest-openid-connect-state"

            before_call = datetime.datetime.now()
            res = client.get("/openid/callback?some=kwarg")
            after_call = datetime.datetime.now()

            assert res.status_code == 302
            assert res.location == '/'

            mock_fetch_access_token.assert_called_once_with(uri='https://example-public.example.com:443/openid/callback?some=kwarg', valid_state='unittest-openid-connect-state')
            mock_create_audit_event.assert_called_once_with(action=terrareg.audit_action.AuditAction.USER_LOGIN, object_type=None, object_id=None, old_value=None, new_value=None)

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session.get('openid_connect_state') == 'unittest-openid-connect-state'
                assert session.get('authentication_type') == 4
                assert session.get('is_admin_authenticated') is True
                assert session.get('openid_connect_access_token') == 'unittest-access-token'
                assert session.get('openid_connect_id_token') == 'unittest-id-token'
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) >= (before_call + datetime.timedelta(seconds=5000))
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) <= (after_call + datetime.timedelta(seconds=5000))
                assert session.get('openid_groups') == ["a-group-1", "second-group"]
                assert session.get('openid_username') == "unittest-username"
                assert session.get('session_id')
                assert session.get('csrf_token')


    def test_endpoint_email_fallback(self, client, mock_models):
        """Test endpoint falling back to email address"""
        with unittest.mock.patch('terrareg.config.Config.SECRET_KEY', 'abcdefg'), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.models.Session.cleanup_old_sessions', create=True) as cleanup_old_sessions_mock, \
                unittest.mock.patch('terrareg.config.Config.PUBLIC_URL', 'https://example-public.example.com'), \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', unittest.mock.MagicMock(return_value={
                    "id_token": "unittest-id-token", "access_token": "unittest-access-token", "expires_in": 5000})) as mock_fetch_access_token, \
                unittest.mock.patch('terrareg.openid_connect.OpenidConnect.get_user_info', unittest.mock.MagicMock(return_value={
                    "groups": ["a-group-1", "second-group"], "email": "some-email@address.com"})):
            # Update real app secret key
            self.SERVER._app.secret_key = 'abcdefg'

            with client.session_transaction() as session:
                # set a user id without going through the login route
                session["openid_connect_state"] = "unittest-openid-connect-state"

            before_call = datetime.datetime.now()
            res = client.get("/openid/callback?some=kwarg")
            after_call = datetime.datetime.now()

            assert res.status_code == 302
            assert res.location == '/'

            mock_fetch_access_token.assert_called_once_with(uri='https://example-public.example.com:443/openid/callback?some=kwarg', valid_state='unittest-openid-connect-state')
            mock_create_audit_event.assert_called_once_with(action=terrareg.audit_action.AuditAction.USER_LOGIN, object_type=None, object_id=None, old_value=None, new_value=None)

            # Ensure session variables have not been set
            with client.session_transaction() as session:
                assert session.get('openid_connect_state') == 'unittest-openid-connect-state'
                assert session.get('authentication_type') == 4
                assert session.get('is_admin_authenticated') is True
                assert session.get('openid_connect_access_token') == 'unittest-access-token'
                assert session.get('openid_connect_id_token') == 'unittest-id-token'
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) >= (before_call + datetime.timedelta(seconds=5000))
                assert datetime.datetime.fromtimestamp(session.get('openid_connect_expires_at')) <= (after_call + datetime.timedelta(seconds=5000))
                assert session.get('openid_groups') == ["a-group-1", "second-group"]
                assert session.get('openid_username') == "some-email@address.com"
                assert session.get('session_id')
                assert session.get('csrf_token')


