
from datetime import datetime
import unittest.mock

import pytest

import terrareg.namespace_type
from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search
from test.integration.terrareg.fixtures import test_github_provider_source
from terrareg.auth.admin_session_auth_method import AdminSessionAuthMethod
import terrareg.auth.github_auth_method
import terrareg.models


class TestApiGithubRefreshNamespace(TerraregIntegrationTest):
    """Test GithubRefreshNamespace endpoint"""

    def test_invalid_provider_source(self, client, test_github_provider_source):
        """Test dndpoint with invalid provider source"""
        res = client.post("/doesnotexist/refresh-namespace", json={"namespace": "initial-providers", "csrf_token": "test-token"})
        assert res.status_code == 404

    def test_unauthenticated(self, client, test_github_provider_source):
        """Test Endpoint without authentication"""
        self._get_current_auth_method_mock.stop()
        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.refresh_namespace_repositories', unittest.mock.MagicMock()) as mock_refresh_namespace_repositories, \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf:
            res = client.post("/test-github-provider/refresh-namespace", json={"namespace": "initial-providers", "csrf_token": "test-token"})
            assert res.status_code == 401
            assert res.json == {
                'message': 'The server could not verify that you are authorized to access the '
                            'URL requested. You either supplied the wrong credentials (e.g. a '
                            "bad password), or your browser doesn't understand how to supply "
                            'the credentials required.',
            }

            mock_refresh_namespace_repositories.assert_not_called()

    def test_invalid_namespace(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with different auth"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
            unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
            unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.refresh_namespace_repositories', unittest.mock.MagicMock()) as mock_refresh_namespace_repositories:

            res = client.post("/test-github-provider/refresh-namespace", json={"namespace": "does-not-exist", "csrf_token": "test-token"})
            assert res.status_code == 404
            assert res.json == {'errors': ['Not Found']}
            mock_check_csrf.assert_called_once_with('test-token')
            mock_refresh_namespace_repositories.assert_not_called()

    def test_authenticated_with_another_auth(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with different auth"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
            unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
            unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.refresh_namespace_repositories', unittest.mock.MagicMock()) as mock_refresh_namespace_repositories:

            res = client.post("/test-github-provider/refresh-namespace", json={"namespace": "initial-providers", "csrf_token": "test-token"})
            assert res.status_code == 200
            assert res.json == []
            mock_check_csrf.assert_called_once_with('test-token')
            mock_refresh_namespace_repositories.assert_called_once_with(namespace=terrareg.models.Namespace.get("initial-providers"))

    def test_github_authenticated(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with github"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.csrf.check_csrf_token', return_value=True) as mock_check_csrf, \
                unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.refresh_namespace_repositories', unittest.mock.MagicMock()) as mock_refresh_namespace_repositories, \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_current_instance',
                                    unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())):

            res = client.post("/test-github-provider/refresh-namespace", json={"namespace": "initial-providers", "csrf_token": "test-token"})
            assert res.status_code == 200
            assert res.json == []
            mock_refresh_namespace_repositories.assert_called_once_with(namespace=terrareg.models.Namespace.get("initial-providers"))
