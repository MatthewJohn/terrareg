
from unittest import mock
import pytest

from terrareg.auth import UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg import MockNamespace, MockUserGroup, MockUserGroupNamespacePermission, TerraregUnitTest, setup_test_data

# Required as this is sued by BaseOpenidConnectAuthMethod
from test import test_request_context

test_data = {
    'first-namespace': {
        'id': 1
    },
    'second-namespace': {
        'id': 2
    },
    'third-namespace': {
        'id': 3
    },
}

user_group_data = {
    'groupwithnopermissions': {
    },
    'validgroup': {
        'namespace_permissions': {
            'first-namespace': UserGroupNamespacePermissionType.FULL
        }
    },
    'secondgroup': {
        'namespace_permissions': {
            'first-namespace': UserGroupNamespacePermissionType.FULL,
            'second-namespace': UserGroupNamespacePermissionType.MODIFY
        }
    },
    'modifyonly': {
        'namespace_permissions': {
            'first-namespace': UserGroupNamespacePermissionType.MODIFY
        }
    },
    'siteadmingroup': {
        'site_admin': True
    }
}


class BaseSsoAuthMethodTests:

    CLS = None

    @pytest.mark.parametrize('is_site_admin,sso_groups,namespace_to_check,permission_type_to_check,expected_result', [
        ## Failure cases
        # No group memberships
        (
            False,
            [],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            False
        ),
        (
            False,
            [],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            False
        ),
        # No group membership, checking for modify permissions
        (
            False,
            [],
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            False
        ),
        # Not part of any valid groups
        (
            False,
            ['doesnotexist'],
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            False
        ),
        # Check full access with only modify access
        (
            False,
            ['modifyonly'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            False
        ),

        ## Pass cases
        # Check full access with full access defined
        (
            False,
            ['validgroup', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        # Check modify access with full access defined
        (
            False,
            ['validgroup', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            True
        ),
        # Check modify access with modify access defined
        (
            False,
            ['modifyonly', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            True
        ),
        # Check full access with multiple group memberships with permission
        (
            False,
            ['groupwithnopermissions', 'modifyonly', 'validgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        # Check site admin always passes
        (
            True,
            [],
            'first-namespace',
            UserGroupNamespacePermissionType.MODIFY,
            True
        ),
        (
            True,
            [],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        # Site admin with group mappings
        (
            True,
            ['validgroup', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        (
            True,
            ['validgroup', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
        (
            True,
            ['validgroup', 'invalidgroup'],
            'first-namespace',
            UserGroupNamespacePermissionType.FULL,
            True
        ),
    ])
    @setup_test_data(test_data, user_group_data=user_group_data)
    def test_check_namespace_access(self, is_site_admin, sso_groups, namespace_to_check, permission_type_to_check, expected_result):
        """Test check_namespace_access method"""

        mock_get_group_memberships = mock.MagicMock(return_value=sso_groups)
        mock_is_admin = mock.MagicMock(return_value=is_site_admin)

        with mock.patch('terrareg.models.UserGroup', MockUserGroup), \
                mock.patch('terrareg.models.UserGroupNamespacePermission',
                           MockUserGroupNamespacePermission), \
                mock.patch('terrareg.models.Namespace', MockNamespace), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.get_group_memberships', mock_get_group_memberships), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.is_admin', mock_is_admin):
            obj = self.CLS()
            assert obj.check_namespace_access(permission_type_to_check, namespace_to_check) is expected_result
