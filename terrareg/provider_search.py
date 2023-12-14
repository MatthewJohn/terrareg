
import datetime

import sqlalchemy
from terrareg.config import Config

from terrareg.database import Database
import terrareg.models
from terrareg.filters import NamespaceTrustFilter
import terrareg.result_data
import terrareg.provider_model


class ProviderSearch:

    @classmethod
    def _get_search_query_filter(cls, query: str):
        """Filter query based on wild-carded match of fields."""

        db = Database.get()
        wheres = []
        point_sum = None
        if query:
            for query_part in query.split():
                wildcarded_query_part = '%{0}%'.format(query_part)               
                point_value = sqlalchemy.cast(
                    sqlalchemy.case(
                            (db.provider.c.name.like(query_part), 20),
                            (db.namespace.c.namespace.like(query_part), 18),
                            (db.provider.c.description.like(query_part), 13),
                            (db.provider.c.name.like(wildcarded_query_part), 5),
                            (db.provider.c.description.like(wildcarded_query_part), 4),
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
                        db.provider.c.name.like(wildcarded_query_part),
                        db.provider.c.description.like(wildcarded_query_part),
                        db.namespace.c.namespace.like(wildcarded_query_part)
                    )
                )

        relevance = sqlalchemy.sql.expression.label('relevance', point_sum)
        select = db.select_provider_joined_latest_provider_version(
            db.provider,
            db.provider_version,
            db.namespace,
            db.provider.c.name.label('provider_name'),
            db.provider_category.c.slug.label('provider_category_slug'),
            relevance
        )
        for where_ in wheres:
            select = select.where(where_)

        # Group by module provider ID
        select = select.group_by(
            db.provider.c.id
        ).order_by(
            sqlalchemy.desc(relevance)
        )

        return select

    @classmethod
    def search_providers(
        cls,
        offset: int,
        limit: int,
        query: str=None,
        namespaces: list=None,
        providers: list=None,
        categories: list=None,
        namespace_trust_filters: list=NamespaceTrustFilter.UNSPECIFIED) -> terrareg.result_data.ResultData:

        # Limit the limits
        limit = 50 if limit > 50 else limit
        limit = 1 if limit < 1 else limit
        offset = 0 if offset < 0 else offset

        db = Database.get()

        select = cls._get_search_query_filter(query)

        # If provider has been supplied, select by that
        if providers:
            select = select.where(
                db.provider.c.name.in_(providers)
            )

        # If namespace has been supplied, select by that
        if namespaces:
            select = select.where(
                db.namespace.c.namespace.in_(namespaces)
            )

        if categories:
            select = select.where(
                db.provider_category.c.slug.in_(categories)
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
            db.provider.c.name
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
                module_providers.append(terrareg.provider_model.Provider(namespace=namespace, name=r['provider_name']))

        return terrareg.result_data.ResultData(
            offset=offset,
            limit=limit,
            rows=module_providers,
            count=count
        )


    @classmethod
    def get_search_filters(cls, query):
        """Get list of search filters and filter counts."""
        db = Database.get()
        main_select = cls._get_search_query_filter(query)

        with db.get_connection() as conn:
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

            main_select_subquery = main_select.subquery()
            category_res = conn.execute(
                sqlalchemy.select(
                    [sqlalchemy.func.count().label('count'), main_select_subquery.c.provider_category_slug]
                ).select_from(
                    main_select_subquery
                ).group_by(main_select_subquery.c.provider_category_slug)
            )

            namespace_subquery = main_select.group_by(
                db.namespace.c.namespace,
                db.provider.c.name
            ).subquery()
            namespace_res = conn.execute(
                sqlalchemy.select(
                    [sqlalchemy.func.count().label('count'), namespace_subquery.c.namespace]
                ).select_from(
                    namespace_subquery
                ).group_by(namespace_subquery.c.namespace)
            )

            return {
                'trusted_namespaces': trusted_count,
                'contributed': contributed_count,
                'provider_categories': {
                    r['provider_category_slug']: r['count']
                    for r in category_res
                },
                'namespaces': {
                    r['namespace']: r['count']
                    for r in namespace_res
                }
            }

