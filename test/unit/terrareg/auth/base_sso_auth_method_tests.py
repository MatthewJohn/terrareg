
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

    @pytest.mark.parametrize('publish_api_key_config,has_namespace_access,expected_result', [
        # With access to namespace
        (None, True, True),
        ([], True, True),
        (['key1'], True, True),
        (['key1', 'key2'], True, True),
        # Without access to namespace
        (None, False, True),
        ([], False, True),
        (['key1'], False, False),
        (['key1', 'key2'], False, False),
    ])
    def test_can_publish_module_version(self, publish_api_key_config, has_namespace_access, expected_result):
        """Test can_publish_module_version method"""
        mock_check_namespace_access = mock.MagicMock(return_value=has_namespace_access)
        with mock.patch(f'terrareg.auth.{self.CLS.__name__}.check_namespace_access', mock_check_namespace_access), \
                mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', publish_api_key_config):
            obj = self.CLS()
            assert obj.can_publish_module_version(namespace='testnamespace') is expected_result

        if publish_api_key_config:
            mock_check_namespace_access.assert_called_once_with(
                namespace='testnamespace',
                permission_type=UserGroupNamespacePermissionType.MODIFY)
        else:
            mock_check_namespace_access.assert_not_called()

    @pytest.mark.parametrize('upload_api_key_config,has_namespace_access,expected_result', [
        # With access to namespace
        (None, True, True),
        ([], True, True),
        (['key1'], True, True),
        (['key1', 'key2'], True, True),
        # Without access to namespace
        (None, False, True),
        ([], False, True),
        (['key1'], False, False),
        (['key1', 'key2'], False, False),
    ])
    def test_can_upload_module_version(self, upload_api_key_config, has_namespace_access, expected_result):
        """Test can_upload_module_version method"""
        mock_check_namespace_access = mock.MagicMock(return_value=has_namespace_access)
        with mock.patch(f'terrareg.auth.{self.CLS.__name__}.check_namespace_access', mock_check_namespace_access), \
                mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_key_config):
            obj = self.CLS()
            assert obj.can_upload_module_version(namespace='testnamespace') is expected_result

        if upload_api_key_config:
            mock_check_namespace_access.assert_called_once_with(
                namespace='testnamespace',
                permission_type=UserGroupNamespacePermissionType.MODIFY)
        else:
            mock_check_namespace_access.assert_not_called()

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = self.CLS()
        assert obj.is_built_in_admin() is False

    @pytest.mark.parametrize('sso_groups,rbac_enabled,expected_result', [
        ([], True, False),
        (['validgroup', True, False]),
        (['validgroup', 'invalidgroup'], True, False),

        # Passing
        (['siteadmingroup'], True, True),
        (['invalidgroup', 'validgroup', 'siteadmingroup'], True, True),

        # RBAC disabled
        ([], False, True),
        (['validgroup', False, True]),
        (['validgroup', 'invalidgroup'], False, True),
        (['siteadmingroup'], False, True),
        (['invalidgroup', 'validgroup', 'siteadmingroup'], False, True)
    ])
    @setup_test_data(None, user_group_data=user_group_data)
    def test_is_admin(self, sso_groups, rbac_enabled, expected_result):
        """Test is_admin method"""
        mock_get_group_memberships = mock.MagicMock(return_value=sso_groups)

        with mock.patch('terrareg.models.UserGroup', MockUserGroup), \
                mock.patch('terrareg.models.UserGroupNamespacePermission',
                           MockUserGroupNamespacePermission), \
                mock.patch('terrareg.models.Namespace', MockNamespace), \
                mock.patch('terrareg.config.Config.ENABLE_ACCESS_CONTROLS', rbac_enabled), \
                mock.patch(f'terrareg.auth.{self.CLS.__name__}.get_group_memberships', mock_get_group_memberships):
            obj = self.CLS()
            assert obj.is_admin() is expected_result

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = self.CLS()
        assert obj.is_authenticated() is True

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = self.CLS()
        assert obj.requires_csrf_tokens is True