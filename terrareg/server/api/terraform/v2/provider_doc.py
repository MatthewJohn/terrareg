
from flask_restful import reqparse

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.auth_wrapper
import terrareg.provider_version_model
import terrareg.provider_version_documentation_model
import terrareg.provider_documentation_type


class ApiV2ProviderDoc(ErrorCatchingResource):
    """Interface for obtain provider doc details"""

    method_decorators = [terrareg.auth_wrapper.auth_wrapper('can_access_read_api')]

    def _get_arg_parser(self):
        """Return argument parser for endpoint"""
        arg_parser = reqparse.RequestParser()
        arg_parser.add_argument(
            'output',
            type=str,
            location='args',
            default='md',
            dest='output',
            help='Content output type, either "html" or "md"'
        )
        return arg_parser

    def _get(self, doc_id):
        """
        Obtain details about provider document
        """
        args = self._get_arg_parser().parse_args()

        document = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_pk(
            pk=doc_id
        )

        if not document:
            return self._get_404_response()

        return {
            "data": document.get_v2_api_details(html=(args.output == "html"))
        }
