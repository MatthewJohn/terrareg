
import datetime

import sqlalchemy
from terrareg.config import TRUSTED_NAMESPACES

from terrareg.database import Database
import terrareg.models
from terrareg.filters import NamespaceTrustFilter


class ModuleSearchResults(object):
    """Object containing search results."""

    @property
    def module_providers(self):
        """Return module providers."""
        return self._module_providers

    @property
    def count(self):
        """Return count."""
        return self._count

    def __init__(self, module_providers: list, count: str):
        """Store member variables"""
        self._module_providers = module_providers
        self._count = count


class ModuleSearch(object):

    @classmethod
    def _get_search_query_filter(cls, select: sqlalchemy.sql.selectable.Select, query: str):
        """Filter query based on wildcarded match of fields."""

        db = Database.get()
        if query:
            for query_part in query.split():
                wildcarded_query_part = '%{0}%'.format(query_part)
                select = select.where(
                    sqlalchemy.or_(
                        db.module_provider.c.namespace.like(query_part),
                        db.module_provider.c.module.like(wildcarded_query_part),
                        db.module_provider.c.provider.like(query_part),
                        db.module_version.c.version.like(query_part),
                        db.module_version.c.description.like(wildcarded_query_part),
                        db.module_version.c.owner.like(wildcarded_query_part)
                    ),
                    db.module_version.c.published == True
                )
        return select

    @classmethod
    def search_module_providers(
        cls,
        offset: int,
        limit: int,
        query: str=None,
        namespace: str=None,
        provider: str=None,
        verified: bool=False,
        namespace_trust_filters: list=NamespaceTrustFilter.UNSPECIFIED):

        db = Database.get()
        select = db.select_module_version_joined_module_provider()

        select = cls._get_search_query_filter(select, query)

        # If provider has been supplied, select by that
        if provider:
            select = select.where(
                db.module_provider.c.provider == provider
            )

        # If namespace has been supplied, select by that
        if namespace:
            select = select.where(
                db.module_provider.c.namespace == namespace
            )

        # Filter by verified modules, if requested
        if verified:
            select = select.where(
                db.module_provider.c.verified == True
            )

        if namespace_trust_filters is not NamespaceTrustFilter.UNSPECIFIED:
            or_query = []
            if NamespaceTrustFilter.TRUSTED_NAMESPACES in namespace_trust_filters:
                or_query.append(db.module_provider.c.namespace.in_(tuple(TRUSTED_NAMESPACES)))
            if NamespaceTrustFilter.CONTRIBUTED in namespace_trust_filters:
                or_query.append(~db.module_provider.c.namespace.in_(tuple(TRUSTED_NAMESPACES)))
            select = select.where(sqlalchemy.or_(*or_query))


        # Group by and order by namespace, module and provider
        select = select.group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).order_by(
            db.module_provider.c.namespace.asc(),
            db.module_provider.c.module.asc(),
            db.module_provider.c.provider.asc()
        )

        limited_search = select.limit(limit).offset(offset)
        count_search = sqlalchemy.select(sqlalchemy.func.count).select_from(select)

        conn = db.get_engine().connect()
        res = conn.execute(limited_search)

        count_result = conn.execute(count_search)
        count = count_result.fetchone()['count']

        module_providers = []
        for r in res:
            namespace = terrareg.models.Namespace(name=r['namespace'])
            module = terrareg.models.Module(namespace=namespace, name=r['module'])
            module_providers.append(terrareg.models.ModuleProvider(module=module, name=r['provider']))

        return ModuleSearchResults(
            module_providers=module_providers,
            count=count
        )

    @classmethod
    def get_search_filters(cls, query):
        """Get list of search filters and filter counts."""
        db = Database.get()
        conn = db.get_engine().connect()
        select = db.select_module_version_joined_module_provider()

        main_select = cls._get_search_query_filter(select, query)

        verified_count = conn.execute(
            sqlalchemy.select(
                [sqlalchemy.func.count().label('count')]
            ).select_from(
                main_select.where(
                    db.module_provider.c.verified==True
                ).subquery()
            )
        ).fetchone()['count']

        trusted_count = conn.execute(
            sqlalchemy.select(
                [sqlalchemy.func.count().label('count')]
            ).select_from(
                main_select.where(
                    db.module_provider.c.namespace.in_(tuple(TRUSTED_NAMESPACES))
                ).subquery()
            )
        ).fetchone()['count']

        contributed_count = conn.execute(
            sqlalchemy.select(
                [sqlalchemy.func.count().label('count')]
            ).select_from(
                main_select.where(
                    ~db.module_provider.c.namespace.in_(tuple(TRUSTED_NAMESPACES))
                ).subquery()
            )
        ).fetchone()['count']

        provider_subquery = main_select.group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).subquery()
        provider_res = conn.execute(
            sqlalchemy.select(
                [sqlalchemy.func.count().label('count'), provider_subquery.c.provider]
            ).select_from(
                provider_subquery
            ).group_by(provider_subquery.c.provider)
        )

        return {
            'verified': verified_count,
            'trusted_namespaces': trusted_count,
            'contributed': contributed_count,
            'providers': {
                r['provider']: r['count']
                for r in provider_res
            }
        }

    @staticmethod
    def get_most_recently_published():
        """Return module with most recent published date."""
        db = Database.get()
        select = db.select_module_version_joined_module_provider().where(
        ).order_by(db.module_version.c.published_at.desc(), 
        ).limit(1)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        row = res.fetchone()

        # If there are no rows, return None
        if not row:
            return None

        namespace = terrareg.models.Namespace(name=row['namespace'])
        module = terrareg.models.Module(namespace=namespace,
                                        name=row['module'])
        module_provider = terrareg.models.ModuleProvider(module=module,
                                                         name=row['provider'])
        return terrareg.models.ModuleVersion(module_provider=module_provider,
                                             version=row['version'])

    @staticmethod
    def get_most_downloaded_module_provider_this_Week():
        """Obtain module provider with most downloads this week."""
        db = Database.get()
        conn = db.get_engine().connect()
        counts = sqlalchemy.select(
            [
                sqlalchemy.func.count().label('download_count'),
                db.module_provider.c.namespace,
                db.module_provider.c.module,
                db.module_provider.c.provider
            ]
        ).select_from(
            db.analytics
        ).join(
            db.module_version,
            db.module_version.c.id == db.analytics.c.parent_module_version
        ).join(
            db.module_provider,
            db.module_provider.c.id == db.module_version.c.module_provider_id
        ).where(
            db.analytics.c.timestamp >= (
                datetime.datetime.now() -
                datetime.timedelta(days=7)
            )
        ).group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).subquery()

        select = counts.select(
        ).order_by(counts.c.download_count.desc()
        ).limit(1)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        row = res.fetchone()

        # If there are no rows, return None
        if not row:
            return None

        namespace = terrareg.models.Namespace(name=row['namespace'])
        module = terrareg.models.Module(namespace=namespace,
                                        name=row['module'])
        return terrareg.models.ModuleProvider(module=module,
                                              name=row['provider'])
