
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import (
    TEST_MODULE_DATA, mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
from terrareg.auth import UserGroupNamespacePermissionType
from test import client, app_context, test_request_context


class TestApiTerraregNamespaces(TerraregUnitTest):
    """Test module provider settings endpoint"""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data()
    def test_create_pre_existing_namespace(
            self, app_context,
            test_request_context, mock_models,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.post(
                '/v1/terrareg/namespaces',
                json={
                    'name': 'moduleextraction',
                    'display_name': 'Test Display Name',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Unittest namespace already exists', 'status': 'Error'}
            assert res.status_code == 500

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')

    @setup_test_data()
    def test_create_without_display_name(
            self, app_context, test_request_context, mock_models,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.post(
                '/v1/terrareg/namespaces',
                json={
                    'name': 'missing-display-name',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {
                'display_name': None,
                'name': 'missing-display-name',
                'view_href': '/modules/missing-display-name'
            }
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')

            ns = terrareg.models.Namespace(name='missing-display-name')
            assert ns._get_db_row() is not None

    @setup_test_data()
    def test_create_without_name(self, app_context, test_request_context, mock_models, client):
        """Test update of repository URL with invalid protocol."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.post(
                '/v1/terrareg/namespaces',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {
                'message': 'Namespace name is invalid',
                'status': 'Error'
            }
            assert res.status_code == 500

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')

    @setup_test_data()
    def test_create_namespace_without_permission(self, app_context, test_request_context, mock_models, client):
        """Test creation of module provider without permission."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.Namespace.create', unittest.mock.MagicMock()) as mock_namespace_create:

            res = client.post(
                '/v1/terrareg/namespaces',
                json={
                    'name': 'no-permissions',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403

            mock_namespace_create.assert_not_called()
