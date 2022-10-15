
import unittest.mock

from test.unit.terrareg import (
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionPublish(TerraregUnitTest):
    """Test module version publish resource."""

    def _get_mock_get_current_auth_method(self, allowed_to_create):
        """Return mock auth method"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_publish_module_version = unittest.mock.MagicMock(return_value=allowed_to_create)
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)
        return mock_get_current_auth_method

    @setup_test_data()
    def test_publish_unpublished_module_version(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'test.unit.terrareg.MockModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/10.2.1/publish')
            assert res.status_code == 200
            assert res.json == {'status': 'Success'}

            mocked_update_attributes.assert_called_once_with(published=True)

    @setup_test_data()
    def test_publish_published_module_version(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'test.unit.terrareg.MockModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/publish')
            assert res.status_code == 200
            assert res.json == {'status': 'Success'}

            mocked_update_attributes.assert_called_once_with(published=True)

    @setup_test_data()
    def test_publish_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_module(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/providerdoesnotexist/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module provider does not exist'}

            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_namespace(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/namespacedoesnotexist/providerdoesnotexist/doesnotexistprovider/5.5.4/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Namespace does not exist'}

            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_non_existent_version(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(True)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/99.21.22/publish')
            assert res.status_code == 400
            assert res.json == {'message': 'Module version does not exist'}

            mocked_update_attributes.assert_not_called()

    @setup_test_data()
    def test_publish_invalid_api_key(self, client, mocked_server_namespace_fixture):
        """Test creating a module version without who's module provider does not contain a repository URL."""
        with unittest.mock.patch(
                    'test.unit.terrareg.MockModuleVersion.update_attributes') as mocked_update_attributes, \
                unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', self._get_mock_get_current_auth_method(False)):

            res = client.post(
                '/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/10.2.1/publish')
            assert res.status_code == 401
            assert res.json == {'message': 'The server could not verify that you are authorized to access the URL requested. You either supplied the wrong credentials (e.g. a bad password), or your browser doesn\'t understand how to supply the credentials required.'}

            mocked_update_attributes.assert_not_called()