
import urllib.parse

from flask import request
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregModuleVersionDetails(ErrorCatchingResource):
    """Interface to obtain module version details."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider, version=None):
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

        if version is not None:
            module_version = terrareg.models.ModuleVersion.get(module_provider=module_provider, version=version)
        else:
            module_version = module_provider.get_latest_version()

        if module_version is None:
            return self._get_404_response()

        return module_version.get_terrareg_api_details(
            request_domain=urllib.parse.urlparse(request.base_url).hostname,
            target_terraform_version=args.target_terraform_version
        )
