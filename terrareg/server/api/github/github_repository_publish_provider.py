from flask_restful import reqparse
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.csrf
import terrareg.auth
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.repository_model
import terrareg.provider_category_model
import terrareg.provider_model
import terrareg.provider_tier
import terrareg.database
from terrareg.errors import NoGithubAppInstallationError
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type


def get_namespace_from_request_repository_id(provider_source, repository_id):
    """Return namespace from request repository ID"""
    provider_source = terrareg.provider_source.factory.ProviderSourceFactory().get_provider_source_by_api_name(api_name=provider_source)
    if not provider_source:
        return None
    repository = terrareg.repository_model.Repository.get_by_provider_source_and_provider_id(provider_source=provider_source, provider_id=repository_id)
    if not repository:
        return None
    return repository.owner

class GithubRepositoryPublishProvider(ErrorCatchingResource):
    """Interface publish a repository as a provider"""

    method_decorators = {
        "post": [
            terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
                terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
                # Obtain namespace for check from post data
                kwarg_values={"namespace": lambda provider_source, repository_id: get_namespace_from_request_repository_id(provider_source=provider_source, repository_id=repository_id)}
            )
        ],
    }

    def _post_arg_parser(self):
        """Return arg paraser for post method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'category_id',
            type=int, location='form',
            help='Provider category ID for provider',
            required=True
        )
        parser.add_argument(
            'csrf_token',
            type=str, location='form',
            help='CSRF Token',
            default=None
        )
        return parser

    def _post(self, provider_source, repository_id):
        """Publish repository provider."""
        args = self._post_arg_parser().parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        # Obtain provider source
        provider_source_factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        provider_source_obj = provider_source_factory.get_provider_source_by_api_name(provider_source)
        if not provider_source_obj:
            return self._get_404_response()

        # Ensure user is authenticated via Github OR is a site admin
        if not (github_auth_method := terrareg.auth.GithubAuthMethod.get_current_instance()) and not terrareg.auth.AuthFactory().get_current_auth_method().is_admin():
            return self._get_401_response()

        use_default_provider_source_auth = False
        if not github_auth_method and terrareg.auth.AuthFactory().get_current_auth_method().is_admin():
            use_default_provider_source_auth = True

        provider_category = terrareg.provider_category_model.ProviderCategoryFactory.get().get_provider_category_by_pk(pk=args.category_id)
        if not provider_category or not provider_category.user_selectable:
            return {'status': 'Error', 'message': 'Provider Category does not exist'}, 400

        repository = terrareg.repository_model.Repository.get_by_provider_source_and_provider_id(provider_source=provider_source_obj, provider_id=repository_id)
        if not repository:
            return {'status': 'Error', 'message': 'Repository does not exist'}, 404

        with terrareg.database.Database.start_transaction() as transaction:
            try:
                provider = terrareg.provider_model.Provider.create(
                    repository=repository,
                    provider_category=provider_category,
                    use_default_provider_source_auth=use_default_provider_source_auth,
                    tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
                )
            except NoGithubAppInstallationError:
                # Obtain installation URL before database transaction is rolled back
                app_installation_url = provider_source_obj.get_app_installation_url()
                transaction.transaction.rollback()
                return {'status': 'Error', 'message': 'no-app-installed', 'link': app_installation_url}, 400

            if not provider:
                transaction.transaction.rollback()
                return {'status': 'Error', 'message': 'An error occurred whilst creating provider'}, 500

            # Attempt to obtain 1 valid version
            try:
                versions = provider.refresh_versions(limit=1)
            except:
                transaction.transaction.rollback()
                return {'status': 'Error', 'message': 'An internal server error occurred'}, 500

            if not versions:
                transaction.transaction.rollback()
                return {'status': 'Error', 'message': 'No valid releases found for provider'}, 400

        return {
            "name": provider.name,
            "namespace": provider.namespace.name
        }, 200
