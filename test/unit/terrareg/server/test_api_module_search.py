
from terrareg.module_search import ModuleSearch, ModuleSearchResults
from terrareg.filters import NamespaceTrustFilter
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    client, setup_test_data, TerraregUnitTest
)
from . import mocked_search_module_providers


class TestApiModuleSearch(TerraregUnitTest):

    def test_with_no_params(self, client, mocked_search_module_providers):
        """Test ApiModuleSearch with no params"""
        res = client.get('/v1/modules/search')
        assert res.status_code == 400
        ModuleSearch.search_module_providers.assert_not_called()

    def test_with_query_string(self, client, mocked_search_module_providers):
        """Call with query param"""
        res = client.get('/v1/modules/search?q=unittestteststring')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='unittestteststring', namespaces=None, providers=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_limit_offset(self, client, mocked_search_module_providers):
        """Call with limit and offset"""
        res = client.get('/v1/modules/search?q=test&offset=23&limit=12')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 23, 'limit': 12, 'prev_offset': 11}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=None, providers=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=23, limit=12)

    def test_with_provider(self, client, mocked_search_module_providers):
        """Call with provider filter"""
        res = client.get('/v1/modules/search?q=test&provider=testprovider')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=None, providers=['testprovider'], verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_multiple_providers(self, client, mocked_search_module_providers):
        """Call with multiple provider filters."""
        res = client.get('/v1/modules/search?q=test&provider=testprovider1&provider=unittestprovider2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=None, providers=['testprovider1', 'unittestprovider2'], verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_namespace(self, client, mocked_search_module_providers):
        """Call with namespace filter"""
        res = client.get('/v1/modules/search?q=test&namespace=testnamespace')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=['testnamespace'], providers=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_multiple_namespaces(self, client, mocked_search_module_providers):
        """Call with namespace filter"""
        res = client.get('/v1/modules/search?q=test&namespace=testnamespace&namespace=unittestnamespace2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=['testnamespace', 'unittestnamespace2'], providers=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_namespace_trust_filters(self, client, mocked_search_module_providers):
        """Call with trusted namespace/contributed filters"""
        for namespace_filter in [['&trusted_namespaces=false', []],
                                ['&trusted_namespaces=true', [NamespaceTrustFilter.TRUSTED_NAMESPACES]],
                                ['&contributed=false', []],
                                ['&contributed=true', [NamespaceTrustFilter.CONTRIBUTED]],
                                ['&trusted_namespaces=false&contributed=false', []],
                                ['&trusted_namespaces=true&contributed=false', [NamespaceTrustFilter.TRUSTED_NAMESPACES]],
                                ['&trusted_namespaces=false&contributed=true', [NamespaceTrustFilter.CONTRIBUTED]],
                                ['&trusted_namespaces=true&contributed=true', [NamespaceTrustFilter.TRUSTED_NAMESPACES, NamespaceTrustFilter.CONTRIBUTED]]]:

            res = client.get('/v1/modules/search?q=test{0}'.format(namespace_filter[0]))

            assert res.status_code == 200
            assert res.json == {
                'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
            }
            ModuleSearch.search_module_providers.assert_called_with(
                query='test', namespaces=None, providers=None, verified=False,
                namespace_trust_filters=namespace_filter[1],
                offset=0, limit=10)

    def test_with_verified_false(self, client, mocked_search_module_providers):
        """Call with verified flag as false"""
        res = client.get('/v1/modules/search?q=test&verified=false')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=None, providers=None, verified=False,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    def test_with_verified_true(self, client, mocked_search_module_providers):
        """Test call with verified as true"""
        res = client.get('/v1/modules/search?q=test&verified=true')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 10}, 'modules': []
        }
        ModuleSearch.search_module_providers.assert_called_with(
            query='test', namespaces=None, providers=None, verified=True,
            namespace_trust_filters=NamespaceTrustFilter.UNSPECIFIED,
            offset=0, limit=10)

    @setup_test_data()
    def test_with_single_module_response(self, client, mocked_search_module_providers):
        """Test return of single module module"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'

        def return_results(*args, **kwargs):
            return ModuleSearchResults(
                offset=0,
                limit=1,
                count=1,
                module_providers=[mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/search?q=test&offset=0&limit=1')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 1}, 'modules': [
                {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'mock-module',
                'version': '1.2.3', 'provider': 'testprovider',
                'description': 'Mock description', 'source': 'http://github.com/testnamespace/mock-module',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }

    @setup_test_data()
    def test_with_multiple_module_response(self, client, mocked_search_module_providers):
        """Test multiple modules in results"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = MockNamespace(name='secondtestnamespace')
        mock_module_2 = MockModule(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = MockModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'

        def return_results(*args, **kwargs):
            return ModuleSearchResults(
                offset=0,
                limit=2,
                count=2,
                module_providers=[mock_module_provider_2, mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/search?q=test&offset=0&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 0, 'limit': 2}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }

    @setup_test_data()
    def test_with_next_offset(self, client, mocked_search_module_providers):
        """Test multiple modules in results"""
        namespace = MockNamespace(name='testnamespace')
        module = MockModule(namespace=namespace, name='mock-module')
        mock_module_provider = MockModuleProvider(module=module, name='testprovider')
        mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
        mock_namespace_2 = MockNamespace(name='secondtestnamespace')
        mock_module_2 = MockModule(namespace=mock_namespace_2, name='mockmodule2')
        mock_module_provider_2 = MockModuleProvider(module=mock_module_2, name='secondprovider')
        mock_module_provider_2.MOCK_LATEST_VERSION_NUMBER = '3.0.0'

        def return_results(*args, **kwargs):
            return ModuleSearchResults(
                offset=4,
                limit=2,
                count=7,
                module_providers=[mock_module_provider_2, mock_module_provider]
            )
        mocked_search_module_providers.side_effect = return_results

        res = client.get('/v1/modules/search?q=test&offset=4&limit=2')

        assert res.status_code == 200
        assert res.json == {
            'meta': {'current_offset': 4, 'limit': 2, 'next_offset': 6, 'prev_offset': 2}, 'modules': [
                mock_module_provider_2.get_latest_version().get_api_outline(),
                mock_module_provider.get_latest_version().get_api_outline(),
            ]
        }
