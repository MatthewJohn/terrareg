
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models


class ApiTerraregAuthUserGroups(ErrorCatchingResource):
    """Interface to list and create user groups."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _get(self):
        """Obtain list of user groups."""
        return [
            {
                'name': user_group.name,
                'site_admin': user_group.site_admin,
                'namespace_permissions': [
                    {
                        'namespace': permission.namespace.name,
                        'permission_type': permission.permission_type.value
                    }
                    for permission in terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group=user_group)
                ]
            }
            for user_group in terrareg.models.UserGroup.get_all_user_groups()
        ]

    def _post(self):
        """Create user group"""
        attributes = request.json
        name = attributes.get('name')
        site_admin = attributes.get('site_admin')

        if site_admin is not True and site_admin is not False:
            return {}, 400

        user_group = terrareg.models.UserGroup.create(name=name, site_admin=site_admin)
        if user_group:
            return {
                'name': user_group.name,
                'site_admin': user_group.site_admin
            }, 201
        else:
            return {}, 400