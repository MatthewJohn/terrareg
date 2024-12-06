from flask_restful import reqparse

import terrareg.analytics
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.csrf
import terrareg.auth_wrapper
import terrareg.provider_model
import terrareg.provider_version_model


class ApiProviderVersions(ErrorCatchingResource):

    method_decorators = {
        "get": [terrareg.auth_wrapper.auth_wrapper('can_access_terraform_api')],
        "post": [terrareg.auth_wrapper.auth_wrapper('can_publish_module_version', request_kwarg_map={'namespace': 'namespace'})],
    }

    def _get(self, namespace, provider):
        """Return provider version details."""

        namespace, _ = terrareg.analytics.AnalyticsEngine.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        provider = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider is None:
            return self._get_404_response()

        return provider.get_versions_api_details()


    def _post_arg_parser(self):
        """Return arg parser for post method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'version',
            type=str, location='json',
            help='Version to be indexed',
            required=True
        )
        parser.add_argument(
            'csrf_token',
            type=str, location='json',
            help='CSRF Token',
            default=None
        )
        return parser

    def _post(self, namespace, provider):
        """Return provider version details."""
        args = self._post_arg_parser().parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        namespace, _ = terrareg.analytics.AnalyticsEngine.extract_analytics_token(namespace)

        namespace_obj = terrareg.models.Namespace.get(name=namespace)
        if namespace_obj is None:
            return self._get_404_response()

        provider = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        if provider is None:
            return self._get_404_response()

        # Index version provided by user
        provider_version = provider.index_version(version=args.version)
        if not provider_version:
            return {"status": "Error", "message": "Unable to create provider version"}
        return {"versions": [provider_version.version]}
