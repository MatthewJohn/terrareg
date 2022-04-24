
import unittest.mock

import pytest

from test.unit.terrareg import (
    client, mocked_server_namespace_fixture,
    test_request_context, app_context,
    setup_test_data
)


class TestApiTerraregModuleProviderSettings:
    """Test module provider settings endpoint"""

    @pytest.mark.parametrize('repository_url', [
        'https://unittest.com/module.git',
        'http://unittest.com/module.git',
        'ssh://unittest.com/module.git'
    ])
    @setup_test_data()
    def test_update_repository_url(
            self, repository_url, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            print(repository_url)
            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': repository_url,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_called_once_with(
                repository_url=repository_url)

    @setup_test_data()
    def test_update_repository_url_invalid_protocol(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with invalid protocol."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'nope://unittest.com/module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL contains an unknown scheme (e.g. https/git/http)'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_invalid_domain(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with an invalid domain."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https:///module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL does not contain a host/domain'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_without_path(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a path."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https://example.com',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repository URL does not contain a path'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_not_called()

    @setup_test_data()
    def test_update_repository_without_csrf(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a CSRF token."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repository_url') as mock_update_repository_url:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repository_url': 'https://example.com/test.git'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with(None)
            mocked_check_admin_authentication.assert_called()
            mock_update_repository_url.assert_called_once_with(
                repository_url='https://example.com/test.git')