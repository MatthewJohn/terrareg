
from flask_restful import Resource


class ApiTerraformWellKnown(Resource):
    """Terraform .well-known discovery"""

    def get(self):
        """Return wellknown JSON"""
        return {
            "modules.v1": "/v1/modules/"
        }
