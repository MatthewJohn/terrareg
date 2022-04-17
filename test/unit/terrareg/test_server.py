
from unittest.mock import MagicMock

from test.unit.terrareg import client


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

