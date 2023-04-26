
import unittest.mock

import pytest

from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionCreate(TerraregUnitTest):
    """Test module version creation resource."""

    def _get_mock_get_current_auth_method(self, allowed_to_create):
        """Return mock auth method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_upload_module_version = unittest.mock.MagicMock(return_value=allowed_to_create)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method

    @setup_test_data()
    def test_creation_with_no_module_provider_repository_url(self, client, mock_models):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch('terrareg.module_version_create.module_version_create') as mocked_module_version_create, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenorepourl/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_module_version_create.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_creation_with_valid_repository_url(self, client, mock_models):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch('terrareg.module_version_create.module_version_create') as mocked_module_version_create, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_module_version_create.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_hook_with_reindexing_published_module(self, client, mock_models):
        """Test hook call whilst re-indexing a published module."""
        with unittest.mock.patch('terrareg.module_version_create.module_version_create') as mocked_module_version_create, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_module_version_create.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_creation_with_non_existent_module_provider(self, client, mock_models):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch('terrareg.module_version_create.module_version_create') as mocked_module_version_create, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/moduledoesnotexist/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mocked_module_version_create.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_creation_with_invalid_authentication(self, client, mock_models):
        """Test creating a module version with invalid API authentication."""
        with unittest.mock.patch('terrareg.module_version_create.module_version_create') as mocked_module_version_create, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(False)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.status_code == 403
            assert res.json == {
                'message': "You don't have the permission to access the requested resource. "
                           "It is either read-protected or not readable by the server."}

            mocked_module_version_create.assert_not_called()
            mocked_process_upload.assert_not_called()
