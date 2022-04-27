
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import (
    TEST_MODULE_DATA, MockModule, MockModuleProvider, MockNamespace, client, mocked_server_namespace_fixture,
    test_request_context, app_context,
    setup_test_data, TerraregUnitTest
)


class TestApiTerraregModuleProviderCreate(TerraregUnitTest):
    """Test module provider settings endpoint"""

    @setup_test_data()
    def test_pre_existing_module_provider(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/create',
                json={
                    'git_tag_format': 'gittag',
                    'repository_url': 'https://github.com/unit/test',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Module provider already exists'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()
            mock_update_git_tag_format.assert_not_called()

    @setup_test_data()
    def test_create_without_repository_details(
            self, app_context, test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            res = client.post(
                '/v1/terrareg/modules/newnamespace/newtestmodule/newprovider/create',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )

            assert res.json == {'id': 'newnamespace/newtestmodule/newprovider'}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()
            mock_update_git_tag_format.assert_not_called()

            ns = MockNamespace(name='newnamespace')
            m = MockModule(namespace=ns, name='newtestmodule')
            p = MockModuleProvider(module=m, name='newprovider')
            assert p._get_db_row() is not None

    @setup_test_data()
    def test_create_module_provider_with_repo_and_tag(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with invalid protocol."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            res = client.post(
                '/v1/terrareg/modules/newns/newm/newp/create',
                json={
                    'repository_url': 'https://unittest.com/module.git',
                    'git_tag_format': 'unitv{version}test',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'id': 'newns/newm/newp'}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_called_once_with(repository_url='https://unittest.com/module.git')
            mock_update_git_tag_format.assert_called_once_with(git_tag_format='unitv{version}test')
