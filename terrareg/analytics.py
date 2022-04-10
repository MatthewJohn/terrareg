
import re

from terrareg.models import ModuleVersion

class AnalyticsEngine:

    def record_module_version_download(
        module_version: ModuleVersion,
        analytics_token: str,
        terraform_version: str,
        user_agent: str):

        # If terraform version not present from header,
        # attempt to determine from user agent
        if not terraform_version:
            user_agent_match = re.match(r'^Terraform/(\d+\.\d+\.\d+)$', user_agent)
            if user_agent_match:
                terraform_version = user_agent_match.group(1)

        print('Moudule {0} downloaded by {1} using terraform {2}'.format(
            module_version.id, analytics_token, terraform_version))
