

from xml.etree.ElementInclude import include
import flask_restful.reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.repository_model
import terrareg.auth_wrapper
import terrareg.csrf


class GithubRefreshNamespace(ErrorCatchingResource):
    """Interface to refresh repositories for a namespaces from a provider source"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _post_arg_parser(self):
        """Return arg parser"""
        parser = flask_restful.reqparse.RequestParser()
        parser.add_argument(
            'namespace',
            type=str,
            required=True,
            help='Namespace to refresh repositories for',
            location='json'
        )
        parser.add_argument(
            'csrf_token',
            type=str,
            required=True,
            help='CSRF token',
            location='json'
        )
        return parser

    def _post(self, provider_source):
        """Refresh repositories for given namespace."""
        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return self._get_404_response()

        # Parse arguments
        args = self._post_arg_parser().parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        namespace_obj = terrareg.models.Namespace.get(name=args.namespace, include_redirect=False)
        if namespace_obj is None:
            return self._get_404_response()

        provider_source_obj.refresh_namespace_repositories(namespace=namespace_obj)

        return [], 200
