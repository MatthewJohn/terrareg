
import hashlib
import hmac
import json
import unittest.mock

import pytest

import terrareg.errors
from test.unit.terrareg import (
    mocked_server_namespace_fixture,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionBitbucketHook(TerraregUnitTest):
    """Test TestApiModuleVersionCreateBitBucketHook resource."""

    @setup_test_data()
    def test_hook_with_full_payload_single_change(self, client, mocked_server_namespace_fixture):
        """Test hook call full payload."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json={
                    "eventKey": "repo:refs_changed", "date": "2022-04-23T21:21:46+0000",
                    "actor": {
                        "name": "admin", "emailAddress": "admin@localhost",
                        "id": 2, "displayName": "admin", "active": True,
                        "slug": "admin", "type": "NORMAL",
                        "links": {"self": [ {"href": "http://localhost:7990/users/admin"}]
                        }
                    },
                    "repository": {
                        "slug": "test-module", "id": 1, "name": "test-module", "hierarchyId": "34098b9e0f8011fcfb25",
                        "scmId": "git", "state": "AVAILABLE", "statusMessage": "Available", "forkable": True,
                        "project": {
                            "key": "BLA", "id": 1, "name": "bla", "public": True, "type": "NORMAL",
                            "links": {"self": [{"href": "http://localhost:7990/projects/BLA"}]}
                        },
                        "public": True,
                        "links": {
                            "clone": [
                                {"href": "ssh://git@localhost:7999/bla/test-module.git", "name": "ssh"},
                                {"href": "http://localhost:7990/scm/bla/test-module.git", "name": "http"}
                            ],
                            "self": [{"href": "http://localhost:7990/projects/BLA/repos/test-module/browse"}]
                        }
                    },
                    "changes": [
                        {
                            "ref": {
                                "id": "refs/tags/v4.0.6",
                                "displayId": "v4.0.6",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v4.0.6",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        }
                    ]
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported all provided tags', 'tags': {'4.0.6': {'status': 'Success'}}}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()


    @setup_test_data()
    def test_hook_with_module_provider_without_repository_url(self, client, mocked_server_namespace_fixture):
        """Test hook call to module provider with no repository url."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/norepourl/hooks/bitbucket',
                json={
                    "changes": [
                        {
                            "ref": {
                                "id": "refs/tags/v4.0.6",
                                "displayId": "v4.0.6",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v4.0.6",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        }
                    ]
                }
            )

            assert res.status_code == 400
            assert res.json == {'message': 'Module provider is not configured with a repository'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_with_prepare_module_exception(self, client, mocked_server_namespace_fixture):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            mocked_prepare_module.side_effect = unittest.mock.Mock(side_effect=terrareg.errors.TerraregError('Unittest error'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json={
                    "changes": [
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

            assert res.status_code == 200
            assert res.json == {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': {
                    '6.2.0': {'status': 'Failed', 'message': 'Unittest error'}
                }
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_with_extraction_exception(self, client, mocked_server_namespace_fixture):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            mocked_process_upload.side_effect = unittest.mock.Mock(side_effect=terrareg.errors.TerraregError('Unittest error'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json={
                    "changes": [
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

            assert res.status_code == 200
            assert res.json == {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': {
                    '6.2.0': {'status': 'Failed', 'message': 'Unittest error'}
                }
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_called()

    @setup_test_data()
    def test_hook_with_multiple_tags(self, client, mocked_server_namespace_fixture):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
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

            assert res.status_code == 200
            assert res.json == {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': {
                    '5.1.2': {'status': 'Success'},
                    '6.2.0': {'status': 'Success'}
                }
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_called()

    @setup_test_data()
    def test_hook_with_invalid_changes(self, client, mocked_server_namespace_fixture):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
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
                        # DELETED TAG
                        {
                            "ref": {
                                "id": "refs/tags/v5.5.2",
                                "displayId": "v5.5.2",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v5.5.2",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "DELETE"
                        },
                        # Commit
                        {
                            'ref': {
                                'id': 'refs/heads/master', 'displayId': 'master', 'type': 'BRANCH'
                            },
                            'refId': 'refs/heads/master', 'fromHash': '1097d939669e3209ff33e6dfe982d84c204f6087',
                            'toHash': '9f492469219b96c807ad0879763ea076689cc322', 'type': 'UPDATE'
                        },
                        # Change without type
                        {
                            "ref": {
                                "id": "refs/tags/v6.0.0",
                                "displayId": "v6.0.0"
                            },
                            "refId": "refs/tags/v6.0.0",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Change invalid ref
                        {
                            "ref": None,
                            "refId": "refs/tags/v6.0.1",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Change without ref
                        {
                            "refId": "refs/tags/v6.0.2",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Change that is not a dict
                        None,
                        # Change with None refId
                        {
                            "ref": {
                                "id": None,
                                "displayId": "v5.5.3",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v5.5.3",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Change without ref Id
                        {
                            "ref": {
                                "displayId": "v5.5.4",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v5.5.4",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Change without type:
                        {
                            "ref": {
                                "id": "refs/tags/v6.0.3",
                                "displayId": "v6.0.3",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v6.0.3",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087"
                        },
                        # Invalid tag format
                        {
                            "ref": {
                                "id": "refs/tags/aa1.2.4",
                                "displayId": "aa1.2.4",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/aa1.2.4",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Invalid symentic tag version
                        {
                            "ref": {
                                "id": "refs/tags/v1.2.3-pre",
                                "displayId": "v1.2.3-pre",
                                "type": "TAG"
                            },
                            "refId": "refs/tags/v1.2.3-pre",
                            "fromHash": "0000000000000000000000000000000000000000",
                            "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                            "type": "ADD"
                        },
                        # Valid tag
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

            assert res.status_code == 200
            assert res.json == {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': {
                    '5.1.2': {'status': 'Success'},
                    '6.2.0': {'status': 'Success'}
                }
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_called()

    def _test_bitbucket_with_no_tag_result_expected(self, client, payload):
        """Test bitbucket call expecting no tags found."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json=payload
            )

            assert res.status_code == 200
            assert res.json == {
                'status': 'Success',
                'message': 'Imported all provided tags',
                'tags': {}
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_with_commit_change(self, client, mocked_server_namespace_fixture):
        """Test hook call with commit."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
                "changes": [
                    {
                        "ref": {
                            "type": "COMMIT"
                        },
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                        "type": "ADD"
                    },
                ]
            }
        )

    @setup_test_data()
    def test_hook_with_change_without_ref_type(self, client, mocked_server_namespace_fixture):
        """Test hook call with with without ref type."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
                "changes": [
                    {
                        "ref": {
                        },
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                        "type": "ADD"
                    }
                ]
            }
        )

    @setup_test_data()
    def test_hook_with_no_type(self, client, mocked_server_namespace_fixture):
        """Test hook call with with without type."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
                "changes": [
                    {
                        "ref": {
                            "id": "refs/tags/v6.2.0",
                            "displayId": "v6.2.0",
                            "type": "TAG"
                        },
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087"
                    },
                ]
            }
        )

    @setup_test_data()
    def test_hook_with_none_ref(self, client, mocked_server_namespace_fixture):
        """Test hook call with with without type."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
                "changes": [
                    {
                        "ref": None,
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                        "type": "ADD"
                    },
                ]
            }
        )

    @setup_test_data()
    def test_hook_with_deleted_tag(self, client, mocked_server_namespace_fixture):
        """Test hook call with with deleted tag."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
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
                        "type": "DELETE"
                    }
                ]
            }
        )

    @setup_test_data()
    def test_hook_without_change_type(self, client, mocked_server_namespace_fixture):
        """Test hook call with without change type."""
        self._test_bitbucket_with_no_tag_result_expected(client,
            {
                "changes": [
                    {
                        "ref": {
                            "id": "refs/tags/v5.1.2",
                            "displayId": "v5.1.2",
                            "type": "TAG"
                        },
                        "refId": "refs/tags/v5.1.2",
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087"
                    }
                ]
            }
        )

    @setup_test_data()
    def test_hook_with_nontag_changes(self, client, mocked_server_namespace_fixture):
        """Test hook call with non tag changes."""
        self._test_bitbucket_with_no_tag_result_expected(
            client,
            {
                "changes": [
                ]
            }
        )

    @pytest.mark.parametrize('signature', [
        # No header
        None,
        # Empty header value
        '',
        # Invalid signature without sha256= prefix
        'isnotavalidsignature',
        # Invalid signature
        'sha256=invalidsignature'
    ])
    @setup_test_data()
    def test_hook_with_invalid_signatures_with_api_keys_enabled(self, signature, client, mocked_server_namespace_fixture):
        """Test hook call with upload API keys enabled with invalid request signature."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ['test-api-key1', 'test-api-key2']):

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                json={},
                headers={'X-Hub-Signature': signature}
            )

            assert res.status_code == 401
            assert res.json == {
                'message': 'The server could not verify that you are authorized to access the '
                           'URL requested. You either supplied the wrong credentials (e.g. a '
                           "bad password), or your browser doesn't understand how to supply "
                           'the credentials required.'
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_with_valid_api_key_signature(self, client, mocked_server_namespace_fixture):
        """Test hook call with valid API key signature."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ['test-api-key1', 'test-api-key2']):

            request_json = {
                "eventKey": "repo:refs_changed", "date": "2022-04-23T21:21:46+0000",
                "actor": {
                    "name": "admin", "emailAddress": "admin@localhost",
                    "id": 2, "displayName": "admin", "active": True,
                    "slug": "admin", "type": "NORMAL",
                    "links": {"self": [ {"href": "http://localhost:7990/users/admin"}]
                    }
                },
                "repository": {
                    "slug": "test-module", "id": 1, "name": "test-module", "hierarchyId": "34098b9e0f8011fcfb25",
                    "scmId": "git", "state": "AVAILABLE", "statusMessage": "Available", "forkable": True,
                    "project": {
                        "key": "BLA", "id": 1, "name": "bla", "public": True, "type": "NORMAL",
                        "links": {"self": [{"href": "http://localhost:7990/projects/BLA"}]}
                    },
                    "public": True,
                    "links": {
                        "clone": [
                            {"href": "ssh://git@localhost:7999/bla/test-module.git", "name": "ssh"},
                            {"href": "http://localhost:7990/scm/bla/test-module.git", "name": "http"}
                        ],
                        "self": [{"href": "http://localhost:7990/projects/BLA/repos/test-module/browse"}]
                    }
                },
                "changes": [
                    {
                        "ref": {
                            "id": "refs/tags/v4.0.6",
                            "displayId": "v4.0.6",
                            "type": "TAG"
                        },
                        "refId": "refs/tags/v4.0.6",
                        "fromHash": "0000000000000000000000000000000000000000",
                        "toHash": "1097d939669e3209ff33e6dfe982d84c204f6087",
                        "type": "ADD"
                    }
                ]
            }

            valid_signature = hmac.new(bytes('test-api-key1', 'utf8'), b'', hashlib.sha256)
            valid_signature.update(json.dumps(request_json).encode('utf8'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/bitbucket',
                data=json.dumps(request_json),
                headers={
                    'Content-Type': 'application/json',
                    'X-Hub-Signature': 'sha256={}'.format(valid_signature.hexdigest())
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported all provided tags', 'tags': {'4.0.6': {'status': 'Success'}}}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()
