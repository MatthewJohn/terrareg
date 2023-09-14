
from flask_restful import Resource

from terrareg.constants import TERRAFORM_REDIRECT_URI_PORT_RANGE


class ApiTerraformWellKnown(Resource):
    """Terraform .well-known discovery"""

    def get(self):
        """Return wellknown JSON"""
        return {
            "modules.v1": "/v1/modules/",
            "login.v1": {
                "client": "terraform-cli",
                "grant_types": ["authz_code"],
                "authz": "/terraform/oauth/authorization",
                "token": "/terraform/oauth/token",
                "ports": TERRAFORM_REDIRECT_URI_PORT_RANGE,
            }
        }
