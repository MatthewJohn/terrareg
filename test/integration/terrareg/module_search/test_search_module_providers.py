
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
            assert len(result.rows) == 1
            assert result.rows[0].id == module_provider.id
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

    @pytest.mark.parametrize('offset,expected_next_offset,expected_results', [
        (0, 2, ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
                'searchbynamespace/searchbymodulename1/searchbyprovidergcp']),
        (2, 4, ['searchbynamespace/searchbymodulename2/published',
                'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws']),
        (3, None, ['searchbynamesp-similar/searchbymodulename3/searchbyprovideraws',
                   'searchbynamesp-similar/searchbymodulename4/aws']),
        (4, None, ['searchbynamesp-similar/searchbymodulename4/aws']),
        (5, None, [])
    ])
    def test_offset_limit(self, offset, expected_next_offset, expected_results):
        """Test offset and limit params of module search."""
        result = ModuleSearch.search_module_providers(
            offset=offset, limit=2,
            query='searchbynamesp'
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

        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_module_provider in expected_results:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('verified_flag,expected_module_provider_ids', [
        # Search with flag unset
        (None, ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
                'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
                'searchbynamespace/searchbymodulename2/published',
                'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws',
                'searchbynamesp-similar/searchbymodulename4/aws']),

        # Search for verified modules
        (True, ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
                'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws']),

        # Search for with verified flag off
        (False, ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
                 'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
                 'searchbynamespace/searchbymodulename2/published',
                 'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws',
                 'searchbynamesp-similar/searchbymodulename4/aws'])
    ])
    def test_search_with_verified_flag(self, verified_flag, expected_module_provider_ids):
        """Test search with verified flag"""
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query='searchbynamesp',
            verified=verified_flag
        )

        assert result.count == len(expected_module_provider_ids)
        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('module_name,expected_versions', [
        ('contributedmodule-withbetaversion', ['1.2.3']),
        ('contributedmodule-onlybeta', []),

        ('verifiedmodule-withbetaversion', ['1.2.3']),
        ('verifiedmodule-onlybeta', [])
    ])
    def test_search_beta_versions(self, module_name, expected_versions):
        """Test search excludes beta versions"""
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query=module_name,
            namespaces=['modulesearch']
        )

        # Ensure that if at least one version is present,
        # the provide is returned
        assert result.count == (1 if len(expected_versions) else 0)

    @pytest.mark.parametrize('namespace,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact namespace match
        ('searchbynamespace',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
          'searchbynamespace/searchbymodulename2/published']
        ),

        # Search with partial namespace match
        ('searchbynamesp',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
          'searchbynamespace/searchbymodulename2/published',
          'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws',
          'searchbynamesp-similar/searchbymodulename4/aws']
        )
    ])
    def test_namespace_search_in_query_string(self, namespace, expected_module_provider_ids):
        """Search based on namespace in query string"""
        # Perform search with namespace in query
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query=namespace
        )

        assert result.count == len(expected_module_provider_ids)
        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('namespace,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact namespace match
        ('searchbynamespace',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
          'searchbynamespace/searchbymodulename2/published']),

        # Search with partial namespace match
        ('searchbynamesp',
         [])
    ])
    def test_namespace_search_in_filter(self, namespace, expected_module_provider_ids):
        """Search based on namespace in filter"""
        # Search with empty query string with namespace filter
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query='',
            namespaces=[namespace]
        )

        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]

        assert result.count == len(expected_module_provider_ids)

        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('module_name_search,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact module name match
        ('searchbymodulename1',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp']
        ),
        ('searchbymodulename2',
         ['searchbynamespace/searchbymodulename2/published']
        ),

        # Search with partial module name match
        ('searchbymodulename',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp',
          'searchbynamespace/searchbymodulename2/published',
          'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws',
          'searchbynamesp-similar/searchbymodulename4/aws']
        )
    ])
    def test_module_name_search_in_query_string(self, module_name_search, expected_module_provider_ids):
        """Search based on module name in query string"""
        # Perform search with module name in query
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query=module_name_search
        )

        assert result.count == len(expected_module_provider_ids)
        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('module_name_search,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact module name match
        ('searchbymodulename1',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamespace/searchbymodulename1/searchbyprovidergcp']
        ),
        ('searchbymodulename2',
         ['searchbynamespace/searchbymodulename2/published']
        ),

        # Search with partial module name match
        ('searchbymodulename', [])
    ])
    def test_module_name_search_in_filter(self, module_name_search, expected_module_provider_ids):
        """Search based on module name in filter"""
        # Search with empty query string with module name filter
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query='',
            modules=[module_name_search]
        )

        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]

        assert result.count == len(expected_module_provider_ids)

        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('provider_name_search,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact provider name match
        ('searchbyprovideraws',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws']
        ),
        ('searchbyprovidergcp',
         ['searchbynamespace/searchbymodulename1/searchbyprovidergcp']
        ),

        # Search with partial provider name match.
        # Provider is not searched by wildcard, so
        # partial match should not yield any results.
        ('searchbyprovider', [])
    ])
    def test_provider_name_search_in_query_string(self, provider_name_search, expected_module_provider_ids):
        """Search based on provider name in query string"""
        # Perform search with provider name in query
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query=provider_name_search
        )

        assert result.count == len(expected_module_provider_ids)
        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]
        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    @pytest.mark.parametrize('provider_name_search,expected_module_provider_ids', [
        ('testnotexist', []),

        # Search with exact provider name match
        ('searchbyprovideraws',
         ['searchbynamespace/searchbymodulename1/searchbyprovideraws',
          'searchbynamesp-similar/searchbymodulename3/searchbyprovideraws']
        ),
        ('searchbyprovidergcp',
         ['searchbynamespace/searchbymodulename1/searchbyprovidergcp']
        ),

        # Search with partial provider name match
        ('searchbyprovider', [])
    ])
    def test_provider_name_search_in_filter(self, provider_name_search, expected_module_provider_ids):
        """Search based on provider name in filter"""
        # Search with empty query string with provider name filter
        result = ModuleSearch.search_module_providers(
            offset=0, limit=50,
            query='',
            providers=[provider_name_search]
        )

        resulting_module_provider_ids = [
            module_provider.id
            for module_provider in result.rows
        ]

        assert result.count == len(expected_module_provider_ids)

        for expected_module_provider in expected_module_provider_ids:
            assert expected_module_provider in resulting_module_provider_ids

    def test_search_in_description(self):
        """Search for module providers based on value in description."""
        result = ModuleSearch.search_module_providers(
            offset=0, limit=10,
            query='DESCRIPTION-Search',
            namespaces=['modulesearch']
        )

        # Ensure that only one result is returned
        assert result.count == 1
        assert result.rows[0].id == 'modulesearch/contributedmodule-oneversion/aws'

    @pytest.mark.parametrize('search_query', [
        # Search by description of non-latest version of module
        'DESCRIPTION-Search-OLDVERSION',
        # Search by description of beta version of module
        'DESCRIPTION-Search-BETAVERSION',
        # Search by description of unpublished version of module
        'DESCRIPTION-Search-UNPUBLISHED'
    ])
    def test_search_in_description_non_latest(self, search_query):
        """Search for module providers based on value in description, expecting description
           of non-latest versions to be ignored."""
        result = ModuleSearch.search_module_providers(
            offset=0, limit=10,
            query=search_query,
            namespaces=['modulesearch']
        )

        # Ensure that no results are returned
        assert result.count == 0
