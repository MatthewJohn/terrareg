
from test.unit.terrareg import (
    MockModuleProvider, MockModuleVersion, MockModule, MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionDetails(TerraregUnitTest):
    """Test ApiModuleVersionDetails resource."""

    @setup_test_data()
    def test_existing_module_version(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/testmodulename/testprovider/1.0.0')

        assert res.json == {
            'id': 'testnamespace/testmodulename/testprovider/1.0.0', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'testmodulename',
            'version': '1.0.0', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': 'http://mock.example.com/mockmodule',
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True, 'trusted': False, 'internal': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'],
            'versions': ['2.4.1', '1.0.0']
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_unverified_module_version(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/unverifiedmodule/testprovider/1.2.3')

        assert res.json == {
            'id': 'testnamespace/unverifiedmodule/testprovider/1.2.3', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'unverifiedmodule',
            'version': '1.2.3', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': None,
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': False, 'trusted': False, 'internal': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['1.2.3']
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_internal_module_version(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/modules/testnamespace/internalmodule/testprovider/5.2.0')

        assert res.json == {
            'id': 'testnamespace/internalmodule/testprovider/5.2.0', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'internalmodule',
            'version': '5.2.0', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': None,
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': False, 'trusted': False, 'internal': True,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['5.2.0']
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_module_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/namespacename/modulename/providername/0.1.2')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.json == test_module_version.get_api_details()
        assert res.status_code == 200
