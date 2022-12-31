
from unittest import mock

import pytest
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
from test import client
import terrareg.models


class TestApiTerraregModuleProviderDetails(TerraregUnitTest):
    """Test ApiTerraregModuleProviderDetails resource."""

    @setup_test_data()
    def test_existing_module_provider_no_custom_urls(self, client, mock_models):
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
    @pytest.mark.parametrize('security_issues_enabled,expected_security_issues,expected_security_results', [
        # When security issues are enabled, 2 should be returned
        (True, 2, [{
            'description': 'Secret explicitly uses the default key.',
            'impact': 'Using AWS managed keys reduces the '
                        'flexibility and control over the encryption '
                        'key',
            'links': ['https://aquasecurity.github.io/tfsec/v1.26.0/checks/aws/ssm/secret-use-customer-key/',
                        'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret#kms_key_id'],
            'location': {'end_line': 4,
                        'filename': 'main.tf',
                        'start_line': 2},
            'long_id': 'aws-ssm-secret-use-customer-key',
            'resolution': 'Use customer managed keys',
            'resource': 'aws_secretsmanager_secret.this',
            'rule_description': 'Secrets Manager should use '
                                'customer managed keys',
            'rule_id': 'AVD-AWS-0098',
            'rule_provider': 'aws',
            'rule_service': 'ssm',
            'severity': 'LOW',
            'status': 0,
            'warning': False},
            {'description': 'Some security issue 2.',
            'impact': 'Entire project is compromised',
            'links': ['https://example.com/issuehere',
                        'https://example.com/docshere'],
            'location': {'end_line': 1,
                        'filename': 'main.tf',
                        'start_line': 6},
            'long_id': 'dodgy-bad-is-bad',
            'resolution': 'Do not use bad code',
            'resource': 'some_data_resource.this',
            'rule_description': 'Dodgy code should be removed',
            'rule_id': 'DDG-ANC-001',
            'rule_provider': 'bad',
            'rule_service': 'code',
            'severity': 'HIGH',
            'status': 0,
            'warning': False}
        ]),
        # When security issues are disabled, 0 should be returned
        (False, 0, None)
    ])
    def test_existing_module_provider_with_security_issues(
            self, security_issues_enabled, expected_security_issues,
            expected_security_results, client, mock_models):
        """Test obtaining details about module provider with security issues"""
        with mock.patch('terrareg.config.Config.ENABLE_SECURITY_SCANNING', security_issues_enabled):
            res = client.get('/v1/terrareg/modules/testnamespace/withsecurityissues/testprovider')

            assert res.json == {
                'id': 'testnamespace/withsecurityissues/testprovider/1.0.0', 'owner': 'Mock Owner',
                'namespace': 'testnamespace', 'name': 'withsecurityissues',
                'version': '1.0.0', 'provider': 'testprovider',
                'description': 'Mock description',
                'source': None,
                'published_at': '2020-01-01T23:18:12',
                'downloads': 0, 'verified': False, 'trusted': False, 'internal': False,
                'root': {
                    'path': '', 'readme': 'Mock module README file',
                    'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                    'provider_dependencies': [], 'resources': []
                },
                'submodules': [], 'providers': ['testprovider'], 'versions': ['1.0.0'],
                'display_source_url': None,
                'git_provider_id': None,
                'git_tag_format': '{version}',
                'module_provider_id': 'testnamespace/withsecurityissues/testprovider',
                'published_at_display': 'January 01, 2020',
                'repo_base_url_template': None,
                'repo_browse_url_template': None,
                'repo_clone_url_template': None,
                'terraform_example_version_string': '1.0.0',
                'beta': False,
                'published': True,
                'security_failures': expected_security_issues,
                'security_results': expected_security_results,
                'git_path': None,
                'additional_tab_files': {}
            }

            assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_git_provider_and_no_versions(self, client, mock_models):
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
            'versions': [],
            'git_path': None
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_custom_repo_urls_and_unpublished_version(self, client, mock_models):
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
            'versions': [],
            'git_path': None
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_provider_with_no_git_provider_or_custom_urls_and_only_beta_version(self, client, mock_models):
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
            'versions': [],
            'git_path': None
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_module_provider(self, client, mock_models):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/emptynamespace/unittestdoesnotexist/unittestproviderdoesnotexist')

        assert res.json == {'message': 'Module provider does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_namespace(self, client, mock_models):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mock_models):
        """Test endpoint with analytics token and ensure it doesn't convert the analytics token."""

        res = client.get('/v1/terrareg/modules/test_token-name__testnamespace/testmodulename/testprovider')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_matches_terrareg_api_details_function(self, client, mock_models):
        """Test endpoint with analytics token"""

        res = client.get('/v1/terrareg/modules/testnamespace/testmodulename/testprovider')

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')

        assert res.json == test_module_provider.get_latest_version().get_terrareg_api_details()
        assert res.status_code == 200
