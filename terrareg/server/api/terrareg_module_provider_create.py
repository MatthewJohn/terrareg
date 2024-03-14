
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf
import terrareg.errors
import terrareg.models
import terrareg.database


class ApiTerraregModuleProviderCreate(ErrorCatchingResource):
    """Provide interface to create module provider."""

    method_decorators = [
        terrareg.auth_wrapper.auth_wrapper('check_namespace_access',
            terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType.FULL,
            request_kwarg_map={'namespace': 'namespace'})]

    def _post_arg_parser(self):
        """Return arg parrser for POST request"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'git_provider_id', type=str,
            required=False,
            default=None,
            help='ID of the git provider to associate to module provider.',
            location='json'
        )
        parser.add_argument(
            'repo_base_url_template', type=str,
            required=False,
            default=None,
            help='Templated base git URL.',
            location='json'
        )
        parser.add_argument(
            'repo_clone_url_template', type=str,
            required=False,
            default=None,
            help='Templated git clone URL.',
            location='json'
        )
        parser.add_argument(
            'repo_browse_url_template', type=str,
            required=False,
            default=None,
            help='Templated URL for browsing repository.',
            location='json'
        )
        parser.add_argument(
            'git_tag_format', type=str,
            required=False,
            default=None,
            help='Module provider git tag format.',
            location='json'
        )
        parser.add_argument(
            'git_path', type=str,
            required=False,
            default=None,
            help='Path within git repository that the module exists.',
            location='json'
        )
        parser.add_argument(
            'csrf_token', type=str,
            required=False,
            help='CSRF token',
            location='json',
            default=None
        )
        return parser

    def _post(self, namespace, name, provider):
        """Handle update to settings."""
        parser = self._post_arg_parser()

        args = parser.parse_args()

        terrareg.csrf.check_csrf_token(args.csrf_token)

        # Update repository URL of module version
        namespace = terrareg.models.Namespace.get(name=namespace)
        if namespace is None:
            return {'message': 'Namespace does not exist'}, 400
        module = terrareg.models.Module(namespace=namespace, name=name)

        # Check if module provider already exists
        module_provider = terrareg.models.ModuleProvider.get(module=module, name=provider)
        if module_provider is not None:
            return {'message': 'Module provider already exists'}, 400

        with terrareg.database.Database.start_transaction() as transaction_context:
            module_provider = terrareg.models.ModuleProvider.create(module=module, name=provider)

            # If git provider ID has been specified,
            # validate it and update attribute of module provider.
            if args.git_provider_id is not None:
                git_provider = terrareg.models.GitProvider.get(id=args.git_provider_id)
                # If a non-empty git provider ID was provided and none
                # were returned, return an error about invalid
                # git provider ID
                if args.git_provider_id and git_provider is None:
                    transaction_context.transaction.rollback()
                    return {'message': 'Git provider does not exist.'}, 400

                module_provider.update_git_provider(git_provider=git_provider)

            # Ensure base repository URL is parsable
            repo_base_url_template = args.repo_base_url_template
            # If the argument is None, assume it's not being updated,
            # as this is the default value for the arg parser.
            if repo_base_url_template is not None:
                if repo_base_url_template == '':
                    # If repository URL is empty, set to None
                    repo_base_url_template = None

                try:
                    module_provider.update_repo_base_url_template(repo_base_url_template=repo_base_url_template)
                except terrareg.errors.RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repository base URL: {}'.format(str(exc))}, 400

            # Ensure repository URL is parsable
            repo_clone_url_template = args.repo_clone_url_template
            # If the argument is None, assume it's not being updated,
            # as this is the default value for the arg parser.
            if repo_clone_url_template is not None:
                if repo_clone_url_template == '':
                    # If repository URL is empty, set to None
                    repo_clone_url_template = None

                try:
                    module_provider.update_repo_clone_url_template(repo_clone_url_template=repo_clone_url_template)
                except terrareg.errors.RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repository clone URL: {}'.format(str(exc))}, 400

            # Ensure repository URL is parsable
            repo_browse_url_template = args.repo_browse_url_template
            if repo_browse_url_template is not None:
                if repo_browse_url_template == '':
                    # If repository URL is empty, set to None
                    repo_browse_url_template = None

                try:
                    module_provider.update_repo_browse_url_template(repo_browse_url_template=repo_browse_url_template)
                except terrareg.errors.RepositoryUrlParseError as exc:
                    transaction_context.transaction.rollback()
                    return {'message': 'Repository browse URL: {}'.format(str(exc))}, 400

            # Update git tag format of object
            git_tag_format = args.git_tag_format
            if git_tag_format is not None:
                module_provider.update_git_tag_format(git_tag_format=git_tag_format)

            # Update git path
            git_path = args.git_path
            if git_path is not None:
                module_provider.update_git_path(git_path=git_path)

        return {
            'id': module_provider.id
        }
