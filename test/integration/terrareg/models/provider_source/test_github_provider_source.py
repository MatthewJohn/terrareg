
# Temp import
from typing import Union, List

import unittest.mock

import pytest

from .base_provider_source_tests import BaseProviderSourceTests
from . import test_provider_source
import terrareg.provider_source
import terrareg.errors
import terrareg.namespace_type
from test.integration.terrareg.fixtures import (
    test_namespace
)


class TestGithubProviderSource(BaseProviderSourceTests):

    _CLASS = terrareg.provider_source.GithubProviderSource
    ADDITIONAL_CONFIG = {
        "login_button_text": "Unit Test Github Login",
        "base_url": "https://github.example-test.com",
        "api_url": "https://api.github.example-test.com",
        "client_id": "unittest-github-client-id",
        "client_secret": "unittest-github-client-secret"
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
        # # Some results
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
