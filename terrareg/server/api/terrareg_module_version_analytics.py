
from flask import request
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.config
import terrareg.analytics
import terrareg.auth_wrapper
import terrareg.auth

class ApiTerraregModuleVersionAnalytics(ErrorCatchingResource):
    """Provide endpoint for interacting with module version analytics."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper("can_access_terraform_api")]

    def _post_arg_parser(self):
        """Get arg parser for POST request"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            "analytics_token",
            type=str,
            required=True,
            help="Analytics token to register.",
            location="json"
        )
        parser.add_argument(
            "terraform_version",
            type=str,
            required=False,
            default=None,
            help="Version of Terraform used.",
            location="json"
        )
        return parser

    def _post(self, namespace, name, provider, version):
        """
        Submit analytics for module version.

        Used as an alternative to passing analytics in module source URL.
        The module execution can perform a http request to this endpoint to register analytics for a module usage.
        """

        if terrareg.config.Config().DISABLE_ANALYTICS:
            return {"errors": ["Analytics is disabled"]}, 400

        # If a version has been provided, get the exact version
        if version:
            _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
            if error:
                return self._get_404_response()

        auth_method = terrareg.auth.AuthFactory().get_current_auth_method()

        # Silenty return early if auth method ignores recording analytics
        if not auth_method.should_record_terraform_analytics():
            return {}

        parser = self._post_arg_parser()
        args = parser.parse_args()

        auth_method = terrareg.auth.AuthFactory().get_current_auth_method()

        # Determine if auth method ignores recording analytics
        if not auth_method.should_record_terraform_analytics():
            return {}

        analytics_token = terrareg.analytics.AnalyticsEngine.sanitise_analytics_token(args.analytics_token)
        if analytics_token is None:
            return {"errors": ["Invalid analytics token"]}, 400

        terrareg.analytics.AnalyticsEngine.record_module_version_download(
            namespace_name=namespace,
            module_name=name,
            provider_name=provider,
            module_version=module_version,
            analytics_token=analytics_token,
            terraform_version=args.terraform_version,
            user_agent=None,
            auth_token=auth_method.get_terraform_auth_token(),
            ignore_user_agent=True,
        )

        return {}
