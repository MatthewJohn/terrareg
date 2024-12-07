
import unittest.mock

import pytest

import terrareg.audit_action
from test.unit.terrareg import (
    TEST_MODULE_DATA,
    mock_models,
    setup_test_data, TerraregUnitTest
)
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context
import terrareg.models


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
        mock_auth_method.get_username.return_value = 'unittest-mock-username'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test creating new permission using endpoint."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event:

            user_group = terrareg.models.UserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

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
            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.USER_GROUP_NAMESPACE_PERMISSION_ADD,
                object_type='UserGroupNamespacePermission',
                object_id='nopermissions/namespace1',
                old_value=None, new_value=None
            )
            permissions = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 1
            assert permissions[0].namespace.name == 'namespace1'
            assert permissions[0].permission_type == UserGroupNamespacePermissionType.MODIFY

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_invalid_permission_type(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test creating new permission with invalid permission type."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/namespace1',
                json={'permission_type': 'NOTREAL'})
            assert res.json == {'message': 'Invalid namespace permission type'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()

            permissions = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_invalid_namespace(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test creating new permission with invalid namespace."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/doesnotexist',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Namespace does not exist.'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()

            permissions = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
            assert len(permissions) == 0


    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_duplicate_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test creating duplicate permission for user group."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.post(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Permission already exists for this user_group/namespace.'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()

            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_duplicate_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.post(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': 'Permission already exists for this user_group/namespace.'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()

            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_add_permission_without_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test creating new permission without permission to perform action."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('nopermissions')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

            res = client.post(
                '/v1/terrareg/user-groups/nopermissions/permissions/namespace1',
                json={'permission_type': 'MODIFY'})
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            mock_create_audit_event.assert_not_called()

            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test deletion of permission."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            user_group = terrareg.models.UserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1')
            assert res.json == {}
            assert res.status_code == 200

            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.USER_GROUP_NAMESPACE_PERMISSION_DELETE,
                object_type='UserGroupNamespacePermission',
                object_id='onepermissiongroup/namespace1',
                old_value=None, new_value=None
            )

            mock_auth_method.is_admin.assert_called_once()
            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 0

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_non_existent_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test deletion of non-existent permission."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            user_group = terrareg.models.UserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace2')
            assert res.json == {'message': 'Permission does not exist.'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()

            mock_auth_method.is_admin.assert_called_once()
            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_delete_without_permission(
            self, app_context,
            test_request_context,
            mock_models, client
        ):
        """Test deletion of permission without authorization."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]):

            user_group = terrareg.models.UserGroup.get_by_group_name('onepermissiongroup')
            # Ensure user group doesn't have any permissions
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
            assert terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)[0].permission_type == UserGroupNamespacePermissionType.FULL

            res = client.delete(
                '/v1/terrareg/user-groups/onepermissiongroup/permissions/namespace1')
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            mock_create_audit_event.assert_not_called()

            # Assert permissions have not been modified
            assert len(terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group(user_group)) == 1
