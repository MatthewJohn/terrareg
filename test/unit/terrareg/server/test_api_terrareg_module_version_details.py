
from unittest import mock
from test.unit.terrareg import (
    MockModuleProvider, MockModule, MockModuleVersion, MockNamespace,
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiTerraregModuleVersionDetails(TerraregUnitTest):
    """Test ApiTerraregModuleVersionDetails resource."""

    @setup_test_data()
    def test_existing_module_version_no_custom_urls(self, client, mocked_server_namespace_fixture):
        res = client.get('/v1/terrareg/modules/testnamespace/lonelymodule/testprovider/1.0.0')

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
            'submodules': [], 'versions': ['1.0.0'], 'providers': ['testprovider'],
            'display_source_url': None,
            'git_provider_id': None,
            'git_tag_format': '{version}',
            'module_provider_id': 'testnamespace/lonelymodule/testprovider',
            'published_at_display': 'January 01, 2020',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'terraform_example_version_string': '1.0.0',
            'beta': False,
            'published': True,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {}
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_git_provider(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - configured with a git provider
         - configured with a tag format
         - has no versions
        """
        res = client.get('/v1/terrareg/modules/moduleextraction/gitextraction/usesgitproviderwithversions/2.2.2')

        assert res.json == {
            'id': 'moduleextraction/gitextraction/usesgitproviderwithversions/2.2.2',
            'namespace': 'moduleextraction',
            'name': 'gitextraction',
            'provider': 'usesgitproviderwithversions',
            'verified': False,
            'trusted': False,
            'git_provider_id': 1,
            'git_tag_format': 'v{version}',
            'module_provider_id': 'moduleextraction/gitextraction/usesgitproviderwithversions',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'versions': ['2.2.2'],
            'description': 'Mock description',
            'display_source_url': 'https://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions/browse/v2.2.2/',
            'downloads': 0,
            'internal': False,
            'owner': 'Mock Owner',
            'published_at': '2020-01-01T23:18:12',
            'published_at_display': 'January 01, 2020',
            'root': {'dependencies': [],
                    'empty': False,
                    'inputs': [],
                    'outputs': [],
                    'path': '',
                    'provider_dependencies': [],
                    'readme': 'Mock module README file',
                    'resources': []},
            'source': 'https://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions',
            'submodules': [],
            'terraform_example_version_string': '2.2.2',
            'version': '2.2.2',
            'providers': [
                'staticrepourl',
                'placeholdercloneurl',
                'usesgitprovider',
                'usesgitproviderwithversions',
                'nogittagformat',
                'complexgittagformat',
                'norepourl'
            ],
            'beta': False,
            'published': True,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {}
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_custom_repo_urls_and_unpublished_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - configured with a custom repo URLs
         - has no published versions
        """
        res = client.get('/v1/terrareg/modules/testnamespace/modulenotpublished/testprovider/10.2.1')

        assert res.json == {
            'id': 'testnamespace/modulenotpublished/testprovider/10.2.1',
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
            'versions': [],
            'providers': ['testprovider'],
            'description': 'Mock description',
            'display_source_url': 'https://custom-localhost.com/testnamespace/modulenotpublished-testprovider/browse/10.2.1/',
            'downloads': 0,
            'internal': False,
            'owner': 'Mock Owner',
            'published_at': '2020-01-01T23:18:12',
            'published_at_display': 'January 01, 2020',
            'root': {'dependencies': [],
                    'empty': False,
                    'inputs': [],
                    'outputs': [],
                    'path': '',
                    'provider_dependencies': [],
                    'readme': 'Mock module README file',
                    'resources': []},
            'source': 'https://custom-localhost.com/testnamespace/modulenotpublished-testprovider',
            'submodules': [],
            'terraform_example_version_string': '10.2.1',
            'version': '10.2.1',
            'beta': False,
            'published': False,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {}
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_no_git_provider_or_custom_urls_and_only_beta_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with module provider that is:
         - no custom repos URLS
         - no git provider
         - only has beta version published
        """
        res = client.get('/v1/terrareg/modules/testnamespace/onlybeta/testprovider/2.2.4-beta')

        assert res.json == {
            'id': 'testnamespace/onlybeta/testprovider/2.2.4-beta',
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
            'versions': [],
            'description': 'Mock description',
            'display_source_url': None,
            'downloads': 0,
            'internal': False,
            'owner': 'Mock Owner',
            'providers': ['testprovider'],
            'published_at': '2020-01-01T23:18:12',
            'published_at_display': 'January 01, 2020',
            'root': {'dependencies': [],
                    'empty': False,
                    'inputs': [],
                    'outputs': [],
                    'path': '',
                    'provider_dependencies': [],
                    'readme': 'Mock module README file',
                    'resources': []},
            'source': None,
            'submodules': [],
            'terraform_example_version_string': '2.2.4-beta',
            'version': '2.2.4-beta',
            'beta': True,
            'published': True,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {}
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_namespace(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent namespace"""

        res = client.get('/v1/terrareg/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/emptynamespace/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0')

        assert res.json == {'message': 'Module provider does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module_version(self, client, mocked_server_namespace_fixture):
        """Test endpoint with non-existent version"""

        res = client.get('/v1/terrareg/modules/testnamespace/lonelymodule/testprovider/52.1.2')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token and ensure it doesn't convert the analytics token."""

        res = client.get('/v1/terrareg/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_matches_terrareg_api_details_function(self, client, mocked_server_namespace_fixture):
        """Test endpoint with analytics token"""

        res = client.get('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1')

        test_namespace = MockNamespace(name='testnamespace')
        test_module = MockModule(namespace=test_namespace, name='testmodulename')
        test_module_provider = MockModuleProvider(module=test_module, name='testprovider')
        test_module_version = MockModuleVersion(version='2.4.1', module_provider=test_module_provider)

        assert res.json == test_module_version.get_terrareg_api_details()
        assert res.status_code == 200

    @setup_test_data()
    def test_additional_tab_files(self, client, mocked_server_namespace_fixture):
        """Test additional tab files in API response"""

        with mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS', '[]'):
            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

            assert res.status_code == 200
            assert res.json['additional_tab_files'] == {}

        with mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS', '[["License", ["first-file", "LICENSE", "second-file"]], ["Changelog", ["CHANGELOG.md"]], ["doesnotexist", ["DOES_NOT_EXIST"]]]'):
            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

            assert res.status_code == 200
            assert res.json['additional_tab_files'] == {'Changelog': 'CHANGELOG.md', 'License': 'LICENSE'}
