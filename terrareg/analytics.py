
import re
import datetime
from distutils.version import StrictVersion
from telnetlib import AUTHENTICATION


import sqlalchemy

from terrareg.database import Database
from terrareg.config import ANALYTICS_AUTH_KEYS


class AnalyticsEngine:

    @staticmethod
    def _join_filter_analytics_table_by_module_provider(db, query, module_provider):
        """Join query to module_version table and filter by module_version."""
        return query.join(
            db.module_version,
            db.module_version.c.id == db.analytics.c.parent_module_version
        ).join(
            db.module_provider,
            db.module_version.c.module_provider_id == db.module_provider.c.id
        ).where(
            db.module_provider.c.namespace == module_provider._module._namespace.name,
            db.module_provider.c.module == module_provider._module.name,
            db.module_provider.c.provider == module_provider.name
        )

    @staticmethod
    def check_auth_token(auth_token):
        """Check if auth token matches required environment analaytics tokens."""
        # If no analytics tokens have been defined, always return to be recorded with empty environment
        if not ANALYTICS_AUTH_KEYS:
            return True, None

        # If there is one auth token provided and does not contain an environment name
        # check and return empty environment
        if len(ANALYTICS_AUTH_KEYS) == 1 and len(ANALYTICS_AUTH_KEYS[0].split(':')) == 1:
            # Check if token matches provided token
            return (ANALYTICS_AUTH_KEYS[0] == auth_token), None

        # Otherwise compile list of environment and return appropriate environment
        for analytics_auth_key in ANALYTICS_AUTH_KEYS:
            analytics_auth_key_split = analytics_auth_key.split(':')
            if analytics_auth_key_split[0] == auth_token:
                return True, analytics_auth_key_split[1]

        # Default to returning to not authenticate
        return False, None

    @staticmethod
    def record_module_version_download(
        module_version,
        analytics_token: str,
        terraform_version: str,
        user_agent: str,
        auth_token: str):
        """Store information about module version download in database."""

        # If terraform version not present from header,
        # attempt to determine from user agent
        if not terraform_version:
            user_agent_match = re.match(r'^Terraform/(\d+\.\d+\.\d+)$', user_agent)
            if user_agent_match:
                terraform_version = user_agent_match.group(1)

        # Check if token is valid and whether it matches a deployment environment to be
        # recorded.
        record, environment = AnalyticsEngine.check_auth_token(auth_token)
        if not record:
            return

        print('Moudule {0} downloaded by {1} using terraform {2}'.format(
            module_version.id, analytics_token, terraform_version))

        # Insert analytics details into DB
        db = Database.get()
        conn = db.get_engine().connect()
        insert_statement = db.analytics.insert().values(
            parent_module_version=module_version.pk,
            timestamp=datetime.datetime.now(),
            terraform_version=terraform_version,
            analytics_token=analytics_token,
            auth_token=auth_token,
            environment=environment
        )
        conn.execute(insert_statement)

    def get_total_downloads():
        """Return number of downloads for a given module version."""
        db = Database.get()
        conn = db.get_engine().connect()
        select = sqlalchemy.select(
            [sqlalchemy.func.count()]
        ).select_from(
            db.analytics
        )
        res = conn.execute(select)
        return res.scalar()

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
            )
            select = AnalyticsEngine._join_filter_analytics_table_by_module_provider(
                db=db, query=select, module_provider=module_provider)

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

        id_subquery = sqlalchemy.select(
            [
                sqlalchemy.func.max(db.analytics.c.id)
            ]
        ).select_from(
            db.analytics
        )
        id_subquery = AnalyticsEngine._join_filter_analytics_table_by_module_provider(
            db=db, query=id_subquery,
            module_provider=module_provider)

        id_subquery = id_subquery.group_by(
            db.analytics.c.analytics_token,
            db.analytics.c.environment
        ).subquery()

        select = sqlalchemy.select([
            db.analytics.c.environment,
            db.analytics.c.analytics_token,
            db.module_version.c.version,
            db.analytics.c.terraform_version,
            db.analytics.c.timestamp
        ]).select_from(
            db.analytics
        )
        select = AnalyticsEngine._join_filter_analytics_table_by_module_provider(
            db=db, query=select, module_provider=module_provider)
        select = select.where(db.analytics.c.id.in_(id_subquery))

        res = conn.execute(select)

        for r in res:
            print(','.join([a for a in r.values() if a]))
        return {}
        token_version_mapping = {}
        for row in res:
            token = row[0] if row[0] else 'No token provided'
            if token not in token_version_mapping:
                token_version_mapping[token] = {
                    'terraform_version': None,
                    'module_version': None,
                    'environment': None
                }
            terraform_version = row[2] if row[2] else '0.0.0'

            # Check if module version is higher than 
            if (not token_version_mapping[token]['module_version'] or
                    StrictVersion(row[1]) > StrictVersion(token_version_mapping[token]['module_version'])):
                token_version_mapping[token]['module_version'] = row[1]
            if (not token_version_mapping[token]['terraform_version'] or
                    StrictVersion(terraform_version) > StrictVersion(token_version_mapping[token]['terraform_version'])):
                token_version_mapping[token]['terraform_version'] = terraform_version

        return token_version_mapping
