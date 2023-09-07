
import hashlib
import hmac
import json
import unittest.mock

import pytest

import terrareg.errors
from test.unit.terrareg import (
    mock_models,
    setup_test_data, TerraregUnitTest
)
from test import client


class TestApiModuleVersionGithubHook(TerraregUnitTest):
    """Test TestApiModuleVersionCreateGithubHook resource."""

    @setup_test_data()
    def test_hook_with_full_payload_single_change(self, client, mock_models):
        """Test hook call full payload."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                json={
                    "action": "published",
                    "release": {
                        "url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases/2",
                        "assets_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases/2/assets",
                        "upload_url": "https://octocoders.github.io/api/uploads/repos/Codertocat/Hello-World/releases/2/assets{?name,label}",
                        "html_url": "https://octocoders.github.io/Codertocat/Hello-World/releases/tag/v4.0.7",
                        "id": 2,
                        "node_id": "MDc6UmVsZWFzZTI=",
                        "tag_name": "v4.0.6",
                        "target_commitish": "master",
                        "name": None,
                        "draft": False,
                        "author": {
                            "login": "Codertocat",
                            "id": 4,
                            "node_id": "MDQ6VXNlcjQ=",
                            "avatar_url": "https://octocoders.github.io/avatars/u/4?",
                            "gravatar_id": "",
                            "url": "https://octocoders.github.io/api/v3/users/Codertocat",
                            "html_url": "https://octocoders.github.io/Codertocat",
                            "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
                            "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
                            "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
                            "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
                            "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
                            "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
                            "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
                            "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
                            "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
                            "type": "User",
                            "site_admin": False
                        },
                        "prerelease": False,
                        "created_at": "2019-05-15T19:37:08Z",
                        "published_at": "2019-05-15T19:38:20Z",
                        "assets": [],
                        "tarball_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/tarball/0.0.1",
                        "zipball_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/zipball/0.0.1",
                        "body": None
                    },
                    "repository": {
                        "id": 118,
                        "node_id": "MDEwOlJlcG9zaXRvcnkxMTg=",
                        "name": "Hello-World",
                        "full_name": "Codertocat/Hello-World",
                        "private": False,
                        "owner": {
                            "login": "Codertocat",
                            "id": 4,
                            "node_id": "MDQ6VXNlcjQ=",
                            "avatar_url": "https://octocoders.github.io/avatars/u/4?",
                            "gravatar_id": "",
                            "url": "https://octocoders.github.io/api/v3/users/Codertocat",
                            "html_url": "https://octocoders.github.io/Codertocat",
                            "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
                            "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
                            "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
                            "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
                            "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
                            "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
                            "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
                            "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
                            "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
                            "type": "User",
                            "site_admin": False
                        },
                        "html_url": "https://octocoders.github.io/Codertocat/Hello-World",
                        "description": None,
                        "fork": False,
                        "url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World",
                        "forks_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/forks",
                        "keys_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/keys{/key_id}",
                        "collaborators_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/collaborators{/collaborator}",
                        "teams_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/teams",
                        "hooks_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/hooks",
                        "issue_events_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues/events{/number}",
                        "events_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/events",
                        "assignees_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/assignees{/user}",
                        "branches_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/branches{/branch}",
                        "tags_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/tags",
                        "blobs_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/blobs{/sha}",
                        "git_tags_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/tags{/sha}",
                        "git_refs_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/refs{/sha}",
                        "trees_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/trees{/sha}",
                        "statuses_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/statuses/{sha}",
                        "languages_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/languages",
                        "stargazers_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/stargazers",
                        "contributors_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/contributors",
                        "subscribers_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/subscribers",
                        "subscription_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/subscription",
                        "commits_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/commits{/sha}",
                        "git_commits_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/commits{/sha}",
                        "comments_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/comments{/number}",
                        "issue_comment_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues/comments{/number}",
                        "contents_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/contents/{+path}",
                        "compare_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/compare/{base}...{head}",
                        "merges_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/merges",
                        "archive_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/{archive_format}{/ref}",
                        "downloads_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/downloads",
                        "issues_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues{/number}",
                        "pulls_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/pulls{/number}",
                        "milestones_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/milestones{/number}",
                        "notifications_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/notifications{?since,all,participating}",
                        "labels_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/labels{/name}",
                        "releases_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases{/id}",
                        "deployments_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/deployments",
                        "created_at": "2019-05-15T19:37:07Z",
                        "updated_at": "2019-05-15T19:38:15Z",
                        "pushed_at": "2019-05-15T19:38:19Z",
                        "git_url": "git://octocoders.github.io/Codertocat/Hello-World.git",
                        "ssh_url": "git@octocoders.github.io:Codertocat/Hello-World.git",
                        "clone_url": "https://octocoders.github.io/Codertocat/Hello-World.git",
                        "svn_url": "https://octocoders.github.io/Codertocat/Hello-World",
                        "homepage": None,
                        "size": 0,
                        "stargazers_count": 0,
                        "watchers_count": 0,
                        "language": "Ruby",
                        "has_issues": True,
                        "has_projects": True,
                        "has_downloads": True,
                        "has_wiki": True,
                        "has_pages": True,
                        "forks_count": 1,
                        "mirror_url": None,
                        "archived": False,
                        "disabled": False,
                        "open_issues_count": 2,
                        "license": None,
                        "forks": 1,
                        "open_issues": 2,
                        "watchers": 0,
                        "default_branch": "master"
                    },
                    "enterprise": {
                        "id": 1,
                        "slug": "github",
                        "name": "GitHub",
                        "node_id": "MDg6QnVzaW5lc3Mx",
                        "avatar_url": "https://octocoders.github.io/avatars/b/1?",
                        "description": None,
                        "website_url": None,
                        "html_url": "https://octocoders.github.io/businesses/github",
                        "created_at": "2019-05-14T19:31:12Z",
                        "updated_at": "2019-05-14T19:31:12Z"
                    },
                    "sender": {
                        "login": "Codertocat",
                        "id": 4,
                        "node_id": "MDQ6VXNlcjQ=",
                        "avatar_url": "https://octocoders.github.io/avatars/u/4?",
                        "gravatar_id": "",
                        "url": "https://octocoders.github.io/api/v3/users/Codertocat",
                        "html_url": "https://octocoders.github.io/Codertocat",
                        "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
                        "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
                        "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
                        "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
                        "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
                        "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
                        "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
                        "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
                        "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
                        "type": "User",
                        "site_admin": False
                    },
                    "installation": {
                        "id": 5,
                        "node_id": "MDIzOkludGVncmF0aW9uSW5zdGFsbGF0aW9uNQ=="
                    }
                }
            )

            assert res.json == {
                'status': 'Success',
                'message': 'Imported provided tag',
                'tag': 'v4.0.6'
            }
            assert res.status_code == 200

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()


    @setup_test_data()
    def test_hook_with_module_provider_without_repository_url(self, client, mock_models):
        """Test hook call to module provider with no repository url."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/norepourl/hooks/github',
                json={
                    "action": "published",
                    "release": {
                        "tag_name": "v4.0.0"
                    }
                }
            )

            assert res.status_code == 400
            assert res.json == {'status': 'Error', 'message': 'Module provider is not configured with a repository'}

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_with_prepare_module_exception(self, client, mock_models):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            mocked_prepare_module.side_effect = unittest.mock.Mock(side_effect=terrareg.errors.TerraregError('Unittest error'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                json={
                    "action": "published",
                    "release": {
                        "tag_name": "v6.2.0"
                    }
                }
            )

            assert res.status_code == 500
            assert res.json == {
                'status': 'Error',
                'message': 'Tag failed to import',
                'tag': 'v6.2.0'
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    @pytest.mark.parametrize('pre_existing_published_module_version', [
        False,
        True
    ])
    def test_hook_with_reindexing_published_module(self, pre_existing_published_module_version, client, mock_models):
        """Test hook call whilst re-indexing a published module."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=pre_existing_published_module_version) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.models.ModuleVersion.publish') as mocked_publish, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/bitbucketexample/testprovider/hooks/github',
                json={
                    "action": "published",
                    "release": {
                        "tag_name": "v4.0.6"
                    }
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported provided tag', 'tag': 'v4.0.6'}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()
            if pre_existing_published_module_version:
                mocked_publish.assert_called_once_with()
            else:
                mocked_publish.assert_not_called()

    @setup_test_data()
    def test_hook_with_extraction_exception(self, client, mock_models):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            mocked_process_upload.side_effect = unittest.mock.Mock(side_effect=terrareg.errors.TerraregError('Unittest error'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                json={
                    "action": "published",
                    "release": {
                        "tag_name": "v6.2.0"
                    }
                }
            )

            assert res.status_code == 500
            assert res.json == {
                'status': 'Error',
                'message': 'Tag failed to import',
                'tag': 'v6.2.0'
            }

            mocked_prepare_module.assert_called()
            mocked_process_upload.assert_called()

    def _test_github_with_client_error_expected(self, client, payload, expected_error=None):
        """Test github call with invalid payload."""
        if expected_error is None:
            expected_error = 'Received a non-release hook request'
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload:

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                json=payload
            )

            assert res.status_code == 400
            assert res.json == {
                'status': 'Error',
                'message': expected_error
            }

            mocked_prepare_module.assert_not_called()
            mocked_process_upload.assert_not_called()

    @setup_test_data()
    def test_hook_without_action(self, client, mock_models):
        """Test hook call without action."""
        self._test_github_with_client_error_expected(client,
            {
                "release": {
                    "tag_name": "v6.2.0"
                }
            },
            expected_error='No action present in request'
        )

    @setup_test_data()
    def test_hook_without_release(self, client, mock_models):
        """Test hook call with without release attribute."""
        self._test_github_with_client_error_expected(client,
            {
                "action": "published"
            }
        )

    @setup_test_data()
    def test_hook_without_tag_name(self, client, mock_models):
        """Test hook call without tag name."""
        self._test_github_with_client_error_expected(client,
            {
                "action": "published",
                "release": {
                }
            },
            expected_error='tag_name not present in request'
        )

    @setup_test_data()
    def test_hook_with_invalid_tag(self, client, mock_models):
        """Test hook call with an invalid tag."""
        self._test_github_with_client_error_expected(client,
            {
                "action": "published",
                "release": {
                    "tag_name": "notatag"
                }
            },
            expected_error='Release tag does not match configured version regex'
        )

    @setup_test_data()
    def test_hook_with_invalid_release(self, client, mock_models):
        """Test hook call with invalid release attribute."""
        self._test_github_with_client_error_expected(client,
            {
                "action": "published",
                "release": None
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
    def test_hook_with_invalid_signatures_with_api_keys_enabled(self, signature, client, mock_models):
        """Test hook call with upload API keys enabled with invalid request signature."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ['test-api-key1', 'test-api-key2']):

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
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
    def test_hook_with_valid_api_key_signature(self, client, mock_models):
        """Test hook call with valid API key signature."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ['test-api-key1', 'test-api-key2']):

            request_json = {
                "action": "published",
                "release": {
                    "tag_name": "v4.0.6"
                }
            }

            valid_signature = hmac.new(bytes('test-api-key1', 'utf8'), b'', hashlib.sha256)
            valid_signature.update(json.dumps(request_json).encode('utf8'))

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                data=json.dumps(request_json),
                headers={
                    'Content-Type': 'application/json',
                    'x-hub-signature-256': 'sha256={}'.format(valid_signature.hexdigest())
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported provided tag', 'tag': 'v4.0.6'}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

    @setup_test_data()
    def test_hook_with_real_payload_and_valid_api_key_signature(self, client, mock_models):
        """Test hook call with valid API key signature."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module', return_value=False) as mocked_prepare_module, \
                unittest.mock.patch(
                    'terrareg.module_extractor.GitModuleExtractor.process_upload') as mocked_process_upload, \
                unittest.mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', ['test-api-key1', 'test-api-key2']):

            request_body = (
                '{"action":"released","release":{"url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/releases/120339891",'
                '"assets_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/releases/120339891/assets",'
                '"upload_url":"https://uploads.github.com/repos/MatthewJohn/test-terrareg-hook/releases/120339891/assets{?name,label}",'
                '"html_url":"https://github.com/MatthewJohn/test-terrareg-hook/releases/tag/v1.0.0","id":120339891,'
                '"author":{"login":"MatthewJohn","id":1266262,"node_id":"MDQ6VXNlcjEyNjYyNjI=",'
                '"avatar_url":"https://avatars.githubusercontent.com/u/1266262?v=4","gravatar_id":"",'
                '"url":"https://api.github.com/users/MatthewJohn","html_url":"https://github.com/MatthewJohn",'
                '"followers_url":"https://api.github.com/users/MatthewJohn/followers",'
                '"following_url":"https://api.github.com/users/MatthewJohn/following{/other_user}",'
                '"gists_url":"https://api.github.com/users/MatthewJohn/gists{/gist_id}",'
                '"starred_url":"https://api.github.com/users/MatthewJohn/starred{/owner}{/repo}",'
                '"subscriptions_url":"https://api.github.com/users/MatthewJohn/subscriptions",'
                '"organizations_url":"https://api.github.com/users/MatthewJohn/orgs",'
                '"repos_url":"https://api.github.com/users/MatthewJohn/repos",'
                '"events_url":"https://api.github.com/users/MatthewJohn/events{/privacy}",'
                '"received_events_url":"https://api.github.com/users/MatthewJohn/received_events",'
                '"type":"User","site_admin":false},"node_id":"RE_kwDOKQvcu84HLD2z","tag_name":"v1.0.0",'
                '"target_commitish":"main","name":"v1.0.0","draft":false,"prerelease":false,'
                '"created_at":"2023-09-07T19:37:35Z","published_at":"2023-09-07T19:43:54Z",'
                '"assets":[],"tarball_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/tarball/v1.0.0",'
                '"zipball_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/zipball/v1.0.0","body":""},'
                '"repository":{"id":688643259,"node_id":"R_kgDOKQvcuw","name":"test-terrareg-hook",'
                '"full_name":"MatthewJohn/test-terrareg-hook","private":false,'
                '"owner":{"login":"MatthewJohn","id":1266262,"node_id":"MDQ6VXNlcjEyNjYyNjI=",'
                '"avatar_url":"https://avatars.githubusercontent.com/u/1266262?v=4","gravatar_id":"",'
                '"url":"https://api.github.com/users/MatthewJohn","html_url":"https://github.com/MatthewJohn",'
                '"followers_url":"https://api.github.com/users/MatthewJohn/followers",'
                '"following_url":"https://api.github.com/users/MatthewJohn/following{/other_user}",'
                '"gists_url":"https://api.github.com/users/MatthewJohn/gists{/gist_id}",'
                '"starred_url":"https://api.github.com/users/MatthewJohn/starred{/owner}{/repo}",'
                '"subscriptions_url":"https://api.github.com/users/MatthewJohn/subscriptions",'
                '"organizations_url":"https://api.github.com/users/MatthewJohn/orgs",'
                '"repos_url":"https://api.github.com/users/MatthewJohn/repos",'
                '"events_url":"https://api.github.com/users/MatthewJohn/events{/privacy}",'
                '"received_events_url":"https://api.github.com/users/MatthewJohn/received_events",'
                '"type":"User","site_admin":false},"html_url":"https://github.com/MatthewJohn/test-terrareg-hook",'
                '"description":null,"fork":false,"url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook",'
                '"forks_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/forks",'
                '"keys_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/keys{/key_id}",'
                '"collaborators_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/collaborators{/collaborator}",'
                '"teams_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/teams",'
                '"hooks_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/hooks",'
                '"issue_events_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/issues/events{/number}",'
                '"events_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/events",'
                '"assignees_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/assignees{/user}",'
                '"branches_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/branches{/branch}",'
                '"tags_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/tags",'
                '"blobs_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/git/blobs{/sha}",'
                '"git_tags_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/git/tags{/sha}",'
                '"git_refs_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/git/refs{/sha}",'
                '"trees_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/git/trees{/sha}",'
                '"statuses_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/statuses/{sha}",'
                '"languages_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/languages",'
                '"stargazers_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/stargazers",'
                '"contributors_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/contributors",'
                '"subscribers_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/subscribers",'
                '"subscription_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/subscription",'
                '"commits_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/commits{/sha}",'
                '"git_commits_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/git/commits{/sha}",'
                '"comments_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/comments{/number}",'
                '"issue_comment_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/issues/comments{/number}",'
                '"contents_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/contents/{+path}",'
                '"compare_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/compare/{base}...{head}",'
                '"merges_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/merges",'
                '"archive_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/{archive_format}{/ref}",'
                '"downloads_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/downloads",'
                '"issues_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/issues{/number}",'
                '"pulls_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/pulls{/number}",'
                '"milestones_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/milestones{/number}",'
                '"notifications_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/notifications{?since,all,participating}",'
                '"labels_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/labels{/name}",'
                '"releases_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/releases{/id}",'
                '"deployments_url":"https://api.github.com/repos/MatthewJohn/test-terrareg-hook/deployments",'
                '"created_at":"2023-09-07T19:37:35Z",'
                '"updated_at":"2023-09-07T19:37:36Z",'
                '"pushed_at":"2023-09-07T19:43:54Z",'
                '"git_url":"git://github.com/MatthewJohn/test-terrareg-hook.git",'
                '"ssh_url":"git@github.com:MatthewJohn/test-terrareg-hook.git",'
                '"clone_url":"https://github.com/MatthewJohn/test-terrareg-hook.git",'
                '"svn_url":"https://github.com/MatthewJohn/test-terrareg-hook",'
                '"homepage":null,"size":0,"stargazers_count":0,"watchers_count":0,"language":null,"has_issues":true,"has_projects":true,"has_downloads":true,"has_wiki":true,"has_pages":false,"has_discussions":false,"forks_count":0,"mirror_url":null,"archived":false,"disabled":false,"open_issues_count":0,"license":null,"allow_forking":true,"is_template":false,"web_commit_signoff_required":false,"topics":[],"visibility":"public",'
                '"forks":0,"open_issues":0,"watchers":0,"default_branch":"main"},"sender":{"login":"MatthewJohn",'
                '"id":1266262,"node_id":"MDQ6VXNlcjEyNjYyNjI=",'
                '"avatar_url":"https://avatars.githubusercontent.com/u/1266262?v=4",'
                '"gravatar_id":"",'
                '"url":"https://api.github.com/users/MatthewJohn",'
                '"html_url":"https://github.com/MatthewJohn",'
                '"followers_url":"https://api.github.com/users/MatthewJohn/followers",'
                '"following_url":"https://api.github.com/users/MatthewJohn/following{/other_user}",'
                '"gists_url":"https://api.github.com/users/MatthewJohn/gists{/gist_id}",'
                '"starred_url":"https://api.github.com/users/MatthewJohn/starred{/owner}{/repo}",'
                '"subscriptions_url":"https://api.github.com/users/MatthewJohn/subscriptions",'
                '"organizations_url":"https://api.github.com/users/MatthewJohn/orgs",'
                '"repos_url":"https://api.github.com/users/MatthewJohn/repos",'
                '"events_url":"https://api.github.com/users/MatthewJohn/events{/privacy}",'
                '"received_events_url":"https://api.github.com/users/MatthewJohn/received_events",'
                '"type":"User",'
                '"site_admin":false}}'
            )

            res = client.post(
                '/v1/terrareg/modules/moduleextraction/githubexample/testprovider/hooks/github',
                data=request_body,
                headers={
                    'Content-Type': 'application/json',
                    "x-hub-signature": "sha1=68b65512af03e57467031a8f88ff81d14d3dd6e2",
                    "x-hub-signature-256": "sha256=9d73318ab08866f74febbbc97971471d1078f75a857cad0f3ba0c537f3a86c04",
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported provided tag', 'tag': 'v1.0.0'}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()

