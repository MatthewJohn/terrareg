
import datetime

import sqlalchemy

from terrareg.database import Database
import terrareg.auth


class AuditEvent:
    
    @classmethod
    def get_events(cls, limit=10, offset=0, descending=True,
                    order_by='timestamp', query=None):
        """Return audit events from database"""
        db = Database.get()
        db_query = sqlalchemy.select(
            db.audit_history
        ).select_from(
            db.audit_history
        )
        filtered = db_query

        # If query string has been provided,
        # match any rows where any of the columns match
        if query:
            filtered = filtered.where(
                sqlalchemy.or_(
                    db.audit_history.c.username.like(f'%{query}%'),
                    db.audit_history.c.action.like(f'%{query}%'),
                    db.audit_history.c.object_id.like(f'%{query}%'),
                    db.audit_history.c.old_value.like(f'%{query}%'),
                    db.audit_history.c.new_value.like(f'%{query}%')
                )
            )

        # Convert name of audit by column to column attribute of table
        order_by_column = getattr(db.audit_history.c, order_by)

        # If ordering by action, which is an enum, cast column
        # to char, as enum is ordered by the order of the enum definitions, not
        # alphabetically
        if order_by_column == db.audit_history.c.action:
            order_by_column = order_by_column.cast(sqlalchemy.CHAR)

        # Create query with ordering, limit and offset applied
        filtered_limit = filtered.order_by(
            sqlalchemy.desc(order_by_column) if descending else sqlalchemy.asc(order_by_column)
        )
        # If timestamp was not originally requested to order by,
        # use timestamp in descender order as secondary ordering
        if order_by != "timestamp":
            filtered_limit = filtered_limit.order_by(
                sqlalchemy.desc(db.audit_history.c.timestamp)
            )

        filtered_limit = filtered_limit.limit(
            limit
        ).offset(
            offset
        )

        # Create query counting all rows in table
        total_count_query = sqlalchemy.select(
            sqlalchemy.func.count('*').label('count')
        ).select_from(db_query.subquery())

        # Create query counting all rows with just filtering
        filtered_count_query = sqlalchemy.select(
            sqlalchemy.func.count('*').label('count')
        ).select_from(filtered.subquery())

        with db.get_connection() as conn:
            res = conn.execute(filtered_limit)
            res = res.fetchall()
            filtered_count = conn.execute(filtered_count_query).fetchone()['count']
            total_count = conn.execute(total_count_query).fetchone()['count']

        return res, total_count, filtered_count

    @classmethod
    def create_audit_event(cls, action,
                           object_type, object_id,
                           old_value, new_value):
        """Create audit event"""
        # Insert audit event into DB
        db = Database.get()
        insert_statement = db.audit_history.insert().values(
            username=terrareg.auth.AuthFactory().get_current_auth_method().get_username(),
            action=action,
            object_type=object_type,
            object_id=object_id,
            old_value=old_value,
            new_value=new_value,
            timestamp=datetime.datetime.now()
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)
