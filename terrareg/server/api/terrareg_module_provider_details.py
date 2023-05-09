
import urllib.parse

from flask import request
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource


class ApiTerraregModuleProviderDetails(ErrorCatchingResource):
    """Interface to obtain module provider details."""

    def _get(self, namespace, name, provider):
        """Return details about module version."""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'target_terraform_version', type=str, location='args', default=None,
            help='Provide terraform version to show compatibility with search results.'
        )
        args = parser.parse_args()

        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        # If a version exists, obtain the details for that
        latest_version = module_provider.get_latest_version()
        if latest_version is not None:
            return latest_version.get_terrareg_api_details(
                request_domain=urllib.parse.urlparse(request.base_url).hostname,
                target_terraform_version=args.target_terraform_version
            )

        # Otherwise, return module provider details
        return module_provider.get_terrareg_api_details()
