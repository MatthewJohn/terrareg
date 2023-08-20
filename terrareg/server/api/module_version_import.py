
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models
import terrareg.database
import terrareg.module_extractor


class ApiModuleVersionImport(ErrorCatchingResource):
    """
    Provide interface to import/index version for git-backed modules.
    """

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_upload_module_version', request_kwarg_map={'namespace': 'namespace'})]

    def _post_arg_parser(self):
        """Obtain argument parser for post request"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'version',
            type=str,
            required=False,
            default=None,
            location='json',
            help=(
                'The semantic version number of the module to be imported.'
                ' This can only be used if the git tag format of the module provider contains a {version} placeholder.'
                ' Conflicts with git_tag'
            )
        )
        parser.add_argument(
            'git_tag',
            type=str,
            required=False,
            default=None,
            location='json',
            help='The git tag of the module to be imported. Conflicts with version.'
        )
        return parser

    def _post(self, namespace, name, provider):
        """Handle creation of module version."""

        args = self._post_arg_parser().parse_args()

        if args.git_tag and args.version:
            return {'status': 'Error', 'message': 'git_tag and version are mutually exclusive - only one can be passed.'}, 400
        if not args.git_tag and not args.version:
            return {'status': 'Error', 'message': 'Either git_tag or version must be provided'}, 400

        with terrareg.database.Database.start_transaction():
            _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
            if error:
                return error[0], 400

            # Ensure that the module provider has a repository url configured.
            if not module_provider.get_git_clone_url():
                return {'status': 'Error', 'message': 'Module provider is not configured with a repository'}, 400

            if args.version and not module_provider.can_index_by_version:
                return {
                    'status': 'Error',
                    'message': 'Module provider is not configured with a git tag format containing a {version} placeholder. '
                               'Indexing must be performed using the git_tag argument'
                }, 400

            version = args.version
            if args.git_tag:
                version = module_provider.get_version_from_tag(tag=args.git_tag)
                if not version:
                    return {
                        'status': 'Error',
                        'message': 'Version could not be derrived from git tag. '
                                'Ensure it matches the git_tag_format template for this module provider'
                    }, 400

            module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

            previous_version_published = module_version.prepare_module()
            with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as me:
                me.process_upload()

            if previous_version_published:
                module_version.publish()

            return {
                'status': 'Success'
            }
