
import unittest.mock
import datetime

import pytest
import werkzeug.exceptions
import jwt

from test.unit.terrareg import TerraregUnitTest
from test import app_context, test_request_context
from terrareg.server import auth_wrapper


class TestAuthWrapper(TerraregUnitTest):
    """Test auth_wrapper wrapper"""

    def _mock_function(self, x, y):
        """Test method to wrap to check arg/kwargs"""
        return x, y

    def test_basic_auth_success(self, app_context, test_request_context):
        """Test basic successful authentication"""
        # Mock server method being protected by auth
        mock_protected_method = unittest.mock.MagicMock(side_effect=self._mock_function)
        # Test auth method
        mock_auth_method = unittest.mock.MagicMock()
        # Test auth checking method, passed to auth_wrapper
        mock_auth_method.test_auth_method = unittest.mock.MagicMock(return_value=True)
        # Mock get_current_auth_method to return mock auth method
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            wrapped_mock = auth_wrapper('test_auth_method')(mock_protected_method)

            # Call wrapped method
            assert wrapped_mock('x-value', y='y-value') == ('x-value', 'y-value')

            mock_get_current_auth_method.assert_called_once_with()
            mock_auth_method.test_auth_method.assert_called_once_with()
            mock_protected_method.assert_called_once_with('x-value', y='y-value')

    def test_basic_auth_failure(self, app_context, test_request_context):
        """Test basic failed authentication"""
        # Mock server method being protected by auth
        mock_protected_method = unittest.mock.MagicMock(side_effect=self._mock_function)
        # Test auth method
        mock_auth_method = unittest.mock.MagicMock()
        # Test auth checking method, passed to auth_wrapper
        mock_auth_method.test_failing_method = unittest.mock.MagicMock(return_value=False)
        # Mock get_current_auth_method to return mock auth method
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            wrapped_mock = auth_wrapper('test_failing_method')(mock_protected_method)

            # Call wrapped method
            with pytest.raises(werkzeug.exceptions.Unauthorized):
                wrapped_mock()

            mock_get_current_auth_method.assert_called_once_with()
            mock_auth_method.test_failing_method.assert_called_once_with()
            mock_protected_method.assert_not_called()

    def test_auth_method_with_arguments(self, app_context, test_request_context):
        """Test authentication method that takes arguments"""
        # Mock server method being protected by auth
        mock_protected_method = unittest.mock.MagicMock()
        # Test auth method
        mock_auth_method = unittest.mock.MagicMock()
        # Test auth checking method, passed to auth_wrapper
        mock_auth_method.test_auth_method = unittest.mock.MagicMock(return_value=True)
        # Mock get_current_auth_method to return mock auth method
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            wrapped_mock = auth_wrapper('test_auth_method', 'arg1', testkwarg='kwargvalue')(mock_protected_method)

            # Call wrapped method
            wrapped_mock('x-value', y='y-value')

            mock_get_current_auth_method.assert_called_once_with()

            # Ensure auth check method was called with arguments provided to wrapper
            mock_auth_method.test_auth_method.assert_called_once_with('arg1', testkwarg='kwargvalue')

    def test_auth_method_with_request_argument_mapping(self, app_context, test_request_context):
        """Test authentication method provides arguments to protected method"""
        # Mock server method being protected by auth
        mock_protected_method = unittest.mock.MagicMock()
        # Test auth method
        mock_auth_method = unittest.mock.MagicMock()
        # Test auth checking method, passed to auth_wrapper
        mock_auth_method.test_auth_method = unittest.mock.MagicMock(return_value=True)
        # Mock get_current_auth_method to return mock auth method
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            wrapped_mock = auth_wrapper('test_auth_method', request_kwarg_map={'request_argument': 'auth_check_argument'})(mock_protected_method)

            # Call wrapped method
            wrapped_mock(request_argument='test')

            mock_get_current_auth_method.assert_called_once_with()

            # Ensure auth check method was called with arguments provided to wrapper
            mock_auth_method.test_auth_method.assert_called_once_with(auth_check_argument='test')
