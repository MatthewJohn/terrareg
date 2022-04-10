
import re
import datetime

from terrareg.models import ModuleVersion
from terrareg.database import Database


class AnalyticsEngine:

    def record_module_version_download(
        module_version: ModuleVersion,
        analytics_token: str,
        terraform_version: str,
        user_agent: str):
        """Store information about module version download in database."""

        # If terraform version not present from header,
        # attempt to determine from user agent
        if not terraform_version:
            user_agent_match = re.match(r'^Terraform/(\d+\.\d+\.\d+)$', user_agent)
            if user_agent_match:
                terraform_version = user_agent_match.group(1)

        print('Moudule {0} downloaded by {1} using terraform {2}'.format(
            module_version.id, analytics_token, terraform_version))

        # Insert analytics details into DB
        db = Database.get()
        conn = db.get_engine().connect()
        insert_statement = db.analytics.insert().values(
            parent_module_version=module_version.pk,
            timestamp=datetime.datetime.now(),
            terraform_version=terraform_version,
            analytics_token=analytics_token
        )
        conn.execute(insert_statement)
