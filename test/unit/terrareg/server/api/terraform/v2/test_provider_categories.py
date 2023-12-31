
from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context


class TestApiProviderCategories(TerraregIntegrationTest):
    """Test ApiProviderCategories get endpoint"""


    def test_endpoint_with_namespace(self, client):
        """Test endpoint with namespace."""

        res = client.get('/v2/categories')
        assert res.status_code == 200
        assert res.json == {
            'data': [
            {'attributes': {'name': 'Default Visible Test',
                            'slug': 'default-visible-test',
                            'user-selectable': True},
                'id': '55',
                'links': {'self': '/v2/categories/55'},
                'type': 'categories'},
            {'attributes': {'name': 'Hidden Database',
                            'slug': 'hidden-database',
                            'user-selectable': False},
                'id': '99',
                'links': {'self': '/v2/categories/99'},
                'type': 'categories'},
            {'attributes': {'name': 'No Slug Provided!',
                            'slug': 'no-slug-provided',
                            'user-selectable': True},
                'id': '100',
                'links': {'self': '/v2/categories/100'},
                'type': 'categories'},
            {'attributes': {'name': 'Second Visible Cloud',
                            'slug': 'second-visible-cloud',
                            'user-selectable': True},
                'id': '54',
                'links': {'self': '/v2/categories/54'},
                'type': 'categories'},
            {'attributes': {'name': 'Unused category',
                            'slug': 'unused-category',
                            'user-selectable': True},
                'id': '101',
                'links': {'self': '/v2/categories/101'},
                'type': 'categories'},
            {'attributes': {'name': 'Visible Monitoring',
                            'slug': 'visible-monitoring',
                            'user-selectable': True},
                'id': '523',
                'links': {'self': '/v2/categories/523'},
                'type': 'categories'}],
        }

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/categories')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)