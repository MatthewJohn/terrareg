
import datetime
from terrareg.database import Database


class AuditEvent:
    
    @classmethod
    def create_audit_event(cls, username, action,
                           object_type, object_id,
                           old_value, new_value):
        """Create audit event"""
        # Insert audit event into DB
        db = Database.get()
        insert_statement = db.audit_history.insert().values(
            username=username,
            action=action,
            object_type=object_type,
            object_id=object_id,
            old_value=old_value,
            new_value=new_value,
            timestamp=datetime.datetime.now()
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)
