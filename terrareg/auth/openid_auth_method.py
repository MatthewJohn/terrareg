
import datetime

import flask

import terrareg.openid_connect
from .base_sso_auth_method import BaseSsoAuthMethod
from .authentication_type import AuthenticationType


class OpenidConnectAuthMethod(BaseSsoAuthMethod):
    """Auth method for OpenID authentication"""
    
    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_OPENID_CONNECT
    
    @classmethod
    def check_session(cls):
        """Check OpenID session"""
        # Check OpenID connect expiry time
        try:
            session_timestamp = float(flask.session.get('openid_connect_expires_at', 0))
        except ValueError:
            return False

        if datetime.datetime.now() >= datetime.datetime.fromtimestamp(session_timestamp):
            return False

        try:
            terrareg.openid_connect.OpenidConnect.validate_session_token(flask.session.get('openid_connect_id_token'))
        except:
            # Catch any exceptions when validating session and return False
            return False
        # If validate did not raise errors, return True
        return True

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return flask.session.get('openid_groups', []) or []

    @classmethod
    def is_enabled(cls):
        return terrareg.openid_connect.OpenidConnect.is_enabled()

    def get_username(self):
        """Get username of current user"""
        return flask.session.get('openid_username')
