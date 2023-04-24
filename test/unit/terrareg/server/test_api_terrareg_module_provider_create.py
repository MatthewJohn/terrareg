
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


class TestApiTerraregModuleProviderCreate(TerraregUnitTest):
    """Test module provider settings endpoint"""

    def _mock_get_current_auth_method(self, has_permission):
        """Return mock method for get_current_auth_method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.check_namespace_access = unittest.mock.MagicMock(return_value=has_permission)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method, mock_auth_method

    @setup_test_data()
    def test_pre_existing_module_provider(
            self, app_context,
            test_request_context, mock_models,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_clone_url_template') as mock_update_repo_clone_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/create',
                json={
                    'git_tag_format': 'gittag',
                    'repo_clone_url_template': 'https://github.com/unit/test',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Module provider already exists'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_update_repo_clone_url_template.assert_not_called()
            mock_update_git_tag_format.assert_not_called()

    @setup_test_data()
    def test_create_without_repository_details(
            self, app_context, test_request_context, mock_models,
            client
        ):
        """Test update of repository URL."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_clone_url_template') as mock_update_repo_clone_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_browse_url_template') as mock_update_repo_browse_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_base_url_template') as mock_update_repo_base_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_path') as mock_update_git_path, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            with unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event'):
                ns = terrareg.models.Namespace.create('newnamespace', None)

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
            mock_update_repo_clone_url_template.assert_not_called()
            mock_update_git_tag_format.assert_not_called()
            mock_update_repo_browse_url_template.assert_not_called()
            mock_update_repo_base_url_template.assert_not_called()
            mock_update_git_path.assert_not_called()

            m = terrareg.models.Module(namespace=ns, name='newtestmodule')
            p = terrareg.models.ModuleProvider(module=m, name='newprovider')
            assert p._get_db_row() is not None

    @setup_test_data()
    def test_create_module_provider_with_repo_and_tag(self, app_context, test_request_context, mock_models, client):
        """Test update of repository URL with git repo URLs, tag format and git path."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_clone_url_template') as mock_update_repo_clone_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_browse_url_template') as mock_update_repo_browse_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_base_url_template') as mock_update_repo_base_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_path') as mock_update_git_path, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            with unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event'):
                terrareg.models.Namespace.create('newnamespace', None)

            res = client.post(
                '/v1/terrareg/modules/newnamespace/newm/newp/create',
                json={
                    'repo_clone_url_template': 'https://unittest.com/module.git',
                    'repo_base_url_template': 'https://baseurl.com/module.git',
                    'repo_browse_url_template': 'https://browseurl.com/',
                    'git_tag_format': 'unitv{version}test',
                    'git_path': '/acustompath',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'id': 'newnamespace/newm/newp'}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='newnamespace')
            mock_update_repo_clone_url_template.assert_called_once_with(repo_clone_url_template='https://unittest.com/module.git')
            mock_update_git_tag_format.assert_called_once_with(git_tag_format='unitv{version}test')
            mock_update_repo_browse_url_template.assert_called_once_with(repo_browse_url_template='https://browseurl.com/')
            mock_update_repo_base_url_template.assert_called_once_with(repo_base_url_template='https://baseurl.com/module.git')
            mock_update_git_path.assert_called_once_with(git_path='/acustompath')

    @setup_test_data()
    def test_create_module_provider_with_empty_repo_and_tag(self, app_context, test_request_context, mock_models, client):
        """Test update of repository URL with empty URL and git tag format."""
        mock_get_current_auth_method, mock_auth_method = self._mock_get_current_auth_method(True)
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_clone_url_template') as mock_update_repo_clone_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_browse_url_template') as mock_update_repo_browse_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_repo_base_url_template') as mock_update_repo_base_url_template, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_path') as mock_update_git_path, \
                unittest.mock.patch('terrareg.models.ModuleProvider.update_git_tag_format') as mock_update_git_tag_format:

            with unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event'):
                terrareg.models.Namespace.create('newnamespace', None)

            res = client.post(
                '/v1/terrareg/modules/newnamespace/newm/newp/create',
                json={
                    'repo_clone_url_template': '',
                    'repo_base_url_template': '',
                    'repo_browse_url_template': '',
                    'git_tag_format': '',
                    'git_path': '',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'id': 'newnamespace/newm/newp'}
            assert res.status_code == 200

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')
            mock_auth_method.check_namespace_access.assert_called_once_with(UserGroupNamespacePermissionType.FULL, namespace='newnamespace')
            mock_update_repo_clone_url_template.assert_called_once_with(repo_clone_url_template=None)
            mock_update_repo_browse_url_template.assert_called_once_with(repo_browse_url_template=None)
            mock_update_repo_base_url_template.assert_called_once_with(repo_base_url_template=None)
            mock_update_git_tag_format.assert_called_once_with(git_tag_format='')
            mock_update_git_path.assert_called_once_with(git_path='')

    @setup_test_data()
    def test_create_module_provider_with_non_existent_namespace(self, app_context, test_request_context, mock_models, client):
        """Test creation of module provider with non-existent namespace."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(True)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.post(
                '/v1/terrareg/modules/doesnotexist/newm/newp/create',
                json={
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': 'Namespace does not exist'}
            assert res.status_code == 400

            # Ensure required checks are called
            mock_check_csrf.assert_called_once_with('unittestcsrf')

    @setup_test_data()
    def test_create_module_provider_without_permission(self, app_context, test_request_context, mock_models, client):
        """Test creation of module provider without permission."""
        with app_context, test_request_context, client, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._mock_get_current_auth_method(False)[0]), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:

            res = client.post(
                '/v1/terrareg/modules/testnamespace/doesnotexist/testprovider/create',
                json={
                    'git_tag_format': 'gittag',
                    'repo_clone_url_template': 'https://github.com/unit/test',
                    'csrf_token': 'unittestcsrf'
                }
            )
            assert res.json == {'message': "You don't have the permission to access the requested resource. "
                                           "It is either read-protected or not readable by the server."}
            assert res.status_code == 403
