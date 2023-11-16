
import re
import datetime


import sqlalchemy

from terrareg.database import Database
from terrareg.config import Config
import terrareg.models


class AnalyticsEngine:

    _ARE_TOKENS_ENABLED = None
    _ARE_ENVIRONMENTS_ENABLED = None
    _TOKEN_ENVIRONMENT_MAPPING = None

    DEFAULT_ENVIRONMENT_NAME = 'Default'

    @classmethod
    def get_datetime_now(cls):
        """Return datetime now"""
        return datetime.datetime.now()

    @classmethod
    def are_tokens_enabled(cls):
        """Determine if tokens are enabled."""
        if AnalyticsEngine._ARE_TOKENS_ENABLED is None:
            AnalyticsEngine._ARE_TOKENS_ENABLED = bool(Config().ANALYTICS_AUTH_KEYS)
        return AnalyticsEngine._ARE_TOKENS_ENABLED

    @classmethod
    def are_environments_enabled(cls):
        """Determine if token environments are enabled."""
        if AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED is None:
            AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED = (
                AnalyticsEngine.are_tokens_enabled() and
                not (len(Config().ANALYTICS_AUTH_KEYS) == 1 and len(Config().ANALYTICS_AUTH_KEYS[0].split(':')) == 1)
            )
        return AnalyticsEngine._ARE_ENVIRONMENTS_ENABLED

    @classmethod
    def get_token_environment_mapping(cls):
        """Determine if token environments are enabled."""
        if AnalyticsEngine._TOKEN_ENVIRONMENT_MAPPING is None:
            AnalyticsEngine._TOKEN_ENVIRONMENT_MAPPING = {
                analytics_auth_key.split(':')[0]: analytics_auth_key.split(':')[1]
                for analytics_auth_key in Config().ANALYTICS_AUTH_KEYS
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
        ).join(
            db.namespace,
            db.module_provider.c.namespace_id == db.namespace.c.id
        ).where(
            db.module_provider.c.id == module_provider.pk
        )

    @staticmethod
    def get_environment_from_token(auth_token):
        """Check if auth token matches required environment analytics tokens."""
        # If no analytics tokens have been defined, return default environment
        if not AnalyticsEngine.are_tokens_enabled():
            return AnalyticsEngine.DEFAULT_ENVIRONMENT_NAME

        # If there is one auth token provided and does not contain an environment name
        # check and return empty environment
        if not AnalyticsEngine.are_environments_enabled():
            # Check if token matches provided token,
            # if so, return default environment name
            return (AnalyticsEngine.DEFAULT_ENVIRONMENT_NAME
                    if (Config().ANALYTICS_AUTH_KEYS[0] == auth_token) else
                    None)

        # Otherwise check if auth token is for an environment
        if auth_token in AnalyticsEngine.get_token_environment_mapping():
            return AnalyticsEngine.get_token_environment_mapping()[auth_token]

        # Default to returning when not using a valid environment
        return None

    @staticmethod
    def record_module_version_download(
        namespace_name: str,
        module_name: str,
        provider_name: str,
        module_version,
        analytics_token: str,
        terraform_version: str,
        user_agent: str,
        auth_token: str):
        """Store information about module version download in database."""

        # If Terraform version not present from header,
        # attempt to determine from user agent
        if not terraform_version:
            user_agent_match = re.match(r'^Terraform/(\d+\.\d+\.\d+)$', user_agent)
            if user_agent_match:
                terraform_version = user_agent_match.group(1)

        # Obtain environment from auth token.
        # If auth token is not provided, 
        environment = AnalyticsEngine.get_environment_from_token(auth_token)

        # Insert analytics details into DB
        db = Database.get()
        insert_statement = db.analytics.insert().values(
            parent_module_version=module_version.pk,
            timestamp=AnalyticsEngine.get_datetime_now(),
            terraform_version=terraform_version,
            analytics_token=analytics_token,
            auth_token=auth_token,
            environment=environment,
            namespace_name=namespace_name,
            module_name=module_name,
            provider_name=provider_name
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)

    def get_total_downloads():
        """Return number of downloads for a given module version."""
        db = Database.get()
        select = sqlalchemy.select(
            [sqlalchemy.func.count()]
        ).select_from(
            db.analytics
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            return res.scalar()

    @staticmethod
    def get_global_module_usage_base_query(include_empty_auth_token=False):
        """Return base query for getting all analytics tokens."""
        db = Database.get()
        # Initial query to select all analytics joined to module version and module provider
        select = sqlalchemy.select(
            db.module_provider.c.id,
            db.namespace.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider,
            db.analytics.c.analytics_token
        ).select_from(
            db.analytics
        ).join(
            db.module_version,
            db.analytics.c.parent_module_version == db.module_version.c.id
        ).join(
            db.module_provider,
            db.module_version.c.module_provider_id == db.module_provider.c.id
        ).join(
            db.namespace,
            db.module_provider.c.namespace_id==db.namespace.c.id
        ).where(
            # Filter unpublished and beta versions
            db.module_version.c.published == True,
            db.module_version.c.beta == False
        )
        # Filter rows with empty auth token, if including them is not enabled
        if not include_empty_auth_token:
            select = select.where(
                db.analytics.c.auth_token != None
            )

        # Group select by analytics token and module provider ID
        select = select.group_by(
            db.analytics.c.analytics_token,
            db.module_provider.c.id
        )
        return select

    @staticmethod
    def get_global_module_usage_counts(include_empty_auth_token=False):
        """Return all analytics tokens, grouped by module provider."""
        db = Database.get()
        # Initial query to select all analytics joined to module version and module provider
        select = AnalyticsEngine.get_global_module_usage_base_query(
            include_empty_auth_token=include_empty_auth_token).subquery()

        # Select the count for each module provider, grouping by the analytics ID
        count_select = sqlalchemy.select(
            sqlalchemy.func.count(),
            select.c.namespace,
            select.c.module,
            select.c.provider
        ).select_from(
            select
        ).group_by(
            select.c.id
        )

        with db.get_connection() as conn:
            res = conn.execute(count_select)
            data = {
                f"{row['namespace']}/{row['module']}/{row['provider']}": row['count']
                for row in res
            }
        return data

    def get_module_version_total_downloads(module_version):
        """Return number of downloads for a given module version."""
        db = Database.get()
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
        with db.get_connection() as conn:
            res = conn.execute(select)
            return res.scalar()

    @staticmethod
    def get_module_provider_download_stats(module_provider):
        """Return number of downloads for intervals."""
        db = Database.get()
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
                from_timestamp = AnalyticsEngine.get_datetime_now() - datetime.timedelta(days=i[0])
                select = select.where(
                    db.analytics.c.timestamp >= from_timestamp
                )

            with db.get_connection() as conn:
                res = conn.execute(select)
                stats[i[1]] = res.scalar()

        return stats

    @staticmethod
    def check_module_provider_redirect_usage(module_provider_redirect):
        """Check for any analytics using a redirect's name"""
        # Create list of source namespace names, with actual namespace name
        # and any redirects
        namespace_names = [
            module_provider_redirect.namespace.name
        ] + [
            namespace_redirect.name
            for namespace_redirect in terrareg.models.NamespaceRedirect.get_by_namespace(module_provider_redirect.namespace)
        ]

        # Obtain all tokens
        db = Database.get()

        # Get all analytics, join to module version and module provider to filter by module_provider_id of
        # the module provider redirect. Group by each analytics token, getting the latest (max ID) for each
        id_subquery = sqlalchemy.select(
            sqlalchemy.func.max(db.analytics.c.id).label('latest_id'),
        ).select_from(
            db.analytics
        ).join(
            db.module_version,
            db.module_version.c.id == db.analytics.c.parent_module_version
        ).join(
            db.module_provider,
            db.module_version.c.module_provider_id == db.module_provider.c.id
        ).where(
            db.module_provider.c.id == module_provider_redirect.module_provider_id,
        ).group_by(
            db.analytics.c.analytics_token
        ).subquery()

        # Pass this query into as a sub-query to filter those analytics that are
        # using the redirect details
        filter_query = sqlalchemy.select(
            db.analytics.c.analytics_token,
            db.analytics.c.timestamp,
            db.analytics.c.provider_name,
            db.analytics.c.namespace_name
        ).select_from(
            db.analytics
        ).join(
            id_subquery,
            id_subquery.c.latest_id==db.analytics.c.id
        ).where(
            db.analytics.c.module_name==module_provider_redirect.module_name,
            db.analytics.c.provider_name==module_provider_redirect.provider_name,
            db.analytics.c.namespace_name.in_(namespace_names)
        )

        # If look-back days has been configured, limit the query
        # to timestamps more recent than the cut-off
        lookback_days = Config().REDIRECT_DELETION_LOOKBACK_DAYS
        if lookback_days >= 0:
            filter_query = filter_query.where(
                db.analytics.c.timestamp>=(AnalyticsEngine.get_datetime_now() - datetime.timedelta(days=lookback_days))
            )

        with db.get_connection() as conn:
            res = conn.execute(filter_query).all()

        return res

    @staticmethod
    def get_module_provider_token_versions(module_provider):
        """Return list of users for module provider."""
        db = Database.get()

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
        )

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

        token_version_mapping = {}
        # Convert list of environments to a map,
        # providing the environment priority (index of the environment
        # in the environment list).
        environment_priorities = {
            env.split(':')[1]: itx
            for itx, env in enumerate(Config().ANALYTICS_AUTH_KEYS)
        }

        with db.get_connection() as conn:
            res = conn.execute(select)

            for row in res:

                # Check if row is usable
                ## Skip any rows without analytics tokens, if they are required.
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

    @classmethod
    def delete_analytics_for_module_version(cls, module_version):
        """Delete all analytics for given module version."""
        db = Database.get()

        with db.get_connection() as conn:
            conn.execute(db.analytics.delete().where(
                db.analytics.c.parent_module_version == module_version.pk
            ))

    @classmethod
    def migrate_analytics_to_new_module_version(cls, old_version_version_pk, new_module_version):
        """Migrate all analytics for old module version ID to new module version."""
        db = Database.get()

        with db.get_connection() as conn:
            conn.execute(db.analytics.update().where(
                db.analytics.c.parent_module_version == old_version_version_pk
            ).values(
                parent_module_version=new_module_version.pk
            ))

    @classmethod
    def get_module_provider_version_statistics(cls):
        """Return number of major, minor and patch releases for a module version"""
        major_count = 0
        minor_count = 0
        patch_count = 0
        for namespace in terrareg.models.Namespace.get_all(only_published=True).rows:
            for module in namespace.get_all_modules():
                for module_provider in module.get_providers():
                    versions = module_provider.get_versions(include_beta=False, include_unpublished=False)
                    # Reverse versions, so that the they increase in value
                    versions.reverse()

                    # Setup variable to hold previous version
                    previous_version = None
                    for module_version in versions:
                        # Split version number by . and convert each version part to integers
                        version_split = [int(v) for v in module_version.version.split('.')]
                        # If this is the first version, count as a major release,
                        # otherwise, check if major version has increased since last seen release
                        if previous_version is None or version_split[0] > previous_version[0]:
                            major_count += 1
                        # Check if version is a minor change
                        elif version_split[1] > previous_version[1]:
                            minor_count += 1
                        # Check if version is a patch change
                        elif version_split[2] > previous_version[2]:
                            patch_count += 1
                        else:
                            print('Unable to determine version change between:', previous_version, 'and', version_split)

                        previous_version = version_split

        # Return all 3 counts
        return major_count, minor_count, patch_count

    @classmethod
    def get_prometheus_metrics(cls):
        """Return Prometheus metrics for modules and usage."""
        prometheus_generator = PrometheusGenerator()
        
        module_count_metric = PrometheusMetric(
            name='module_providers_count',
            type_='counter',
            help='Total number of module providers with a published version'
        )
        module_count_metric.add_data_row(value=terrareg.models.ModuleProvider.get_total_count(only_published=True))
        prometheus_generator.add_metric(module_count_metric)

        major_count, minor_count, patch_count = cls.get_module_provider_version_statistics()
        version_major_count_metric = PrometheusMetric(
            name='module_version_major_count',
            type_='counter',
            help='Total number of major versions released'
        )
        version_major_count_metric.add_data_row(value=major_count)
        prometheus_generator.add_metric(version_major_count_metric)
        version_minor_count_metric = PrometheusMetric(
            name='module_version_minor_count',
            type_='counter',
            help='Total number of minor versions released'
        )
        version_minor_count_metric.add_data_row(value=minor_count)
        prometheus_generator.add_metric(version_minor_count_metric)
        version_patch_count_metric = PrometheusMetric(
            name='module_version_patch_count',
            type_='counter',
            help='Total number of patch versions released'
        )
        version_patch_count_metric.add_data_row(value=patch_count)
        prometheus_generator.add_metric(version_patch_count_metric)

        module_provider_usage_metric = PrometheusMetric(
            'module_provider_usage',
            type_='counter',
            help='Analytics tokens used in a module provider'
        )
        db = Database.get()
        with db.get_connection() as conn:
            rows = conn.execute(AnalyticsEngine.get_global_module_usage_base_query(include_empty_auth_token=True)).fetchall()
        for row in rows:
            module_provider_usage_metric.add_data_row(
                value='1',
                labels={'module_provider_id': '{}/{}/{}'.format(row['namespace'], row['module'], row['provider']),
                        'analytics_token': row['analytics_token']}
            )
        prometheus_generator.add_metric(module_provider_usage_metric)

        return prometheus_generator.generate()


class PrometheusMetric:
    """Prometheus metric"""

    def __init__(self, name, type_, help):
        """Store member variables and initialise help and type lines."""
        self._name = name
        self._type = type_
        self._help = help
        self._lines = [
            f'# HELP {self._name} {self._help}',
            f'# TYPE {self._name} {self._type}'
        ]

    def add_data_row(self, value, labels=None):
        """Add data row, with optional labels"""
        labels = {} if labels is None else labels
        label_strings = [f'{key}="{labels[key]}"' for key in labels]
        label_string = ', '.join(label_strings)
        if label_string:
            label_string = '{' + label_string + '}'

        self._lines.append(f'{self._name}{label_string} {value}')

    def generate(self):
        """Return generated lines for metric."""
        return self._lines

class PrometheusGenerator:
    """Generate Prometheus metrics output"""

    def __init__(self):
        """Initialise empty data"""
        self._lines = []

    def add_metric(self, metric: PrometheusMetric):
        """Add metrics from a PromethiusMetric object."""
        self._lines += metric.generate()

    def generate(self):
        """Generate output for Prometheus metrics"""
        return '\n'.join(self._lines)
