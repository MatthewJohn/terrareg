
import contextlib
import unittest.mock
import datetime

import pytest
import werkzeug.exceptions
import jwt

from test.unit.terrareg import MockSession, TerraregUnitTest, mocked_server_session_fixture
from test import client, app_context, test_request_context
from terrareg.server import auth_wrapper
from terrareg.auth import AuthFactory, AuthenticationType


class TestAuthFactory(TerraregUnitTest):
    """Test AuthFactory class"""

    def _get_auth_method_mocks(self, enabled_mocks):
        """Setup mocks for each auth method"""
        auth_method_class_mocks = {}
        auth_method_instance_mocks = {}

        for mock_name in ['AdminApiKeyAuthMethod', 'AdminSessionAuthMethod', 'UploadApiKeyAuthMethod',
                          'PublishApiKeyAuthMethod', 'SamlAuthMethod', 'OpenidConnectAuthMethod',
                          'NotAuthenticated']:
            auth_method_class_mocks[mock_name] = unittest.mock.MagicMock()
            auth_method_class_mocks[mock_name].is_enabled = unittest.mock.MagicMock(return_value=(mock_name in enabled_mocks))
            auth_method_instance_mocks[mock_name] = unittest.mock.MagicMock()
            auth_method_class_mocks[mock_name].get_current_instance = unittest.mock.MagicMock(return_value=auth_method_instance_mocks[mock_name])

        return auth_method_class_mocks, auth_method_instance_mocks

    @pytest.mark.parametrize('enabled_mocks,expected_class', [
        # Test each of the methods individually
        (['AdminApiKeyAuthMethod'], 'AdminApiKeyAuthMethod'),
        (['AdminSessionAuthMethod'], 'AdminSessionAuthMethod'),
        (['UploadApiKeyAuthMethod'], 'UploadApiKeyAuthMethod'),
        (['PublishApiKeyAuthMethod'], 'PublishApiKeyAuthMethod'),
        (['SamlAuthMethod'], 'SamlAuthMethod'),
        (['OpenidConnectAuthMethod'], 'OpenidConnectAuthMethod'),
        (['NotAuthenticated'], 'NotAuthenticated'),

        # Enable all auth methods and remove one at a time to ensure they are checked in order
        (['AdminApiKeyAuthMethod', 'AdminSessionAuthMethod', 'UploadApiKeyAuthMethod',
          'PublishApiKeyAuthMethod', 'SamlAuthMethod', 'OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'AdminApiKeyAuthMethod'),
        (['AdminSessionAuthMethod', 'UploadApiKeyAuthMethod',
          'PublishApiKeyAuthMethod', 'SamlAuthMethod', 'OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'AdminSessionAuthMethod'),
        (['UploadApiKeyAuthMethod',
          'PublishApiKeyAuthMethod', 'SamlAuthMethod', 'OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'UploadApiKeyAuthMethod'),
        (['PublishApiKeyAuthMethod', 'SamlAuthMethod', 'OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'PublishApiKeyAuthMethod'),
        (['SamlAuthMethod', 'OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'SamlAuthMethod'),
        (['OpenidConnectAuthMethod',
          'NotAuthenticated'],
          'OpenidConnectAuthMethod'),

        # Check random selection to ensure there's no hidden trickery
        (['NotAuthenticated', 'SamlAuthMethod'], 'SamlAuthMethod')
    ])
    def test_get_current_auth_method(self, enabled_mocks, expected_class, app_context):
        """Test get_current_auth_method method."""
        auth_method_class_mocks, auth_method_instance_mocks = self._get_auth_method_mocks(enabled_mocks)

        # Enter mock context to enable 'g'
        with app_context:

            # Enable all mocks of auth method classes
            with contextlib.ExitStack() as stack:
                for mock_ctx in [unittest.mock.patch(f'terrareg.auth.{mock_name}', auth_method_class_mocks[mock_name])
                                for mock_name in auth_method_class_mocks]:
                    stack.enter_context(mock_ctx)

                auth_method = AuthFactory().get_current_auth_method()
                assert auth_method == auth_method_instance_mocks[expected_class]

    def test_get_current_auth_method_non_enabled(self, app_context):
        """Test get_current_auth_method when no methods are enabled."""
        auth_method_class_mocks, _ = self._get_auth_method_mocks([])

        # Enter mock context to enable 'g'
        with app_context:

            # Enable all mocks of auth method classes
            with contextlib.ExitStack() as stack:
                for mock_ctx in [unittest.mock.patch(f'terrareg.auth.{mock_name}', auth_method_class_mocks[mock_name])
                                for mock_name in auth_method_class_mocks]:
                    stack.enter_context(mock_ctx)

                with pytest.raises(Exception) as exc:
                    AuthFactory().get_current_auth_method()
                    assert exc.value == 'Unable to determine current auth type - not caught by NotAuthenticated'
