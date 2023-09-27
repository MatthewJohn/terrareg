
from .base_terraform_static_token import BaseTerraformStaticToken
import terrareg.config


class TerraformAnalyticsAuthKeyAuthMethod(BaseTerraformStaticToken):
    """Auth method for handling Terraform authentication using an 'analytics auth key' deployment token"""

    @classmethod
    def get_valid_terraform_tokens(cls):
        """Obtain list of valid tokens"""
        # Split each auth key 'xxxxxx:dev', 'yyyyyy:prod' by colon to obtain the auth key
        return [
            auth_key.split(':')[0]
            for auth_key in terrareg.config.Config().ANALYTICS_AUTH_KEYS
            if auth_key.split(':')[0]
        ]

    def get_username(self):
        """Return username"""
        return "Terraform deployment analytics token"
