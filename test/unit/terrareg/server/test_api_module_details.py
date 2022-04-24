
from test.unit.terrareg import (
    client, mocked_server_namespace_fixture,
    setup_test_data
)


class TestApiModuleDetails:
    """Test ApiModuleDetails resource."""

    @setup_test_data()
    def test_existing_module(self, client, mocked_server_namespace_fixture):
        """Test endpoint with existing module"""

        res = client.get('/v1/modules/testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 5, 'offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }
        assert res.status_code == 200

    def test_non_existent_module(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/doesnotexist/unittestdoesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/lonelymodule')

        assert res.json == {
            'meta': {'limit': 5, 'offset': 0}, 'modules': [
                {'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'lonelymodule', 'version': '1.0.0',
                'provider': 'testprovider', 'description': 'Mock description',
                'source': 'http://mock.example.com/mockmodule',
                'published_at': '2020-01-01T23:18:12', 'downloads': 0, 'verified': True}
            ]
        }
        assert res.status_code == 200

