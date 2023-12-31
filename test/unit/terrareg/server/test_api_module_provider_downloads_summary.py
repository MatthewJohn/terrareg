
from test.unit.terrareg import (
    mock_models,
    TerraregUnitTest, setup_test_data
)
from test import client
from . import mock_server_get_module_provider_download_stats



class TestApiModuleProviderDownloadsSummary(TerraregUnitTest):
 
    @setup_test_data()
    def test_existing_module(self, client, mock_models, mock_server_get_module_provider_download_stats):
        """Test endpoint with existing module"""
        res = client.get('/v1/modules/testnamespace/testmodulename/testprovider/downloads/summary')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {'month': 58, 'total': 226, 'week': 10, 'year': 127},
                'id': 'testnamespace/testmodulename/testprovider',
                'type': 'module-downloads-summary'
            }
        }

    @setup_test_data()
    def test_non_existing_namespace(self, client, mock_models, mock_server_get_module_provider_download_stats):
        """Test endpoint with a non-existent namespace"""
        res = client.get('/v1/modules/doesnotexist/testmodule/testprovider/downloads/summary')
        assert res.status_code == 400
        assert res.json == {'message': 'Namespace does not exist'}

    @setup_test_data()
    def test_non_existing_module(self, client, mock_models, mock_server_get_module_provider_download_stats):
        """Test endpoint with a non-existent module"""
        res = client.get('/v1/modules/testnamespace/doesnotexist/testprovider/downloads/summary')
        assert res.status_code == 400
        assert res.json == {'message': 'Module provider does not exist'}


    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/modules/testnamespace/testmodule/testprovider/downloads/summary')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
