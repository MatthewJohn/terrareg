
import unittest.mock
import datetime

import pytest
import werkzeug.exceptions
import jwt

from test.unit.terrareg import MockSession, TerraregUnitTest, mocked_server_session_fixture
from test import client, app_context, test_request_context
from terrareg.server import (
    require_admin_authentication, AuthenticationType,
    get_current_authentication_type
)


class TestRequireAdminAuthenticationWrapper(TerraregUnitTest):
    """Test require_admin_authentication wrapper"""

    def _mock_function(self, x, y):
        """Test method to wrap to check arg/kwargs"""
        return x, y

    def _run_authentication_test(
        self,
        app_context,
        test_request_context,
        config_secret_key,
        config_admin_authentication_token,
        expect_fail,
        expected_authentication_type=None,
        mock_headers=None,
        mock_session=None,
        mock_sessions={}):
        """Perform authentication test."""
        with test_request_context, \
                app_context, \
                unittest.mock.patch('terrareg.config.Config.SECRET_KEY', config_secret_key), \
                unittest.mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', config_admin_authentication_token), \
                unittest.mock.patch.dict(MockSession.MOCK_SESSIONS, dict(mock_sessions)):

            # Fake mock_headers and mock_session
            if mock_headers:
                test_request_context.request.headers = mock_headers
            if mock_session:
                test_request_context.session = mock_session

            wrapped_mock = require_admin_authentication(self._mock_function)

            # Ensure before calling authentication, that current authentication
            # type is shown as not checked
            assert get_current_authentication_type() is AuthenticationType.NOT_CHECKED

            if expect_fail:
                with pytest.raises(werkzeug.exceptions.Unauthorized):
                    wrapped_mock()
            else:
                assert wrapped_mock('x-value', y='y-value') == ('x-value', 'y-value')

                # Check authentication_type has been set correctly. 
                if expected_authentication_type:
                    assert get_current_authentication_type() is expected_authentication_type

    def test_unauthenticated(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure 401 without an API key or mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='asecrethere',
            config_admin_authentication_token='testpassword',
            expect_fail=True
        )

    def test_mock_session_authentication_with_no_app_secret(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure 401 with valid authentication without an APP SECRET."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_401_with_expired_mock_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Test with an expired session."""
        # @TODO This is currently testing the functionality of the MockSession.
        # This test should be removed, since the checking functionality can't be put
        # into a non-overridden method, as it's part of the SQL query
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'session_id': 'unittestssessionid',
                'is_admin_authenticated': True,
                'authentication_type': AuthenticationType.SESSION_PASSWORD.value
            },
            mock_sessions={
                'unittestssessionid': datetime.datetime.now() - datetime.timedelta(minutes=1)
            }
        )

    def test_401_with_nonexistent_mock_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Test with expired or non-existent session ID."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'session_id': 'nonexistentsessionid',
                'is_admin_authenticated': True,
                'authentication_type': AuthenticationType.SESSION_PASSWORD.value
            },
            mock_sessions={
            }
        )

    def test_invalid_authentication_with_empty_api_key(self, app_context, test_request_context, mocked_server_session_fixture):
        """Test authentication with an empty API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='',
            expect_fail=True,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': ''
            }
        )

    def test_authentication_with_mock_session_without_authentication_type(self, app_context, test_request_context, mocked_server_session_fixture):
        """Test authentication without an authentication type in session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            expected_authentication_type=AuthenticationType.SESSION_PASSWORD,
            mock_session={
                'session_id': 'unittestsessionid',
                'is_admin_authenticated': True
            },
            mock_sessions={
                'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_authentication_with_mock_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure resource is called with valid password session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=False,
            expected_authentication_type=AuthenticationType.SESSION_PASSWORD,
            mock_session={
                'session_id': 'unittestsessionid',
                'is_admin_authenticated': True,
                'authentication_type': AuthenticationType.SESSION_PASSWORD.value
            },
            mock_sessions={
                'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_authentication_openid_without_expiry_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure authentication via OpenID connect without a valid expiry is rejected."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            expected_authentication_type=AuthenticationType.SESSION_OPENID_CONNECT,
            mock_session={
                'session_id': 'unittestsessionid',
                'is_admin_authenticated': True,
                'authentication_type': AuthenticationType.SESSION_OPENID_CONNECT.value
            },
            mock_sessions={
                'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_authentication_with_expired_openid_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure expired OpenID connect session is rejected."""
        mock_validate_session_token = unittest.mock.MagicMock(return_value=False)

        with unittest.mock.patch('terrareg.openid_connect.OpenidConnect.validate_session_token', mock_validate_session_token):
            self._run_authentication_test(
                app_context=app_context,
                test_request_context=test_request_context,
                config_secret_key='testsecret',
                config_admin_authentication_token='testpassword',
                expect_fail=True,
                expected_authentication_type=AuthenticationType.SESSION_OPENID_CONNECT,
                mock_session={
                    'session_id': 'unittestsessionid',
                    'is_admin_authenticated': True,
                    'authentication_type': AuthenticationType.SESSION_OPENID_CONNECT.value,
                    'openid_connect_expires_at': 500,
                    'openid_connect_id_token': 'unittest-openid-connect-id-token'
                },
                mock_sessions={
                    'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
                }
            )

            mock_validate_session_token.assert_not_called()


    def test_authentication_with_openid_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Test authentication with a valid OpenID session."""
        mock_validate_session_token = unittest.mock.MagicMock(return_value=True)

        with unittest.mock.patch('terrareg.openid_connect.OpenidConnect.validate_session_token', mock_validate_session_token):
            self._run_authentication_test(
                app_context=app_context,
                test_request_context=test_request_context,
                config_secret_key='testsecret',
                config_admin_authentication_token='testpassword',
                expect_fail=False,
                expected_authentication_type=AuthenticationType.SESSION_OPENID_CONNECT,
                mock_session={
                    'session_id': 'unittestsessionid',
                    'is_admin_authenticated': True,
                    'authentication_type': AuthenticationType.SESSION_OPENID_CONNECT.value,
                    'openid_connect_expires_at': 33197904000,
                    'openid_connect_id_token': 'unittest-openid-connect-id-token'
                },
                mock_sessions={
                    'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
                }
            )

        mock_validate_session_token.assert_called_once_with('unittest-openid-connect-id-token')

    def test_authentication_with_invalid_openid_session(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure an invalid OpenID conect ID token is causes a failed login."""
        def raise_exception():
            raise jwt.exceptions.DecodeError
        mock_validate_session_token = unittest.mock.MagicMock(side_effect=raise_exception)

        with unittest.mock.patch('terrareg.openid_connect.OpenidConnect.validate_session_token', mock_validate_session_token):
            self._run_authentication_test(
                app_context=app_context,
                test_request_context=test_request_context,
                config_secret_key='testsecret',
                config_admin_authentication_token='testpassword',
                expect_fail=True,
                expected_authentication_type=AuthenticationType.SESSION_OPENID_CONNECT,
                mock_session={
                    'session_id': 'unittestsessionid',
                    'is_admin_authenticated': True,
                    'authentication_type': AuthenticationType.SESSION_OPENID_CONNECT.value,
                    'openid_connect_expires_at': 33197904000,
                    'openid_connect_id_token': 'invalid-openid-connect-id-token'
                },
                mock_sessions={
                    'unittestsessionid': datetime.datetime.now() + datetime.timedelta(hours=5)
                }
            )

        mock_validate_session_token.assert_called_once_with('invalid-openid-connect-id-token')

    def test_authentication_with_api_key(self, app_context, test_request_context, mocked_server_session_fixture):
        """Ensure resource is called with an API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=False,
            expected_authentication_type=AuthenticationType.AUTHENTICATION_TOKEN,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': 'testpassword'
            }
        )
