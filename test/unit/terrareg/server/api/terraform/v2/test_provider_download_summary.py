
from datetime import datetime, timedelta
import pytest
import unittest.mock

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.models
import terrareg.analytics


class TestApiProviderDownloadSummary(TerraregIntegrationTest):
    """Test ApiProviderDownloadSummary endpoint"""

    def test_endpoint(self, client, test_request_context):
        """Test endpoint."""

        with test_request_context:
            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
            provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")
            for date, count in [(datetime.now() - timedelta(weeks=75), 42),
                                (datetime.now() - timedelta(weeks=8), 25),
                                (datetime.now() - timedelta(weeks=3), 12),
                                (datetime.now() - timedelta(days=4), 5),
                                (datetime.now() - timedelta(hours=2), 3)]:
                with unittest.mock.patch('terrareg.analytics.AnalyticsEngine.get_datetime_now', unittest.mock.MagicMock(return_value=date)):
                    for _ in range(0, count):
                        terrareg.analytics.ProviderAnalytics.record_provider_version_download(
                            namespace_name=namespace.name,
                            provider_name=provider.name,
                            provider_version=provider_version,
                            terraform_version="1.4.5",
                            user_agent="terraform hcl/1.4.5"
                        )

        res = client.get(f'/v2/providers/{provider.pk}/downloads/summary')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'month': 20,
                    'total': 87,
                    'week': 8,
                    'year': 45
                },
                'id': str(provider.pk),
                'type': 'provider-downloads-summary'
            },
        }

    def test_endpoint_with_non_existent_provider(self, client):
        """Test endpoint with non-existent provider."""
        res = client.get('/v2/providers/6145135/downloads/summary')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/providers/6145135/downloads/summary')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)