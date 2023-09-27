
import sqlalchemy

import terrareg.models
import terrareg.config
import terrareg.auth
from terrareg.database import Database
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from .base_session_auth_method import BaseSessionAuthMethod


class BaseSsoAuthMethod(BaseSessionAuthMethod):
    """Base methods for SSO based authentication"""

    def get_group_memberships(self):
        """Return list of groups that the user is a member of"""
        raise NotImplementedError

    def is_admin(self):
        """Check if user is an admin"""
        # Check if RBAC is enabled, if not, all authenticated users
        # are treated as admins
        if not terrareg.config.Config().ENABLE_ACCESS_CONTROLS:
            return True

        # Obtain list of user's groups
        for group in self.get_group_memberships():
            user_group = terrareg.models.UserGroup.get_by_group_name(group)
            if user_group is not None and user_group.site_admin:
                return True
        return False

    def can_publish_module_version(self, namespace):
        """Determine if user can publish a module version to given namespace."""
        return (
            # If PUBLISH API keys have not been enabled and
            # RBAC has not been enabled,
            # allow user to publish module versions, as this
            # can be performed without authentication
            ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not terrareg.auth.PublishApiKeyAuthMethod.is_enabled())) or
            # Otherwise, check for MODIFY namespace access
            self.check_namespace_access(namespace=namespace, permission_type=UserGroupNamespacePermissionType.MODIFY)
        )

    def can_upload_module_version(self, namespace):
        """Determine if user can upload a module version to given namespace."""
        return (
            # If UPLOAD API keys have not been enabled and
            # RBAC has not been enabled,
            # allow user to publish module versions, as this
            # can be performed without authentication
            ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not terrareg.auth.UploadApiKeyAuthMethod.is_enabled())) or
            # Otherwise, check for MODIFY namespace access
            self.check_namespace_access(namespace=namespace, permission_type=UserGroupNamespacePermissionType.MODIFY)
        )

    def get_all_namespace_permissions(self):
        """Obtain all namespace permissions for user."""
        # Obtain list of user's groups
        groups = self.get_group_memberships()

        # Find any permissions
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    db.user_group_namespace_permission.c.permission_type,
                    db.namespace.c.namespace
                ).join(
                    db.user_group,
                    db.user_group_namespace_permission.c.user_group_id==db.user_group.c.id
                ).join(
                    db.namespace,
                    db.user_group_namespace_permission.c.namespace_id==db.namespace.c.id
                ).where(
                    db.user_group.c.name.in_(groups)
                )
            )
            return {
                terrareg.models.Namespace(row['namespace']): row['permission_type']
                for row in res
            }

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        # Check admin access
        if self.is_admin():
            return True

        namespace_obj = terrareg.models.Namespace.get(namespace)
        if not namespace_obj:
            return False

        # Obtain list of user's groups
        user_groups = []
        for group in self.get_group_memberships():
            user_group = terrareg.models.UserGroup.get_by_group_name(group)
            # If user group object was found for SSO group,
            # add to list
            if user_group:
                user_groups.append(user_group)

        user_group_permissions = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            user_groups=user_groups,
            namespace=namespace_obj
        )
        for user_group_permission in user_group_permissions:
            if (user_group_permission.permission_type == permission_type or
                    user_group_permission.permission_type == UserGroupNamespacePermissionType.FULL):
                return True
        return False
