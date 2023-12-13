
from unittest import mock
import pytest
from terrareg.filters import NamespaceTrustFilter

from terrareg.models import Namespace
from terrareg.provider_search import ProviderSearch
from terrareg.provider_model import Provider

from test.integration.terrareg import TerraregIntegrationTest

class TestSearchProviders(TerraregIntegrationTest):

    @pytest.mark.parametrize('offset,limit,expected_offset,expected_limit,expected_prev', [
        (0, 1, 0, 1, None),
        (0, 10, 0, 10, None),
        # Test max allowed limit
        (0, 50, 0, 50, None),
        # Test exceeding max limit
        (0, 51, 0, 50, None),
        # Test with expected previous offset
        (10, 2, 10, 2, 8),
        # Test with limit and offset that would
        # mean a negative previous offset
        (10, 20, 10, 20, 0),
        # Test with negative offset
        (-5, 1, 0, 1, None),
        # Test with limit of 0
        (5, 0, 5, 1, 4)
    ])
    def test_offset_without_next(self, offset, limit, expected_offset, expected_limit, expected_prev):
        """Test search with partial module name match with multiple matches."""

        namespace = Namespace(name='providersearch')
        provider = Provider(namespace=namespace, name='contributedprovider-oneversion')

        result = ProviderSearch.search_providers(
            query='contributedprovider-oneversion',
            offset=offset,
            limit=limit
        )

        expected_meta = {
            'limit': expected_limit,
            'current_offset': expected_offset,
        }
        if expected_prev is not None:
            expected_meta['prev_offset'] = expected_prev

        assert result.meta == expected_meta

        assert result.count == 1

        if result.meta['current_offset'] == 0:
            assert len(result.rows) == 1
            assert result.rows[0].id == provider.id
        else:
            assert result.rows == []

    @pytest.mark.parametrize('namespace_trust_filter,expected_result_count', [
        (NamespaceTrustFilter.UNSPECIFIED, 5),
        ([NamespaceTrustFilter.TRUSTED_NAMESPACES], 3),
        ([NamespaceTrustFilter.CONTRIBUTED], 2),
        ([NamespaceTrustFilter.TRUSTED_NAMESPACES, NamespaceTrustFilter.CONTRIBUTED], 5)
    ])
    def test_namespace_filter(self, namespace_trust_filter, expected_result_count):
        """Test search with different trust filters."""

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted']):
            result = ProviderSearch.search_providers(
                query='mixedsearch',
                offset=0,
                limit=1,
                namespace_trust_filters=namespace_trust_filter
            )

        assert result.count == expected_result_count

    @pytest.mark.parametrize('offset,expected_next_offset,expected_results', [
        (0, 2, ['modulesearch-trusted/mixedsearch-trusted-result',
                'modulesearch-trusted/mixedsearch-trusted-second-result']),
        (2, 4, ['modulesearch-trusted/mixedsearch-trusted-result-multiversion',
                'contributed-providersearch/mixedsearch-result']),
        (3, None, ['contributed-providersearch/mixedsearch-result',
                   'contributed-providersearch/mixedsearch-result-multiversion']),
        (4, None, ['contributed-providersearch/mixedsearch-result-multiversion']),
        (5, None, [])
    ])
    def test_offset_limit(self, offset, expected_next_offset, expected_results):
        """Test offset and limit params of module search."""
        result = ProviderSearch.search_providers(
            offset=offset, limit=2,
            query='mixedsearch'
        )

        assert result.count == 5

        assert result.meta['current_offset'] == offset

        if offset == 0:
            assert 'prev_offset' not in result.meta
        else:
            assert result.meta['prev_offset'] == (offset - 2)

        if offset in [0, 2]:
            assert result.meta['next_offset'] == expected_next_offset
        else:
            assert 'next_offset' not in result.meta

        resulting_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_provider in expected_results:
            assert expected_provider in resulting_provider_ids

    @pytest.mark.parametrize('namespace,expected_provider_ids', [
        ('testnotexist', []),

        # Search with exact namespace match
        ('modulesearch-trusted',
         ['modulesearch-trusted/mixedsearch-trusted-result',
          'modulesearch-trusted/mixedsearch-trusted-second-result',
          'modulesearch-trusted/mixedsearch-trusted-result-multiversion']),

        # Search with partial namespace match
        ('modulesearch-trust',
         ['modulesearch-trusted/mixedsearch-trusted-result',
          'modulesearch-trusted/mixedsearch-trusted-second-result',
          'modulesearch-trusted/mixedsearch-trusted-result-multiversion']),
    ])
    def test_namespace_search_in_query_string(self, namespace, expected_provider_ids):
        """Search based on namespace in query string"""
        # Perform search with namespace in query
        result = ProviderSearch.search_providers(
            offset=0, limit=50,
            query=namespace
        )

        assert result.count == len(expected_provider_ids)
        resulting_provider_ids = [
            provider.id
            for provider in result.rows
        ]
        for expected_provider in expected_provider_ids:
            assert expected_provider in expected_provider_ids

    @pytest.mark.parametrize('namespace,expected_provider_ids', [
        ('testnotexist', []),

        # Search with exact namespace match
        ('modulesearch-trusted',
         ['modulesearch-trusted/mixedsearch-trusted-result',
          'modulesearch-trusted/mixedsearch-trusted-second-result',
          'modulesearch-trusted/mixedsearch-trusted-result-multiversion']),

        # Search with partial namespace match
        ('modulesearch-trust',
         [])
    ])
    def test_namespace_search_in_filter(self, namespace, expected_provider_ids):
        """Search based on namespace in filter"""
        # Search with empty query string with namespace filter
        result = ProviderSearch.search_providers(
            offset=0, limit=50,
            query='',
            namespaces=[namespace]
        )

        resulting_provider_ids = [
            provider.id
            for provider in result.rows
        ]

        assert result.count == len(expected_provider_ids)

        for expected_provider in resulting_provider_ids:
            assert expected_provider in expected_provider_ids

    @pytest.mark.parametrize('provider_name_search,expected_provider_ids', [
        ('testnotexist', []),

        # Search with exact module name match
        ('mixedsearch-result',
         ['contributed-providersearch/mixedsearch-result',
          'contributed-providersearch/mixedsearch-result-multiversion']
        ),

        # Search with partial module name match
        ('mixedsearch',
         ['modulesearch-trusted/mixedsearch-trusted-result',
          'modulesearch-trusted/mixedsearch-trusted-second-result',
          'modulesearch-trusted/mixedsearch-trusted-result-multiversion',
          'contributed-providersearch/mixedsearch-result',
          'contributed-providersearch/mixedsearch-result-multiversion'
         ]
        )
    ])
    def test_provider_name_search_in_query_string(self, provider_name_search, expected_provider_ids):
        """Search based on module name in query string"""
        # Perform search with module name in query
        result = ProviderSearch.search_providers(
            offset=0, limit=50,
            query=provider_name_search
        )

        assert result.count == len(expected_provider_ids)
        resulting_provider_ids = [
            provider.id
            for provider in result.rows
        ]
        for expected_provider in expected_provider_ids:
            assert expected_provider in resulting_provider_ids

    @pytest.mark.parametrize('provider_name_search,expected_provider_ids', [
        ('testnotexist', []),

        # Search with exact module name match
        ('mixedsearch-result',
         ['contributed-providersearch/mixedsearch-result']
        ),

        # Search with partial module name match
        ('mixedsearch', [])
    ])
    def test_provider_name_search_in_filter(self, provider_name_search, expected_provider_ids):
        """Search based on provider name in filter"""
        # Search with empty query string with module name filter
        result = ProviderSearch.search_providers(
            offset=0, limit=50,
            query='',
            providers=[provider_name_search]
        )

        resulting_provider_ids = [
            provider.id
            for provider in result.rows
        ]

        assert result.count == len(expected_provider_ids)

        for expected_provider in expected_provider_ids:
            assert expected_provider in resulting_provider_ids

    def test_search_in_description(self):
        """Search for module providers based on value in description."""
        result = ProviderSearch.search_providers(
            offset=0, limit=10,
            query='DESCRIPTION-Search'
        )

        # Ensure that only one result is returned
        assert result.count == 1
        assert result.rows[0].id == 'providersearch/contributedprovider-oneversion'

    @pytest.mark.parametrize('search_query', [
        # Search by description with no versions
        'DESCRIPTION-NoVersion',
    ])
    def test_search_in_description_no_version(self, search_query):
        """Search for module providers based on value in description, expecting description
           of provider without a version to be ignored."""
        result = ProviderSearch.search_providers(
            offset=0, limit=10,
            query=search_query
        )

        # Ensure that no results are returned
        assert result.count == 0
