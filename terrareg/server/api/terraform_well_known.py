
from flask_restful import Resource

from terrareg.constants import TERRAFORM_REDIRECT_URI_PORT_RANGE
import terrareg.terraform_idp


class ApiTerraformWellKnown(Resource):
    """Terraform .well-known discovery"""

    def get(self):
        """Return wellknown JSON"""
        data = {
            "modules.v1": "/v1/modules/",
        }
        if terrareg.terraform_idp.TerraformIdp.get().is_enabled:
            data["login.v1"] = {
                "client": "terraform-cli",
                "grant_types": ["authz_code", "token"],
                "authz": "/terraform/oauth/authorization",
                "token": "/terraform/oauth/token",
                "ports": TERRAFORM_REDIRECT_URI_PORT_RANGE,
            }
        return data
