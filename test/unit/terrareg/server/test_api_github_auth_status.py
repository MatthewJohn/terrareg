# class GithubAuthStatus(ErrorCatchingResource):
#     """Interface to provide details about current authentication status with Github"""

#     def _get(self, provider_source: str):
#         """Provide authentication status."""

#         # Obtain provider source
#         provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
#         provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
#         if not provider_source_obj:
#             return self._get_404_response()

#         github_authenticated = False
#         username = None
#         if auth_method := terrareg.auth.GithubAuthMethod.get_current_instance():
#             github_authenticated = True
#             username = auth_method.get_username()

#         return {
#             "auth": github_authenticated,
#             "username": username
#         }


from datetime import datetime
import unittest.mock

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search
from test.integration.terrareg.fixtures import test_github_provider_source
from terrareg.auth.admin_session_auth_method import AdminSessionAuthMethod
import terrareg.auth.github_auth_method


class TestApiGithubAuthStatus(TerraregIntegrationTest):
    """Test GithubAuthStatus endpoint"""

    def test_unauthenticated(self, client, test_github_provider_source):
        """Test Endpoint without authentication"""
        res = client.get("/test-github-provider/auth/status")
        assert res.status_code == 200
        assert res.json == {'auth': False, 'username': None}

    def test_authenticated_with_another_auth(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with different auth"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod)

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.get("/test-github-provider/auth/status")
            assert res.status_code == 200
            assert res.json == {'auth': False, 'username': None}

    def test_authenticated(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with github"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod)

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_username', unittest.mock.MagicMock(return_value='unittestusername')), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_current_instance',
                                    unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())):

            res = client.get("/test-github-provider/auth/status")
            assert res.status_code == 200
            assert res.json == {'auth': True, 'username': "unittestusername"}



