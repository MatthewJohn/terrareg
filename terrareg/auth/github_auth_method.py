
import flask

import terrareg.github
import terrareg.config
import terrareg.models
import terrareg.user_group_namespace_permission_type
import terrareg.namespace_type
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
        """Return map of organisations that the user is an owner of with the namespace type"""
        organisations = flask.session.get('organisations', {})
        if not isinstance(organisations, dict):
            return {}
        return {
            org: terrareg.namespace_type.NamespaceType(organisations[org])
            for org in organisations
        }

    def get_github_organisations(self):
        """Return map of Github organisations that the user is an owner of to namespace type"""
        return self._get_organisation_memeberships()

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return [org for org in self._get_organisation_memeberships()]

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
            memberships = self._get_organisation_memeberships()
            for namespace_name in memberships:
                if namespace_name not in namespace_permissions:
                    # Attempt to get namespace object, and create
                    # if it doesn't exist
                    namespace_obj = terrareg.models.Namespace.get(name=namespace_name, include_redirect=False)
                    if not namespace_obj:
                        namespace_obj = terrareg.models.Namespace.create(name=namespace_name, type_=memberships[namespace_name])
                    else:
                        # If the namespace does exist, ensure the mapping of namespace type is correct
                        if namespace_obj.namespace_type is not memberships[namespace_name]:
                            namespace_obj.update_attributes(namespace_type=memberships[namespace_name])

                    namespace_permissions[namespace_obj] = terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL

        return namespace_permissions
