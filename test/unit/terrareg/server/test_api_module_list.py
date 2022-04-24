
from terrareg.module_search import ModuleSearch

from . import mocked_search_module_providers
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    client, setup_test_data
)


class TestApiModuleList:

    def test_with_no_params(self, client, mocked_search_module_providers):
        """Call with no parameters"""
        res = client.get('/v1/modules')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=0, limit=10)

    def test_with_limit_offset(self, client, mocked_search_module_providers):
        """Call with limit and offset"""
        res = client.get('/v1/modules?offset=23&limit=12')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 23, 'limit': 12, 'next_offset': 35, 'prev_offset': 11}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=23, limit=12)

    def test_with_max_limit(self, client, mocked_search_module_providers):
        """Call with limit higher than max"""
        res = client.get('/v1/modules?offset=65&limit=55')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 65, 'limit': 50, 'next_offset': 115, 'prev_offset': 15}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=65, limit=50)

    def test_with_provider_filter(self, client, mocked_search_module_providers):
        """Call with provider limit"""
        res = client.get('/v1/modules?provider=testprovider')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }

        ModuleSearch.search_module_providers.assert_called_with(provider='testprovider', verified=False, offset=0, limit=10)

    def test_with_verified_false(self, client, mocked_search_module_providers):
        """Call with verified flag as false"""
        res = client.get('/v1/modules?verified=false')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=0, limit=10)


    def test_with_verified_true(self, client, mocked_search_module_providers):
        """Call with verified flag as true"""
        res = client.get('/v1/modules?verified=true')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=True, offset=0, limit=10)

    @setup_test_data()
    def test_with_module_response(self, client, mocked_search_module_providers):
        """Test return of single module module"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        ModuleSearch.search_module_providers.return_value = [mock_module_provider]

        res = client.get('/v1/modules?offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1, 'prev_offset': 0}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }

    @setup_test_data()
    def test_with_multiple_modules_response(self, client, mocked_search_module_providers):
        """Test multiple modules in results"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = MockNamespace(name='secondtestnamespace')
        mock_module_2 = MockModule(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = MockModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'
        ModuleSearch.search_module_providers.return_value = [mock_module_provider_2, mock_module_provider]

        res = client.get('/v1/modules?offset=0&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 2, 'next_offset': 2, 'prev_offset': 0}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }

