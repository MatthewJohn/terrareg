
from unittest import mock
import unittest.mock

from test.unit.terrareg import TerraregUnitTest, setup_test_data, mock_models
from test import client
from test import app_context, test_request_context
import terrareg.models
import terrareg.errors


class TestApiTerraregGitProviders(TerraregUnitTest):
    """Test TestApiTerraregGitProviders resource."""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method."""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.is_admin = unittest.mock.MagicMock(return_value=has_permission)
        mock_auth_method.can_access_read_api = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    def test_with_no_git_providers_configured(self, mock_models, client):
        """Test endpoint when no git providers are configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == []

    @setup_test_data()
    def test_with_git_providers_configured(self, mock_models, client):
        """Test endpoint with git providers configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == [
            {
                'id': 1,
                'name': 'testgitprovider',
                'base_url_template': 'https://localhost.com/{namespace}/{module}-{provider}',
                'clone_url_template': 'ssh://localhost.com/{namespace}/{module}-{provider}',
                'browse_url_template': 'https://localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
                'git_path_template': None
            },
            {
                'id': 2,
                'name': 'second-git-provider',
                'base_url_template': 'https://localhost2.example/{namespace}-{module}-{provider}',
                'clone_url_template': 'ssh://localhost2.com/{namespace}/{module}-{provider}',
                'browse_url_template': 'https://localhost2.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
                'git_path_template': '/modules/{module}/'
            },
            {
                'id': 3,
                'name': 'third-git-provider',
                'base_url_template': 'https://localhost2.example/{namespace}-{module}-{provider}',
                'clone_url_template': 'https://localhost2.com/{namespace}/{module}-{provider}',
                'browse_url_template': 'https://localhost2.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
                'git_path_template': '/modules/{module}/'
            },
        ]

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/terrareg/git_providers')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)

    def test_create_git_provider(self, client, app_context, test_request_context):
        """Test creating a git provider."""
        created_provider = mock.MagicMock(
            pk=9,
            name='runtime-provider',
            base_url_template='https://example.com/{namespace}/{module}',
            clone_url_template='ssh://git@example.com/{namespace}/{module}.git',
            browse_url_template='https://example.com/{namespace}/{module}/tree/{tag}/{path}',
            git_path_template='/{provider}'
        )
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.GitProvider.create', return_value=created_provider) as mock_create:
            res = client.post('/v1/terrareg/git_providers', json={
                'name': 'runtime-provider',
                'base_url_template': 'https://example.com/{namespace}/{module}',
                'clone_url_template': 'ssh://git@example.com/{namespace}/{module}.git',
                'browse_url_template': 'https://example.com/{namespace}/{module}/tree/{tag}/{path}',
                'git_path_template': '/{provider}',
                'csrf_token': 'csrf-token'
            })

        assert res.status_code == 201
        assert res.json == {
            'id': 9,
            'name': 'runtime-provider',
            'base_url_template': 'https://example.com/{namespace}/{module}',
            'clone_url_template': 'ssh://git@example.com/{namespace}/{module}.git',
            'browse_url_template': 'https://example.com/{namespace}/{module}/tree/{tag}/{path}',
            'git_path_template': '/{provider}'
        }
        mock_check_csrf.assert_called_once_with('csrf-token')
        mock_create.assert_called_once_with(
            name='runtime-provider',
            base_url_template='https://example.com/{namespace}/{module}',
            clone_url_template='ssh://git@example.com/{namespace}/{module}.git',
            browse_url_template='https://example.com/{namespace}/{module}/tree/{tag}/{path}',
            git_path_template='/{provider}'
        )

    def test_update_git_provider(self, client, app_context, test_request_context):
        """Test updating a git provider."""
        git_provider = mock.MagicMock(
            pk=4,
            name='updated-provider',
            base_url_template='https://example.com/scm/{namespace}/{module}',
            clone_url_template='ssh://git@example.com/scm/{namespace}/{module}.git',
            browse_url_template='https://example.com/scm/{namespace}/{module}/tree/{tag}/{path}',
            git_path_template=None
        )
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.GitProvider.get', return_value=git_provider) as mock_get:
            res = client.post('/v1/terrareg/git_providers/4', json={
                'name': 'updated-provider',
                'base_url_template': 'https://example.com/scm/{namespace}/{module}',
                'clone_url_template': 'ssh://git@example.com/scm/{namespace}/{module}.git',
                'browse_url_template': 'https://example.com/scm/{namespace}/{module}/tree/{tag}/{path}',
                'git_path_template': '',
                'csrf_token': 'csrf-token'
            })

        assert res.status_code == 200
        git_provider.update.assert_called_once_with(
            name='updated-provider',
            base_url_template='https://example.com/scm/{namespace}/{module}',
            clone_url_template='ssh://git@example.com/scm/{namespace}/{module}.git',
            browse_url_template='https://example.com/scm/{namespace}/{module}/tree/{tag}/{path}',
            git_path_template=''
        )
        mock_check_csrf.assert_called_once_with('csrf-token')
        mock_get.assert_called_once_with(id=4)

    def test_delete_git_provider_in_use(self, client, app_context, test_request_context):
        """Test delete returns validation error when provider is in use."""
        git_provider = mock.MagicMock()
        git_provider.delete.side_effect = terrareg.errors.GitProviderInUseError('provider in use')

        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.models.GitProvider.get', return_value=git_provider):
            res = client.delete('/v1/terrareg/git_providers/2')

        assert res.status_code == 400
        assert res.json == {'status': 'Error', 'message': 'provider in use'}
