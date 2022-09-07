
import datetime

import pytest

import terrareg.errors
from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context
from terrareg.server import (
    AuthenticationType, check_csrf_token
)


class TestCSRFFunctions(TerraregUnitTest):
    """Test CSRF functions."""

    def test_valid_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking a valid CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = 'testcsrftoken'
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            assert check_csrf_token('testcsrftoken') == True

    def test_incorrect_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking an incorrect CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = 'testcsrftoken'
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.IncorrectCSRFTokenError):
                check_csrf_token('doesnotmatch')

    def test_empty_csrf_with_session(self, app_context, test_request_context, client):
        """Test checking an incorrect CSRF token with a session."""
        self.SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Create fake session
            test_request_context.session['csrf_token'] = ''
            test_request_context.session['is_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token('')

    def test_invalid_csrf_without_session(self, app_context, test_request_context, client):
        """Test checking a invalid CSRF token with a session is not established."""
        self.SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token('doesnotmatter')

    def test_csrf_ignored_with_authentication_token(self, app_context, test_request_context, client):
        """Test checking a CSRF token is ignored when using authentication token."""
        self.SERVER._app.secret_key = 'averysecretkey'
        with app_context, test_request_context:

            # Set global context as authentication token
            app_context.g.authentication_type = AuthenticationType.AUTHENTICATION_TOKEN

            assert check_csrf_token(None) == False

    @pytest.mark.parametrize('authentication_type', [
        (AuthenticationType.NOT_AUTHENTICATED,),
        (AuthenticationType.NOT_CHECKED, ),
        (AuthenticationType.SESSION_PASSWORD,),
        (AuthenticationType.SESSION_OPENID_CONNECT,)]
    )
    def test_csrf_not_ignored_with_non_authentication_token(self, authentication_type, app_context, test_request_context, client):
        """Test that all authentication types throw errors when CSRF is not passed."""
        self.SERVER._app.secret_key = 'averysecretkey'

        # Test that no session is thrown when no session is present
        with app_context, test_request_context:

            app_context.g.authentication_type = authentication_type

            with pytest.raises(terrareg.errors.NoSessionSetError):
                check_csrf_token(None)

        # Test that incorrect CSRF token is thrown, when incorrect token is provided
        with app_context, test_request_context:

            app_context.g.authentication_type = authentication_type

            # Create fake session
            test_request_context.session['csrf_token'] = 'iscorrect'
            test_request_context.session['is_admin_authenticated'] = True
            test_request_context.session['expires'] = datetime.datetime.now() + datetime.timedelta(minutes=1)
            test_request_context.session.modified = True

            with pytest.raises(terrareg.errors.IncorrectCSRFTokenError):
                check_csrf_token(None)
