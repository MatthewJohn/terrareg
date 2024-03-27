
from unittest import mock

from test.unit.terrareg import TerraregUnitTest, setup_test_data, mock_models
from test import client
import terrareg.models


class TestApiTerraregGitProviders(TerraregUnitTest):
    """Test TestApiTerraregGitProviders resource."""

    def test_with_no_git_providers_configured(self, mock_models, client):
        """Test endpoint when no git providers are configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == []

    @setup_test_data()
    def test_with_git_providers_configured(self, mock_models, client):
        """Test endpoint with git providers configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == [
            {'id': 1, 'name': 'testgitprovider', 'git_path_template': None},
            {'id': 2, 'name': 'second-git-provider', 'git_path_template': '/modules/{module}/'}
        ]

    def test_unauthenticated(self, client, mock_models):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/terrareg/git_providers')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
