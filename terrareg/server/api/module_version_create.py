
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models
import terrareg.database
import terrareg.module_extractor


class ApiModuleVersionCreate(ErrorCatchingResource):
    """
    Provide interface to create release for git-backed modules.

    **DEPRECATION NOTICE**

    This API maybe removed in future.
    This deprecation is still in discussion.

    Consider migrating to '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/import'
    """

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_upload_module_version', request_kwarg_map={'namespace': 'namespace'})]

    def _post(self, namespace, name, provider, version):
        """Handle creation of module version."""
        with terrareg.database.Database.start_transaction():
            _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
            if error:
                return error[0], 400

            # Ensure that the module provider has a repository url configured.
            if not module_provider.get_git_clone_url():
                return {'message': 'Module provider is not configured with a repository'}, 400

            # Check if module provider can be indexed using a version
            if not module_provider.can_index_by_version:
                return {
                    'status': 'Error',
                    'message': 'Module provider is not configured with a git tag format containing a {version} placeholder. '
                               'Indexing must be performed using the git_tag argument'
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
