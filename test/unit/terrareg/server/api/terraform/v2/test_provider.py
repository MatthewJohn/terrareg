
from datetime import datetime, timedelta
import pytest
import unittest.mock

from test.integration.terrareg import TerraregIntegrationTest
from test import AnyDateString, client, app_context, test_request_context
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.models
import terrareg.analytics


class TestApiProvider(TerraregIntegrationTest):
    """Test ApiProvider endpoint"""

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

        res = client.get(f'/v2/providers/{namespace.name}/{provider.name}')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'alias': None,
                    'description': 'Test Multiple Versions',
                    'downloads': 87,
                    'featured': False,
                    'full-name': 'terraform-provider-multiple-versions',
                    'logo-url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
                    'name': 'multiple-versions',
                    'namespace': 'initial-providers',
                    'owner-name': 'initial-providers',
                    'repository-id': provider.repository.pk,
                    'robots-noindex': False,
                    'source': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions',
                    'tier': 'community',
                    'unlisted': False,
                    'warning': ''
                },
                'id': provider.pk,
                'links': {'self': f'/v2/providers/{provider.pk}'},
                'type': 'providers'
            }
        }

    def test_with_provider_version_includes(self, client, test_request_context):
        """Test endpoint with provider version include."""

        with test_request_context:
            namespace = terrareg.models.Namespace.get("contributed-providersearch")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="mixedsearch-result-multiversion")
            provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.2.3")
            provider_version_v2 = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="2.0.0")

        res = client.get(f'/v2/providers/{namespace.name}/{provider.name}?include=provider-versions')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'alias': None,
                    'description': 'Test Multiple Versions',
                    'downloads': 0,
                    'featured': False,
                    'full-name': 'terraform-provider-mixedsearch-result-multiversion',
                    'logo-url': 'https://git.example.com/contributed-providersearch-namespace/terraform-provider-test-initial.png',
                    'name': 'mixedsearch-result-multiversion',
                    'namespace': 'contributed-providersearch',
                    'owner-name': 'contributed-providersearch',
                    'repository-id': provider.repository.pk,
                    'robots-noindex': False,
                    'source': 'https://github.example.com/contributed-providersearch/terraform-provider-mixedsearch-result-multiversion',
                    'tier': 'community',
                    'unlisted': False,
                    'warning': ''
                },
                'id': provider.pk,
                'links': {'self': f'/v2/providers/{provider.pk}'},
                'type': 'providers'
            },
            'included': [
                {
                    'attributes': {
                        'description': 'Test Multiple Versions',
                        'downloads': 0,
                        'published-at': AnyDateString(),
                        'tag': 'v2.0.0',
                        'version': '2.0.0'
                    },
                    'id': str(provider_version_v2.pk),
                    'links': {'self': f'/v2/provider-versions/{provider_version_v2.pk}'},
                    'type': 'provider-versions'
                },
                {
                    'attributes': {
                        'description': 'Test Multiple Versions',
                        'downloads': 0,
                        'published-at': AnyDateString(),
                        'tag': 'v1.2.3',
                        'version': '1.2.3'
                    },
                    'id': str(provider_version.pk),
                    'links': {'self': f'/v2/provider-versions/{provider_version.pk}'},
                    'type': 'provider-versions'
                }
            ]
        }

    def test_with_category_includes(self, client, test_request_context):
        """Test endpoint with category include."""

        with test_request_context:
            namespace = terrareg.models.Namespace.get("contributed-providersearch")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="mixedsearch-result-multiversion")

        res = client.get(f'/v2/providers/{namespace.name}/{provider.name}?include=categories')
        assert res.status_code == 200
        assert res.json == {
            'data': {
                'attributes': {
                    'alias': None,
                    'description': 'Test Multiple Versions',
                    'downloads': 0,
                    'featured': False,
                    'full-name': 'terraform-provider-mixedsearch-result-multiversion',
                    'logo-url': 'https://git.example.com/contributed-providersearch-namespace/terraform-provider-test-initial.png',
                    'name': 'mixedsearch-result-multiversion',
                    'namespace': 'contributed-providersearch',
                    'owner-name': 'contributed-providersearch',
                    'repository-id': provider.repository.pk,
                    'robots-noindex': False,
                    'source': 'https://github.example.com/contributed-providersearch/terraform-provider-mixedsearch-result-multiversion',
                    'tier': 'community',
                    'unlisted': False,
                    'warning': ''
                },
                'id': provider.pk,
                'links': {'self': f'/v2/providers/{provider.pk}'},
                'type': 'providers'
            },
            'included': [
                {
                    'attributes': {
                        'name': 'Visible Monitoring',
                        'slug': 'visible-monitoring',
                        'user-selectable': True
                    },
                    'id': '523',
                    'links': {'self': '/v2/categories/523'},
                    'type': 'categories'
                }
            ]
        }

    def test_endpoint_with_non_existent_provider(self, client):
        """Test endpoint with non-existent provider."""
        res = client.get('/v2/providers/initial-providers/doesnotexist')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_endpoint_with_non_existent_namespace(self, client):
        """Test endpoint with non-existent namespace."""
        res = client.get('/v2/providers/doesnotexist/doesnotexist')
        assert res.status_code == 404
        assert res.json == {'errors': ['Not Found']}

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/providers/initial-providers/multiple-versions')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)
