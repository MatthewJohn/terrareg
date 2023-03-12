
from flask import request

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.models
import terrareg.database
import terrareg.module_extractor
import terrareg.module_version_create


class ApiModuleVersionCreate(ErrorCatchingResource):
    """Provide interface to create release for git-backed modules."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_upload_module_version', request_kwarg_map={'namespace': 'namespace'})]

    def _post(self, namespace, name, provider, version):
        """Handle creation of module version."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error[0], 400

        request_id = None
        # If a request that contains data that can be loaded
        # by json, extract request_id
        try:
            request_id = request.json().get('request_id')
        except:
            pass

        # Ensure that the module provider has a repository url configured.
        if not module_provider.get_git_clone_url():
            return {'message': 'Module provider is not configured with a repository'}, 400

        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version=version)

        with terrareg.module_version_create.module_version_create(module_version):
            with terrareg.module_extractor.GitModuleExtractor(module_version=module_version, request_id=request_id) as me:
                me.process_upload()

        return {
            'status': 'Success'
        }
