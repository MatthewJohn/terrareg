

from terrareg.analytics import AnalyticsEngine
from . import AnalyticsIntegrationTest


class TestGetGlobalModuleUsage(AnalyticsIntegrationTest):
    """Test get_global_module_usage function."""

    def test_get_global_module_usage_counts_with_no_analytics(self):
        """Test function with no analytics recorded."""
        assert AnalyticsEngine.get_global_module_usage_counts() == {}

    def test_get_global_module_usage_counts_excluding_no_environment(self):
        """Test function with default functionality, excluding stats for analytics without an API token"""
        self._import_test_analaytics(self._TEST_ANALYTICS_DATA)

        assert AnalyticsEngine.get_global_module_usage_counts() == {
            'testnamespace/publishedmodule/testprovider': 4,
            'testnamespace/publishedmodule/secondprovider': 2,
            'testnamespace/secondmodule/testprovider': 2,
            'secondnamespace/othernamespacemodule/anotherprovider': 1
        }

    def test_get_global_module_usage_counts_including_empty_auth_token(self):
        """Test function including stats for analytics without an auth token"""
        self._import_test_analaytics(self._TEST_ANALYTICS_DATA)

        assert AnalyticsEngine.get_global_module_usage_counts(include_empty_auth_token=True) == {
            'testnamespace/publishedmodule/testprovider': 5,
            'testnamespace/publishedmodule/secondprovider': 2,
            'testnamespace/secondmodule/testprovider': 2,
            'secondnamespace/othernamespacemodule/anotherprovider': 1,
            'testnamespace/noanalyticstoken/testprovider': 1
        }


