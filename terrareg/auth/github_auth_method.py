
import flask

import terrareg.github
import terrareg.config
import terrareg.models
import terrareg.user_group_namespace_permission_type
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

    def _get_organisation_memeberships(self):
        """Return list of organisations that the user is an owner of"""
        organisations = flask.session.get('organisations', [])
        if not isinstance(organisations, list):
            return []
        return organisations

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return self._get_organisation_memeberships()

    def get_username(self):
        """Get username of current user"""
        if username := flask.session.get('github_username'):
            return username
        return None

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        # If github automatic namespace generation is enabled,
        # allow access to these namespaces
        if terrareg.config.Config().AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES:
            if namespace.lower() in self._get_organisation_memeberships():
                return True

        # If not enabled, or namespace does not match,
        # revert to parent method
        return super().check_namespace_access(permission_type=permission_type, namespace=namespace)

    def get_all_namespace_permissions(self):
        """Get all namespace permissions for authenticated user"""
        namespace_permissions = super().get_all_namespace_permissions()

        # If Github auto-generated namespaces is enabled,
        # add additional permissions for the namespaces
        if terrareg.config.Config().AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES:
            for namespace_name in self._get_organisation_memeberships():
                if namespace_name not in namespace_permissions:
                    # Attempt to get namespace object, and create
                    # if it doesn't exist
                    namespace_obj = terrareg.models.Namespace.get(name=namespace_name, include_redirect=False)
                    if not namespace_obj:
                        namespace_obj = terrareg.models.Namespace.create(name=namespace_name)

                    namespace_permissions[namespace_obj] = terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL

        return namespace_permissions
