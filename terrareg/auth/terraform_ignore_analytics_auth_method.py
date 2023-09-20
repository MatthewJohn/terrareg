
from .base_terraform_static_token import BaseTerraformStaticToken
import terrareg.config


class TerraformIgnoreAnalyticsAuthMethod(BaseTerraformStaticToken):
    """Auth method for handling Terraform authentication using an 'ignore analytics' token"""

    @classmethod
    def get_valid_terraform_tokens(cls):
        """Obtain list of valid tokens"""
        return terrareg.config.Config().IGNORE_ANALYTICS_TOKEN_AUTH_KEYS

    def get_username(self):
        """Return username"""
        return "Terraform ignore analytics token"

    def should_record_terraform_analytics(self):
        """Whether Terraform downloads by the user should be recorded"""
        return False
