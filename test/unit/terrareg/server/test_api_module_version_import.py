
import unittest.mock

import pytest

from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
import terrareg.models
from test import client


class TestApiModuleVersionImport(TerraregUnitTest):
    """Test module version import resource."""

    def _get_mock_get_current_auth_method(self, allowed_to_create):
        """Return mock auth method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_upload_module_version = unittest.mock.MagicMock(return_value=allowed_to_create)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method

    @setup_test_data()
    def test_import_with_no_module_provider_repository_url(self, client, mock_models):
        """Test importing a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenorepourl/testprovider/import',
                json={'version': '5.5.4'}
            )
            assert res.status_code == 400
            assert res.json == {
                'status': 'Error',
                'message': 'Module provider is not configured with a repository'
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_import_by_version_with_valid_repository_url(self, client, mock_models):
        """Import a module version by version"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'version': '5.5.4'}
            )
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_import_by_git_tag_with_valid_repository_url(self, client, mock_models):
        """Import a module version by git_tag"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            test_module_provider = terrareg.models.ModuleProvider.get(
                module=terrareg.models.Module(
                    namespace=terrareg.models.Namespace.get(name='testnamespace'),
                    name='modulewithrepourl'
                ),
                name='testprovider'
            )
            test_module_provider.update_attributes(git_tag_format='v{version}')

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'git_tag': 'v5.5.4'}
            )
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    @pytest.mark.parametrize('pre_existing_published_module_version', [
        False,
        True
    ])
    def test_hook_with_reindexing_published_module(self, pre_existing_published_module_version, client, mock_models):
        """Test hook call whilst re-indexing a published module."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=pre_existing_published_module_version) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.models.ModuleVersion.publish') as mocked_publish, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'version': '5.5.4'}
            )
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()
            if pre_existing_published_module_version:
                mocked_publish.assert_called_once_with()
            else:
                mocked_publish.assert_not_called()

    @setup_test_data()
    def test_import_with_non_existent_module_provider(self, client, mock_models):
        """Attempt to import a module with a non-existent module provider"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/moduledoesnotexist/testprovider/import',
                json={'version': '5.5.4'}
            )
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_import_with_invalid_authentication(self, client, mock_models):
        """Test importing a module version with invalid API authentication."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(False)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'version': '5.5.4'}
            )
            assert res.status_code == 403
            assert res.json == {
                'message': "You don't have the permission to access the requested resource. "
                           "It is either read-protected or not readable by the server."}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    @pytest.mark.parametrize('post_data', [
        {},
        {'git_tag': ''}, {'git_tag': None},
        {'version': ''}, {'version': None},
        {'version': ''}, {'version': None},
        {'git_tag': '', 'version': ''},
        {'git_tag': None, 'version': None},
    ])
    def test_import_without_specifying_version(self, post_data, client, mock_models):
        """Attempt to import without specifying a git_tag or version"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/moduledoesnotexist/testprovider/import',
                json=post_data
            )
            assert res.status_code == 400
            assert res.json == {'status': 'Error', 'message': 'Either git_tag or version must be provided'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @pytest.mark.parametrize('git_tag_format, tag', [
        ('v{version}', 'notvalid'),
        ('v{major}.{minor}', 'v1.2.3'),
        ('v{major}.{minor}.{patch}', 'v1.2.'),
        ('{version}', 'v1.2.3'),
    ])
    @setup_test_data()
    def test_import_with_invalid_tag(self, git_tag_format, tag, client, mock_models):
        """Attempt import with a git tag that does not match the git tag format"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            test_module_provider = terrareg.models.ModuleProvider.get(
                module=terrareg.models.Module(
                    namespace=terrareg.models.Namespace.get(name='testnamespace'),
                    name='modulewithrepourl'
                ),
                name='testprovider'
            )
            test_module_provider.update_attributes(git_tag_format=git_tag_format)

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'git_tag': tag}
            )
            assert res.status_code == 400
            assert res.json == {
                'status': 'Error',
                'message': 'Version could not be derrived from git tag. Ensure it matches the git_tag_format template for this module provider'
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @pytest.mark.parametrize('git_tag_format', [
        'v{major}',
        'v{minor}',
        'v{patch}',
        '{major}.{minor}.{patch}',
    ])
    @setup_test_data()
    def test_import_by_version_when_using_non_semantic_git_tag_format(self, git_tag_format, client, mock_models):
        """Attempt import by version for a module provider that does not have a git tag format with {version} placeholder."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            test_module_provider = terrareg.models.ModuleProvider.get(
                module=terrareg.models.Module(
                    namespace=terrareg.models.Namespace.get(name='testnamespace'),
                    name='modulewithrepourl'
                ),
                name='testprovider'
            )
            test_module_provider.update_attributes(git_tag_format=git_tag_format)

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'version': '1.2.3'}
            )
            assert res.status_code == 400
            assert res.json == {
                'status': 'Error',
                'message': 'Module provider is not configured with a git tag format containing a {version} placeholder. '
                           'Indexing must be performed using the git_tag argument'
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_import_with_invalid_version_format(self, client, mock_models):
        """Attempt import by version with invalid version string"""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/import',
                json={'version': 'thisisinvalid'}
            )
            assert res.json == {'status': 'Error', 'message': 'Version is invalid'}
            assert res.status_code == 500

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()
