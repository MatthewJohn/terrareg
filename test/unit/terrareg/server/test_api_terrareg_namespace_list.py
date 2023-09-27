
from unittest import mock
from test.unit.terrareg import TerraregUnitTest, setup_test_data, mock_models
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
    def test_with_namespaces_present(self, client, mock_models):
        """Test endpoint with existing namespaces."""
        res = client.get('/v1/terrareg/namespaces')
        assert res.status_code == 200
        assert res.json == [
            {'name': 'testnamespace', 'view_href': '/modules/testnamespace', 'display_name': None},
            {'name': 'moduledetails', 'view_href': '/modules/moduledetails', 'display_name': None},
            {'name': 'secondtestnamespace', 'view_href': '/modules/secondtestnamespace', 'display_name': None},
            {'name': 'smallernamespacelist', 'view_href': '/modules/smallernamespacelist', 'display_name': 'Smaller Namespace List'},
            {'name': 'moduleextraction', 'view_href': '/modules/moduleextraction', 'display_name': None},
            {'name': 'emptynamespace', 'view_href': '/modules/emptynamespace', 'display_name': None}
        ]

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/terrareg/namespaces')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
