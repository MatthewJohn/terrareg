
from unittest import mock

import pytest
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiTerraregNamespaceModules(TerraregUnitTest):
    """Test ApiTerraregNamespaceModules resource."""

    @setup_test_data()
    def test_existing_namespace_with_mixed_modules(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/terrareg/modules/smallernamespacelist')

        assert res.json == {
            'meta': {
                'current_offset': 0,
                'limit': 10
            },
            'modules': [
                {
                    # Ensure normal published module is shown
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/publishedone/testprovider',
                    'module_provider_id': 'smallernamespacelist/publishedone/testprovider',
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'trusted': False,
                    'verified': False,
                    'versions': ['2.1.1']
                },
                {
                    # Second published module provider in same module
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/publishedone/secondnamespace',
                    'module_provider_id': 'smallernamespacelist/publishedone/secondnamespace',
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'provider': 'secondnamespace',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'trusted': False,
                    'verified': False,
                    'versions': ['2.2.2']
                },
                {
                    # Ensure published module provider that contains only a beta version is shown
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/onlybeta/testprovider',
                    'module_provider_id': 'smallernamespacelist/onlybeta/testprovider',
                    'name': 'onlybeta',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'trusted': False,
                    'verified': False,
                    'versions': []
                },
                {
                    # Ensure published module provider that contains only an unpublished version is shown
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/onlyunpublished/testprovider',
                    'module_provider_id': 'smallernamespacelist/onlyunpublished/testprovider',
                    'name': 'onlyunpublished',
                    'namespace': 'smallernamespacelist',
                    'provider': 'testprovider',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'trusted': False,
                    'verified': False,
                    'versions': []
                }
            ]
        }

        assert res.status_code == 200

    def test_non_existent_namespace(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/doesnotexist')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token and ensure it doesn't convert the analytics token."""

        res = client.get('/v1/terrareg/modules/test_token-name__testnamespace/testmodulename/testprovider')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_matches_terrareg_api_details_function(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/terrareg/modules/testnamespace/testmodulename/testprovider')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')

        assert res.json == test_module_provider.get_latest_version().get_terrareg_api_details()
        assert res.status_code == 200
