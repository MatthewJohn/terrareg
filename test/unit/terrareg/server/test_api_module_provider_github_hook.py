
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
    def test_hook_with_extraction_exception(self, client, mock_models):
        """Test hook call with multiple tag changes."""
        with unittest.mock.patch(
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'terrareg.models.ModuleVersion.prepare_module') as mocked_prepare_module, \
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
                    'X-Hub-Signature': 'sha256={}'.format(valid_signature.hexdigest())
                }
            )

            assert res.status_code == 200
            assert res.json == {'status': 'Success', 'message': 'Imported provided tag', 'tag': 'v4.0.6'}

            mocked_prepare_module.assert_called_once()
            mocked_process_upload.assert_called_once()
