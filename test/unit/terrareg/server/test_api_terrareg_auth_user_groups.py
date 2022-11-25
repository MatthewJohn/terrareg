
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import (
    TEST_MODULE_DATA, mock_server_user_groups_fixture,
    setup_test_data, TerraregUnitTest
)
from terrareg.auth import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context

test_data = {
    'namespace1': {'id': 1},
    'namespace2': {'id': 2}
}

test_usergroup_data = {
    'siteadmingroup': {
        'id': 1,
        'site_admin': True,
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


class TestApiTerraregAuthUserGroups(TerraregUnitTest):
    """Test user groups endpoint"""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data(test_data, user_group_data=test_usergroup_data)
    def test_get(
            self, app_context,
            test_request_context, mock_server_user_groups_fixture,
            client
        ):
        """Test update of repository URL."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.get('/v1/terrareg/user-groups')
            assert res.json == [
                {
                    'name': 'siteadmingroup',
                    'namespace_permissions': [],
                    'site_admin': True
                },
                {
                    'name': 'onepermissiongroup',
                    'namespace_permissions': [
                        {'namespace': 'namespace1',
                         'permission_type': 'FULL'}
                    ],
                    'site_admin': False
                },
                {
                    'name': 'multipermissiongroup',
                    'namespace_permissions': [
                        {'namespace': 'namespace1',
                         'permission_type': 'MODIFY'},
                        {'namespace': 'namespace2',
                         'permission_type': 'FULL'}
                    ],
                    'site_admin': False
                },
            ]
            assert res.status_code == 200

            # Ensure required checks are called
            mock_auth_method.is_admin.assert_called_once()

    @setup_test_data()
    def test_create_module_provider_without_permission(self, app_context, test_request_context, mock_server_user_groups_fixture, client):
        """Test creation of module provider without permission."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]), \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.get('/v1/terrareg/user-groups')
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           'It is either read-protected or not readable by the server.'}
            assert res.status_code == 403
