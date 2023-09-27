
import json
from unittest import mock

import pytest

from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest,
)
import terrareg.models
import terrareg.version_constraint
from terrareg.constants import EXTRACTION_VERSION
from test import client


class TestApiTerraregModuleVersionDetails(TerraregUnitTest):
    """Test ApiTerraregModuleVersionDetails resource."""

    @setup_test_data()
    def test_existing_module_version_no_custom_urls(self, client, mock_models):
        res = client.get('/v1/terrareg/modules/testnamespace/lonelymodule/testprovider/1.0.0')

        assert res.json == {
            'id': 'testnamespace/lonelymodule/testprovider/1.0.0', 'owner': 'Mock Owner',
            'namespace': 'testnamespace', 'name': 'lonelymodule',
            'version': '1.0.0', 'provider': 'testprovider',
            'description': 'Mock description',
            'source': None,
            'custom_links': [],
            'published_at': '2020-01-01T23:18:12',
            'downloads': 0, 'verified': True, 'trusted': False, 'internal': False,
            'root': {
                'path': '', 'readme': 'Mock module README file',
                'empty': False, 'inputs': [], 'outputs': [], 'dependencies': [],
                'provider_dependencies': [], 'resources': [], "modules": []
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
            'terraform_version_constraint': None,
            'terraform_example_version_comment': [],
            'beta': False,
            'published': True,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {},
            'graph_url': '/modules/testnamespace/lonelymodule/testprovider/1.0.0/graph',
            'module_extraction_up_to_date': True,
            'usage_example': (
                'module "lonelymodule" {\n'
                '  source  = '
                '"localhost/my-tf-application__testnamespace/lonelymodule/testprovider"\n'
                '  version = "1.0.0"\n'
                '\n'
                '  # Provide variables here\n'
                '}'
            )
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_fully_populated_module(self, client, mock_models):
        res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

        assert res.json == {
            "id": "moduledetails/fullypopulated/testprovider/1.5.0",
            "namespace": "moduledetails",
            "name": "fullypopulated",
            "provider": "testprovider",
            "verified": False,
            "trusted": False,
            "versions": ["1.2.0", "1.5.0"],
            "module_provider_id": "moduledetails/fullypopulated/testprovider",
            "git_provider_id": None,
            "git_tag_format": "{version}",
            "git_path": None,
            "repo_base_url_template": "https://mp-base-url.com/{namespace}/{module}-{provider}",
            "repo_clone_url_template": "ssh://mp-clone-url.com/{namespace}/{module}-{provider}",
            "repo_browse_url_template": "https://mp-browse-url.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix",
            "owner": "This is the owner of the module",
            "version": "1.5.0",
            "description": "This is a test module version for tests.",
            "source": "https://link-to.com/source-code-here",
            "published_at": "2022-01-05T22:53:12",
            "downloads": 0,
            "internal": False,
            "root": {
                "path": "",
                "readme": "# This is an exaple README!",
                "empty": False,
                "inputs": [
                    {
                        "name": "name_of_application",
                        "type": "string",
                        "description": "Enter the application name",
                        "default": None,
                        "required": True,
                    },
                    {
                        "name": "string_with_default_value",
                        "type": "string",
                        "description": "Override the default string",
                        "default": "this is the default",
                        "required": False,
                    },
                    {
                        "name": "example_boolean_input",
                        "type": "bool",
                        "description": "Override the truthful boolean",
                        "default": True,
                        "required": False,
                    },
                    {
                        "name": "example_list_input",
                        "type": "list",
                        "description": "Override the stringy list",
                        "default": ["value 1", "value 2"],
                        "required": False,
                    },
                ],
                "outputs": [
                    {"name": "generated_name", "description": "Name with randomness"},
                    {"name": "no_desc_output", "description": None},
                ],
                "dependencies": [
                    {
                        "name": "hashicorp-registry-module",
                        "source": "matthewjohn/test-module/null",
                        "version": "1.5.0",
                    },
                    {
                        "name": "local-registry-module",
                        "source": "my-registry.example.com/matthewjohn/test-module/null",
                        "version": "2.1.3",
                    },
                ],
                "provider_dependencies": [
                    {
                        "name": "random",
                        "namespace": "hashicorp",
                        "source": "",
                        "version": ">= 5.2.1, < 6.0.0",
                    },
                    {
                        "name": "unsafe",
                        "namespace": "someothercompany",
                        "source": "",
                        "version": "2.0.0",
                    },
                ],
                "resources": [
                    {
                        "type": "string",
                        "name": "random_suffix",
                        "provider": "random",
                        "source": "hashicorp/random",
                        "mode": "managed",
                        "version": "latest",
                        "description": None,
                    }
                ],
                "modules": [
                    {
                        "name": "local-path-module",
                        "source": "./modules/local-module",
                        "version": None,
                        "description": "A local module",
                    },
                    {
                        "name": "hashicorp-registry-module",
                        "source": "matthewjohn/test-module/null",
                        "version": "1.5.0",
                        "description": "A module from hashicorp",
                    },
                    {
                        "name": "local-registry-module",
                        "source": "my-registry.example.com/matthewjohn/test-module/null",
                        "version": "2.1.3",
                        "description": "A module from local registry",
                    },
                    {
                        "name": "hashicorp-registry-module.hashicorp-sub-module",
                        "source": "matthewjohn/test-module/null//modules/hashi-submodule",
                        "version": "1.5.0",
                        "description": "A module from hashicorp",
                    },
                    {
                        "name": "local-path-module.local-child-module",
                        "source": "./modules/second-local",
                        "version": None,
                        "description": "A local module",
                    },
                    {
                        "name": "local-registry-module.alternative-sub-module",
                        "source": "my-registry.example.com/matthewjohn/test-module/null//modules/alternative-submodule",
                        "version": "2.1.3",
                        "description": "A module from local registry",
                    },
                ],
            },
            "submodules": [],
            "providers": ["testprovider"],
            "published_at_display": "January 05, 2022",
            "display_source_url": "https://mp-browse-url.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/suffix",
            "terraform_example_version_string": "1.5.0",
            "terraform_example_version_comment": [],
            "beta": False,
            "published": True,
            "security_failures": 0,
            "security_results": None,
            "additional_tab_files": {"Release Notes": "CHANGELOG.md", "License": "LICENSE"},
            "custom_links": [],
            "graph_url": "/modules/moduledetails/fullypopulated/testprovider/1.5.0/graph",
            "terraform_version_constraint": ">= 1.0, < 2.0.0",
            "module_extraction_up_to_date": True,
            "usage_example": 'module "fullypopulated" {\n  source  = "localhost/my-tf-application__moduledetails/fullypopulated/testprovider"\n  version = "1.5.0"\n\n  # Provide variables here\n}',
        }


        assert res.status_code == 200

    @setup_test_data()
    def test_terraform_version_constraint(self, client, mock_models):
        """Test terraform_version_constraint attribute"""
        res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

        assert res.json["terraform_version_constraint"] == ">= 1.0, < 2.0.0"
        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_git_provider(self, client, mock_models):
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
            'custom_links': [],
            'git_provider_id': 1,
            'git_tag_format': 'v{version}',
            'module_provider_id': 'moduleextraction/gitextraction/usesgitproviderwithversions',
            'repo_base_url_template': None,
            'repo_browse_url_template': None,
            'repo_clone_url_template': None,
            'versions': ['2.2.2', '2.1.0'],
            'description': 'Mock description',
            'display_source_url': 'https://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions/browse/v2.2.2/',
            'downloads': 0,
            'internal': False,
            'owner': 'Mock Owner',
            'published_at': '2020-01-01T23:18:12',
            'published_at_display': 'January 01, 2020',
            'root': {
                'dependencies': [],
                'empty': False,
                'inputs': [],
                'outputs': [],
                'path': '',
                'provider_dependencies': [],
                'readme': 'Mock module README file',
                'resources': [],
                'modules': []
            },
            'source': 'https://localhost.com/moduleextraction/gitextraction-usesgitproviderwithversions',
            'submodules': [],
            'terraform_example_version_string': '2.2.2',
            'terraform_version_constraint': None,
            'terraform_example_version_comment': [],
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
            'additional_tab_files': {},
            'graph_url': '/modules/moduleextraction/gitextraction/usesgitproviderwithversions/2.2.2/graph',
            'module_extraction_up_to_date': True,
            'usage_example': (
                'module "gitextraction" {\n'
                '  source  = '
                '"localhost/my-tf-application__moduleextraction/gitextraction/usesgitproviderwithversions"\n'
                '  version = "2.2.2"\n'
                '\n'
                '  # Provide variables here\n'
                '}'
            )
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_custom_repo_urls_and_unpublished_version(self, client, mock_models):
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
            'custom_links': [],
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
            'root': {
                'dependencies': [],
                'empty': False,
                'inputs': [],
                'outputs': [],
                'path': '',
                'provider_dependencies': [],
                'readme': 'Mock module README file',
                'resources': [],
                'modules': []
            },
            'source': 'https://custom-localhost.com/testnamespace/modulenotpublished-testprovider',
            'submodules': [],
            'terraform_example_version_string': '10.2.1',
            'terraform_version_constraint': None,
            'terraform_example_version_comment': [
                'This version of this module has not yet been published,',
                'meaning that it cannot yet be used by Terraform'
            ],
            'version': '10.2.1',
            'beta': False,
            'published': False,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {},
            'graph_url': '/modules/testnamespace/modulenotpublished/testprovider/10.2.1/graph',
            'module_extraction_up_to_date': True,
            'usage_example': (
                'module "modulenotpublished" {\n'
                '  source  = '
                '"localhost/my-tf-application__testnamespace/modulenotpublished/testprovider"\n'
                '  # This version of this module has not yet been '
                'published,\n'
                '  # meaning that it cannot yet be used by Terraform\n'
                '  version = "10.2.1"\n'
                '\n'
                '  # Provide variables here\n'
                '}'
            )
        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_no_git_provider_or_custom_urls_and_only_beta_version(self, client, mock_models):
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
            'custom_links': [],
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
            'root': {
                'dependencies': [],
                'empty': False,
                'inputs': [],
                'outputs': [],
                'path': '',
                'provider_dependencies': [],
                'readme': 'Mock module README file',
                'resources': [],
                'modules': []
            },
            'source': None,
            'submodules': [],
            'terraform_example_version_string': '2.2.4-beta',
            'terraform_version_constraint': None,
            'terraform_example_version_comment': [
                'This version of the module is a beta version.',
                'To use this version, it must be pinned in Terraform'
            ],
            'version': '2.2.4-beta',
            'beta': True,
            'published': True,
            'security_failures': 0,
            'security_results': None,
            'git_path': None,
            'additional_tab_files': {},
            'graph_url': '/modules/testnamespace/onlybeta/testprovider/2.2.4-beta/graph',
            'module_extraction_up_to_date': True,
            'usage_example': (
                'module "onlybeta" {\n'
                '  source  = '
                '"localhost/my-tf-application__testnamespace/onlybeta/testprovider"\n'
                '  # This version of the module is a beta version.\n'
                '  # To use this version, it must be pinned in Terraform\n'
                '  version = "2.2.4-beta"\n'
                '\n'
                '  # Provide variables here\n'
                '}'
            ),

        }

        assert res.status_code == 200

    @setup_test_data()
    def test_existing_module_version_with_custom_links(self, client, mock_models):
        """Test endpoint with custom links configured"""

        with mock.patch('terrareg.config.Config.MODULE_LINKS', json.dumps([
                    {"text": "Placeholders in text module:{module} provider:{provider} ns:{namespace}",
                     "url": "https://example.com/placeholders-in-link/{namespace}/{module}-{provider}/end"},
                    {"text": "Link that does not apply",
                     "url": "https://mydomain.example.com/",
                     "namespaces": ["not-the-namespace", "another-namespace"]},
                    {"text": "Link that applies to this namespace",
                     "url": "https://applies-to-this-module.com",
                     "namespaces": ["another-namespace", "testnamespace", "another-another-one"]}
                ])):

            res = client.get('/v1/terrareg/modules/testnamespace/onlybeta/testprovider/2.2.4-beta')

            assert res.json.get("custom_links") == [
                {"text": "Placeholders in text module:onlybeta provider:testprovider ns:testnamespace",
                 "url": "https://example.com/placeholders-in-link/testnamespace/onlybeta-testprovider/end"},
                {"text": "Link that applies to this namespace",
                 "url": "https://applies-to-this-module.com"},
            ]

            assert res.status_code == 200

    @setup_test_data()
    def test_non_existent_namespace(self, client, mock_models):
        """Test endpoint with non-existent namespace"""

        res = client.get('/v1/terrareg/modules/doesnotexist/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module(self, client, mock_models):
        """Test endpoint with non-existent module"""

        res = client.get('/v1/terrareg/modules/emptynamespace/unittestdoesnotexist/unittestproviderdoesnotexist/1.0.0')

        assert res.json == {'message': 'Module provider does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_non_existent_module_version(self, client, mock_models):
        """Test endpoint with non-existent version"""

        res = client.get('/v1/terrareg/modules/testnamespace/lonelymodule/testprovider/52.1.2')

        assert res.json == {'errors': ['Not Found']}
        assert res.status_code == 404

    @setup_test_data()
    def test_analytics_token_not_converted(self, client, mock_models):
        """Test endpoint with analytics token and ensure it doesn't convert the analytics token."""

        res = client.get('/v1/terrareg/modules/test_token-name__testnamespace/testmodulename/testprovider/2.4.1')

        assert res.json == {'message': 'Namespace does not exist'}
        assert res.status_code == 400

    @setup_test_data()
    def test_matches_terrareg_api_details_function(self, client, mock_models):
        """Test endpoint with analytics token"""

        res = client.get('/v1/terrareg/modules/testnamespace/testmodulename/testprovider/2.4.1')

        test_namespace = terrareg.models.Namespace(name='testnamespace')
        test_module = terrareg.models.Module(namespace=test_namespace, name='testmodulename')
        test_module_provider = terrareg.models.ModuleProvider(module=test_module, name='testprovider')
        test_module_version = terrareg.models.ModuleVersion(version='2.4.1', module_provider=test_module_provider)

        assert res.json == test_module_version.get_terrareg_api_details(request_domain="localhost")
        assert res.status_code == 200

    @setup_test_data()
    def test_additional_tab_files(self, client, mock_models):
        """Test additional tab files in API response"""

        with mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS', '[]'):
            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

            assert res.status_code == 200
            assert res.json['additional_tab_files'] == {}

        with mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS', '[["License", ["first-file", "LICENSE", "second-file"]], ["Changelog", ["CHANGELOG.md"]], ["doesnotexist", ["DOES_NOT_EXIST"]]]'):
            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

            assert res.status_code == 200
            assert res.json['additional_tab_files'] == {'Changelog': 'CHANGELOG.md', 'License': 'LICENSE'}

    @setup_test_data()
    @pytest.mark.parametrize('extraction_version,expected_update_to_date_flag', [
        (0, False),
        (EXTRACTION_VERSION, True)
    ])
    def test_update_to_date_flag(self, extraction_version, expected_update_to_date_flag, client, mock_models):
        """Test additional tab files in API response"""

        terrareg.models.ModuleVersion(
            terrareg.models.ModuleProvider(
                terrareg.models.Module(
                    terrareg.models.Namespace('moduledetails'),
                    "fullypopulated"
                ),
                "testprovider"
            ),
            "1.5.0"
        ).update_attributes(extraction_version=extraction_version)

        res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

        assert res.status_code == 200
        assert res.json['module_extraction_up_to_date'] == expected_update_to_date_flag

    @setup_test_data()
    def test_terraform_example_version_comment(self, client, mock_models):
        """Test example version comment is passed to API correctly"""
        with mock.patch("terrareg.models.ModuleVersion.get_terraform_example_version_comment",
                        mock.MagicMock(return_value=["a unit test", "comment value"])):
            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

            assert res.json["terraform_example_version_comment"] == ["a unit test", "comment value"]

    @setup_test_data()
    def test_terraform_version_compatibility(self, client, mock_models):
        """Test example version comment is passed to API correctly"""
        with mock.patch("terrareg.version_constraint.VersionConstraint.is_compatible",
                        mock.MagicMock(return_value=terrareg.version_constraint.VersionCompatibilityType.COMPATIBLE)) as mock_is_compatible, \
                mock.patch("terrareg.models.ModuleVersion.get_terraform_version_constraints", mock.MagicMock(return_value="> 5.1.7, <= 9.2.2")):

            res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0?target_terraform_version=4.2.1')

            assert res.json["version_compatibility"] == "compatible"

            mock_is_compatible.assert_called_once_with(constraint='> 5.1.7, <= 9.2.2', target_version='4.2.1')

    @setup_test_data()
    def test_submodules(self, client, mock_models):
        """Test 'modules' attribute of module details"""
        res = client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')
        assert res.json["root"]["modules"] == [
            {
                'name': 'local-path-module',
                'source': './modules/local-module',
                'version': None,
                'description': 'A local module'
            },
            {
                'name': 'hashicorp-registry-module',
                'source': 'matthewjohn/test-module/null',
                'version': '1.5.0',
                'description': 'A module from hashicorp'
            },
            {
                'name': 'local-registry-module',
                'source': 'my-registry.example.com/matthewjohn/test-module/null',
                'version': '2.1.3',
                'description': 'A module from local registry'
            },
            {
                'name': 'hashicorp-registry-module.hashicorp-sub-module',
                'source': 'matthewjohn/test-module/null//modules/hashi-submodule',
                'version': '1.5.0',
                'description': 'A module from hashicorp'
            },
            {
                'name': 'local-path-module.local-child-module',
                'source': './modules/second-local',
                'version': None,
                'description': 'A local module'
            },
            {
                'name': 'local-registry-module.alternative-sub-module',
                'source': 'my-registry.example.com/matthewjohn/test-module/null//modules/alternative-submodule',
                'version': '2.1.3',
                'description': 'A module from local registry'
            }
        ]

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/1.5.0')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
