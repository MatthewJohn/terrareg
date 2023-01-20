
from flask import make_response

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.analytics


class PrometheusMetrics(ErrorCatchingResource):
    """Provide usage anayltics for Prometheus scraper"""

    def _get(self):
        """
        Return Prometheus metrics for global statistics and module provider statistics
        """
        response = make_response(terrareg.analytics.AnalyticsEngine.get_prometheus_metrics())
        response.headers['content-type'] = 'text/plain; version=0.0.4'

        return response
