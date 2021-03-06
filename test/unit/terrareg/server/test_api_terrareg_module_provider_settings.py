
import unittest.mock

import pytest

from test.unit.terrareg import (
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client, app_context, test_request_context



class TestApiTerraregModuleProviderSettings(TerraregUnitTest):
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
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': repository_url,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_called_once_with(repo_clone_url_template=repository_url)

    @setup_test_data()
    def test_update_repository_url_invalid_protocol(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with invalid protocol."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'nope://unittest.com/module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL contains an unknown scheme (e.g. https/ssh/http)'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_invalid_domain(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL with an invalid domain."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https:///module.git',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL does not contain a host/domain'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_not_called()

    @setup_test_data()
    def test_update_repository_url_without_path(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a path."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https://example.com',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Repo clone URL: Repository URL does not contain a path'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_not_called()

    @setup_test_data()
    def test_update_repository_without_csrf(self, app_context, test_request_context, mocked_server_namespace_fixture, client):
        """Test update of repository URL without a CSRF token."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'repo_clone_url_template': 'https://example.com/test.git'
                }
            )
            assert res.json == {'message': 'No session is presesnt to check CSRF token', 'status': 'Error'}
            assert res.status_code == 500

            # Ensure required checks are called
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_not_called()

    @setup_test_data()
    def test_update_git_tag_format(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of git tag format."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'git_tag_format': 'newgittagformat',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_called_with('newgittagformat')
            mock_update_attributes.assert_not_called()

    @setup_test_data()
    def test_update_empty_git_tag_format(
            self, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of git tag format with empty value."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'git_tag_format': '',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_called_with('')
            mock_update_attributes.assert_not_called()

    @pytest.mark.parametrize('verified_state', [True, False])
    @setup_test_data()
    def test_update_verified_flag(
            self, verified_state, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of verified flag."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'verified': verified_state,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_called_with(verified=verified_state)

    @pytest.mark.parametrize('verified_state', ['', 'isastring'])
    @setup_test_data()
    def test_update_verified_flag_invalid_value(
            self, verified_state, app_context,
            test_request_context, mocked_server_namespace_fixture,
            client
        ):
        """Test update of verified flag with invalid value."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.server.check_admin_authentication', return_value=True) as mocked_check_admin_authentication, \
                unittest.mock.patch('terrareg.server.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_attributes') as mock_update_attributes:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/settings',
                json={
                    'verified': verified_state,
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': {'verified': 'Whether module provider is marked as verified.'}}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_not_called()
            mocked_check_admin_authentication.assert_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_attributes.assert_not_called()
