
from operator import mod
import unittest.mock

import pytest

from test.unit.terrareg import TerraregUnitTest
from test import client, app_context, test_request_context


class TestPrometheusMetrics(TerraregUnitTest):
    """Test global usage stats endpoint"""

    def test_prometheus_metrics(
            self, app_context,
            test_request_context,
            client
        ):
        """Test update of repository URL."""
        with client, \
                unittest.mock.patch('terrareg.analytics.AnalyticsEngine.get_prometheus_metrics') as mock_get_prometheus_metrics:

            mock_get_prometheus_metrics.return_value = """
# HELP unittest_output_count Unittest test output
# # TYPE unittest_output_count counter
# unittest_output_count 5
""".strip()

            res = client.get('/metrics')

            assert res.data.decode('utf-8') == """
# HELP unittest_output_count Unittest test output
# # TYPE unittest_output_count counter
# unittest_output_count 5
""".strip()
            assert res.status_code == 200
            assert res.headers['Content-Type'] == 'text/plain; version=0.0.4'

            mock_get_prometheus_metrics.assert_called_once()
