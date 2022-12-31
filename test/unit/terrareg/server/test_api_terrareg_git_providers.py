
from unittest import mock

from test.unit.terrareg import TerraregUnitTest, setup_test_data, mocked_git_provider_fixture
from test import client
import terrareg.models


class TestApiTerraregGitProviders(TerraregUnitTest):
    """Test TestApiTerraregGitProviders resource."""

    def test_with_no_git_providers_configured(self, mocked_git_provider_fixture, client):
        """Test endpoint when no git providers are configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == []

    @setup_test_data()
    def test_with_git_providers_configured(self, mocked_git_provider_fixture, client):
        """Test endpoint with git providers configured."""
        res = client.get('/v1/terrareg/git_providers')
        assert res.status_code == 200
        assert res.json == [
            {'id': 1, 'name': 'testgitprovider'},
            {'id': 2, 'name': 'second-git-provider'}
        ]
