
import flask

import terrareg.github
from .base_sso_auth_method import BaseSsoAuthMethod
from .authentication_type import AuthenticationType


class GithubAuthMethod(BaseSsoAuthMethod):
    """Auth method for Github authentication"""

    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_GITHUB

    @classmethod
    def check_session(cls):
        """Check session is valid"""
        return bool(flask.session.get('github_username'))

    @classmethod
    def is_enabled(cls):
        return terrareg.github.Github.is_enabled()

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return flask.session.get('organisations', [])

    def get_username(self):
        """Get username of current user"""
        return flask.session.get('github_username')
