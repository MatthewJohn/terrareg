
from flask_restful import reqparse, inputs

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper


class ApiTerraregModuleProviderVersions(ErrorCatchingResource):
    """Interface to obtain module provider versions"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get(self, namespace, name, provider):
        """Return list of module versions for module provider"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'include-beta', type=inputs.boolean,
            default=False, help='Whether to include beta versions',
            location='args',
            dest='include_beta'
        )
        parser.add_argument(
            'include-unpublished', type=inputs.boolean,
            default=False, help='Whether to include unpublished versions',
            location='args',
            dest='include_unpublished'
        )
        args = parser.parse_args()

        namespace, module, module_provider, error = self.get_module_provider_by_names(namespace, name, provider)
        if error:
            return error

        return [
            {
                'version': module_version.version,
                'published': module_version.published,
                'beta': module_version.beta
            } for module_version in module_provider.get_versions(
                include_beta=args.include_beta, include_unpublished=args.include_unpublished)
        ]
