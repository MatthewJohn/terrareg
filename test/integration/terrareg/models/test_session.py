
import datetime
import secrets
from unittest import mock
import pytest
from terrareg.database import Database

from terrareg.models import Namespace, Session
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

class TestSession(TerraregIntegrationTest):
    """Test Session model class"""

    def test_create_session(self):
        """Test creating a session."""
        db = Database.get()
        with db.get_connection() as conn:
            # Delete any pre-existing sessions
            conn.execute(db.session.delete())

        with mock.patch('terrareg.config.Config.ADMIN_SESSION_EXPIRY_MINS', 60):
            session_obj = Session.create_session()

        assert session_obj is not None
        assert isinstance(session_obj, Session)

        # Check session ID is a truthful and is a string
        assert session_obj.id
        assert isinstance(session_obj.id, str)

        # Ensure session ID is in database
        with db.get_connection() as conn:
            res = conn.execute(db.session.select().where(
                db.session.c.id==session_obj.id
            ))
            row = res.fetchone()
        assert row

        # Check expiry is about the correct time (within 2 minutes)
        assert row['expiry'] > (datetime.datetime.now() + datetime.timedelta(minutes=58))
        assert row['expiry'] < (datetime.datetime.now() + datetime.timedelta(minutes=60))

    def test_delete_session(self):
        """Test deleting a session."""
        db = Database.get()
        session_id = secrets.token_urlsafe(Session.SESSION_ID_LENGTH)

        with db.get_connection() as conn:
            conn.execute(db.session.insert().values(
                id=session_id,
                expiry=(datetime.datetime.now() + datetime.timedelta(minutes=1))
            ))
        
        # Create session object
        session_obj = Session.check_session(session_id)
        assert session_obj

        # Use delete method to remove session
        session_obj.delete()

        # Ensure row no longer present in databases
        with db.get_connection() as conn:
            res = conn.execute(db.session.select().where(
                db.session.c.id==session_id
            ))
            row = res.fetchone()
        assert row is None

        # Ensure session can no longer be obtained
        session_obj = Session.check_session(session_id)
        assert session_obj is None

    def test_checking_valid_session(self):
        """Test checking a valid session."""
        # Insert session into database
        db = Database.get()
        session_id = secrets.token_urlsafe(Session.SESSION_ID_LENGTH)

        with db.get_connection() as conn:
            conn.execute(db.session.insert().values(
                id=session_id,
                expiry=(datetime.datetime.now() + datetime.timedelta(minutes=1))
            ))
        
        # Check session ID using check_session method
        session_obj = Session.check_session(session_id)
        # Ensure a valid instance of Session is returned
        assert session_obj is not None
        assert isinstance(session_obj, Session)

        # Ensure session ID matches
        assert session_obj.id == session_id

    def test_checking_expired_session(self):
        """Test checking an expired session"""
        # Insert session into database
        db = Database.get()
        session_id = secrets.token_urlsafe(Session.SESSION_ID_LENGTH)

        with db.get_connection() as conn:
            conn.execute(db.session.insert().values(
                id=session_id,
                expiry=(datetime.datetime.now() - datetime.timedelta(minutes=1))
            ))

        # Check session ID using check_session method
        session_obj = Session.check_session(session_id)

        # Ensure a no session object is returned
        assert session_obj is None

    @pytest.mark.parametrize('session_id', [
        # Empty values
        None,
        '',

        # Non-exist in valid form
        'NMBfiFLW3EQjHVFZlM5T6Tomzcj3fEGe87Hc1u38afA',

        # Invalid characters
        '`"\'@$#}+='
    ])
    def test_checking_invalid_session(self, session_id):
        """Test checking invalid session IDs"""
        assert Session.check_session(session_id) == None

    def test_cleanup_old_sessions(self):
        """Test cleaning up old sessions."""
        db = Database.get()

        with db.get_connection() as conn:
            conn.execute(db.session.insert().values(
                [
                    {'id': 'expiredsessionid', 'expiry': datetime.datetime.now() - datetime.timedelta(minutes=10)},
                    {'id': 'notexpired', 'expiry': datetime.datetime.now() + datetime.timedelta(minutes=1)}
                ]
            ))

        with mock.patch('terrareg.config.Config.ADMIN_SESSION_EXPIRY_MINS', 5):
            Session.cleanup_old_sessions()

        with db.get_connection() as conn:
            rows = conn.execute(db.session.select()).fetchall()

        assert len(rows) == 1
        assert rows[0]['id'] == 'notexpired'
