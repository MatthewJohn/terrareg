
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.analytics
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregGlobalUsageStats(ErrorCatchingResource):
    """Provide interface to obtain statistics about global module usage."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """
        Return stats on total module providers,
        total unique analytics tokens per module
        (with and without auth token).
        """
        module_usage_with_auth_token = terrareg.analytics.AnalyticsEngine.get_global_module_usage_counts()
        module_usage_including_empty_auth_token = terrareg.analytics.AnalyticsEngine.get_global_module_usage_counts(include_empty_auth_token=True)
        total_analytics_token_with_auth_token = sum(module_usage_with_auth_token.values())
        total_analytics_token_including_empty_auth_token = sum(module_usage_including_empty_auth_token.values())
        return {
            'module_provider_count': terrareg.models.ModuleProvider.get_total_count(),
            'module_provider_usage_breakdown_with_auth_token': module_usage_with_auth_token,
            'module_provider_usage_count_with_auth_token': total_analytics_token_with_auth_token,
            'module_provider_usage_including_empty_auth_token': module_usage_including_empty_auth_token,
            'module_provider_usage_count_including_empty_auth_token': total_analytics_token_including_empty_auth_token
        }