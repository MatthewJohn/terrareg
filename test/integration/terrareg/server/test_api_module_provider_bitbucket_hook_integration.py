

import os
import unittest.mock

import pytest

import terrareg.errors
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace

from test import client

from test.integration.terrareg import TerraregIntegrationTest

class TestApiModuleProviderBitbucketHookIntegration(TerraregIntegrationTest):
    """Perform integration tests of ApiModuleProviderBitbucketHook"""

    @pytest.mark.parametrize('import_failures,response_code,response', [
        # with no failures
        ([], 200, {
            'status': 'Success',
            'message': 'Imported all provided tags',
            'tags': {
                '5.1.2': {'status': 'Success'},
                '6.2.0': {'status': 'Success'}
            }
        }),

        # With 1 failure
        (['5.1.2'], 500, {
            'status': 'Error',
            'message': 'One or more tags failed to import',
            'tags': {
                '5.1.2': {'message': 'Unittest clone error', 'status': 'Failed'},
                '6.2.0': {'status': 'Success'}
            }
        }),

        # With all failures,
        (['5.1.2', '6.2.0'], 500, {
            'status': 'Error',
            'message': 'One or more tags failed to import',
            'tags': {
                '5.1.2': {'message': 'Unittest clone error', 'status': 'Failed'},
                '6.2.0': {'message': 'Unittest clone error', 'status': 'Failed'}
            }
        })
    ])
    def test_hook_with_multiple_tags_with_failure(self, import_failures, response_code, response, client):
        """Test hook call with multiple tag changes."""

        def clone_repository_side_effect(self):
            if self._module_version.version in import_failures:
                raise terrareg.errors.GitCloneError('Unittest clone error')
            else:
                with open(os.path.join(self.extract_directory, 'main.tf'), 'w') as fh:
                    fh.write('output "test" { value = "test" }')

        with unittest.mock.patch(
                'terrareg.module_extractor.GitModuleExtractor._clone_repository', clone_repository_side_effect) as mocked_clone_repository, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ''):

            module = Module(Namespace('moduleextraction'), 'bitbucketmultipletags')

            # Remove module provider, if it already exists
            module_provider = ModuleProvider.get(module, 'onefailure')
            if module_provider:
                module_provider.delete()

            module_provider = ModuleProvider.get(
                module,
                'onefailure',
                create=True
            )
            module_provider.update_repo_clone_url_template('https://localhost/test.git')
            module_provider.update_git_tag_format('v{version}')

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketmultipletags/onefailure/hooks/bitbucket',
                json={
                    "changes": [
                        {
                            "ref": {
                                "id": "refs/tags/v5.1.2",
                                "displayId": "v5.1.2",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v5.1.2",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        {
                            "ref": {
                                "id": "refs/tags/v6.2.0",
                                "displayId": "v6.2.0",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v6.2.0",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        }
                    ]
                }
            )

            assert res.status_code == response_code
            assert res.json == response

            namespace = Namespace('moduleextraction')
            module = Module(namespace, 'bitbucketmultipletags')
            provider = ModuleProvider.get(module, 'onefailure')
            assert provider is not None

            # Ensure the versions were created based on whether they imported successfully
            module_version = ModuleVersion.get(module_provider=provider, version='5.1.2')
            if '5.1.2' in import_failures:
                assert module_version is None
            else:
                assert module_version is not None

            module_version = ModuleVersion.get(module_provider=provider, version='6.2.0')
            if '6.2.0' in import_failures:
                assert module_version is None
            else:
                assert module_version is not None
