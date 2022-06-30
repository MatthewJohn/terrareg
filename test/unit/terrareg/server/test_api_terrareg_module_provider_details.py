
from unittest import mock
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiTerraregModuleProviderDetails(TerraregUnitTest):
    """Test ApiTerraregModuleProviderDetails resource."""

    @setup_test_data()
    def test_existing_module_provider_no_custom_urls(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/terrareg/modules/testnamespace/lonelymodule/testprovider')

        assert res.json == {
            'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'lonelymodule',
            'version': '1.0.0', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': None,
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True, 'trusted': False, 'internal': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': []
            },
            'submodules': [], 'providers': ['testprovider'], 'versions': ['1.0.0'],
            'display_source_url': None,
            'git_provider_id': None,
            'git_tag_format': '{version}',
            'module_provider_id': 'testnamespace/lonelymodule/testprovider',
            'published_at_display': 'January 01, 2020',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'terraform_example_version_string': '1.0.0'
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_git_provider_and_no_versions(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - configured with a git provider
         - configured with a tag format
         - has no versions
        """
        res = client.get('/v1/terrareg/modules/moduleextraction/gitextraction/usesgitprovider')

        assert res.json == {
            'id': 'moduleextraction/gitextraction/usesgitprovider',
            'namespace': 'moduleextraction',
            'name': 'gitextraction',
            'provider': 'usesgitprovider',
            'verified': False,
            'trusted': False,
            'git_provider_id': 1,
            'git_tag_format': 'v{version}',
            'module_provider_id': 'moduleextraction/gitextraction/usesgitprovider',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'versions': []
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_custom_repo_urls_and_unpublished_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - configured with a custom repo URLs
         - has no published versions
        """
        res = client.get('/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider')

        assert res.json == {
            'id': 'testnamespace/modulenotpublished/testprovider',
            'namespace': 'testnamespace',
            'name': 'modulenotpublished',
            'provider': 'testprovider',
            'verified': False,
            'trusted': False,
            'git_provider_id': None,
            'git_tag_format': '{version}',
            'module_provider_id': 'testnamespace/modulenotpublished/testprovider',
            'repo_base_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}',
            'repo_browse_url_template': 'https://custom-localhost.com/{namespace}/{module}-{provider}/browse/{tag}/{path}',
            'repo_clone_url_template': 'ssh://custom-localhost.com/{namespace}/{module}-{provider}',
            'versions': []
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_no_git_provider_or_custom_urls_and_only_beta_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - no custom repos URLS
         - no git provider
         - only has beta version published
        """
        res = client.get('/v1/terrareg/modules/testnamespace/onlybeta/testprovider')

        assert res.json == {
            'id': 'testnamespace/onlybeta/testprovider',
            'namespace': 'testnamespace',
            'name': 'onlybeta',
            'provider': 'testprovider',
            'verified': False,
            'trusted': False,
            'git_provider_id': None,
            'git_tag_format': '{version}',
            'module_provider_id': 'testnamespace/onlybeta/testprovider',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'versions': []
        }

        assert res.status_code == 200

    def test_non_existent_module_provider(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist')

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
