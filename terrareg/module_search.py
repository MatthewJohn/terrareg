
import datetime

import sqlalchemy
from terrareg.config import Config

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

    @property
    def meta(self):
        """Return API meta for limit/offsets."""
        # Setup base metadata with current offset and limit
        meta_data = {
            "limit": self._limit,
            "current_offset": self._offset,
        }

        # If current offset is not 0,
        # Set previous offset as current offset minus the current limit,
        # or 0, depending on whichever is higher.
        if self._offset > 0:
            meta_data['prev_offset'] = (self._offset - self._limit) if (self._offset >= self._limit) else 0

        # If the current count of results is greater than the next offset,
        # provide the next offset in the metadata
        next_offset = (self._offset + self._limit)
        if self.count > next_offset:
            meta_data['next_offset'] = next_offset

        return meta_data

    def __init__(self, offset: int, limit: int, module_providers: list, count: str):
        """Store member variables"""
        self._offset = offset
        self._limit = limit
        self._module_providers = module_providers
        self._count = count


class ModuleSearch(object):

    @classmethod
    def _get_search_query_filter(cls, query: str):
        """Filter query based on wildcarded match of fields."""

        db = Database.get()
        wheres = []
        point_sum = None
        if query:
            for query_part in query.split():

                wildcarded_query_part = '%{0}%'.format(query_part)               
                point_value = sqlalchemy.cast(
                    sqlalchemy.case(
                            (db.module_provider.c.module.like(query_part), 20),
                            (db.namespace.c.namespace.like(query_part), 18),
                            (db.module_provider.c.provider.like(query_part), 14),
                            (db.module_version.c.description.like(query_part), 13),
                            (db.module_version.c.owner.like(query_part), 12),
                            (db.module_provider.c.module.like(wildcarded_query_part), 5),
                            (db.module_version.c.description.like(wildcarded_query_part), 4),
                            (db.module_version.c.owner.like(wildcarded_query_part), 3),
                            (db.namespace.c.namespace.like(wildcarded_query_part), 2),
                        else_=0
                    ),
                    sqlalchemy.Integer
                )
                if point_sum is None:
                    point_sum = point_value
                else:
                    point_sum += point_value
                wheres.append(
                    sqlalchemy.or_(
                        db.module_provider.c.provider.like(query_part),
                        db.module_provider.c.module.like(wildcarded_query_part),
                        db.module_version.c.description.like(wildcarded_query_part),
                        db.module_version.c.owner.like(wildcarded_query_part),
                        db.namespace.c.namespace.like(wildcarded_query_part)
                    )
                )

        relevance = sqlalchemy.sql.expression.label('relevance', point_sum)
        select = db.select_module_provider_joined_latest_module_version(
            db.module_provider,
            db.module_version,
            db.namespace,
            relevance
        )
        for where_ in wheres:
            select = select.where(where_)

        # Filter search by published module versions,
        # remove beta versions
        # and group by module provider ID
        select = select.where(
            db.module_version.c.published == True,
            db.module_version.c.beta == False
        ).group_by(
            db.module_provider.c.id
        ).order_by(
            sqlalchemy.desc(relevance)
        )

        return select

    @classmethod
    def search_module_providers(
        cls,
        offset: int,
        limit: int,
        query: str=None,
        namespaces: list=None,
        modules: list=None,
        providers: list=None,
        verified: bool=False,
        include_internal: bool=False,
        namespace_trust_filters: list=NamespaceTrustFilter.UNSPECIFIED):

        # Limit the limits
        limit = 50 if limit > 50 else limit
        limit = 1 if limit < 1 else limit
        offset = 0 if offset < 0 else offset

        db = Database.get()

        select = cls._get_search_query_filter(query)

        # If provider has been supplied, select by that
        if providers:
            select = select.where(
                db.module_provider.c.provider.in_(providers)
            )

        # If namespace has been supplied, select by that
        if namespaces:
            select = select.where(
                db.namespace.c.namespace.in_(namespaces)
            )

        # If namespace has been supplied, select by that
        if modules:
            select = select.where(
                db.module_provider.c.module.in_(modules)
            )

        # Filter by verified modules, if requested
        if verified:
            select = select.where(
                db.module_provider.c.verified == True
            )

        # Filter internal modules, if not being included
        if not include_internal:
            select = select.where(
                db.module_version.c.internal == False
            )

        if namespace_trust_filters is not NamespaceTrustFilter.UNSPECIFIED:
            or_query = []
            if NamespaceTrustFilter.TRUSTED_NAMESPACES in namespace_trust_filters:
                or_query.append(db.namespace.c.namespace.in_(tuple(Config().TRUSTED_NAMESPACES)))
            if NamespaceTrustFilter.CONTRIBUTED in namespace_trust_filters:
                or_query.append(~db.namespace.c.namespace.in_(tuple(Config().TRUSTED_NAMESPACES)))
            select = select.where(sqlalchemy.or_(*or_query))


        # Group by and order by namespace, module and provider
        select = select.group_by(
            db.namespace.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        )

        limited_search = select.limit(limit).offset(offset)
        count_search = sqlalchemy.select(sqlalchemy.func.count().label('count')).select_from(select.subquery())

        with db.get_connection() as conn:
            res = conn.execute(limited_search)
            count_result = conn.execute(count_search)

            count = count_result.fetchone()['count']

            module_providers = []
            for r in res:
                namespace = terrareg.models.Namespace(name=r['namespace'])
                module = terrareg.models.Module(namespace=namespace, name=r['module'])
                module_providers.append(terrareg.models.ModuleProvider(module=module, name=r['provider']))

        return ModuleSearchResults(
            offset=offset,
            limit=limit,
            module_providers=module_providers,
            count=count
        )

    @classmethod
    def get_search_filters(cls, query):
        """Get list of search filters and filter counts."""
        db = Database.get()
        main_select = cls._get_search_query_filter(query)

        # Remove any internal modules
        main_select = main_select.where(
            db.module_version.c.internal == False
        )

        with db.get_connection() as conn:
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
                        db.namespace.c.namespace.in_(tuple(Config().TRUSTED_NAMESPACES))
                    ).subquery()
                )
            ).fetchone()['count']

            contributed_count = conn.execute(
                sqlalchemy.select(
                    [sqlalchemy.func.count().label('count')]
                ).select_from(
                    main_select.where(
                        ~db.namespace.c.namespace.in_(tuple(Config().TRUSTED_NAMESPACES))
                    ).subquery()
                )
            ).fetchone()['count']

            provider_subquery = main_select.group_by(
                db.namespace.c.namespace,
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

            namespace_subquery = main_select.group_by(
                db.namespace.c.namespace,
                db.module_provider.c.module,
                db.module_provider.c.provider
            ).subquery()
            namespace_res = conn.execute(
                sqlalchemy.select(
                    [sqlalchemy.func.count().label('count'), namespace_subquery.c.namespace]
                ).select_from(
                    namespace_subquery
                ).group_by(namespace_subquery.c.namespace)
            )

            return {
                'verified': verified_count,
                'trusted_namespaces': trusted_count,
                'contributed': contributed_count,
                'providers': {
                    r['provider']: r['count']
                    for r in provider_res
                },
                'namespaces': {
                    r['namespace']: r['count']
                    for r in namespace_res
                }
            }

    @staticmethod
    def get_most_recently_published():
        """Return module with most recent published date."""
        db = Database.get()
        select = db.select_module_provider_joined_latest_module_version(
            db.module_version,
            db.module_provider,
            db.namespace
        ).where(
            db.module_version.c.published == True,
            db.module_version.c.beta == False,
            db.module_version.c.internal == False,
            db.module_version.c.extraction_complete == True
        ).order_by(db.module_version.c.published_at.desc(), 
        ).limit(1)

        with db.get_connection() as conn:
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
        counts = sqlalchemy.select(
            [
                sqlalchemy.func.count().label('download_count'),
                db.namespace.c.namespace,
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
        ).join(
            db.namespace,
            db.module_provider.c.namespace_id == db.namespace.c.id
        ).where(
            db.analytics.c.timestamp >= (
                datetime.datetime.now() -
                datetime.timedelta(days=7)
            ),
            db.module_version.c.published == True,
            db.module_version.c.beta == False,
            db.module_version.c.internal == False
        ).group_by(
            db.namespace.c.id,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).subquery()

        select = counts.select(
        ).order_by(
            counts.c.download_count.desc()
        ).limit(1)

        with db.get_connection() as conn:
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
