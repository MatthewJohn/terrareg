
from flask import session

import terrareg.auth
from terrareg.errors import NoSessionSetError, IncorrectCSRFTokenError


def get_csrf_token():
    """Return current session CSRF token."""
    return session.get('csrf_token', '')


def check_csrf_token(csrf_token):
    """Check CSRF token."""
    # If user is authenticated using authentication token,
    # do not required CSRF token
    if not terrareg.auth.AuthFactory().get_current_auth_method().requires_csrf_tokens:
        return False

    session_token = get_csrf_token()
    if not session_token:
        raise NoSessionSetError('No session is presesnt to check CSRF token')
    elif session_token != csrf_token:
        raise IncorrectCSRFTokenError('CSRF token is incorrect')
    else:
        return True
