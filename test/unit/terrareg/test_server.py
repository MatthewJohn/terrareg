
from unittest.mock import MagicMock

from test.unit.terrareg import client


def test_api_module_list(client):

    from terrareg.module_search import ModuleSearch
    ModuleSearch.search_module_providers = MagicMock(return_value=[])

    res = client.get('/v1/modules')

    assert res.status_code == 200
    assert res.json == {
        'meta': {'current_offset': 0, 'limit': 10, 'next_offset': 10, 'prev_offset': 0}, 'modules': []
    }


