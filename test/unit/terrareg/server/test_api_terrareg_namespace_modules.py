
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
                    'beta': False,
                    'description': 'Test description',
                    'display_source_url': None,
                    'downloads': 0,
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/publishedone/testprovider/2.1.1',
                    'internal': False,
                    'module_provider_id': 'smallernamespacelist/publishedone/testprovider',
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'owner': 'Mock Owner',
                    'provider': 'testprovider',
                    'providers': ['testprovider',
                                    'secondnamespace'],
                    'published': True,
                    'published_at': '2020-01-01T23:18:12',
                    'published_at_display': 'January 01, 2020',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'root': {'dependencies': [],
                            'empty': False,
                            'inputs': [],
                            'outputs': [],
                            'path': '',
                            'provider_dependencies': [],
                            'readme': 'Mock module README file',
                            'resources': []},
                    'security_failures': 0,
                    'source': None,
                    'submodules': [],
                    'terraform_example_version_string': '2.1.1',
                    'trusted': False,
                    'verified': False,
                    'version': '2.1.1',
                    'versions': ['2.1.1']
                },
                {
                    # Second published module provider in same module
                    'beta': False,
                    'description': 'Description of second provider in module',
                    'display_source_url': None,
                    'downloads': 0,
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/publishedone/secondnamespace/2.2.2',
                    'internal': False,
                    'module_provider_id': 'smallernamespacelist/publishedone/secondnamespace',
                    'name': 'publishedone',
                    'namespace': 'smallernamespacelist',
                    'owner': 'Mock Owner',
                    'provider': 'secondnamespace',
                    'providers': ['testprovider',
                                    'secondnamespace'],
                    'published': True,
                    'published_at': '2020-01-01T23:18:12',
                    'published_at_display': 'January 01, 2020',
                    'repo_base_url_template': None,
                    'repo_browse_url_template': None,
                    'repo_clone_url_template': None,
                    'root': {'dependencies': [],
                            'empty': False,
                            'inputs': [],
                            'outputs': [],
                            'path': '',
                            'provider_dependencies': [],
                            'readme': 'Mock module README file',
                            'resources': []},
                    'security_failures': 0,
                    'source': None,
                    'submodules': [],
                    'terraform_example_version_string': '2.2.2',
                    'trusted': False,
                    'verified': False,
                    'version': '2.2.2',
                    'versions': ['2.2.2']
                },
                {
                    # Ensure module provider with no versions is returned correctly
                    'git_provider_id': None,
                    'git_tag_format': '{version}',
                    'id': 'smallernamespacelist/noversions/testprovider',
                    'module_provider_id': 'smallernamespacelist/noversions/testprovider',
                    'name': 'noversions',
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
