
# Temp import
import json
import tempfile
from typing import Union, List

import unittest.mock

import pytest

from .base_provider_source_tests import BaseProviderSourceTests
from . import test_provider_source, test_repository, test_provider
import terrareg.provider_source
import terrareg.errors
import terrareg.namespace_type
import terrareg.database
import terrareg.repository_model
import terrareg.repository_kind
import terrareg.provider_model
import terrareg.provider_source.repository_release_metadata
import terrareg.provider_version_model
from test.integration.terrareg.fixtures import (
    test_namespace, test_provider_category, test_gpg_key
)


TEST_GITHUB_PRIVATE_KEY = """
-----BEGIN RSA PRIVATE KEY-----
MIICXQIBAAKBgQDfUVYMRr22c+KwSt8PFxP+uDbe5thCcfPV+IxtTf/F2LAoCcNd
XdqFJ+hpokKYx6KjKDlGUIA+kf9a+CODcXs6OLqiom+ml/k47MD4eYmaoW7Sw+7Q
IzLiCwEDkOhmt/MIFDOxjSHC34jGSVleQeT8xuaIWgpiTzSv+1dMQb+V+wIDAQAB
AoGBAKx2z0J524etpbNKj0vDIfEE6XNpyjg+cvabli/QHij4eMrjB1ry4ZEWSfpS
kqYU/ziMFveDshcgf5oMqriXinbxhNX5AmdpgsxS/9Qk7vwHWSjqT/2fsAyFTT5B
fyv6/f/wLVW49lHpsr/2OT2fv7gQTb2MLPfD5I65SXxQ0t9ZAkEA/CzhpKV3Xuk1
qOhlik8UfFHnp0/cyWj3592LqCAcZRPUbL/9idF9MaVBQOYssXiScakwPhWtzB6J
ru9ce8h0hQJBAOK0aNtQCCpgzgpzxHvXDcSpRW4PX/UfCNE3Kpv//IqHtDeZJJZM
aytB/UwExLN71o8DUdZ03WBSQrw6GmaFKH8CQDko1zChzO/3fpE9tB5olGUlj5Ou
F4aTw3WMEybVuHn0x7aqwgZmNLF3GtZiFglYIiGfTu8TrORSm7TKTrVEF50CQQCl
joKUxplv+Un+sBRpK9/OIp+lhGzbIVLbFqJzUjonIHsnrxrc9+m7qXFFNqY/PMyv
nAkDyExyryA1PWlSPSQZAkB8JBd321fxU6uJegsyWHQalfadzALKuQeVoQ9603Eu
2KZ2AFT0zXVdkE0D7VlcxXWMIUn9fUkzUQFxjbHqf9SG
-----END RSA PRIVATE KEY-----
""".strip()



class TestGithubProviderSource(BaseProviderSourceTests):

    _CLASS = terrareg.provider_source.GithubProviderSource
    ADDITIONAL_CONFIG = {
        "login_button_text": "Unit Test Github Login",
        "base_url": "https://github.example-test.com",
        "api_url": "https://api.github.example-test.com",
        "client_id": "unittest-github-client-id",
        "client_secret": "unittest-github-client-secret",
        "private_key_path": "./unittest-path-to-private-key.pem",
        "app_id": 954956
    }

    def test_generate_db_config_from_source_config(self):
        """Test generate_db_config_from_source_config"""
        assert self._CLASS.generate_db_config_from_source_config({
            "base_url": "https://github.example.com",
            "api_url": "https://api.github.example.com",
            "client_id": "unittest-client-id",
            "client_secret": "unittest-client-secret",
            "login_button_text": "Login via Github using this unit test",
            "private_key_path": "./path/to/key.pem",
            "app_id": "1234appid",
            "default_access_token": "pa-test-personal-access-token",
            "default_installation_id": "ut-default-installation-id-here",
            "auto_generate_github_organisation_namespaces": False
        }) == {
            "base_url": "https://github.example.com",
            "api_url": "https://api.github.example.com",
            "client_id": "unittest-client-id",
            "client_secret": "unittest-client-secret",
            "login_button_text": "Login via Github using this unit test",
            "private_key_path": "./path/to/key.pem",
            "app_id": "1234appid",
            "default_access_token": "pa-test-personal-access-token",
            "default_installation_id": "ut-default-installation-id-here",
            "auto_generate_github_organisation_namespaces": False
        }

    @pytest.mark.parametrize('missing_argument', [
        "base_url",
        "api_url",
        "client_id",
        "client_secret",
        "login_button_text",
        "private_key_path",
        "app_id",
        "auto_generate_github_organisation_namespaces"
    ])
    def test_generate_db_config_from_source_config_missing_required_argument(self, missing_argument):
        """Test generate_db_config_from_source_config"""
        config = {
            "base_url": "https://github.example.com",
            "api_url": "https://api.github.example.com",
            "client_id": "unittest-client-id",
            "client_secret": "unittest-client-secret",
            "login_button_text": "Login via Github using this unit test",
            "private_key_path": "./path/to/key.pem",
            "app_id": "1234appid",
            "default_access_token": "pa-test-personal-access-token",
            "default_installation_id": "ut-default-installation-id-here",
            "auto_generate_github_organisation_namespaces": False
        }
        # Ensure config works before removing argument
        self._CLASS.generate_db_config_from_source_config(config)

        # Delete missing parameter, re-run and ensure it raises an exception
        del config[missing_argument]
        with pytest.raises(terrareg.errors.InvalidProviderSourceConfigError):
            self._CLASS.generate_db_config_from_source_config(config)

    @pytest.mark.parametrize('missing_argument', [
        "default_access_token",
        "default_installation_id",
    ])
    def test_generate_db_config_from_source_config_missing_optional_argument(self, missing_argument):
        """Test generate_db_config_from_source_config with missing optional argument"""
        config = {
            "base_url": "https://github.example.com",
            "api_url": "https://api.github.example.com",
            "client_id": "unittest-client-id",
            "client_secret": "unittest-client-secret",
            "login_button_text": "Login via Github using this unit test",
            "private_key_path": "./path/to/key.pem",
            "app_id": "1234appid",
            "default_access_token": "pa-test-personal-access-token",
            "default_installation_id": "ut-default-installation-id-here",
            "auto_generate_github_organisation_namespaces": False
        }
        # Ensure config works before removing argument
        self._CLASS.generate_db_config_from_source_config(config)

        # Delete missing parameter, re-run and ensure it raises an exception
        del config[missing_argument]
        result_config = self._CLASS.generate_db_config_from_source_config(config)
        assert missing_argument not in result_config

    def test_login_button_text(self, test_provider_source):
        """Test login_button_text property"""
        assert test_provider_source.login_button_text == "Unit Test Github Login"

    @pytest.mark.parametrize('request_response_code, request_response, expected_response', [
        (200, "first_param=125132&access_token=unittest-access-token&someotherparam=123", "unittest-access-token"),
        # Invalid response code
        (400, None, None),
        # Invalid response data
        (200, "first_param=125132&someotherparam=123", None)
    ])
    def test_get_user_access_token(self, request_response_code, request_response, expected_response, test_provider_source):
        """Test get_user_access_token"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = request_response_code
        mock_response.text = request_response

        with unittest.mock.patch('terrareg.provider_source.github.requests.post', unittest.mock.MagicMock()) as mock_requests_post:
            mock_requests_post.return_value = mock_response

            assert test_provider_source.get_user_access_token(code="abcdef-inputcode") == expected_response

        mock_requests_post.assert_called_once_with(
            # Not API URL!
            # As per https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authenticating-to-the-rest-api-with-an-oauth-app
            "https://github.example-test.com/login/oauth/access_token",
            data={
                "client_id": "unittest-github-client-id",
                "client_secret": "unittest-github-client-secret",
                "code": "abcdef-inputcode"
            }
        )

    @pytest.mark.parametrize('result_count, expected_pages', [
        # No results
        (0, 1),
        # Some results
        (1, 1),
        (50, 1),
        # Page count boundary tests
        (99, 1),
        (100, 2),
        (101, 2),
        (250, 3),
    ])
    def test_update_repositories(self, result_count, expected_pages, test_provider_source):
        """Test update_repositories"""
        response_data_values = [
            [
                {"test_repository_itx": row_itx + (page_itx * 100)}
                for row_itx in range(0, min((result_count - (page_itx * 100)), 100))
            ]
            for page_itx in range(0, expected_pages)
        ]

        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 200
        mock_response.json.side_effect = response_data_values

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock()) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._add_repository') as mock__add_repository:
            mock_requests_get.return_value = mock_response

            test_provider_source.update_repositories(access_token="abcdef-test-access-token")

        expected_add_repository_calls = [
            unittest.mock.call(repository_metadata={"test_repository_itx": itx})
            for itx in range(result_count)
        ]
        mock__add_repository.assert_has_calls(
            calls=expected_add_repository_calls
        )

        expected_request_get_calls = []
        for page_itx in range(expected_pages):
            expected_request_get_calls.append(unittest.mock.call(
                'https://api.github.example-test.com/user/repos',
                params={
                    'visibility': 'public',
                    'affiliation': 'owner,organization_member',
                    'sort': 'created',
                    'direction': 'desc',
                    'per_page': '100',
                    'page': str(page_itx + 1)
                },
                headers={
                    'X-GitHub-Api-Version': '2022-11-28',
                    'Accept': 'application/vnd.github+json',
                    'Authorization': 'Bearer abcdef-test-access-token'
                }
            ))
            expected_request_get_calls.append(unittest.mock.call().json())
        mock_requests_get.assert_has_calls(expected_request_get_calls)

    def test_update_repositories_invalid_response(self, test_provider_source):
        """Test update_repositories with invalid response"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 500

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock()) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._add_repository') as mock__add_repository:
            mock_requests_get.return_value = mock_response

            test_provider_source.update_repositories(access_token="abcdef-test-access-token")

        mock__add_repository.assert_not_called()
        mock_requests_get.assert_called_once()
        mock_response.json.assert_not_called()

    def test_refresh_namespace_repositories_no_access_token(self, test_provider_source, test_namespace):
        """Test refresh_namespace_repositories with no default access token"""
        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_default_access_token',
                    unittest.mock.MagicMock()) as mock__get_default_access_token, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._is_entity_org_or_user',
                    unittest.mock.MagicMock()) as mock__is_entity_org_or_user, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock()) as mock_requests_get:

            mock__get_default_access_token.return_value = None

            with pytest.raises(terrareg.errors.ProviderSourceDefaultAccessTokenNotConfiguredError):
                test_provider_source.refresh_namespace_repositories(namespace=test_namespace)

            mock__get_default_access_token.assert_called_once_with()
            mock__is_entity_org_or_user.assert_not_called()
            mock_requests_get.assert_not_called()

    def test_refresh_namespace_repositories_no_type(self, test_provider_source, test_namespace):
        """Test refresh_namespace_repositories and unable to determine github user type"""
        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_default_access_token',
                    unittest.mock.MagicMock()) as mock__get_default_access_token, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._is_entity_org_or_user',
                    unittest.mock.MagicMock()) as mock__is_entity_org_or_user, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock()) as mock_requests_get:

            mock__get_default_access_token.return_value = "abcdefg-access-token"
            mock__is_entity_org_or_user.return_value = None

            with pytest.raises(terrareg.errors.GithubEntityDoesNotExistError):
                test_provider_source.refresh_namespace_repositories(namespace=test_namespace)

            mock__get_default_access_token.assert_called_once_with()
            mock__is_entity_org_or_user.assert_called_once_with('some-organisation', access_token='abcdefg-access-token')
            mock_requests_get.assert_not_called()

    @pytest.mark.parametrize('result_count, expected_pages', [
        # No results
        (0, 1),
        # # Some results
        (1, 1),
        (50, 1),
        # Page count boundary tests
        (99, 1),
        (100, 2),
        (101, 2),
        (250, 3),
    ])
    @pytest.mark.parametrize('namespace_type, expected_api_endpoint', [
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, '/users/some-organisation/repos'),
        (terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION, '/orgs/some-organisation/repos'),
    ])
    def test_refresh_namespace_repositories(self, result_count, expected_pages, namespace_type, expected_api_endpoint, test_provider_source, test_namespace):
        """Test refresh_namespace_repositories"""
        response_data_values = [
            [
                {"test_repository_itx": row_itx + (page_itx * 100)}
                for row_itx in range(0, min((result_count - (page_itx * 100)), 100))
            ]
            for page_itx in range(0, expected_pages)
        ]

        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 200
        mock_response.json.side_effect = response_data_values

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_default_access_token',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_default_access_token, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._is_entity_org_or_user',
                    unittest.mock.MagicMock(return_value=namespace_type)) as mock__is_entity_org_or_user, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._add_repository') as mock__add_repository:

            test_provider_source.refresh_namespace_repositories(namespace=test_namespace)

        mock__get_default_access_token.assert_called_once_with()
        mock__is_entity_org_or_user.assert_called_once_with('some-organisation', access_token='abcdef-test-access-token')

        expected_add_repository_calls = [
            unittest.mock.call(repository_metadata={"test_repository_itx": itx})
            for itx in range(result_count)
        ]
        mock__add_repository.assert_has_calls(
            calls=expected_add_repository_calls
        )

        expected_request_get_calls = [
            unittest.mock.call(
                f'https://api.github.example-test.com{expected_api_endpoint}',
                params={
                    "sort": "created",
                    "direction": "desc",
                    "per_page": "100",
                    'page': str(page_itx + 1)
                },
                headers={
                    'X-GitHub-Api-Version': '2022-11-28',
                    'Accept': 'application/vnd.github+json',
                    'Authorization': 'Bearer abcdef-test-access-token'
                }
            )
            for page_itx in range(expected_pages)
        ]
        mock_requests_get.assert_has_calls(expected_request_get_calls)

    def test_refresh_namespace_repositories_invalid_response_code(self, test_provider_source, test_namespace):
        """Test refresh_namespace_repositories with invalid response code"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 500

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_default_access_token',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_default_access_token, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._is_entity_org_or_user',
                    unittest.mock.MagicMock(return_value=terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION)) as mock__is_entity_org_or_user, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._add_repository') as mock__add_repository:

            test_provider_source.refresh_namespace_repositories(namespace=test_namespace)

        mock__get_default_access_token.assert_called_once_with()
        mock__is_entity_org_or_user.assert_called_once_with('some-organisation', access_token='abcdef-test-access-token')
        mock__add_repository.assert_not_called()

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/orgs/some-organisation/repos',
            params={
                "sort": "created",
                "direction": "desc",
                "per_page": "100",
                'page': "1"
            },
            headers={
                'X-GitHub-Api-Version': '2022-11-28',
                'Accept': 'application/vnd.github+json',
                'Authorization': 'Bearer abcdef-test-access-token'
            }
        )

    @pytest.mark.parametrize('release_version, git_tag, release_id, github_release_metadata, get_commit_hash_result, should_call_get_commit_hash, already_exists, get_release_artifacts_metadata, get_release_artifacts_metadata_should_be_called, expected_response', [
        # Example returning valid metadata
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "name": "Unit Test Release v5.2.3", "tag_name": "v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, should commit hash be called, does release already exist
         "abcefg12345", True, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should be called
         True,
         # Expected response
         terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
             name="Unit Test Release v5.2.3", tag="v5.2.3", archive_url="https://example.com/release/v5.2.3.zip",
             commit_hash="abcefg12345", provider_id="unittest-release-id", release_artifacts=[
                 terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")
             ])
        ),

        # Already exists
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "name": "Unit Test Release v5.2.3", "tag_name": "v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, should commit hash be called, does release already exist
         "abcefg12345", True, True,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should be called
         False,
         # Expected response
         terrareg.provider_version_model.ProviderVersion
        ),

        # Missing required metadata from
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"name": "Unit Test Release v5.2.3", "tag_name": "v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, commit hash should not be called, does release already exist
         "abcefg12345", False, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should not be called
         False,
         # Expected response
         None,
        ),
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "tag_name": "v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, commit hash should not be called, does release already exist
         "abcefg12345", False, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should not be called
         False,
         # Expected response
         None,
        ),
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "name": "Unit Test Release v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, commit hash should not be called, does release already exist
         "abcefg12345", False, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should not be called
         False,
         # Expected response
         None,
        ),
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "name": "Unit Test Release v5.2.3", "tag_name": "v5.2.3"},
         # Commit hash, commit hash should not be called, does release already exist
         "abcefg12345", False, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should not be called
         False,
         # Expected response
         None,
        ),

        # No commit hash found
        ("5.2.3", "v5.2.3", "unittest-release-id",
         # Metadata
         {"id": "unittest-release-id", "name": "Unit Test Release v5.2.3", "tag_name": "v5.2.3", "tarball_url": "https://example.com/release/v5.2.3.zip"},
         # Commit hash, commit hash should not be called, does release already exist
         None, True, False,
         # Release artifacts
         [terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="release-artifact1", provider_id="relase-artifact-provider-id")],
         # Release artifacts should not be called
         False,
         # Expected response
         None,
        ),
    ])
    def test__process_release(self, release_version, git_tag, release_id, github_release_metadata,
                              get_commit_hash_result, should_call_get_commit_hash, already_exists,
                              get_release_artifacts_metadata, get_release_artifacts_metadata_should_be_called, expected_response,
                              test_provider_source, test_provider, test_repository, test_gpg_key):
        """Test _process_release method"""

        # If expected response data is ProviderVersion class, convert to
        # instance of provider version, using fixtures
        if expected_response is terrareg.provider_version_model.ProviderVersion:
            expected_response = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=release_version)

        pre_existing_provider_version = None
        if already_exists:
            pre_existing_provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version=release_version)
            pre_existing_provider_version._create_db_row(git_tag=git_tag, gpg_key=test_gpg_key)

        try:

            with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._get_release_artifacts_metadata',
                                     unittest.mock.MagicMock(return_value=get_release_artifacts_metadata)) as mock__get_release_artifacts_metadata, \
                    unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._get_commit_hash_by_release',
                                        unittest.mock.MagicMock(return_value=get_commit_hash_result)) as mock__get_commit_hash_by_release:
                assert test_provider_source._process_release(
                    provider=test_provider,
                    repository=test_repository,
                    access_token="unittest-access-token",
                    github_release_metadata=github_release_metadata
                ) == expected_response

                if should_call_get_commit_hash:
                    mock__get_commit_hash_by_release.assert_called_once_with(
                        repository=test_repository,
                        tag_name=git_tag,
                        access_token="unittest-access-token"
                    )
                else:
                    mock__get_commit_hash_by_release.assert_not_called()

                if get_release_artifacts_metadata_should_be_called:    
                    mock__get_release_artifacts_metadata.assert_called_once_with(
                        release_id=release_id,
                        repository=test_repository,
                        access_token="unittest-access-token"
                    )
                else:
                    mock__get_release_artifacts_metadata.assert_not_called()
        finally:
            if pre_existing_provider_version:
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    conn.execute(db.provider_version.delete(
                        db.provider_version.c.provider_id==test_provider.pk
                    ))

    @pytest.mark.parametrize('result_count, expected_pages', [
        # No results
        (0, 1),
        # # Some results
        (1, 1),
        (50, 1),
        # Page count boundary tests
        (99, 1),
        (100, 2),
        (101, 2),
        (250, 3),
    ])
    def test_get_new_releases(self, result_count, expected_pages, test_provider_source, test_provider):
        """Test get_new_releases with pagination"""
        response_data_values = [
            [
                {"test_repository_itx": row_itx + (page_itx * 100)}
                for row_itx in range(0, min((result_count - (page_itx * 100)), 100))
            ]
            for page_itx in range(0, expected_pages)
        ]

        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 200
        mock_response.json.side_effect = response_data_values

        mock_release_metadata = [
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name=f"Release v{itx}", tag="v{itx}", archive_url=f"https://example.com/release/{itx}.zip",
                commit_hash=f"abcef{itx}", provider_id=f"test-provider-id-{itx}",
                release_artifacts=[]
            )
            for itx in range(result_count)
        ]
        mock__process_release = unittest.mock.MagicMock(side_effect=mock_release_metadata)

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_access_token_for_provider, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._process_release', mock__process_release):

            assert test_provider_source.get_new_releases(provider=test_provider) == mock_release_metadata

        mock__get_access_token_for_provider.assert_called_once_with(provider=test_provider)

        expected_process_release_calls = [
            unittest.mock.call(
                provider=test_provider,
                repository=test_provider.repository,
                access_token="abcdef-test-access-token",
                github_release_metadata={"test_repository_itx": itx}
            )
            for itx in range(result_count)
        ]
        mock__process_release.assert_has_calls(
            calls=expected_process_release_calls
        )

        expected_request_get_calls = [
            unittest.mock.call(
                f'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases',
                params={
                    "per_page": "100",
                    'page': str(page_itx + 1)
                },
                headers={
                    'X-GitHub-Api-Version': '2022-11-28',
                    'Accept': 'application/vnd.github+json',
                    'Authorization': 'Bearer abcdef-test-access-token'
                }
            )
            for page_itx in range(expected_pages)
        ]
        mock_requests_get.assert_has_calls(expected_request_get_calls)

    def test_get_new_releases_break_if_already_exists(self, test_provider_source, test_provider):
        """Test get_new_releases with breaking from processing new releases one is found that already exists"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"already_exists": True},
            {"new_release": True}
        ]

        mock_release_metadata = [
            # Return ProviderVersion for pre-existing release
            terrareg.provider_version_model.ProviderVersion(provider=test_provider, version="1.5.2"),
            # Return metadata for new valid release
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name=f"Release v2.0.0", tag="v2.0.0", archive_url=f"https://example.com/release/2.0.0.zip",
                commit_hash=f"abcef200", provider_id=f"test-provider-id-2.0.0",
                release_artifacts=[]
            )
        ]
        mock__process_release = unittest.mock.MagicMock(side_effect=mock_release_metadata)

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_access_token_for_provider, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._process_release', mock__process_release):

            assert test_provider_source.get_new_releases(provider=test_provider) == []

        mock__get_access_token_for_provider.assert_called_once_with(provider=test_provider)

        mock__process_release.assert_called_once_with(
            provider=test_provider,
            repository=test_provider.repository,
            access_token="abcdef-test-access-token",
            github_release_metadata={"already_exists": True}
        )

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases',
            params={
                "per_page": "100",
                'page': "1"
            },
            headers={
                'X-GitHub-Api-Version': '2022-11-28',
                'Accept': 'application/vnd.github+json',
                'Authorization': 'Bearer abcdef-test-access-token'
            }
        )

    def test_get_new_releases_skip_invalid_releases(self, test_provider_source, test_provider):
        """Test get_new_releases with skipping invalid releases and continuing with remaining releases"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 200
        mock_response.json.return_value = [
            {"already_exists": True},
            {"new_release": True}
        ]

        mock_release_metadata = [
            # Return None for invalid release
            None,
            # Return valid release
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name=f"Release v2.0.0", tag="v2.0.0", archive_url=f"https://example.com/release/2.0.0.zip",
                commit_hash=f"abcef200", provider_id=f"test-provider-id-2.0.0",
                release_artifacts=[]
            )
        ]
        mock__process_release = unittest.mock.MagicMock(side_effect=mock_release_metadata)

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_access_token_for_provider, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._process_release', mock__process_release):

            # Ensure only the release metadata is returned
            assert test_provider_source.get_new_releases(provider=test_provider) == mock_release_metadata[1:]

        mock__get_access_token_for_provider.assert_called_once_with(provider=test_provider)

        mock__process_release.assert_has_calls(calls=[
            unittest.mock.call(
                provider=test_provider,
                repository=test_provider.repository,
                access_token="abcdef-test-access-token",
                github_release_metadata={"already_exists": True}
            ),
            unittest.mock.call(
                provider=test_provider,
                repository=test_provider.repository,
                access_token="abcdef-test-access-token",
                github_release_metadata={"new_release": True}
            ),
        ])

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases',
            params={
                "per_page": "100",
                'page': "1"
            },
            headers={
                'X-GitHub-Api-Version': '2022-11-28',
                'Accept': 'application/vnd.github+json',
                'Authorization': 'Bearer abcdef-test-access-token'
            }
        )

    def test_get_new_releases_invalid_response(self, test_provider_source, test_provider):
        """Test get_new_releases with skipping invalid releases and continuing with remaining releases"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = 404
        mock_response.json.return_value = [
            {"new_release": True}
        ]

        mock_release_metadata = [
            # Return valid release
            terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name=f"Release v2.0.0", tag="v2.0.0", archive_url=f"https://example.com/release/2.0.0.zip",
                commit_hash=f"abcef200", provider_id=f"test-provider-id-2.0.0",
                release_artifacts=[]
            )
        ]
        mock__process_release = unittest.mock.MagicMock(side_effect=mock_release_metadata)

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value='abcdef-test-access-token')) as mock__get_access_token_for_provider, \
                unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._process_release', mock__process_release):

            # Ensure no metadata is returned
            assert test_provider_source.get_new_releases(provider=test_provider) == []

        mock__get_access_token_for_provider.assert_called_once_with(provider=test_provider)

        mock__process_release.assert_not_called()

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases',
            params={
                "per_page": "100",
                'page': "1"
            },
            headers={
                'X-GitHub-Api-Version': '2022-11-28',
                'Accept': 'application/vnd.github+json',
                'Authorization': 'Bearer abcdef-test-access-token'
            }
        )

    @pytest.mark.parametrize('response_code, content, get_access_token_for_provider_response, expected_result', [
        (200, b'test-content\nhere', 'unittest-access-token', b'test-content\nhere'),
        (200, None, 'unittest-access-token', None),
        (404, b'test-content\nhere', 'unittest-access-token', None),

        # No access token
        (200, b'test-content\nhere', None, None),
    ])
    def test_get_release_artifact(self, response_code, content, get_access_token_for_provider_response, expected_result, test_provider_source, test_provider):
        """Test get_release_artifact"""

        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.content = content

        artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
            name='unittest-artifact-name', provider_id='unittest-artifact-id-12345')
        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name='v5.2.3', tag='v5.2.3', archive_url='https://git.example.com/release/v5.2.3.tgz',
            commit_hash='1235abcdef', provider_id='unittest-5.2.3-release-id',
            release_artifacts=[artifact_metadata]
        )

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value=get_access_token_for_provider_response)) as mock_get_access_token_for_provider, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get:

            assert test_provider_source.get_release_artifact(
                provider=test_provider,
                release_metadata=release_metadata,
                artifact_metadata=artifact_metadata
            ) == expected_result

            mock_get_access_token_for_provider.assert_called_once_with(provider=test_provider)
            if get_access_token_for_provider_response:
                mock_requests_get.assert_called_once_with(
                    'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases/assets/unittest-artifact-id-12345',
                    headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/octet-stream', 'Authorization': 'Bearer unittest-access-token'},
                    allow_redirects=True
                )
            else:
                mock_requests_get.assert_not_called()

    @pytest.mark.parametrize('response_code, content, get_access_token_for_provider_response, expected_result', [
        (200, b'test-content\nhere', 'unittest-access-token', (b'test-content\nhere', 'some-organisation-terraform-provider-unittest-create-1235abc')),
        (200, None, 'unittest-access-token', (None, 'some-organisation-terraform-provider-unittest-create-1235abc')),
        (404, b'test-content\nhere', 'unittest-access-token', (None, 'some-organisation-terraform-provider-unittest-create-1235abc')),

        # No access token
        (200, b'test-content\nhere', None, (None, 'some-organisation-terraform-provider-unittest-create-1235abc')),
    ])
    def test_get_release_archive(self, response_code, content, get_access_token_for_provider_response, expected_result, test_provider_source, test_provider):
        """Test get_release_archive"""

        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.content = content

        artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
            name='unittest-artifact-name', provider_id='unittest-artifact-id-12345')
        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name='v5.2.3', tag='v5.2.3', archive_url='https://git.example.com/release/v5.2.3.tgz',
            commit_hash='1235abcdef', provider_id='unittest-5.2.3-release-id',
            release_artifacts=[artifact_metadata]
        )

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_access_token_for_provider',
                    unittest.mock.MagicMock(return_value=get_access_token_for_provider_response)) as mock_get_access_token_for_provider, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get:

            assert test_provider_source.get_release_archive(
                provider=test_provider,
                release_metadata=release_metadata
            ) == expected_result

            mock_get_access_token_for_provider.assert_called_once_with(provider=test_provider)
            if get_access_token_for_provider_response:
                mock_requests_get.assert_called_once_with(
                    'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/tarball/v5.2.3',
                    headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/json', 'Authorization': 'Bearer unittest-access-token'},
                    allow_redirects=True
                )
            else:
                mock_requests_get.assert_not_called()

    def test_get_public_source_url(self, test_provider_source, test_repository):
        """Test get_public_source_url"""
        assert test_provider_source.get_public_source_url(repository=test_repository) == "https://github.example-test.com/some-organisation/terraform-provider-unittest-create"

    def test_get_public_artifact_download_url(self, test_provider_source, test_provider, test_gpg_key):
        """Test get_public_artifact_download_url"""
        provider_version = terrareg.provider_version_model.ProviderVersion(provider=test_provider, version="2.3.1")
        provider_version._create_db_row(git_tag=f"v2.3.1", gpg_key=test_gpg_key)
        provider_version.publish()
        assert test_provider_source.get_public_artifact_download_url(
            provider_version=provider_version,
            artifact_name="unittest-artifact-v2.3.1.tar.gz"
        ) == "https://github.example-test.com/some-organisation/terraform-provider-unittest-create/releases/download/v2.3.1/unittest-artifact-v2.3.1.tar.gz"


    # Test Custom properties/methods

    def test_github__init__(self, test_provider_source):
        """Test github __init__ method"""
        assert test_provider_source._private_key_content is None

    def test__client_id(self, test_provider_source):
        """Test _client_id property"""
        assert test_provider_source._client_id == "unittest-github-client-id"

    def test__client_secret(self, test_provider_source):
        """Test _client_secret property"""
        assert test_provider_source._client_secret == "unittest-github-client-secret"

    def test__base_url(self, test_provider_source):
        """Test _base_url property"""
        assert test_provider_source._base_url == "https://github.example-test.com"

    def test__api_url(self, test_provider_source):
        """Test _api_url property"""
        assert test_provider_source._api_url == "https://api.github.example-test.com"

    @pytest.mark.parametrize('value, expected_result', [
        (None, False),
        (False, False),
        (True, True),
    ])
    def test_auto_generate_github_organisation_namespaces(self, value, expected_result, test_provider_source):
        """Test auto_generate_github_organisation_namespaces property"""
        test_provider_source._cache_db_row = {
            "config": terrareg.database.Database.encode_blob(json.dumps({
                "auto_generate_github_organisation_namespaces": value
            } if value else {}))
        }

        assert test_provider_source.auto_generate_github_organisation_namespaces == expected_result

    def test__private_key_path(self, test_provider_source):
        """Test _private_key_path property"""
        assert test_provider_source._private_key_path == "./unittest-path-to-private-key.pem"

    @pytest.mark.parametrize('file_exists', [
        True,
        False
    ])
    def test__private_key(self, file_exists, test_provider_source):
        """Test _private_key property"""
        with tempfile.NamedTemporaryFile(delete=True) as pem_file:
            pem_file.write("test\nprivate-key\nContent".encode('utf-8'))
            pem_file.flush()

            if file_exists:
                test_provider_source._cache_db_row = {
                    "config": terrareg.database.Database.encode_blob(json.dumps({
                        "private_key_path": pem_file.name
                    }))
                }

            if file_exists:
                assert test_provider_source._private_key == "test\nprivate-key\nContent".encode('utf-8')
                test_provider_source._private_key_content == "test\nprivate-key\nContent".encode('utf-8')
            else:
                assert test_provider_source._private_key is None
                assert test_provider_source._private_key_content is None
        
        # Ensure cached value is used
        test_provider_source._private_key_content = "cached\ntest\nprivate-key\nContent".encode('utf-8')
        assert test_provider_source._private_key == "cached\ntest\nprivate-key\nContent".encode('utf-8')

    def test_github_app_id(self, test_provider_source):
        """Test github_app_id property"""
        assert test_provider_source.github_app_id == 954956

    @pytest.mark.parametrize('client_id, client_secret, base_url, api_url, expected_result', [
        (None, None, None, None, False),
        ('test client id', 'test client secret', None, 'https://api.github.com', False),
        ('test client id', 'test client secret', 'https://github.com', None, False),
        (None, 'test client secret', 'https://github.com', 'https://api.github.com', False),
        ('test client id', None, 'https://github.com', 'https://api.github.com', False),
        ('test client id', 'test client secret', 'https://github.com', 'https://api.github.com', True),
    ])
    def test_is_enabled(self, client_id, client_secret, base_url, api_url, expected_result, test_provider_source):
        """Test is_enabled property"""
        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._client_id', client_id), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._client_secret', client_secret), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._base_url', base_url), \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._api_url', api_url):
            assert test_provider_source.is_enabled == expected_result

    def test_get_login_redirect_url(self, test_provider_source):
        """Test get_login_redirect_url"""
        assert test_provider_source.get_login_redirect_url() == "https://github.example-test.com/login/oauth/authorize?client_id=unittest-github-client-id"

    @pytest.mark.parametrize('default_installation_id, default_access_token, generate_app_installation_token_response, expected_response', [
        # no tokens
        (None, None, None, None),

        # Default installation token, but no access token returned
        ("unittest-installation-token", None, None, None),

        # Default installation token returns value
        ("unittest-installation-token", None, "unitttest-installation-access-key", "unitttest-installation-access-key"),

        # Default installation token preferred over default access token
        ("unittest-installation-token", "unittest-default-access-token", "unitttest-installation-access-key", "unitttest-installation-access-key"),
        ("unittest-installation-token", "unittest-default-access-token", None, None),
        # Fallback to default access token
        (None, "unittest-default-access-token", None, "unittest-default-access-token"),

    ])
    def test__get_default_access_token(self, default_installation_id, default_access_token, generate_app_installation_token_response, expected_response, test_provider_source):
        """Test _get_default_access_token method"""
        test_provider_source._cache_db_row = {
            "config": terrareg.database.Database.encode_blob(json.dumps({
                "default_installation_id": default_installation_id,
                "default_access_token": default_access_token
            }))
        }
        with unittest.mock.patch(
                'terrareg.provider_source.GithubProviderSource.generate_app_installation_token',
                unittest.mock.MagicMock(return_value=generate_app_installation_token_response)) as mock_generate_app_installation_token:
            assert test_provider_source._get_default_access_token() == expected_response

        if default_installation_id:
            mock_generate_app_installation_token.assert_called_once_with(default_installation_id)
        else:
            mock_generate_app_installation_token.assert_not_called()

    @pytest.mark.parametrize('access_token, expect_call, response_code, response_data, expected_response', [
        ('abcdefg', True, 200, {'login': 'testusername'}, 'testusername'),
        ('abcdefg', True, 200, {'somethingelse': 'sometingelse'}, None),
        ('abcdefg', True, 400, {}, None),
        (None, False, 200, {'login': 'testusername'}, None)
    ])
    def test_get_username(self, access_token, expect_call, response_code, response_data, expected_response, test_provider_source):
        """Test get_username method"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.json = unittest.mock.MagicMock(return_value=response_data)
        mock_request_get = unittest.mock.MagicMock(return_value=mock_response)
        with unittest.mock.patch("terrareg.provider_source.github.requests.get", mock_request_get):
            assert test_provider_source.get_username(access_token) == expected_response

        if expect_call:
            mock_request_get.assert_called_once_with(
                'https://api.github.example-test.com/user',
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )
        else:
            mock_request_get.assert_not_called()

    @pytest.mark.parametrize('access_token, expect_call, response_code, response_data, expected_response', [
        ('abcdefg', True, 200, [], []),
        ('abcdefg', True, 200, [
            {
                "organization": {"login": "valid1"},
                "state": "active",
                "role": "admin"
            },
            {
                "organization": {"login": "invalid"},
            },
            {
                "organization": {"login": "multiplematch"},
                "state": "active",
                "role": "admin"
            }
        ], ["valid1", "multiplematch"]),
        ('abcdefg', True, 200, [
            {
                "organization": {"login": "invalidstate"},
                "state": "bad",
                "role": "admin"
            }
        ], []),
        ('abcdefg', True, 200, [
            {
                "organization": {"login": "invalidstate"},
                "role": "admin"
            }
        ], []),
        ('abcdefg', True, 200, [
            {
                "organization": {"login": "invalidstate"},
                "state": "active",
                "role": "badrole"
            }
        ], []),
        ('abcdefg', True, 200, [
            {
                "organization": {"login": "invalidstate"},
                "state": "bad"
            }
        ], []),
        ('abcdefg', True, 200, [
            {
                "state": "bad",
                "role": "admin"
            }
        ], []),
        ('abcdefg', True, 400, {}, []),
        (None, False, 200, None, [])
    ])
    def test_get_user_organisations(self, access_token, expect_call, response_code, response_data, expected_response, test_provider_source):
        """Test get_user_organisations"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.json = unittest.mock.MagicMock(return_value=response_data)
        mock_request_get = unittest.mock.MagicMock(return_value=mock_response)
        with unittest.mock.patch("terrareg.provider_source.github.requests.get", mock_request_get):
            assert test_provider_source.get_user_organisations(access_token) == expected_response

        if expect_call:
            mock_request_get.assert_called_once_with(
                'https://api.github.example-test.com/user/memberships/orgs',
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )
        else:
            mock_request_get.assert_not_called()

    @pytest.mark.parametrize('repository_metadata, expect_create, expected_row', [
        # Working example
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "owner": {"login": "MatthewJohn", "avatar_url": "https://cdn.example.com/MatthewJohn.png"},
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         True,
         {"provider_id": "unittest-github-repo-id-001", "name": "terraform-provider-laptop",
          "description": terrareg.database.Database.encode_blob("Terraform Provider to create a laptop. Don't ask how!"),
          "owner": "MatthewJohn",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git",
          "logo_url": "https://cdn.example.com/MatthewJohn.png"}),

        # Missing optional arguments
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "owner": {"login": "MatthewJohn", "avatar_url": "https://cdn.example.com/MatthewJohn.png"},
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         True,
         {"provider_id": "unittest-github-repo-id-001", "name": "terraform-provider-laptop",
          "description": terrareg.database.Database.encode_blob(""),
          "owner": "MatthewJohn",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git",
          "logo_url": "https://cdn.example.com/MatthewJohn.png"}),
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "owner": {"login": "MatthewJohn"},
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         True,
         {"provider_id": "unittest-github-repo-id-001", "name": "terraform-provider-laptop",
          "description": terrareg.database.Database.encode_blob("Terraform Provider to create a laptop. Don't ask how!"),
          "owner": "MatthewJohn",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git",
          "logo_url": None}),

        # Missing required attributes
        ## Missing ID
        ({"name": "terraform-provider-laptop",
          "owner": {"login": "MatthewJohn", "avatar_url": "https://cdn.example.com/MatthewJohn.png"},
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         False, {}),
        ## Missing name
        ({"id": "unittest-github-repo-id-001",
          "owner": {"login": "MatthewJohn", "avatar_url": "https://cdn.example.com/MatthewJohn.png"},
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         False, {}),
         ## Missing owner
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "owner": {},
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         False, {}),
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "description": "Terraform Provider to create a laptop. Don't ask how!",
          "clone_url": "https://example.github.com/clone/MatthewJohn/terraform-provider-laptop.git"},
         False, {}),
        ## Missing clone URL
        ({"id": "unittest-github-repo-id-001",
          "name": "terraform-provider-laptop",
          "owner": {"login": "MatthewJohn", "avatar_url": "https://cdn.example.com/MatthewJohn.png"},
          "description": "Terraform Provider to create a laptop. Don't ask how!"},
         False, {}),
    ])
    def test__add_repository(self, repository_metadata, expect_create, expected_row, test_provider_source) -> None:
        """Test _add_repository method"""
        # Delete any pre-existing repositories
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.repository.delete())

        test_provider_source._add_repository(repository_metadata=repository_metadata)

        with db.get_connection() as conn:
            row = conn.execute(db.repository.select()).first()

            if expect_create:
                assert row is not None
                row = dict(row)
                del row["id"]

                assert row["provider_source_name"] == test_provider_source.name
                del row["provider_source_name"]

                assert row == expected_row
            else:
                assert row is None

    @pytest.mark.parametrize('status_code, response_json, expected_response', [
        # Valid response
        (200, {"object": {"sha": "abcdefunittestsha"}}, "abcdefunittestsha"),
        # Missing sha/object
        (200, {"object": {}}, None),
        (200, {}, None),
        # Invalid response code
        (404, {"object": {"sha": "abcdefunittestsha"}}, None)
    ])
    def test__get_commit_hash_by_release(self, status_code, response_json, expected_response, test_provider_source):
        """Test _get_commit_hash_by_release method"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json = unittest.mock.MagicMock(return_value=response_json)

        # Delete any pre-existing repositories
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.repository.delete())

        repository = terrareg.repository_model.Repository.create(
            provider_source=test_provider_source,
            provider_id="unittest-repository-id-001",
            name="terraform-provider-unittest",
            description="Test",
            owner="MatthewJohn",
            clone_url="https://git.example.com/MatthewJohn/terraform-provider-unittest",
            logo_url="https://example.com/logo.png"
        )

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get:
            assert test_provider_source._get_commit_hash_by_release(
                repository=repository,
                tag_name="v5.2.1",
                access_token="unittest-access-token"
            ) == expected_response

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/repos/MatthewJohn/terraform-provider-unittest/git/ref/tags/v5.2.1',
            headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/vnd.github+json', 'Authorization': 'Bearer unittest-access-token'}
        )

    @pytest.mark.parametrize(('get_github_app_installation_id_response, generate_app_installation_token_response, '
                              'generate_app_installation_token_should_be_called, get_default_access_token_response, '
                              'get_default_access_token_should_be_called, use_default_provider_source_auth, '
                              'should_raise_exception, expected_response'), [
        # Obtains installation ID from namespace and able to obtain access token
        ('1234-namespace-installation-id', 'unittest-namespace-installation-access-token', True,
         None, False, False, False, 'unittest-namespace-installation-access-token'),
        # Obtains installation ID from namespace and returns None access token
        ('1234-namespace-installation-id', None, True,
         None, False, False, False, None),
        # No namespace installation ID and use_default_provider_source_auth is disabled
        (None, 'abcdefg', False,
         None, False, False, False, None),
        # No namespace installation ID and use_default_provider_source_auth is enabled, returning valid default access token
        (None, 'abcdefg', False,
         "default-auth-access-token-unittest-123", True, True, False, "default-auth-access-token-unittest-123"),
        # No namespace installation ID and use_default_provider_source_auth is enabled, and no default access token,
        # raising an exception
        (None, 'abcdefg', False,
         None, True, True, True, None),
    ])
    def test__get_access_token_for_provider(self, get_github_app_installation_id_response, generate_app_installation_token_response,
                                            generate_app_installation_token_should_be_called, get_default_access_token_response,
                                            get_default_access_token_should_be_called, use_default_provider_source_auth, should_raise_exception,
                                            expected_response,
                                            test_provider_source, test_provider, test_namespace):
        """Test _get_access_token_for_provider"""

        with unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource.get_github_app_installation_id',
                    unittest.mock.MagicMock(return_value=get_github_app_installation_id_response)) as mock_get_github_app_installation_id, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource.generate_app_installation_token',
                    unittest.mock.MagicMock(return_value=generate_app_installation_token_response)) as mock_generate_app_installation_token, \
                unittest.mock.patch(
                    'terrareg.provider_source.github.GithubProviderSource._get_default_access_token',
                    unittest.mock.MagicMock(return_value=get_default_access_token_response)) as mock_get_default_access_token, \
                unittest.mock.patch(
                    'terrareg.provider_model.Provider.use_default_provider_source_auth', use_default_provider_source_auth):

            if should_raise_exception:
                with pytest.raises(terrareg.errors.ProviderSourceDefaultAccessTokenNotConfiguredError):
                    test_provider_source._get_access_token_for_provider(provider=test_provider)
            else:
                assert test_provider_source._get_access_token_for_provider(provider=test_provider) == expected_response

        mock_get_github_app_installation_id.assert_called_once_with(namespace=test_namespace)

        if generate_app_installation_token_should_be_called:
            mock_generate_app_installation_token.assert_called_once_with(installation_id='1234-namespace-installation-id')
        else:
            mock_generate_app_installation_token.assert_not_called()

        if get_default_access_token_should_be_called:
            mock_get_default_access_token.assert_called_once_with()
        else:
            mock_get_default_access_token.assert_not_called()

    @pytest.mark.parametrize('status_code, response_data, expected_response', [
        # Invalid response code
        (500, [{"name": "test-artifact", "id": "test-artifact-id"}], []),
        # No releases
        (200, [], []),
        # Single artifact
        (200, [{"name": "test-artifact", "id": "test-artifact-id"}], [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="test-artifact", provider_id="test-artifact-id")
        ]),
        # Multiple artifacts
        (200, [{"name": "test-artifact1", "id": "test-artifact-id1"}, {"name": "test-artifact2", "id": "test-artifact-id2"}], [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="test-artifact1", provider_id="test-artifact-id1"),
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="test-artifact2", provider_id="test-artifact-id2")
        ]),

        # Artifacts missing properties
        (200, [{"name": "test-artifact"}, {"id": "test-artifact-id"}, {"name": None, "id": None}], []),
    ])
    def test__get_release_artifacts_metadata(self, status_code, response_data, expected_response, test_provider_source, test_repository):
        """Test _get_release_artifacts_metadata"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json.return_value = response_data

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get:
            res = test_provider_source._get_release_artifacts_metadata(
                repository=test_repository,
                release_id=173729,
                access_token='unit-test-acccess-token-get_release_artifacts'
            )

            assert isinstance(res, list)
            assert len(res) == len(expected_response)
            for itx, res_itx in enumerate(res):
                assert isinstance(res_itx, terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata)
                assert res_itx.name == expected_response[itx].name
                assert res_itx.provider_id == expected_response[itx].provider_id

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/repos/some-organisation/terraform-provider-unittest-create/releases/173729/assets',
            params={'per_page': '100', 'page': '1'},
            headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/vnd.github+json', 'Authorization': 'Bearer unit-test-acccess-token-get_release_artifacts'}
        )

    @pytest.mark.parametrize('installation_id, status_code, json_res, should_raise, expected_repsonse', [
        ('unittest-installation-id1', 201, {'token': 'unittest-access-token'}, False, 'unittest-access-token'),

        # No access token
        ('unittest-installation-id2', 201, {'token': None}, False, None),
        ('unittest-installation-id3', 201, {}, False, None),
        # Invalid response code
        ('unittest-installation-id4', 403, {}, True, None),

        # No installation ID
        (None, 201, {'token': 'unittest-access-token'}, False, None),
    ])
    def test_generate_app_installation_token(self, installation_id, status_code, json_res, should_raise, expected_repsonse, test_provider_source):
        """Test generate_app_installation_token"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json.return_value = json_res

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._generate_jwt', unittest.mock.MagicMock(return_value='unittest JWT Auth')), \
                unittest.mock.patch('terrareg.provider_source.github.requests.post', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_post:

            if should_raise:
                with pytest.raises(terrareg.errors.UnableToGenerateGithubInstallationAccessTokenError):
                    test_provider_source.generate_app_installation_token(installation_id=installation_id)
            else:
                assert test_provider_source.generate_app_installation_token(installation_id=installation_id) == expected_repsonse
        
        if installation_id:
            mock_requests_post.assert_called_once_with(
                f'https://api.github.example-test.com/app/installations/{installation_id}/access_tokens',
                headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/vnd.github+json', 'Authorization': 'Bearer unittest JWT Auth'}
            )
        else:
            mock_requests_post.assert_not_called()

    @pytest.mark.parametrize('private_key, expected_result', {
        (None, None),
        (TEST_GITHUB_PRIVATE_KEY.encode('utf-8'),
         ("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2OTkyNTU4OTEsImV4cCI6MTY5OTI1NjQ5MSwiaXNzIjo5NTQ5NTZ9."
          "C2SGwXuObOpi-4jE59fjkFSIAXkCK0tm68JQgi6mbU_SILlYksvb5E106X2B3qrJIQZNmndTYo2fGT_2RwY_kqKHl0xYw0p0Pdqqny"
          "T8jetEK6IiQggyEm-DxlzGE7t3oyuFjq25zbrEjnmlIwnJapboU9V7amyH8WE7dl3MNo4"))
    })
    def test__generate_jwt(self, private_key, expected_result, test_provider_source):
        """Test _generate_jwt"""
        with unittest.mock.patch("terrareg.provider_source.github.GithubProviderSource._private_key", private_key), \
            unittest.mock.patch('terrareg.provider_source.github.time.time', unittest.mock.MagicMock(return_value=1699255891.0833983)):

            assert test_provider_source._generate_jwt() == expected_result

    @pytest.mark.parametrize('status_code, response_data, expected_response, should_raise', [
        (200, {'id': 'abcd-1234', 'name': 'test-org'}, {'id': 'abcd-1234', 'name': 'test-org'}, False),
        (403, {'id': 'abcd-1234', 'name': 'test-org'}, None, True),
    ])
    def test__get_app_metadata(self, status_code, response_data, expected_response, should_raise, test_provider_source):
        """Test _get_app_metadata"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json.return_value = response_data

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._generate_jwt', unittest.mock.MagicMock(return_value='unittest-mock-jwt')):

            if should_raise:
                with pytest.raises(terrareg.errors.InvalidGithubAppMetadataError):
                    test_provider_source._get_app_metadata()
            else:
                assert test_provider_source._get_app_metadata() == expected_response

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/app',
            headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/json', 'Authorization': 'Bearer unittest-mock-jwt'}
        )

    def test_get_app_installation_url(self, test_provider_source):
        """Test get_app_installation_url"""
        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._get_app_metadata',
                    unittest.mock.MagicMock(return_value={'html_url': 'https://example.github.com/apps/my-special-app'})):
            assert test_provider_source.get_app_installation_url() == "https://example.github.com/apps/my-special-app/installations/new"

    @pytest.mark.parametrize('namespace_type, expected_url, status_code, response_data, expected_return_value', [
        # Valid responses
        (terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION, 'https://api.github.example-test.com/orgs/some-organisation/installation',
         200, {"id": "unit-test-installation-id"}, "unit-test-installation-id"),
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, 'https://api.github.example-test.com/users/some-organisation/installation',
         200, {"id": "unit-test-installation-id"}, "unit-test-installation-id"),

        # Invalid status codes
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, 'https://api.github.example-test.com/users/some-organisation/installation',
         404, {"id": "unit-test-installation-id"}, None),
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, 'https://api.github.example-test.com/users/some-organisation/installation',
         500, {"id": "unit-test-installation-id"}, None),

        # Invalid response data
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, 'https://api.github.example-test.com/users/some-organisation/installation',
         200, {"id": None}, None),
        (terrareg.namespace_type.NamespaceType.GITHUB_USER, 'https://api.github.example-test.com/users/some-organisation/installation',
         200, {}, None),

        # Invalid namespace type
        (terrareg.namespace_type.NamespaceType.NONE, None,
         200, {}, None),
    ])
    def test_get_github_app_installation_id(self, namespace_type, expected_url, status_code, response_data, expected_return_value, test_namespace, test_provider_source):
        """Test get_github_app_installation_id"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json.return_value = response_data

        test_namespace.update_attributes(namespace_type=namespace_type)

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource._generate_jwt', unittest.mock.MagicMock(return_value='unittest-mock-jwt')):
            assert test_provider_source.get_github_app_installation_id(
                namespace=test_namespace
            ) == expected_return_value

        if namespace_type is not terrareg.namespace_type.NamespaceType.NONE:
            mock_requests_get.assert_called_once_with(
                expected_url,
                headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/json', 'Authorization': 'Bearer unittest-mock-jwt'}
            )
        else:
            mock_requests_get.assert_not_called()

    @pytest.mark.parametrize('status_code, response_data, expected_value', [
        (200, {"type": "User"}, terrareg.namespace_type.NamespaceType.GITHUB_USER),
        (200, {"type": "Organization"}, terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION),

        # Invalid response code
        (404, {"type": "Organization"}, None),
        # Invalid type
        (200, {"type": "SomethingElse"}, None),
        # No type
        (200, {"type": None}, None),
        (200, {}, None),
    ])
    def test__is_entity_org_or_user(self, status_code, response_data, expected_value, test_provider_source):
        """Test _is_entity_org_or_user"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = status_code
        mock_response.json.return_value = response_data

        with unittest.mock.patch('terrareg.provider_source.github.requests.get', unittest.mock.MagicMock(return_value=mock_response)) as mock_requests_get:
            assert test_provider_source._is_entity_org_or_user(
                identity="unit-test-identity-name",
                access_token="unittest-access-token"
            ) == expected_value

        mock_requests_get.assert_called_once_with(
            'https://api.github.example-test.com/users/unit-test-identity-name',
            headers={'X-GitHub-Api-Version': '2022-11-28', 'Accept': 'application/json', 'Authorization': 'Bearer unittest-access-token'}
        )
