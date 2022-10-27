
from unittest import mock
import pytest

from terrareg.auth import UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg import MockNamespace, MockUserGroup

# Required as this is sued by BaseOpenidConnectAuthMethod
from test import test_request_context


test_data = {
    'first-namespace': {
    },
    'second-namespace': {
    },
    'third-namespace': {
    },
}


class BaseSsoAuthMethodTests:

    CLS = None
    _TEST_DATA = test_data

    @pytest.mark.parametrize('is_site_admin,sso_groups,user_groups_config,namespace_to_check,permission_type_to_check,expected_result', [
        ## Failure cases
        # No group memberships
        (
            False,
            [],
            {},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            False
        ),
        (
            False,
            [],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL},
             'secondgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            False
        ),
        # No group membership, checking for modify permissions
        (
            False,
            [],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL},
             'secondgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            False
        ),
        # Not part of any valid groups
        (
            False,
            ['doesnotexist'],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL},
             'secondgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            False
        ),

        ## Pass cases
        (
            False,
            ['validgroup', 'invalidgroup'],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        # Check site admin always passes
        (
            True,
            [],
            {},
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            True
        ),
        (
            True,
            [],
            {},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        # Site admin with group mappings
        (
            True,
            ['validgroup', 'invalidgroup'],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        (
            True,
            ['validgroup', 'invalidgroup'],
            {'validgroup': {'first-namespace': UserGroupNamespacePermissionType.MODIFY}},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        (
            True,
            ['validgroup', 'invalidgroup'],
            {'validgroup': {'second-namespace': UserGroupNamespacePermissionType.FULL}},
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
    ])
    def test_check_namespace_access(self, is_site_admin, sso_groups, user_groups_config, namespace_to_check, permission_type_to_check, expected_result):
        """Test check_namespace_access method"""
        def mock_get_by_group_name_side_effect(name):
            if name in user_groups_config:
                return MockUserGroup(name)
            return None
        mock_get_by_group_name = mock.MagicMock(side_effect=mock_get_by_group_name_side_effect)

        def mock_get_permissions_by_user_groups_and_namespace_side_effect(user_groups, namespace):
            permissions = []
            for user_group in user_groups:
                if user_group.name in user_groups_config and namespace.name in user_groups_config[user_group.name]:
                    permissions.append(user_groups_config[user_group.name][namespace.name])
            return permissions
        mock_get_permissions_by_user_groups_and_namespace = mock.MagicMock(side_effect=mock_get_permissions_by_user_groups_and_namespace_side_effect)

        mock_get_group_memberships = mock.MagicMock(return_value=sso_groups)
        mock_is_admin = mock.MagicMock(return_value=is_site_admin)

        with mock.patch('terrareg.models.UserGroup.get_by_group_name', mock_get_by_group_name), \
                mock.patch('terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace',
                           mock_get_permissions_by_user_groups_and_namespace), \
                mock.patch('terrareg.models.Namespace', MockNamespace), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.get_group_memberships', mock_get_group_memberships), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.is_admin', mock_is_admin):
            obj = self.CLS()
            assert obj.check_namespace_access(permission_type_to_check, namespace_to_check) is expected_result
