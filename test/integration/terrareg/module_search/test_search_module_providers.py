
from unittest import mock
import pytest
from terrareg.filters import NamespaceTrustFilter

from terrareg.models import Module, ModuleProvider, Namespace
from terrareg.module_search import ModuleSearch
from test.integration.terrareg import TerraregIntegrationTest

class TestSearchModuleProviders(TerraregIntegrationTest):

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

        namespace = Namespace(name='modulesearch')
        module = Module(namespace=namespace, name='contributedmodule-oneversion')
        module_provider = ModuleProvider(module=module, name='aws')

        result = ModuleSearch.search_module_providers(
            query='contributedmodule-oneversion',
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
            assert len(result.module_providers) == 1
            assert result.module_providers[0].id == module_provider.id
        else:
            assert result.module_providers == []

    @pytest.mark.parametrize('namespace_trust_filter,expected_result_count', [
        (NamespaceTrustFilter.UNSPECIFIED, 5),
        ([NamespaceTrustFilter.TRUSTED_NAMESPACES], 3),
        ([NamespaceTrustFilter.CONTRIBUTED], 2),
        ([NamespaceTrustFilter.TRUSTED_NAMESPACES, NamespaceTrustFilter.CONTRIBUTED], 5)
    ])
    def test_namespace_filter(self, namespace_trust_filter, expected_result_count):
        """Test search with different trust filters."""

        namespace = Namespace(name='modulesearch')
        module = Module(namespace=namespace, name='contributedmodule-oneversion')
        module_provider = ModuleProvider(module=module, name='aws')

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted']):
            result = ModuleSearch.search_module_providers(
                query='mixedsearch',
                offset=0,
                limit=1,
                namespace_trust_filters=namespace_trust_filter
            )

        assert result.count == expected_result_count


    @pytest.mark.parametrize('namespace,expected_query_string_result_ids,expected_namespace_filter_result_ids', [
        ('testnotexist', [], []),

        # Search with exact namespace match
        ('searchbynamespace',
         ['searchbynamespace/module1/aws',
          'searchbynamespace/module1/gcp',
          'searchbynamespace/module2/published'],
         ['searchbynamespace/module1/aws',
          'searchbynamespace/module1/gcp',
          'searchbynamespace/module2/published']),

        # Search with partial namespace match
        ('searchbynamesp',
         ['searchbynamespace/module1/aws',
          'searchbynamespace/module1/gcp',
          'searchbynamespace/module2/published',
          'searchbynamesp-similar/module3/aws'],
         [])
    ])
    def test_namespace_search(self, namespace, expected_query_string_result_ids, expected_namespace_filter_result_ids):
        """Search based on namespace"""
        # Perform search with namespace in query
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query=namespace
        )

        assert result.count == len(expected_query_string_result_ids)
        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.module_providers
        ]
        for expected_module_provider in expected_query_string_result_ids:
            assert expected_module_provider in resulting_module_provider_ids


        # Search with empty query string with namespace filter
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query='',
            namespace=namespace
        )

        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.module_providers
        ]
        print(resulting_module_provider_ids)

        assert result.count == len(expected_namespace_filter_result_ids)

        for expected_module_provider in expected_namespace_filter_result_ids:
            assert expected_module_provider in resulting_module_provider_ids
