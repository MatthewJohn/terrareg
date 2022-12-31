
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.models


class ApiTerraregAuthUserGroupNamespacePermissions(ErrorCatchingResource):
    """Interface to create user groups namespace permissions."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _post(self, user_group, namespace):
        """Create user group namespace permission"""
        attributes = request.json
        permission_type = attributes.get('permission_type')
        try:
            permission_type_enum = terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType(permission_type)
        except ValueError:
            return {'message': 'Invalid namespace permission type'}, 400

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if not namespace_obj:
            return {'message': 'Namespace does not exist.'}, 400
        user_group_obj = terrareg.models.UserGroup.get_by_group_name(user_group)
        if not user_group_obj:
            return {'message': 'User group does not exist.'}, 400

        user_group_namespace_permission = terrareg.models.UserGroupNamespacePermission.create(
            user_group=user_group_obj,
            namespace=namespace_obj,
            permission_type=permission_type_enum
        )
        if user_group_namespace_permission:
            return {
                'user_group': user_group_obj.name,
                'namespace': namespace_obj.name,
                'permission_type': permission_type_enum.value
            }, 201
        else:
            return {'message': 'Permission already exists for this user_group/namespace.'}, 400

    def _delete(self, user_group, namespace):
        """Delete user group namespace permission"""
        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if not namespace_obj:
            return {'message': 'Namespace does not exist.'}, 400
        user_group_obj = terrareg.models.UserGroup.get_by_group_name(user_group)
        if not user_group_obj:
            return {'message': 'User group does not exist.'}, 400

        user_group_namespace_permission = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group_and_namespace(
            user_group=user_group_obj,
            namespace=namespace_obj
        )
        if not user_group_namespace_permission:
            return {'message': 'Permission does not exist.'}, 400

        user_group_namespace_permission.delete()
        return {}, 200
