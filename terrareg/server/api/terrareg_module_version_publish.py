
from terrareg.server.error_catching_resource import ErrorCatchingResource
from terrareg.auth_wrapper import auth_wrapper


class ApiTerraregModuleVersionPublish(ErrorCatchingResource):
    """Provide interface to publish module version."""

    method_decorators = [auth_wrapper('can_publish_module_version', request_kwarg_map={'namespace': 'namespace'})]

    def _post(self, namespace, name, provider, version):
        """Publish module."""
        _, _, _, module_version, error = self.get_module_version_by_name(namespace, name, provider, version)
        if error:
            return error

        module_version.publish()
        return {
            'status': 'Success'
        }
