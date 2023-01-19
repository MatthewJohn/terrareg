
from unittest import mock

import pytest

from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest,
)
from test import client


class TestApiTerraregGraphData(TerraregUnitTest):
    """Test ApiTerraregGraphData resource."""

    @setup_test_data()
    def test_endpoint_default_arguments(self, client, mock_models):
        """Test endpoint, mocking get_graph_json methods"""
        with mock.patch("terrareg.models.ModuleDetails.get_graph_json", mock.MagicMock(return_value={"some_mock_data": "mock_value"})) as mock_get_graph_json:
            res = client.get("/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/graph/data")
            assert res.json == {"some_mock_data": "mock_value"}
            assert res.status_code == 200
            mock_get_graph_json.assert_called_once_with(full_module_names=False, full_resource_names=False)

    @setup_test_data()
    @pytest.mark.parametrize("full_module_names,full_resource_names", [
        (False, False),
        (True, False),
        (False, True),
        (True, True)
    ])
    def test_endpoint_arguments(self, full_module_names, full_resource_names, client, mock_models):
        """Test endpoint, mocking get_graph_json methods"""
        with mock.patch("terrareg.models.ModuleDetails.get_graph_json", mock.MagicMock(return_value={"some_mock_data": "mock_value"})) as mock_get_graph_json:
            res = client.get(
                f"/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1/graph/data?full_module_names={full_module_names}&full_resource_names={full_resource_names}"
            )
            assert res.json == {"some_mock_data": "mock_value"}
            assert res.status_code == 200
            mock_get_graph_json.assert_called_once_with(full_module_names=full_module_names, full_resource_names=full_resource_names)

    @setup_test_data()
    def test_non_existent_namespace(self, client, mock_models):
        """Test endpoint with non-existent namespace"""

        res = client.get("/v1/terrareg/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0/graph/data")

        assert res.json == {"message": "Namespace does not exist"}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module(self, client, mock_models):
        """Test endpoint with non-existent module"""

        res = client.get("/v1/terrareg/modules/emptynamespace/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0/graph/data")

        assert res.json == {"message": "Module provider does not exist"}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module_version(self, client, mock_models):
        """Test endpoint with non-existent version"""

        res = client.get("/v1/terrareg/modules/testnamespace/lonelymodule/testprovider/52.1.2/graph/data")

        assert res.json == {"errors": ["Not Found"]}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mock_models):
        """Test endpoint with analytics token and ensure it doesn"t convert the analytics token."""

        res = client.get("/v1/terrareg/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1/graph/data")

        assert res.json == {"message": "Namespace does not exist"}
        assert res.status_code == 400
