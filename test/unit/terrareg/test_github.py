
import unittest.mock

import pytest

from test.unit.terrareg import TerraregUnitTest
import terrareg.github


class TestGithub(TerraregUnitTest):
    """Test Github class"""

    @pytest.mark.parametrize('github_app_client_id, github_app_client_secret, github_url, expected_result', [
        (None, None, None, False),
        ('test client id', 'test client secret', None, False),
        (None, 'test client secret', 'https://github.com', False),
        ('test client id', None, 'https://github.com', False),
        ('test client id', 'test client secret', 'https://github.com', True),
    ])
    def test_is_enabled(self, github_app_client_id, github_app_client_secret, github_url, expected_result):
        """Test is_enabled"""
        with unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_ID', github_app_client_id), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_SECRET', github_app_client_secret), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_URL', github_url):
            assert terrareg.github.Github().is_enabled() is expected_result

    def test_get_login_redirect_url(self):
        """Test get_login_redirect_url"""
        with unittest.mock.patch('terrareg.config.Config.GITHUB_URL', 'https://example-github.com'), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_ID', 'test-github-client-id'):
            assert terrareg.github.Github().get_login_redirect_url() == "https://example-github.com/login/oauth/authorize?client_id=test-github-client-id"


    @pytest.mark.parametrize('code, response_code, response, should_request_be_made, expected_result', [
        ('1234', 200, "access_token=reallysecretaccesstoken&expiry=abcdef", True, "reallysecretaccesstoken"),
        ('1234', 200, "expiry=abcdef", True, None),
        ('1234', 500, "expiry=abcdef", True, None),
        ("", 200, "expiry=abcdef", False, None),
        (None, 200, "expiry=abcdef", False, None),
    ])
    def test_get_access_token(self, code, response_code, response, should_request_be_made, expected_result):
        """Test get_access_token"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.text = response
        mock_request_post = unittest.mock.MagicMock(return_value=mock_response)
        with unittest.mock.patch("terrareg.github.requests.post", mock_request_post), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_ID', "mock-github-client-id"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_SECRET', "mock-github-client-secret"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_URL', "https://example-github.com"):
            assert terrareg.github.Github.get_access_token(code) == expected_result

        if should_request_be_made:
            mock_request_post.assert_called_once_with(
                'https://example-github.com/login/oauth/access_token',
                data={'client_id': 'mock-github-client-id', 'client_secret': 'mock-github-client-secret', 'code': code}
            )
        else:
            mock_request_post.assert_not_called()


    @pytest.mark.parametrize('access_token, expect_call, response_code, response_data, expected_response', [
        ('abcdefg', True, 200, {'login': 'testusername'}, 'testusername'),
        ('abcdefg', True, 200, {'somethingelse': 'sometingelse'}, None),
        ('abcdefg', True, 400, {}, None),
        (None, False, 200, {'login': 'testusername'}, None)
    ])
    def test_get_username(self, access_token, expect_call, response_code, response_data, expected_response):
        """Test get_username"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.json = unittest.mock.MagicMock(return_value=response_data)
        mock_request_get = unittest.mock.MagicMock(return_value=mock_response)
        with unittest.mock.patch("terrareg.github.requests.get", mock_request_get), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_ID', "mock-github-client-id"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_SECRET', "mock-github-client-secret"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_API_URL', "https://api.example-github.com"):
            assert terrareg.github.Github.get_username(access_token) == expected_response

        if expect_call:
            mock_request_get.assert_called_once_with(
                'https://api.example-github.com/user',
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
    def test_get_user_organisations(self, access_token, expect_call, response_code, response_data, expected_response):
        """Test get_user_organisations"""
        mock_response = unittest.mock.MagicMock()
        mock_response.status_code = response_code
        mock_response.json = unittest.mock.MagicMock(return_value=response_data)
        mock_request_get = unittest.mock.MagicMock(return_value=mock_response)
        with unittest.mock.patch("terrareg.github.requests.get", mock_request_get), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_ID', "mock-github-client-id"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_APP_CLIENT_SECRET', "mock-github-client-secret"), \
                unittest.mock.patch('terrareg.config.Config.GITHUB_API_URL', "https://api.example-github.com"):
            assert terrareg.github.Github.get_user_organisations(access_token) == expected_response

        if expect_call:
            mock_request_get.assert_called_once_with(
                'https://api.example-github.com/user/memberships/orgs',
                headers={
                    "X-GitHub-Api-Version": "2022-11-28",
                    "Accept": "application/vnd.github+json",
                    "Authorization": f"Bearer {access_token}"
                }
            )
        else:
            mock_request_get.assert_not_called()