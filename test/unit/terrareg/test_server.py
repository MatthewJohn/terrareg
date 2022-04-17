
from unittest.mock import MagicMock

from terrareg.models import Namespace, Module
from test.unit.terrareg import MockModuleProvider, client


def test_api_module_list(client):

    from terrareg.module_search import ModuleSearch

    # Setup for no modules returned
    ModuleSearch.search_module_providers = MagicMock(return_value=[])

    ## Call with no parameters
    res = client.get('/v1/modules')

    assert res.status_code == 200
    assert res.json == {
        'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
    }

    ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=0, limit=10)

    ## Call with limit and offset
    res = client.get('/v1/modules?offset=23&limit=12')

    assert res.status_code == 200
    assert res.json == {
        'meta': {'current_offset': 23, 'limit': 12, 'next_offset': 35, 'prev_offset': 11}, 'modules': []
    }

    ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=23, limit=12)

    ## Call with limit higher than max
    res = client.get('/v1/modules?offset=65&limit=55')

    assert res.status_code == 200
    assert res.json == {
        'meta': {'current_offset': 65, 'limit': 50, 'next_offset': 115, 'prev_offset': 15}, 'modules': []
    }

    ModuleSearch.search_module_providers.assert_called_with(provider=None, verified=False, offset=65, limit=50)

    # Test return of single module module
    namespace = Namespace(name='testnamespace')
    module = Module(namespace=namespace, name='mock-module')
    mock_module_provider = MockModuleProvider(module=module, name='testprovider')
    mock_module_provider.MOCK_LATEST_VERSION_NUMBER = '1.2.3'
    ModuleSearch.search_module_providers.return_value = [mock_module_provider]

    res = client.get('/v1/modules?offset=0&limit=1')

    assert res.status_code == 200
    print(res.json)
    assert res.json == {
        'meta': {'current_offset': 0, 'limit': 1, 'next_offset': 1, 'prev_offset': 0}, 'modules': [
            {'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
             'namespace': 'testnamespace', 'name': 'mock-module',
             'version': '1.2.3', 'provider': 'testprovider',
             'description': 'Mock description', 'source': 'http://mock.example.com/mockmodule',
             'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
        ]
    }
