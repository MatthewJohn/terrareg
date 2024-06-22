
import urllib.parse

from flask import request
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.auth_wrapper


class ApiTerraregModuleVersionDetails(ErrorCatchingResource):
    """Interface to obtain module provider/version details."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get_arg_parser(self) -> reqparse.RequestParser:
        """Return argument parser for GET method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'target_terraform_version', type=str, location='args', default=None,
            help='Provide terraform version to show compatibility with search results.'
        )
        parser.add_argument(
            'output',
            type=str,
            location='args',
            default='md',
            dest='output',
            help='Variable/Output description format, either "html" or "md"'
        )
        return parser

    def _get(self, namespace, name, provider, version=None):
        """Return details about module version."""
        parser = self._get_arg_parser()
        args = parser.parse_args()

        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        if version is not None:
            module_version = terrareg.models.ModuleVersion.get(module_provider=module_provider, version=version)
        else:
            # If version has not been specified, attempt to get latest version
            module_version = module_provider.get_latest_version()

            # Otherwise, show provider provider details
            if module_version is None:
                return module_provider.get_terrareg_api_details()

        if module_version is None:
            return self._get_404_response()

        return module_version.get_terrareg_api_details(
            request_domain=urllib.parse.urlparse(request.base_url).hostname,
            target_terraform_version=args.target_terraform_version,
            html=(args.output == "html")
        )
