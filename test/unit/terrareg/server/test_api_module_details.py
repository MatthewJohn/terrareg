
import unittest.mock

from test.unit.terrareg import (
    TerraregUnitTest,
    mock_models,
    setup_test_data
)
import terrareg.models
from test import client
from test.unit.terrareg.server import mocked_search_module_providers
from terrareg.result_data import ResultData


class TestApiModuleDetails(TerraregUnitTest):
    """Test ApiModuleDetails resource."""

    @setup_test_data()
    def test_existing_module(self, client, mock_models,
                             mocked_search_module_providers):
        """Test endpoint with existing module"""

        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='lonelymodule')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')

        def return_results(*args, **kwargs):
            return ResultData(
                offset=0,
                limit=10,
                count=1,
                rows=[mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 10, 'current_offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': None,
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True, 'trusted': False,
                'internal': False}
            ]
        }
        assert res.status_code == 200

        mocked_search_module_providers.assert_called_once_with(
            offset=0, limit=10, namespaces=['testnamespace'], modules=['lonelymodule'])

    @setup_test_data()
    def test_unverified_module(self, client, mock_models,
                               mocked_search_module_providers):
        """Test endpoint with existing module"""

        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='unverifiedmodule')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')

        def return_results(*args, **kwargs):
            return ResultData(
                offset=0,
                limit=10,
                count=1,
                rows=[mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/testnamespace/unverifiedmodule')

        assert res.json == {
            'meta': {'limit': 10, 'current_offset': 0}, 'modules': [
                {'id': 'testnamespace/unverifiedmodule/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'unverifiedmodule', 'version': '1.2.3',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': None,
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': False, 'trusted': False,
                'internal': False}
            ]
        }
        assert res.status_code == 200

        mocked_search_module_providers.assert_called_once_with(
            offset=0, limit=10, namespaces=['testnamespace'], modules=['unverifiedmodule'])

    def test_non_existent_module(self, client, mock_models,
                                 mocked_search_module_providers):
        """Test endpoint with non-existent module"""

        def return_results(*args, **kwargs):
            return ResultData(
                offset=0,
                limit=10,
                count=0,
                rows=[]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/doesnotexist/unittestdoesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

        mocked_search_module_providers.assert_called_once_with(
            offset=0, limit=10, namespaces=['doesnotexist'], modules=['unittestdoesnotexist'])

    @setup_test_data()
    def test_analytics_token(self, client, mock_models,
                             mocked_search_module_providers):
        """Test endpoint with analytics token and trusted namespace"""

        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='lonelymodule')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')

        def return_results(*args, **kwargs):
            return ResultData(
                offset=0,
                limit=10,
                count=1,
                rows=[mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        def return_results(*args, **kwargs):
            return ResultData(
                offset=0,
                limit=10,
                count=1,
                rows=[mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        with unittest.mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/modules/test_token-name__testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 10, 'current_offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': None,
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True, 'trusted': True,
                'internal': False}
            ]
        }
        assert res.status_code == 200
        mocked_search_module_providers.assert_called_once_with(
            offset=0, limit=10, namespaces=['testnamespace'], modules=['lonelymodule'])

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/modules/test_token-name__testnamespace/lonelymodule')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
