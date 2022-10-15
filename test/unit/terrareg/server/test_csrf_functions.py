
import datetime
from unittest import mock

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context
from terrareg.server import check_csrf_token
from terrareg.auth import AuthenticationType
import terrareg.auth


class TestCSRFFunctions(TerraregUnitTest):
    """Test CSRF functions."""

    def test_valid_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking a valid CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'

        mock_auth_method = mock.MagicMock()
        mock_auth_method.requires_csrf_tokens = True

        # Mock get_current_auth_method to return  mock auth type
        with mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_auth_method)):
            with app_context, test_request_context:

                # Create fake session
                test_request_context.session['csrf_token'] = 'testcsrftoken'
                test_request_context.session.modified = True

                assert check_csrf_token('testcsrftoken') == True

    def test_incorrect_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking an incorrect CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'

        mock_auth_method = mock.MagicMock()
        mock_auth_method.requires_csrf_tokens = True

        # Mock get_current_auth_method to return  mock auth type
        with mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_auth_method)):

            with app_context, test_request_context:

                # Create fake session
                test_request_context.session['csrf_token'] = 'testcsrftoken'
                test_request_context.session.modified = True

                with pytest.raises(terrareg.errors.IncorrectCSRFTokenError):
                    check_csrf_token('doesnotmatch')

    def test_missing_csrf_token(self, app_context, test_request_context, client):
        """Test checking an missing CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'

        mock_auth_method = mock.MagicMock()
        mock_auth_method.requires_csrf_tokens = True

        # Mock get_current_auth_method to return  mock auth type
        with mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_auth_method)):

            with app_context, test_request_context:

                # Create fake session
                test_request_context.session.modified = True

                with pytest.raises(terrareg.errors.NoSessionSetError):
                    check_csrf_token('doesnotmatch')

    def test_missing_csrf_ignored_with_non_authenticated_sessions(self, app_context, test_request_context, client):
        """Test that only session-based authentication types throw errors when CSRF is not passed."""
        self.SERVER._app.secret_key = 'averysecretkey'

        mock_auth_method = mock.MagicMock()
        mock_auth_method.requires_csrf_tokens = False

        # Mock get_current_auth_method to return  mock auth type
        with mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_auth_method)):
            # Test that incorrect CSRF token is thrown, when incorrect token is provided
            with app_context, test_request_context:

                # Create fake CSRF token in session
                test_request_context.session['csrf_token'] = 'iscorrect'
                test_request_context.session.modified = True

                assert check_csrf_token(None) == False

    def test_incorrect_csrf_ignored_with_non_authenticated_sessions(self, app_context, test_request_context, client):
        """Test that only session-based authentication types throw errors when CSRF is not passed."""
        self.SERVER._app.secret_key = 'averysecretkey'

        mock_auth_method = mock.MagicMock()
        mock_auth_method.requires_csrf_tokens = False

        # Mock get_current_auth_method to return  mock auth type
        with mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock.MagicMock(return_value=mock_auth_method)):
            # Test that incorrect CSRF token is thrown, when incorrect token is provided
            with app_context, test_request_context:

                # Create fake CSRF token in session
                test_request_context.session['csrf_token'] = 'iscorrect'
                test_request_context.session.modified = True

                assert check_csrf_token(None) == False
