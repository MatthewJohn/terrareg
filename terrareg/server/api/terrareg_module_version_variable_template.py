
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper


class ApiTerraregModuleVersionVariableTemplate(ErrorCatchingResource):
    """Provide variable template for module version."""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get_arg_parser(self):
        """Return argument parser for GET method"""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'output',
            type=str,
            location='args',
            default='md',
            dest='output',
            help='Variable/Output description format, either "html" or "md"'
        )
        return parser

    def _get(self, namespace, name, provider, version):
        """Return variable template."""
        parser = self._get_arg_parser()
        args = parser.parse_args()

        _, _, _, module_version, error = self.get_module_version_by_name(
            namespace, name, provider, version)
        if error:
            return error
        return module_version.get_variable_template(html=(args.output == "html"))
