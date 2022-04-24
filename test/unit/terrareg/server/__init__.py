
import unittest.mock

import pytest

from terrareg.module_search import ModuleSearch



@pytest.fixture
def mocked_search_module_providers(request):
    """Create mocked instance of search_module_providers method."""
    unmocked_search_module_providers = ModuleSearch.search_module_providers
    def cleanup_mocked_search_provider():
        ModuleSearch.search_module_providers = unmocked_search_module_providers
    request.addfinalizer(cleanup_mocked_search_provider)

    ModuleSearch.search_module_providers = unittest.mock.MagicMock(return_value=[])


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