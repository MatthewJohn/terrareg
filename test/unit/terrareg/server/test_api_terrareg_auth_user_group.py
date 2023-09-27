


from operator import mod
import unittest.mock

import pytest

import terrareg.audit_action
from test.unit.terrareg import (
    TEST_MODULE_DATA,
    setup_test_data, TerraregUnitTest,
    mock_models
)
import terrareg.models
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context


test_usergroup_data = {
    'todelete': {
        'id': 1,
        'site_admin': False,
        'namespace_permissions': {}
    }
}


class TestApiTerraregAuthUserGroup(TerraregUnitTest):
    """Test user group endpoint"""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data(user_group_data=test_usergroup_data)
    def test_delete(
            self, app_context,
            test_request_context, mock_models,
            client
        ):
        """Test update of repository URL."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            # Ensure user group exists
            assert len(terrareg.models.UserGroup.get_all_user_groups()) == 1

            res = client.delete('/v1/terrareg/user-groups/todelete')
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_auth_method.is_admin.assert_called_once()
            assert terrareg.models.UserGroup.get_all_user_groups() == []

            mock_create_audit_event.assert_called_once_with(
                action=terrareg.audit_action.AuditAction.USER_GROUP_DELETE,
                object_type='UserGroup', object_id='todelete',
                old_value=None, new_value=None
            )

    @setup_test_data(user_group_data=test_usergroup_data)
    def test_delete_non_existent(self, app_context, test_request_context, mock_models, client):
        """Test creation of module provider without permission."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.delete('/v1/terrareg/user-groups/doesnotexist')
            assert res.json == {'message': 'User group does not exist.'}
            assert res.status_code == 400

            mock_create_audit_event.assert_not_called()


    @setup_test_data(user_group_data=test_usergroup_data)
    def test_delete_without_permissions(self, app_context, test_request_context, mock_models, client):
        """Test creation of module provider without permission."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event') as mock_create_audit_event, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.delete('/v1/terrareg/user-groups/todelete')
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           'It is either read-protected or not readable by the server.'}
            assert res.status_code == 403

            mock_create_audit_event.assert_not_called()
