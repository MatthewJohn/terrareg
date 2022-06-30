
from unittest import mock
from test.unit.terrareg import MockGitProvider, TerraregUnitTest, setup_test_data
from test import client


class TestApiTerraregHealth(TerraregUnitTest):
    """Test ApiTerraregHealth resource."""

    def test_with_no_git_providers_configured(self, client):
        """Test endpoint when no git providers are configured."""
        with mock.patch('terrareg.models.GitProvider.get_all') as mocked_git_providers_get_all:
            mocked_git_providers_get_all.return_value = []
            res = client.get('/v1/terrareg/git_providers')
            assert res.status_code == 200
            assert res.json == []
            mocked_git_providers_get_all.assert_called_once()

    @setup_test_data()
    def test_with_git_providers_configured(self, client):
        """Test endpoint with git providers configured."""
        with mock.patch('terrareg.models.GitProvider.get_all') as mocked_git_providers_get_all:
            mocked_git_providers_get_all.side_effect = MockGitProvider.get_all

            res = client.get('/v1/terrareg/git_providers')
            assert res.status_code == 200
            assert res.json == [
                {'id': 1, 'name': 'testgitprovider'},
                {'id': 2, 'name': 'git-provider-only-clone'}
            ]
            mocked_git_providers_get_all.assert_called_once()
