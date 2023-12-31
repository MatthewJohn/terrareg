
import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test import client, app_context, test_request_context
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.models


class TestApiProviderDocs(TerraregIntegrationTest):
    """Test ApiV2ProviderDocs endpoint"""

    def test_endpoint(self, client, test_request_context):
        """Test endpoint."""

        with test_request_context:
            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
            provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        res = client.get(f'/v2/provider-docs?filter[provider-version]={provider_version.pk}&filter[category]=resources&filter[slug]=some_new_resource&filter[language]=hcl&page[size]=1')
        assert res.status_code == 200
        assert res.json == {
            'data': [
                {
                    'attributes': {
                        'category': 'resources',
                        'language': 'hcl',
                        'path': 'resources/new-thing.md',
                        'slug': 'some_new_resource',
                        'subcategory': 'some-second-subcategory',
                        'title': 'multiple_versions_thing_new',
                        'truncated': False
                    },
                    'id': '6347',
                    'links': {'self': '/v2/provider-docs/6347'},
                    'type': 'provider-docs'
                }
            ]
        }

    @pytest.mark.parametrize('valid_provider_version, category, slug, language', [
        # Invalid provider version
        (False, 'resources', 'some_new_resource', 'hcl'),
        # Non-existent doc based on category
        (True, 'data-sources', 'some_new_resource', 'hcl'),
        # Non-existent resource
        (True, 'resources', 'does_not_exist', 'hcl'),
        # Non-existent language
        (True, 'resources', 'some_new_resource', 'python'),
    ])
    def test_endpoint_with_non_existent_document(self, valid_provider_version, category, slug, language, client, test_request_context):
        """Test endpoint with non-existent documents."""
        with test_request_context:
            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
            provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        res = client.get(f'/v2/provider-docs?filter[provider-version]={provider_version.pk if valid_provider_version else 513513}&filter[category]={category}&filter[slug]={slug}&filter[language]={language}&page[size]=1')
        assert res.status_code == 200
        assert res.json == {'data': []}

    def test_endpoint_with_invalid_category(self, client, test_request_context):
        """Test endpoint with invalid category."""
        with test_request_context:
            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
            provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        res = client.get(f'/v2/provider-docs?filter[provider-version]={provider_version.pk}&filter[category]=blah&filter[slug]=some_new_resource&filter[language]=hcl&page[size]=1')
        assert res.status_code == 400
        assert res.json == {'errors': ['unsupported filter category']}

    def test_unauthenticated(self, client):
        """Test unauthenticated call to API"""
        def call_endpoint():
            return client.get('/v2/provider-docs/6347')

        self._test_unauthenticated_read_api_endpoint_test(call_endpoint)