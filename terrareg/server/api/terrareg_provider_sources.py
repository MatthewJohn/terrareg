from flask import request
from flask_restful import Resource

import terrareg.database
import terrareg.models
import terrareg.errors
from terrareg.server.base_handler import BaseHandler


class ApiTerraregProviderSources(Resource):
    """Handle GET request to list provider sources."""

    def get(self):
        """Return list of provider sources."""
        provider_sources = terrareg.models.ProviderSource.get_all()
        return [
            {
                'name': provider_source.name,
                'api_name': provider_source.api_name,
                'provider_source_type': provider_source.provider_source_type.name,
                'clone_url_template': provider_source.clone_url_template,
                'base_url': provider_source.base_url,
                'browse_url': provider_source.browse_url,
                'git_path': provider_source.git_path,
            }
            for provider_source in provider_sources
        ]


class ApiTerraregProviderSource(Resource):
    """Handle GET and DELETE requests for a specific provider source."""

    def get(self, provider_source_name):
        """Return details of a specific provider source."""
        try:
            provider_source = terrareg.models.ProviderSource.get(name=provider_source_name)
            if provider_source is None:
                raise terrareg.errors.ProviderSourceNotFoundError(
                    f'Provider source with name {provider_source_name} not found'
                )
            return {
                'name': provider_source.name,
                'api_name': provider_source.api_name,
                'provider_source_type': provider_source.provider_source_type.name,
                'clone_url_template': provider_source.clone_url_template,
                'base_url': provider_source.base_url,
                'browse_url': provider_source.browse_url,
                'git_path': provider_source.git_path,
            }
        except terrareg.errors.ProviderSourceNotFoundError as e:
            raise terrareg.errors.ResourceNotFoundError(str(e))

    def delete(self, provider_source_name):
        """Delete a provider source."""
        # Check if user is authenticated
        if not terrareg.auth.AuthFactory().get_current_auth_method().can_access_read_api():
            raise terrareg.errors.AuthenticationRequiredError()

        try:
            provider_source = terrareg.models.ProviderSource.get(name=provider_source_name)
            if provider_source is None:
                raise terrareg.errors.ProviderSourceNotFoundError(
                    f'Provider source with name {provider_source_name} not found'
                )

            # Check if provider source is in use
            if provider_source.is_in_use():
                raise terrareg.errors.ProviderSourceInUseError(
                    f'Cannot delete provider source {provider_source_name} as it is in use by modules or providers'
                )

            # Delete the provider source
            provider_source.delete()

            return {'status': 'success', 'message': f'Provider source {provider_source_name} deleted successfully'}
        except terrareg.errors.ProviderSourceNotFoundError as e:
            raise terrareg.errors.ResourceNotFoundError(str(e))
        except terrareg.errors.ProviderSourceInUseError as e:
            raise terrareg.errors.BadRequestError(str(e))
