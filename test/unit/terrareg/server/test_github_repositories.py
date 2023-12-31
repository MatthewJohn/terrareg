
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


class TestApiGithubRepositories(TerraregIntegrationTest):
    """Test GithubRepositories endpoint"""

    def test_invalid_provider_source(self, client, test_github_provider_source):
        """Test dndpoint with invalid provider source"""
        res = client.get("/doesnotexist/repositories")
        assert res.status_code == 404

    def test_unauthenticated(self, client, test_github_provider_source):
        """Test Endpoint without authentication"""
        self._get_current_auth_method_mock.stop()
        res = client.get("/test-github-provider/repositories")
        assert res.status_code == 200
        assert res.json == []

    def test_authenticated_with_another_auth(self, client, test_github_provider_source):
        """Test Endpoint whilst authenticated with different auth"""
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=AdminSessionAuthMethod())

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):

            res = client.get("/test-github-provider/repositories")
            assert res.status_code == 200
            assert res.json == [
                {
                    'full_name': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result',
                    'id': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result',
                    'kind': 'provider',
                    'owner_login': 'modulesearch-trusted',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-second-result',
                    'id': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-second-result',
                    'kind': 'provider',
                    'owner_login': 'modulesearch-trusted',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result-multiversion',
                    'id': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result-multiversion',
                    'kind': 'provider',
                    'owner_login': 'modulesearch-trusted',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result-no-versions',
                    'id': 'modulesearch-trusted/terraform-provider-mixedsearch-trusted-result-no-versions',
                    'kind': 'provider',
                    'owner_login': 'modulesearch-trusted',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-test-initial',
                    'id': 'initial-providers/terraform-provider-test-initial',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-to-delete',
                    'id': 'initial-providers/terraform-provider-to-delete',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-update-attributes',
                    'id': 'initial-providers/terraform-provider-update-attributes',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-empty-provider-publish',
                    'id': 'initial-providers/terraform-provider-empty-provider-publish',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-multiple-versions',
                    'id': 'initial-providers/terraform-provider-multiple-versions',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'second-provider-namespace/terraform-provider-multiple-versions',
                    'id': 'second-provider-namespace/terraform-provider-multiple-versions',
                    'kind': 'provider',
                    'owner_login': 'second-provider-namespace',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'providersearch/terraform-provider-contributedprovider-oneversion',
                    'id': 'providersearch-namespace/terraform-provider-contributedprovider-oneversion',
                    'kind': 'provider',
                    'owner_login': 'providersearch',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'providersearch/terraform-provider-contributedprovider-multiversion',
                    'id': 'providersearch-namespace/terraform-provider-contributedprovider-multiversion',
                    'kind': 'provider',
                    'owner_login': 'providersearch',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'contributed-providersearch/terraform-provider-mixedsearch-result',
                    'id': 'contributed-providersearch-namespace/terraform-provider-mixedsearch-result',
                    'kind': 'provider',
                    'owner_login': 'contributed-providersearch',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'contributed-providersearch/terraform-provider-mixedsearch-result-multiversion',
                    'id': 'contributed-providersearch-namespace/terraform-provider-mixedsearch-result-multiversion',
                    'kind': 'provider',
                    'owner_login': 'contributed-providersearch',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'contributed-providersearch/terraform-provider-mixedsearch-result-no-version',
                    'id': 'contributed-providersearch-namespace/terraform-provider-mixedsearch-result-no-version',
                    'kind': 'provider',
                    'owner_login': 'contributed-providersearch',
                    'owner_type': 'owner',
                    'published_id': None
                },
            ]

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

            res = client.get("/test-github-provider/repositories")
            assert res.status_code == 200
            assert res.json == [
                {
                    'full_name': 'initial-providers/terraform-provider-test-initial',
                    'id': 'initial-providers/terraform-provider-test-initial',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-to-delete',
                    'id': 'initial-providers/terraform-provider-to-delete',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-update-attributes',
                    'id': 'initial-providers/terraform-provider-update-attributes',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-empty-provider-publish',
                    'id': 'initial-providers/terraform-provider-empty-provider-publish',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
                {
                    'full_name': 'initial-providers/terraform-provider-multiple-versions',
                    'id': 'initial-providers/terraform-provider-multiple-versions',
                    'kind': 'provider',
                    'owner_login': 'initial-providers',
                    'owner_type': 'owner',
                    'published_id': None
                },
            ]
