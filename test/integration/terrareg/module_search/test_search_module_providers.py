
from unittest import mock
import pytest

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

        with mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', 'modulesearch'):
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