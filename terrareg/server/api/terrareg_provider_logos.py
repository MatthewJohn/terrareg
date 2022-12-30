
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.models import ProviderLogo


class ApiTerraregProviderLogos(ErrorCatchingResource):
    """Provide interface to obtain all provider logo details"""

    def _get(self):
        """Return all details about provider logos."""
        return {
            provider_logo.provider: {
                'source': provider_logo.source,
                'alt': provider_logo.alt,
                'tos': provider_logo.tos,
                'link': provider_logo.link
            }
            for provider_logo in ProviderLogo.get_all()
        }
