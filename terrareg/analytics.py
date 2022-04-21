
import re
import datetime
from distutils.version import StrictVersion
from telnetlib import AUTHENTICATION


import sqlalchemy

from terrareg.database import Database
from terrareg.config import ANALYTICS_AUTH_KEYS


class AnalyticsEngine:

    _ARE_TOKENS_ENABLED = None
    _ARE_ENVIRONMENTS_ENABLED = None
    _TOKEN_ENVIRONMENT_MAPPING = None

    @classmethod
    def are_tokens_enabled(cls):
        """Determine if tokens are enabled."""
        if AnalyticsEngine._ARE_TOKENS_ENABLED is None:
            AnalyticsEngine._ARE_TOKENS_ENABLED = bool(ANALYTICS_AUTH_KEYS)
        return AnalyticsEngine._ARE_TOKENS_ENABLED

    @classmethod
    def are_environments_enabled(cls):
        """Determine if token environments are enabled."""
        if AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED is None:
            AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED = (
                AnalyticsEngine.are_tokens_enabled() and
                not (len(ANALYTICS_AUTH_KEYS) == 1 and len(ANALYTICS_AUTH_KEYS[0].split(':')) == 1)
            )
        return AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED

    @classmethod
    def get_token_environment_mapping(cls):
        """Determine if token environments are enabled."""
        if AnalyticsEngine._TOKEN_ENVIRONMENT_MAPPING is None:
            AnalyticsEngine._TOKEN_ENVIRONMENT_MAPPING = {
                analytics_auth_key.split(':')[0]: analytics_auth_key.split(':')[1]
                for analytics_auth_key in ANALYTICS_AUTH_KEYS
            } if AnalyticsEngine.are_environments_enabled() else {}
        return AnalyticsEngine._TOKEN_ENVIRONMENT_MAPPING

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
        if not AnalyticsEngine.are_tokens_enabled():
            return True, None

        # If there is one auth token provided and does not contain an environment name
        # check and return empty environment
        if not AnalyticsEngine.are_environments_enabled():
            # Check if token matches provided token
            return (ANALYTICS_AUTH_KEYS[0] == auth_token), None

        # Otherwise check if auth token is for an environment
        if auth_token in AnalyticsEngine.get_token_environment_mapping():
            return True, AnalyticsEngine.get_token_environment_mapping()[auth_token]

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

        # Obtain a list of the MAX (latest) analytics row IDs,
        # grouped by analytics token and environment.
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

        # Select all required fields for the given IDs
        # obtained from subquery.
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

        token_version_mapping = {}
        # Convert list of environments to a map,
        # providing the environment priority (index of the environment
        # in the environment list).
        environment_priorities = {
            env.split(':')[0]: itx
            for itx, env in enumerate(ANALYTICS_AUTH_KEYS)
        }

        for row in res:

            # Check if row is usable
            ## Skip any rows without ananlytics tokens, if they are required.
            if AnalyticsEngine.are_tokens_enabled() and not row['analytics_token']:
                continue
            ## Skip any rows without an environment, if they are required.
            if AnalyticsEngine.are_environments_enabled() and not row['environment']:
                continue

            token = row['analytics_token'] if row['analytics_token'] else 'No token provided'

            # Populate map with empty details for this analytics token,
            # if it doesn't already exist.
            if token not in token_version_mapping:
                token_version_mapping[token] = {
                    'terraform_version': None,
                    'module_version': None,
                    'environment': None
                }
            terraform_version = row['terraform_version'] if row['terraform_version'] else '0.0.0'

            # Check if environment of current download is 'higher than' previous
            ## Use this row if the current 'highest' value has an empty environment (this will always
            ## match the first row)
            if (token_version_mapping[token]['environment'] is None or
                ## Ignore any future rows with an empty environment. If there aren't
                ## environments in use, there will only be one row per analytics token
                (row['environment'] is not None and
                 ## Ensure that the environment (still) exists
                 row['environment'] in environment_priorities and
                 ## Ensure the environment token appears higher in the
                 ## environment priorities than the current 'highest' row
                 environment_priorities[row['environment']] >
                 environment_priorities[token_version_mapping[token]['environment']])):

                token_version_mapping[token]['environment'] = row['environment']
                token_version_mapping[token]['module_version'] = row['version']
                token_version_mapping[token]['terraform_version'] = terraform_version

        return token_version_mapping
