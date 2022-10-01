
from operator import mod
import unittest.mock

import pytest

import terrareg.models
from test.unit.terrareg import (
    TEST_MODULE_DATA, MockModule, MockModuleProvider, MockNamespace, mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client, app_context, test_request_context


class TestApiTerraregModuleProviderDelete(TerraregUnitTest):
    """Test module version delete endpoint"""

    @setup_test_data()
    def test_delete(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test deletion of module version."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch.object(terrareg.models.ModuleVersion, 'delete', autospec=True) as mock_module_version_delete:

            res = client.delete(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/delete',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_module_version_delete.assert_called_once()

            object_to_delete = mock_module_version_delete.mock_calls[0].args[0]
            assert object_to_delete.version == '2.4.1'
            assert object_to_delete._module_provider.name == 'testprovider'
            assert object_to_delete._module_provider._module.name == 'testmodulename'
            assert object_to_delete._module_provider._module._namespace.name == 'testnamespace'

    @setup_test_data()
    def test_delete_non_existing_version(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test delete of non-existant module version."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleVersion.delete') as mock_module_version_delete:

            res = client.delete(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/9.93.0/delete',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Module version does not exist'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_module_version_delete.assert_not_called()

    @setup_test_data()
    def test_delete_non_existing_module_provider(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test delete of module version of non-existent module provider."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleVersion.delete') as mock_module_version_delete:

            res = client.delete(
                '/v1/terrareg/modules/emptynamespace/doesnotexist/doesnotexist/9.93.0/delete',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Module provider does not exist'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_module_version_delete.assert_not_called()


    @setup_test_data()
    def test_delete_non_existing_namespace(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test delete of module version of non-existent namespace."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleVersion.delete') as mock_module_version_delete:

            res = client.delete(
                '/v1/terrareg/modules/doesnotexist/doesnotexist/doesnotexist/9.93.0/delete',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Namespace does not exist'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_module_version_delete.assert_not_called()
