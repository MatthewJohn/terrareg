
import unittest.mock
import datetime

import pytest
import werkzeug.exceptions

from test.unit.terrareg import (
    client, test_request_context,
    app_context
)
from terrareg.server import (
    require_admin_authentication, AuthenticationType,
    get_current_authentication_type
)


class TestRequireAdminAuthenticationWrapper:
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
        mock_session=None):
        """Perform authentication test."""
        with test_request_context, \
                app_context, \
                unittest.mock.patch('terrareg.config.SECRET_KEY', config_secret_key), \
                unittest.mock.patch('terrareg.config.ADMIN_AUTHENTICATION_TOKEN', config_admin_authentication_token):

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

    def test_unauthenticated(self, app_context, test_request_context):
        """Ensure 401 without an API key or mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='asecrethere',
            config_admin_authentication_token='testpassword',
            expect_fail=True
        )

    def test_mock_session_authentication_with_no_app_secret(self, app_context, test_request_context):
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

    def test_401_with_expired_mock_session(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=True,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() - datetime.timedelta(minutes=1)
            }
        )

    def test_invalid_authentication_with_empty_api_key(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
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

    def test_authentication_with_mock_session(self, app_context, test_request_context):
        """Ensure resource is called with valid mock_session."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            config_secret_key='testsecret',
            config_admin_authentication_token='testpassword',
            expect_fail=False,
            expected_authentication_type=AuthenticationType.SESSION,
            mock_session={
                'is_admin_authenticated': True,
                'expires': datetime.datetime.now() + datetime.timedelta(hours=5)
            }
        )

    def test_authentication_with_api_key(self, app_context, test_request_context):
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
