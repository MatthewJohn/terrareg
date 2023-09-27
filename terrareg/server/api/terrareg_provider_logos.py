
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregProviderLogos(ErrorCatchingResource):
    """Provide interface to obtain all provider logo details"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self):
        """Return all details about provider logos."""
        return {
            provider_logo.provider: {
                'source': provider_logo.source,
                'alt': provider_logo.alt,
                'tos': provider_logo.tos,
                'link': provider_logo.link
            }
            for provider_logo in terrareg.models.ProviderLogo.get_all()
        }
