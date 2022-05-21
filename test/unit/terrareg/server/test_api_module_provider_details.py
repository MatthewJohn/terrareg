
from unittest import mock
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    client, mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)


class TestApiModuleProviderDetails(TerraregUnitTest):
    """Test ApiModuleProviderDetails resource."""

    @setup_test_data()
    def test_existing_module_provider(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/mock-module/testprovider')

        assert res.json == {
            'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'mock-module',
            'version': '1.2.3', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': 'http://github.com/testnamespace/mock-module',
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True, 'trusted': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['1.2.3']
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_trusted_module_provider(self, client, mocked_server_namespace_fixture):
        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['testnamespace']):
            res = client.get('/v1/modules/testnamespace/mock-module/testprovider')

        assert res.json == {
            'id': 'testnamespace/mock-module/testprovider/1.2.3', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'mock-module',
            'version': '1.2.3', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': 'http://github.com/testnamespace/mock-module',
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True, 'trusted': True,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['1.2.3']
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_unverified_module_provider(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/unverifiedmodule/testprovider')

        assert res.json == {
            'id': 'testnamespace/unverifiedmodule/testprovider/1.2.3', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'unverifiedmodule',
            'version': '1.2.3', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': None,
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': False, 'trusted': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['1.2.3']
        }

        assert res.status_code == 200

    def test_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/testmodulename/testprovider')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')

        assert res.json == test_module_provider.get_latest_version().get_api_details()
        assert res.status_code == 200
