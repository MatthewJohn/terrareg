
import unittest.mock

from test.unit.terrareg import (
    client, mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)


class TestApiModuleVersionCreate(TerraregUnitTest):
    """Test module version creation resource."""

    @setup_test_data()
    def test_creation_with_no_module_provider_repository_url(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.server.check_api_key_authentication') as mocked_check_api_key_authentication:

            mocked_check_api_key_authentication.return_value = True

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenorepourl/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_creation_with_valid_repository_url(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.server.check_api_key_authentication') as mocked_check_api_key_authentication:

            mocked_check_api_key_authentication.return_value = True

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.json == {'status': 'Success'}
            assert res.status_code == 200

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_creation_with_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.server.check_api_key_authentication') as mocked_check_api_key_authentication:

            mocked_check_api_key_authentication.return_value = True

            res = client.post(
                '/v1/terrareg/modules/testnamespace/moduledoesnotexist/testprovider/5.5.4/import')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_check_api_key_authentication.assert_called_once()
            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_creation_with_invalid_authentication(self, client, mocked_server_namespace_fixture):
        """Test creating a module version with invalid API authentication."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                        unittest.mock.patch('terrareg.server.check_api_key_authentication') as mocked_check_api_key_authentication:

            mocked_check_api_key_authentication.return_value = False

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulewithrepourl/testprovider/5.5.4/import')
            assert res.status_code == 401
            assert res.json == {
                'message': "The server could not verify that you are authorized to access the URL requested. You either supplied the wrong credentials (e.g. a bad password), or your browser doesn't understand how to supply the credentials required."}

            mocked_check_api_key_authentication.assert_called_once()
            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()
