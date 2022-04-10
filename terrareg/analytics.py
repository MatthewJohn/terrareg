
import re
import datetime
from distutils.version import StrictVersion


import sqlalchemy

from terrareg.database import Database


class AnalyticsEngine:

    def record_module_version_download(
        module_version,
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

    def get_module_version_total_downloads(module_version):
        """Return number of downloads for a given module version."""
        db = Database.get()
        conn = db.get_engine().connect()
        select = sqlalchemy.select(
            [sqlalchemy.func.count()]
        ).select_from(
            db.analytics
        ).join(
            db.module_version,
            db.module_version.c.id == db.analytics.c.parent_module_version
        ).where(
            db.module_version.c.id == module_version.pk
        )
        res = conn.execute(select)
        return res.scalar()

    @staticmethod
    def get_module_provider_download_stats(module_provider):
        """Return number of downloads for intervals."""
        db = Database.get()
        conn = db.get_engine().connect()
        stats = {}
        for i in [(7, 'week'), (31, 'month'), (365, 'year'), (None, 'total')]:

            select = sqlalchemy.select(
                [sqlalchemy.func.count()]
            ).select_from(
                db.analytics
            ).join(
                db.module_version,
                db.module_version.c.id == db.analytics.c.parent_module_version
            ).where(
                db.module_version.c.namespace == module_provider._module._namespace.name,
                db.module_version.c.module == module_provider._module.name,
                db.module_version.c.provider == module_provider.name
            )

            # If a checking a given time frame, limit by number of days
            if i[0]:
                from_timestamp = datetime.datetime.now() - datetime.timedelta(days=i[0])
                select = select.where(
                    db.analytics.c.timestamp >= from_timestamp
                )

            res = conn.execute(select)
            stats[i[1]] = res.scalar()

        return stats


    @staticmethod
    def get_module_provider_token_versions(module_provider):
        """Return list of users for module provider."""
        db = Database.get()
        conn = db.get_engine().connect()

        select = sqlalchemy.select(
            [db.analytics.c.analytics_token, db.module_version.c.version]
        ).select_from(
            db.analytics
        ).join(
            db.module_version,
            db.module_version.c.id == db.analytics.c.parent_module_version
        ).where(
            db.module_version.c.namespace == module_provider._module._namespace.name,
            db.module_version.c.module == module_provider._module.name,
            db.module_version.c.provider == module_provider.name
        ).group_by(
            db.analytics.c.analytics_token,
            db.module_version.c.version
        )

        res = conn.execute(select)

        token_version_mapping = {}
        for row in res:
            token = row[0] if row[0] else 'No token provided'
            if (token not in token_version_mapping or
                    StrictVersion(row[1]) > StrictVersion(token_version_mapping[token])):
                token_version_mapping[token] = row[1]

        return token_version_mapping