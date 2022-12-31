
from terrareg.module_search import ModuleSearch, ModuleSearchResults

from . import mocked_search_module_providers
from test import client
from test.unit.terrareg import (
    TerraregUnitTest,
    setup_test_data,
    mock_models
)
import terrareg.models


class TestApiModuleList(TerraregUnitTest):

    def test_with_no_params(self, client, mocked_search_module_providers, mock_models):
        """Call with no parameters"""
        res = client.get('/v1/modules')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(providers=None, verified=False, offset=0, limit=10)

    def test_with_limit_offset(self, client, mocked_search_module_providers, mock_models):
        """Call with limit and offset"""
        res = client.get('/v1/modules?offset=23&limit=12')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 23, 'limit': 12, 'prev_offset': 11}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(providers=None, verified=False, offset=23, limit=12)

    def test_with_provider_filter(self, client, mocked_search_module_providers, mock_models):
        """Call with provider limit"""
        res = client.get('/v1/modules?provider=testprovider')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(providers=['testprovider'], verified=False, offset=0, limit=10)

    def test_with_verified_false(self, client, mocked_search_module_providers, mock_models):
        """Call with verified flag as false"""
        res = client.get('/v1/modules?verified=false')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(providers=None, verified=False, offset=0, limit=10)


    def test_with_verified_true(self, client, mocked_search_module_providers, mock_models):
        """Call with verified flag as true"""
        res = client.get('/v1/modules?verified=true')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(providers=None, verified=True, offset=0, limit=10)

    @setup_test_data()
    def test_with_module_response(self, client, mocked_search_module_providers, mock_models):
        """Test return of single module module"""
        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='mock-module')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')

        def side_effect(*args, **kwargs):
            return ModuleSearchResults(
                offset=0, limit=1,
                count=1, module_providers=[mock_module_provider]
            )
        ModuleSearch.search_module_providers.side_effect = side_effect

        res = client.get('/v1/modules?offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://github.com/testnamespace/mock-module',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True, 'trusted': False,
                'internal': False}
            ]
        }

    @setup_test_data()
    def test_with_module_response_with_more_results_available(self, client, mocked_search_module_providers, mock_models):
        """Test return of single module module"""
        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='mock-module')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')

        def side_effect(*args, **kwargs):
            return ModuleSearchResults(
                offset=0, limit=1,
                count=2, module_providers=[mock_module_provider]
            )
        ModuleSearch.search_module_providers.side_effect = side_effect

        res = client.get('/v1/modules?offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://github.com/testnamespace/mock-module',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True, 'trusted': False,
                'internal': False}
            ]
        }

    @setup_test_data()
    def test_with_multiple_modules_response(self, client, mocked_search_module_providers, mock_models):
        """Test multiple modules in results"""
        namespace = terrareg.models.Namespace(name='testnamespace')
        module = terrareg.models.Module(namespace=namespace, name='mock-module')
        mock_module_provider = terrareg.models.ModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = terrareg.models.Namespace(name='secondtestnamespace')
        mock_module_2 = terrareg.models.Module(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = terrareg.models.ModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'

        def side_effect(*args, **kwargs):
            return ModuleSearchResults(
                offset=0, limit=2,
                count=3, module_providers=[mock_module_provider_2, mock_module_provider]
            )
        ModuleSearch.search_module_providers.side_effect = side_effect

        res = client.get('/v1/modules?offset=0&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 2, 'next_offset': 2}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }

