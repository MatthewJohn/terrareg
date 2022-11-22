
import datetime

import sqlalchemy

from terrareg.database import Database
import terrareg.auth


class AuditEvent:
    
    @classmethod
    def get_events(cls):
        """Return audit events from database"""
        db = Database.get()
        query = sqlalchemy.select(
            db.audit_history
        ).select_from(
            db.audit_history
        )

        with db.get_connection() as conn:
            res = conn.execute(query).fetchall()
        return res

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
