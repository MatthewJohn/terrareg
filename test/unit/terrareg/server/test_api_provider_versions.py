
from datetime import datetime
import unittest.mock

import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_search


class TestApiProviderList(TerraregIntegrationTest):
    """Test ApiNamespaceProviders endpoint"""

    def test_endpoint(self, client):
        """Test endpoint."""
        res = client.get('/v1/providers/initial-providers/multiple-versions')
        assert res.status_code == 200
        assert res.json == {
            'alias': None,
            'description': 'Test Multiple Versions',
            'docs': [],
            'downloads': 0,
            'id': 'initial-providers/multiple-versions/2.0.1',
            'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
            'name': 'multiple-versions',
            'namespace': 'initial-providers',
            'owner': 'initial-providers',
            'published_at': '2023-10-01T12:05:56',
            'source': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions',
            'tag': 'v2.0.1',
            'tier': 'community',
            'version': '2.0.1',
            'versions': [
                '2.0.1',
                '2.0.0',
                '1.5.0',
                '1.1.0',
                '1.1.0-beta',
                '1.0.0'
            ]
        }

    def test_endpoint_with_provider_without_versions(self, client):
        """Test endpoint with provider that doesn't have any versions"""
        res = client.get('/v1/providers/initial-providers/empty-provider-publish')
        assert res.status_code == 404

    def test_endpoint_non_existent_provider(self, client):
        """Test endpoint with non-existent provider"""
        res = client.get('/v1/providers/initial-providers/doesnotexist')
        assert res.status_code == 404

    def test_endpoint_non_existent_namespace(self, client):
        """Test endpoint with non-existent namespace"""
        res = client.get('/v1/providers/doesnotexist/doesnotexist')
        assert res.status_code == 404

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v1/providers/initial-providers/multiple-versions')

        self._test_unauthenticated_terraform_api_endpoint_test(call_endpoint)
