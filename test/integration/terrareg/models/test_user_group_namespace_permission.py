
import pytest
from terrareg.database import Database

from terrareg.models import Namespace, Session, UserGroup, UserGroupNamespacePermission
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test.integration.terrareg import TerraregIntegrationTest

class TestUserGroup(TerraregIntegrationTest):
    """Test UserGroup model class"""

    def setup_method(self, method):
        """Remove any pre-existing user groups and permissions before running each test."""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group_namespace_permission.delete())
            conn.execute(db.user_group.delete())
        super(TestUserGroup, self).setup_method(method)

    @pytest.mark.parametrize('permission_type', [
        UserGroupNamespacePermissionType.FULL,
        UserGroupNamespacePermissionType.MODIFY
    ])
    def test_create_permission(self, permission_type):
        """Test create method, successfully creating user group namespace permission"""
        user_group = UserGroup.create('unittest-usergroup', site_admin=False)
        namespace = Namespace.get('moduledetails')
        permission = UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=namespace,
            permission_type=permission_type
        )
        assert permission.namespace == namespace
        assert permission.user_group == user_group
        assert permission.permission_type == permission_type

    def test_create_duplicate_permission(self):
        """Test creating a permission that has a duplicate namespace/usergroup"""
        user_group = UserGroup.create('duplicate', site_admin=True)

        namespace = Namespace.get('moduledetails')
        permission = UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=namespace,
            permission_type=UserGroupNamespacePermissionType.FULL
        )
        initial_db_row = permission._get_db_row()

        second_permission = UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=namespace,
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        assert second_permission == None

        # Check user group was not created and original user group
        # has been unchanged
        all_permissions = UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
        assert len(all_permissions) == 1
        assert all_permissions[0]._get_db_row() == initial_db_row

    def test_get_permissions_by_user_group(self):
        """Test get_by_group_name method"""
        # Test without any user groups setup
        user_group = UserGroup.create('hasnopermissions', site_admin=True)
        assert UserGroupNamespacePermission.get_permissions_by_user_group(user_group) == []

        # Setup test permissions
        UserGroupNamespacePermission.create(
            user_group=user_group, namespace=Namespace.get('moduledetails'),
            permission_type=UserGroupNamespacePermissionType.FULL)
        UserGroupNamespacePermission.create(
            user_group=user_group, namespace=Namespace.get('testnamespace'),
            permission_type=UserGroupNamespacePermissionType.MODIFY)
        UserGroupNamespacePermission.create(
            user_group=user_group, namespace=Namespace.get('moduleextraction'),
            permission_type=UserGroupNamespacePermissionType.FULL)

        # Test getting permissions by user group
        permissions = UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
        assert len(permissions) == 3
        assert {
            permission.namespace.name: permission.permission_type
            for permission in permissions
        } == {
            'moduledetails': UserGroupNamespacePermissionType.FULL,
            'testnamespace': UserGroupNamespacePermissionType.MODIFY,
            'moduleextraction': UserGroupNamespacePermissionType.FULL
        }
        assert [
            permission.user_group
            for permission in permissions
        ] == [
            user_group, user_group, user_group
        ]

    def test_get_permissions_by_user_group_and_namespace(self):
        """Test get_permissions_by_user_group_and_namespace method"""
        # Setup other data
        other_user_group = UserGroup.create('anotherusergroup', site_admin=False)
        UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('moduleextraction'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('testnamespace'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )

        # Test without any permissions
        user_group = UserGroup.create('usergroup', site_admin=False)

        assert UserGroupNamespacePermission.get_permissions_by_user_group_and_namespace(
            user_group=user_group,
            namespace=Namespace.get('moduleextraction')
        ) == None

        # Test with permissions to other namespaces setup
        UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('testnamespace'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        assert UserGroupNamespacePermission.get_permissions_by_user_group_and_namespace(
            user_group=user_group,
            namespace=Namespace.get('moduleextraction')
        ) == None

        # Test with permissions present
        permission = UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=Namespace.get('moduleextraction'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        assert UserGroupNamespacePermission.get_permissions_by_user_group_and_namespace(
            user_group=user_group,
            namespace=Namespace.get('moduleextraction')
        ) == permission

    def test_get_permissions_by_user_groups_and_namespace(self):
        """Test get_permissions_by_user_groups_and_namespace method"""

        # Test with no groups
        assert UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            [],
            Namespace.get('moduleextraction')
        ) == []

        # Setup other data
        other_user_group = UserGroup.create('anotherusergroup', site_admin=False)

        # Test without any permissions
        user_group = UserGroup.create('usergroup', site_admin=False)

        # Test with one group and no permissions
        assert UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            [user_group],
            Namespace.get('moduleextraction')
        ) == []
        # Test with two groups and no permissions
        assert UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            [user_group, other_user_group],
            Namespace.get('moduleextraction')
        ) == []

        permission1 = UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('moduleextraction'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('testnamespace'),
            permission_type=UserGroupNamespacePermissionType.MODIFY
        )
        UserGroupNamespacePermission.create(
            user_group=other_user_group,
            namespace=Namespace.get('testnamespace'),
            permission_type=UserGroupNamespacePermissionType.FULL
        )

        # Test with two groups and no permissions
        permissions = UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            [user_group, other_user_group],
            Namespace.get('moduleextraction')
        )
        assert len(permissions) == 1
        assert permission1 in permissions

        # Test with permissions present
        permission2 = UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=Namespace.get('moduleextraction'),
            permission_type=UserGroupNamespacePermissionType.FULL
        )
        permissions = UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            [user_group, other_user_group],
            namespace=Namespace.get('moduleextraction')
        )
        assert len(permissions) == 2
        assert permission1 in permissions
        assert permission2 in permissions
