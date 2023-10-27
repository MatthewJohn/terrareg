from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.csrf
import terrareg.auth
import terrareg.models
import terrareg.namespace_type
import terrareg.provider_source.factory
import terrareg.repository_model
import terrareg.provider_category_model
import terrareg.provider_model
import terrareg.database


class GithubRepositoryPublishProvider(ErrorCatchingResource):
    """Interface publish a repository as a provider"""

    # @TODO Add permissions

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

        if not (auth_method := terrareg.auth.GithubAuthMethod.get_current_instance()):
            return self._get_401_response()


        provider_category = terrareg.provider_category_model.ProviderCategoryFactory.get().get_provider_category_by_pk(pk=args.category_id)
        if not provider_category:
            return {'errors': ['Provider Category does not exist']}, 400

        repository = terrareg.repository_model.Repository.get_by_provider_source_and_provider_id(provider_source=provider_source_obj, provider_id=repository_id)
        if not repository:
            return {'errors': ['Repository does not exist']}, 404


        with terrareg.database.Database.start_transaction() as transaction:
            provider = terrareg.provider_model.Provider.create(
                repository=repository,
                provider_category=provider_category
            )
            if not provider:
                transaction.transaction.rollback()
                return {'errors': ['An error occurred whilst creating provider']}, 500

            current_session = terrareg.auth.AuthFactory.get_current_session()
            if not current_session:
                transaction.transaction.rollback()
                return {'errors': ['An internal error accessing token occurred']}, 500

            versions = provider.refresh_versions()
            if not versions:
                transaction.transaction.rollback()
                return {'errors': ['No valid releases found for provider']}, 400

        return {
            "name": provider.name,
            "namespace": provider.namespace.name
        }, 200
