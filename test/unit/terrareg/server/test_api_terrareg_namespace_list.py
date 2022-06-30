
from unittest import mock
from test.unit.terrareg import MockNamespace, TerraregUnitTest, setup_test_data, mocked_server_namespace_fixture
from test import client


class TestApiTerraregNamespaceList(TerraregUnitTest):
    """Test ApiTerraregNamespaceList resource."""

    def test_with_no_namespaces(self, client):
        """Test endpoint when no namespaces are present."""
        with mock.patch('terrareg.models.Namespace.get_all') as mocked_namespace_get_all:
            mocked_namespace_get_all.return_value = []

            res = client.get('/v1/terrareg/namespaces')

            assert res.status_code == 200
            assert res.json == []

            mocked_namespace_get_all.assert_called_once()

    @setup_test_data()
    def test_with_namespaces_present(self, client, mocked_server_namespace_fixture):
        """Test endpoint with existing namespaces."""
        res = client.get('/v1/terrareg/namespaces')
        assert res.status_code == 200
        assert res.json == ['testnamespace']
