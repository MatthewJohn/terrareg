
from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models
import terrareg.database
import terrareg.module_extractor


class ApiModuleVersionCreate(ErrorCatchingResource):
    """Provide interface to create release for git-backed modules."""

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

            module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

            module_version.prepare_module()
            with terrareg.module_extractor.GitModuleExtractor(module_version=module_version) as me:
                me.process_upload()

            return {
                'status': 'Success'
            }

