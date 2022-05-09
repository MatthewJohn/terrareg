
import unittest.mock

import pytest
from terrareg.filters import NamespaceTrustFilter

from terrareg.module_search import ModuleSearch, ModuleSearchResults



@pytest.fixture
def mocked_search_module_providers(request):
    """Create mocked instance of search_module_providers method."""

    def search_results_func(
            offset: int,
            limit: int,
            query: str=None,
            namespaces: list=None,
            providers: list=None,
            verified: bool=False,
            namespace_trust_filters: list=NamespaceTrustFilter.UNSPECIFIED):
        return ModuleSearchResults(offset=offset, limit=limit, count=0, module_providers=[])

    magic_mock = unittest.mock.MagicMock(
        side_effect=search_results_func
    )
    mock = unittest.mock.patch('terrareg.server.ModuleSearch.search_module_providers', magic_mock)

    def cleanup_mocked_search_module_providers():
        mock.stop()
    request.addfinalizer(cleanup_mocked_search_module_providers)
    mock.start()

    yield magic_mock


@pytest.fixture
def mock_record_module_version_download(request):
    """Mock record_module_version_download function of AnalyticsEngine class."""
    magic_mock = unittest.mock.MagicMock(return_value=None)
    mock = unittest.mock.patch('terrareg.server.AnalyticsEngine.record_module_version_download', magic_mock)

    def cleanup_mocked_record_module_version_download():
        mock.stop()
    request.addfinalizer(cleanup_mocked_record_module_version_download)
    mock.start()


@pytest.fixture
def mock_server_get_module_provider_download_stats(request):
    """Mock get_module_provider_download_stats function of AnalyticsEngine class."""
    magic_mock = unittest.mock.MagicMock(return_value={
        'week': 10,
        'month': 58,
        'year': 127,
        'total': 226
    })
    mock = unittest.mock.patch('terrareg.server.AnalyticsEngine.get_module_provider_download_stats', magic_mock)

    def cleanup_mocked_record_module_version_download():
        mock.stop()
    request.addfinalizer(cleanup_mocked_record_module_version_download)
    mock.start()