
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.user_group_namespace_permission_type
import terrareg.csrf
import terrareg.errors
import terrareg.models
import terrareg.database


class ApiTerraregModuleVersionExtractionStatus(ErrorCatchingResource):
    """Provide interface to get status of module version extraction."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper(
        'can_upload_module_version',
        request_kwarg_map={'namespace': 'namespace'})
    ]

    def _post(self, namespace, name, provider, request_id):
        """Handle request to get module version extraction data."""
        _, _, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        extraction_status = terrareg.models.ModuleVersionExtractionStatus.get_by_request_id(module_provider, request_id)
        if not extraction_status:
            return {'message': 'Extraction request does not exist'}, 400

        return extraction_status.json, 200
