
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.server.error_catching_resource import api_error
import terrareg.models
import terrareg.auth_wrapper
import terrareg.csrf
from terrareg.errors import (
    GitProviderInUseError,
    GitProviderManagedByConfigurationError,
    InvalidGitProviderConfigError,
    RepositoryUrlParseError,
)


class ApiTerraregGitProviders(ErrorCatchingResource):
    """Interface to obtain git provider configurations."""

    method_decorators = {
        'get': [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')],
        'post': [terrareg.auth_wrapper.auth_wrapper('is_admin')]
    }

    def _post_arg_parser(self):
        """Return argument parser for create request."""
        parser = reqparse.RequestParser()
        parser.add_argument('name', type=str, required=True, location='json')
        parser.add_argument('base_url_template', type=str, required=True, location='json')
        parser.add_argument('clone_url_template', type=str, required=True, location='json')
        parser.add_argument('browse_url_template', type=str, required=True, location='json')
        parser.add_argument('git_path_template', type=str, required=False, location='json', default=None)
        parser.add_argument('csrf_token', type=str, required=False, location='json', default=None)
        return parser

    def _get(self):
        """Return list of git providers"""
        return [
            {
                "id": git_provider.pk,
                "name": git_provider.name,
                "base_url_template": git_provider.base_url_template,
                "clone_url_template": git_provider.clone_url_template,
                "browse_url_template": git_provider.browse_url_template,
                "git_path_template": git_provider.git_path_template,
            }
            for git_provider in terrareg.models.GitProvider.get_all()
        ]

    def _post(self):
        """Create a git provider."""
        args = self._post_arg_parser().parse_args()
        terrareg.csrf.check_csrf_token(args.csrf_token)

        try:
            git_provider = terrareg.models.GitProvider.create(
                name=args.name,
                base_url_template=args.base_url_template,
                clone_url_template=args.clone_url_template,
                browse_url_template=args.browse_url_template,
                git_path_template=args.git_path_template,
            )
        except (GitProviderManagedByConfigurationError, InvalidGitProviderConfigError, RepositoryUrlParseError) as exc:
            return api_error(str(exc)), 400

        return {
            "id": git_provider.pk,
            "name": git_provider.name,
            "base_url_template": git_provider.base_url_template,
            "clone_url_template": git_provider.clone_url_template,
            "browse_url_template": git_provider.browse_url_template,
            "git_path_template": git_provider.git_path_template,
        }, 201


class ApiTerraregGitProvider(ErrorCatchingResource):
    """Interface to update and delete git provider configurations."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('is_admin')]

    def _post_arg_parser(self):
        """Return argument parser for update request."""
        parser = reqparse.RequestParser()
        parser.add_argument('name', type=str, required=True, location='json')
        parser.add_argument('base_url_template', type=str, required=True, location='json')
        parser.add_argument('clone_url_template', type=str, required=True, location='json')
        parser.add_argument('browse_url_template', type=str, required=True, location='json')
        parser.add_argument('git_path_template', type=str, required=False, location='json', default=None)
        parser.add_argument('csrf_token', type=str, required=False, location='json', default=None)
        return parser

    def _post(self, git_provider_id):
        """Update an existing git provider."""
        args = self._post_arg_parser().parse_args()
        terrareg.csrf.check_csrf_token(args.csrf_token)

        git_provider = terrareg.models.GitProvider.get(id=git_provider_id)
        if git_provider is None:
            return {'message': 'Git provider does not exist.'}, 400

        try:
            git_provider.update(
                name=args.name,
                base_url_template=args.base_url_template,
                clone_url_template=args.clone_url_template,
                browse_url_template=args.browse_url_template,
                git_path_template=args.git_path_template,
            )
        except (GitProviderManagedByConfigurationError, InvalidGitProviderConfigError, RepositoryUrlParseError) as exc:
            return api_error(str(exc)), 400

        return {
            "id": git_provider.pk,
            "name": git_provider.name,
            "base_url_template": git_provider.base_url_template,
            "clone_url_template": git_provider.clone_url_template,
            "browse_url_template": git_provider.browse_url_template,
            "git_path_template": git_provider.git_path_template,
        }

    def _delete(self, git_provider_id):
        """Delete an existing git provider."""
        git_provider = terrareg.models.GitProvider.get(id=git_provider_id)
        if git_provider is None:
            return {'message': 'Git provider does not exist.'}, 400

        try:
            git_provider.delete()
        except (GitProviderInUseError, GitProviderManagedByConfigurationError) as exc:
            return api_error(str(exc)), 400

        return {}, 204
