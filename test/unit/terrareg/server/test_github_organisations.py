
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


class TestApiGithubOrganisations(TerraregIntegrationTest):
    """Test GithubOrganisations endpoint"""

    def test_invalid_provider_source(self, client, test_github_provider_source):
        """Test dndpoint with invalid provider source"""
        res = client.get("/doesnotexist/organizations")
        assert res.status_code == 404

    def test_unauthenticated(self, client, test_github_provider_source):
        """Test Endpoint without authentication"""
        res = client.get("/test-github-provider/organizations")
        assert res.status_code == 200
        assert res.json == []

    def test_authenticated_with_another_auth(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with different auth"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.get("/test-github-provider/organizations")
            assert res.status_code == 200
            assert res.json == []

    def test_authenticated(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with github"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_username', unittest.mock.MagicMock(return_value='unittestusername')), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_github_organisations', unittest.mock.MagicMock(return_value={
                    "moduleextraction": terrareg.namespace_type.NamespaceType.GITHUB_USER,
                    "initial-providers": terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION,
                    "does-not-exist": terrareg.namespace_type.NamespaceType.GITHUB_ORGANISATION
                })), \
                unittest.mock.patch('terrareg.auth.github_auth_method.GithubAuthMethod.get_current_instance',
                                    unittest.mock.MagicMock(return_value=terrareg.auth.github_auth_method.GithubAuthMethod())):

            res = client.get("/test-github-provider/organizations")
            assert res.status_code == 200
            assert res.json == [
                {
                    'admin': True,
                    'can_publish_providers': False,
                    'name': 'moduleextraction',
                    'type': 'organization'
                },
                {
                    'admin': True,
                    'can_publish_providers': True,
                    'name': 'initial-providers',
                    'type': 'organization'
                }
            ]
