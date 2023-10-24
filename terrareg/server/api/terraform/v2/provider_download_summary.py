
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.analytics


class ApiProviderProviderDownloadSummary(ErrorCatchingResource):
    """Interface for providing download summary for providers"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, provider_id: int):
        """Return download summary."""

        provider_obj = terrareg.provider_model.Provider.get_by_pk(provider_id)
        if provider_obj is None:
            return self._get_404_response()

        download_stats = terrareg.analytics.ProviderAnalytics.get_provider_download_stats(provider_obj)

        return {
            "data": {
                "type":"provider-downloads-summary",
                "id": str(provider_obj.pk),
                "attributes": download_stats
            }
        }
