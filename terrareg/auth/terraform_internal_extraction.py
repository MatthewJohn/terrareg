
from .base_terraform_static_token import BaseTerraformStaticToken
import terrareg.config


class TerraformInternalExtractionAuthMethod(BaseTerraformStaticToken):
    """Auth method for handling Terraform authentication for internal extraction"""

    @classmethod
    def get_valid_terraform_tokens(cls):
        """Obtain list of valid tokens"""
        config = terrareg.config.Config()
        return [config.INTERNAL_EXTRACTION_ANALYTICS_TOKEN] if config.INTERNAL_EXTRACTION_ANALYTICS_TOKEN else []

    def get_username(self):
        """Return username"""
        return "Terraform internal extraction"

    def should_record_terraform_analytics(self):
        """Whether Terraform downloads by the user should be recorded"""
        return False
