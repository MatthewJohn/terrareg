
from lib2to3.pgen2 import grammar
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import (
    TEST_MODULE_DATA, MockUserGroup, MockUserGroupNamespacePermission,
    mock_server_user_groups_fixture, mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from terrareg.auth import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context


test_data = {
    'namespace1': {'id': 1},
    'namespace2': {'id': 2}
}

test_usergroup_data = {
    'nopermissions': {
        'id': 1,
        'site_admin': False,
        'namespace_permissions': {
        }
    },
    'onepermissiongroup': {
        'id': 2,
        'site_admin': False,
        'namespace_permissions': {
            'namespace1': UserGroupNamespacePermissionType.FULL
        }
    },
    'multipermissiongroup': {
        'id': 3,
        'site_admin': False,
        'namespace_permissions': {
            'namespace1': UserGroupNamespacePermissionType.MODIFY,
            'namespace2': UserGroupNamespacePermissionType.FULL
        }
    }
}


class TestApiTerraregAuthUserGroupNamespacePermissions(TerraregUnitTest):
    """Test user group endpoint"""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            user_group = MockUserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {
                'namespace': 'namespace1',
                'permission_type': 'MODIFY',
                'user_group': 'nopermissions',
            }

            assert res.status_code == 201

            # Ensure required checks are called
            mock_auth_method.is_admin.assert_called_once()
            permissions = MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 1
            assert permissions[0].namespace.name == 'namespace1'
            assert permissions[0].permission_type == UserGroupNamespacePermissionType.MODIFY

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_invalid_permission_type(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = MockUserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/namespace1',
                json={'permission_type': 'NOTREAL'})
            assert res.json == {'message': 'Invalid namespace permission type'}
            assert res.status_code == 400

            permissions = MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_invalid_namespace(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = MockUserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/doesnotexist',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Namespace does not exist.'}
            assert res.status_code == 400

            permissions = MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 0


    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_duplicate_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = MockUserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.post(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Permission already exists for this user_group/namespace.'}
            assert res.status_code == 400

            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_duplicate_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = MockUserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.post(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Permission already exists for this user_group/namespace.'}
            assert res.status_code == 400

            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_without_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]):

            user_group = MockUserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            user_group = MockUserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1')
            assert res.json == {}
            assert res.status_code == 200

            mock_auth_method.is_admin.assert_called_once()
            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_non_existent_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            user_group = MockUserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace2')
            assert res.json == {'message': 'Permission does not exist.'}
            assert res.status_code == 400

            mock_auth_method.is_admin.assert_called_once()
            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_without_permission(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            mocked_server_namespace_fixture, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]):

            user_group = MockUserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1')
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            # Assert permissions have not been modified
            assert len(MockUserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
