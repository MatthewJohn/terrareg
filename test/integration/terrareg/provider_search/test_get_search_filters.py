
from unittest import mock
import pytest

from terrareg.models import Module, ModuleProvider, Namespace
from terrareg.provider_search import ProviderSearch
from test.integration.terrareg import TerraregIntegrationTest


class TestGetSearchFilters(TerraregIntegrationTest):

    def test_non_search_no_results(self):
        """Test search with no results"""

        results = ProviderSearch.get_search_filters(query='this-search-does-not-exist-at-all')
        assert results == {'namespaces': {}, 'contributed': 0, 'trusted_namespaces': 0, 'provider_categories': {}}

    def test_contributed_provider_one_version(self):
        """Test search with one contributed provider with one version"""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', []):
            results = ProviderSearch.get_search_filters(query='contributedprovider-oneversion')

        assert results == {'namespaces': {'providersearch': 1},
                           'contributed': 1, 'trusted_namespaces': 0,
                           'provider_categories': {'visible-monitoring': 1}}

    def test_contributed_provider_multi_version(self):
        """Test search with one provider with multiple versions."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', []):
            results = ProviderSearch.get_search_filters(query='contributedprovider-multiversion')

        assert results == {'namespaces': {'providersearch': 1},
                           'contributed': 1, 'trusted_namespaces': 0,
                           'provider_categories': {'second-visible-cloud': 1}}

    def test_contributed_multiple_categories(self):
        """Test search with partial provider name match with multiple matches."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', []):
            results = ProviderSearch.get_search_filters(query='contributedprovider')

        assert results == {'namespaces': {'providersearch': 2},
                           'contributed': 2, 'trusted_namespaces': 0,
                           'provider_categories': {'visible-monitoring': 1, 'second-visible-cloud': 1}}

    def test_no_provider_version(self):
        """Test search with provider without a version."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', []):
            results = ProviderSearch.get_search_filters(query='empty-provider-publish')

        assert results == {'namespaces': {},
                           'contributed': 0, 'trusted_namespaces': 0,
                           'provider_categories': {}}

    def test_trusted_provider_one_version(self):
        """Test search with one contributed provider with one version"""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['providersearch']):
            results = ProviderSearch.get_search_filters(query='contributedprovider-oneversion')

        assert results == {'namespaces': {'providersearch': 1},
                           'contributed': 0, 'trusted_namespaces': 1,
                           'provider_categories': {'visible-monitoring': 1}}

    def test_trusted_provider_multi_version(self):
        """Test search with one provider with multiple versions."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['providersearch']):
            results = ProviderSearch.get_search_filters(query='contributedprovider-multiversion')

        assert results == {'namespaces': {'providersearch': 1},
                           'contributed': 0, 'trusted_namespaces': 1,
                           'provider_categories': {'second-visible-cloud': 1}}

    def test_trusted_multiple_providers(self):
        """Test search with partial provider name match with multiple matches."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['doestexist','providersearch','nordoesthis']):
            results = ProviderSearch.get_search_filters(query='contributedprovider')

        assert results == {'namespaces': {'providersearch': 2},
                           'contributed': 0, 'trusted_namespaces': 2,
                           'provider_categories': {'second-visible-cloud': 1, 'visible-monitoring': 1}}

    def test_trusted_no_version_provider(self):
        """Test search with provider without versions version."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['doestexist','providersearch']):
            results = ProviderSearch.get_search_filters(query='contributedprovider-unpublished')

        assert results == {'namespaces': {},
                           'contributed': 0, 'trusted_namespaces': 0,
                           'provider_categories': {}}

    def test_mixed_results(self):
        """Test results containing all results."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['doestexist','providersearch', 'providersearch-trusted']):
            # Search based on partial namespace match
            results = ProviderSearch.get_search_filters(query='providersearch')

        assert results == {'namespaces': {'providersearch': 2, 'contributed-providersearch': 2},
                           'contributed': 2, 'trusted_namespaces': 2,
                           'provider_categories': {'second-visible-cloud': 1, 'visible-monitoring': 3}}
