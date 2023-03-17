
from typing import List

from flask_restful import reqparse, fields
from flask_restful_swagger import swagger

from terrareg.server.error_catching_resource import ErrorCatchingResource
import terrareg.models
import terrareg.module_search

@swagger.model
class TerraformModuleOutline:
    resource_fields = {
        "id": fields.String, "namespace": fields.String, "name": fields.String, "provider": fields.String, "verified": fields.Boolean, "trusted": fields.Boolean
    }
    def __init__(self, id: str=None, namespace: str=None, name: str=None, provider: str=None, verified: bool=False, trusted: bool=False):
        pass

@swagger.model
class MetaResponseModel:

    resource_fields = {
        "limit": fields.Integer(),
        "current_offset": fields.Integer,
        "next_offset": fields.Integer
    }

    def __init__(self, limit: int, current_offset: int, next_offset: int):
        pass

@swagger.model
@swagger.nested(
    meta=MetaResponseModel.__name__,
    modules=TerraformModuleOutline.__name__,
)
class TerraformSearchResponseModel:

    resource_fields = {
        "meta": fields.Nested(MetaResponseModel.resource_fields),
        "modules": fields.List(fields.Nested(TerraformModuleOutline.resource_fields))
    }

    def __init__(self, meta: MetaResponseModel, modules: List[TerraformModuleOutline]):
        pass


class ApiModuleDetails(ErrorCatchingResource):
    """Provide interface to module providers within a module"""

    @swagger.operation(
        responseClass=TerraformSearchResponseModel,
        parameters=[
            {
                "name": "offset",
                "description": "Offset of results",
                "type": "int",
                # "required": False,
                "paramType": "query",
                "dataType": "int"
            },
            {
                "name": "limit",
                "description": "Result limit",
                # "required": False,
                "paramType": "query",
                "dataType": "int"
            },
        ],
        responseMessages=[
            {
                "code": 200,
                "message": "Modules found."
            },
            {
                "code": 404,
                "message": "Namespace does not exist or no modules found."
            }
        ]
    )
    def get(self, namespace, name):
        """Return latest version for each module provider."""

        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int, location='args',
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int, location='args',
            default=10, help='Pagination limit'
        )
        args = parser.parse_args()

        namespace, _ = terrareg.models.Namespace.extract_analytics_token(namespace)

        search_results = terrareg.module_search.ModuleSearch.search_module_providers(
            offset=args.offset,
            limit=args.limit,
            namespaces=[namespace],
            modules=[name]
        )

        if not search_results.module_providers:
            return self._get_404_response()

        return {
            "meta": search_results.meta,
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in search_results.module_providers
            ]
        }

