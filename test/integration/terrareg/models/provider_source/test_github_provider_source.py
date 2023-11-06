
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
from test.integration.terrareg.fixtures import (
    test_namespace, test_provider_category
)


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
