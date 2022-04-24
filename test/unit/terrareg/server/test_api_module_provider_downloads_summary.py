
from test.unit.terrareg import (
    client, mocked_server_namespace_fixture
)
from . import mock_server_get_module_provider_download_stats



class TestApiModuleProviderDownloadsSummary:
 
    def test_existing_module(self, client, mocked_server_namespace_fixture, mock_server_get_module_provider_download_stats):
        """Test endpoint with existing module"""
        res = client.get('/v1/modules/testnamespace/testmodule/testprovider/downloads/summary')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {'month': 58, 'total': 226, 'week': 10, 'year': 127},
                'id': 'testnamespace/testmodule/testprovider',
                'type': 'module-downloads-summary'
            }
        }

