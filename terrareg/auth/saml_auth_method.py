
import flask

import terrareg.saml
from .base_sso_auth_method import BaseSsoAuthMethod
from .authentication_type import AuthenticationType


class SamlAuthMethod(BaseSsoAuthMethod):
    """Auth method for SAML authentication"""

    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_SAML

    @classmethod
    def check_session(cls):
        """Check SAML session is valid"""
        return bool(flask.session.get('samlUserdata'))

    @classmethod
    def is_enabled(cls):
        return terrareg.saml.Saml2.is_enabled()

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        user_data_groups = flask.session.get('samlUserdata', None)
        if user_data_groups and isinstance(user_data_groups, dict):
            groups = user_data_groups.get(terrareg.config.Config().SAML2_GROUP_ATTRIBUTE)
            if isinstance(groups, list):
                return groups
        return []

    def get_username(self):
        """Get username of current user"""
        return flask.session.get('samlNameId')
