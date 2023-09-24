
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
from test import client
import terrareg.models


class TestApiModuleVersions(TerraregUnitTest):
    """Test ApiModuleVersions resource."""

    @setup_test_data()
    def test_existing_module_version(self, client, mock_models):
        res = client.get('/v1/modules/moduledetails/fullypopulated/testprovider/versions')

        assert res.json == {
            'modules': [
                {
                    'source': 'moduledetails/fullypopulated/testprovider',
                    'versions': [
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '1.2.0'
                        },
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '1.6.1-beta'
                        },
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '1.5.0'
                        }
                    ]
                }
            ]
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_unverified_module_version(self, client, mock_models):
        res = client.get('/v1/modules/testnamespace/unverifiedmodule/testprovider/versions')

        assert res.json == {
            'modules': [
                {
                    'source': 'testnamespace/unverifiedmodule/testprovider',
                    'versions': [
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '1.2.3'
                        }
                    ]
                }
            ]
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_internal_module_version(self, client, mock_models):
        res = client.get('/v1/modules/testnamespace/internalmodule/testprovider/versions')

        assert res.json == {
            'modules': [
                {
                    'source': 'testnamespace/internalmodule/testprovider',
                    'versions': [
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '5.2.0'
                        }
                    ]
                }
            ]
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_module_version(self, client, mock_models):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/modules/namespacename/modulename/doesnotexist/versions')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token(self, client, mock_models):
        """Test endpoint with analytics token"""

        res = client.get('/v1/modules/test_token-name__testnamespace/testmodulename/testprovider/versions')

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(module_provider=test_module_provider, version='2.4.1')

        assert res.json == {
            'modules': [
                {
                    'source': 'testnamespace/testmodulename/testprovider',
                    'versions': [
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '2.4.1'
                        },
                        {
                            'root': {'dependencies': [], 'providers': []},
                            'submodules': [],
                            'version': '1.0.0'
                        }
                    ]
                }
            ]
        }
        assert res.status_code == 200

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/modules/moduledetails/fullypopulated/testprovider/versions')

        self._test_unauthenticated_terraform_api_endpoint_test(call_endpoint)
