
import unittest.mock
import datetime

import pytest
import werkzeug.exceptions

from test.unit.terrareg import (
    client, test_request_context,
    app_context, TerraregUnitTest
)
from terrareg.server import (
    AuthenticationType,
    get_current_authentication_type, require_api_authentication
)


class TestRequireApiAuthenticationWrapper(TerraregUnitTest):
    """Test require_admin_authentication wrapper"""

    def _mock_function(self, x, y):
        """Test method to wrap to check arg/kwargs"""
        return x, y

    def _run_authentication_test(
        self,
        app_context,
        test_request_context,
        allowed_api_keys,
        expect_fail,
        admin_authentication_pass,
        expected_authentication_type=None,
        mock_headers=None):
        """Perform authentication test."""
        with test_request_context, \
                app_context, \
                unittest.mock.patch('terrareg.server.check_admin_authentication') as check_admin_authentication_mock:

            # Fake mock_headers and mock_session
            if mock_headers:
                test_request_context.request.headers = mock_headers

            # Mock result of check admin authentication
            check_admin_authentication_mock.return_value = admin_authentication_pass

            wrapped_mock = require_api_authentication(allowed_api_keys)(self._mock_function)

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
            
            check_admin_authentication_mock.assert_called_once()

    def test_unauthenticated_with_api_keys_set(self, app_context, test_request_context):
        """Ensure 401 without an API key or valid admin authentication."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['test-api-key'],
            admin_authentication_pass=False,
            expect_fail=True
        )

    def test_admin_authenticated_with_api_keys_set(self, app_context, test_request_context):
        """Ensure authentication passes with valid admin authentication."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['test-api-key'],
            admin_authentication_pass=True,
            expect_fail=False
        )

    def test_allowed_without_api_keys_set(self, app_context, test_request_context):
        """Ensure authentication passes when no API keys set."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=[],
            admin_authentication_pass=False,
            expect_fail=False
        )

    def test_allowed_without_api_keys_set_with_admin_authentication(self, app_context, test_request_context):
        """Ensure authentication passes when no API keys set and admin authentication passes."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=[],
            admin_authentication_pass=True,
            expect_fail=False
        )

    def test_invalid_authentication_with_empty_api_key(self, app_context, test_request_context):
        """Ensure 403 is returned with an empty API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['testapikey'],
            expect_fail=True,
            admin_authentication_pass=False,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': ''
            }
        )

    def test_invalid_authentication_with_incorrect_api_key(self, app_context, test_request_context):
        """Ensure 403 is returned when incorrect API key is provided."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['testapikey'],
            admin_authentication_pass=False,
            expect_fail=True,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': 'incorrect'
            }
        )

    def test_valid_api_key(self, app_context, test_request_context):
        """Ensure resource is called with valid API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['testapikey'],
            admin_authentication_pass=False,
            expect_fail=False,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': 'testapikey'
            }
        )

    def test_valid_api_key_with_multiple_defined(self, app_context, test_request_context):
        """Ensure resource is called with valid API key."""
        self._run_authentication_test(
            app_context=app_context,
            test_request_context=test_request_context,
            allowed_api_keys=['validkey1', 'testapikey', 'validkey3'],
            admin_authentication_pass=False,
            expect_fail=False,
            mock_headers={
                'Host': 'localhost',
                'X-Terrareg-ApiKey': 'testapikey'
            }
        )
