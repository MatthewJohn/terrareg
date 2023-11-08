
from datetime import datetime, timedelta
import unittest.mock
import pytest

import sqlalchemy

from terrareg.database import Database
from test.integration.terrareg import TerraregIntegrationTest
from terrareg.audit import AuditEvent
from terrareg.audit_action import AuditAction


class TestAuditEvent(TerraregIntegrationTest):
    """Test AuditEvent class"""

    def create_audit_event_test(
            self,
            audit_action=AuditAction.MODULE_PROVIDER_CREATE,
            old_value='old',
            new_value='new',
            username='test-user'):
        """Test create audit event"""
        # Delete any pre-existing audit events
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.audit_history.delete())

        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.get_username.return_value = username
        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method',
                                 return_value=mock_auth_method) as mock_get_auth_method:

            AuditEvent.create_audit_event(
                action=audit_action,
                object_type='unittest-object-type',
                object_id='unittest/object/id',
                old_value=old_value,
                new_value=new_value
            )

            mock_get_auth_method.assert_called_once_with()
            mock_auth_method.get_username.assert_called_once_with()


        # Obtain audit event from database and check values
        with db.get_connection() as conn:
            query = sqlalchemy.select(db.audit_history)
            res = conn.execute(query)
            rows = res.fetchall()
        assert len(rows) == 1
        audit_event = rows[0]

        assert audit_event.username == username
        assert audit_event.action == audit_action
        assert audit_event.object_type == 'unittest-object-type'
        assert audit_event.object_id == 'unittest/object/id'
        assert audit_event.old_value == (str(old_value) if old_value is not None else None)
        assert audit_event.new_value == (str(new_value) if new_value is not None else None)
        assert audit_event.timestamp >= (datetime.now() - timedelta(minutes=1))

    @pytest.mark.parametrize('audit_action', [
        (AuditAction.NAMESPACE_CREATE),
        (AuditAction.NAMESPACE_MODIFY_NAME),
        (AuditAction.NAMESPACE_MODIFY_DISPLAY_NAME),
        (AuditAction.NAMESPACE_DELETE),
        (AuditAction.MODULE_PROVIDER_CREATE),
        (AuditAction.MODULE_PROVIDER_DELETE),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_PROVIDER),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_PATH),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL),
        (AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL),
        (AuditAction.MODULE_PROVIDER_UPDATE_VERIFIED),
        (AuditAction.MODULE_PROVIDER_UPDATE_NAMESPACE),
        (AuditAction.MODULE_PROVIDER_UPDATE_MODULE_NAME),
        (AuditAction.MODULE_PROVIDER_UPDATE_PROVIDER_NAME),
        (AuditAction.MODULE_PROVIDER_REDIRECT_DELETE),
        (AuditAction.MODULE_VERSION_INDEX),
        (AuditAction.MODULE_VERSION_PUBLISH),
        (AuditAction.MODULE_VERSION_DELETE),
        (AuditAction.USER_GROUP_CREATE),
        (AuditAction.USER_GROUP_DELETE),
        (AuditAction.USER_GROUP_NAMESPACE_PERMISSION_ADD),
        (AuditAction.USER_GROUP_NAMESPACE_PERMISSION_MODIFY),
        (AuditAction.USER_GROUP_NAMESPACE_PERMISSION_DELETE),
        (AuditAction.USER_LOGIN),
        (AuditAction.GPG_KEY_CREATE),
        (AuditAction.GPG_KEY_DELETE),
        (AuditAction.PROVIDER_CREATE),
        (AuditAction.PROVIDER_DELETE),
        (AuditAction.PROVIDER_VERSION_INDEX),
        (AuditAction.PROVIDER_VERSION_DELETE),
        (AuditAction.REPOSITORY_CREATE),
        (AuditAction.REPOSITORY_UPDATE),
        (AuditAction.REPOSITORY_DELETE),
    ])
    def test_create_audit_event_audit_action(self, audit_action):
        """Test create audit event, testing audit_action field"""
        self.create_audit_event_test(audit_action=audit_action)

    @pytest.mark.parametrize('old_value', [
        None,
        '',
        'testvalue',
        0,
        1234,
        '1234'        
    ])
    def test_create_audit_event_old_value(self, old_value):
        """Test create audit event, testing old_value field"""
        self.create_audit_event_test(old_value=old_value)

    @pytest.mark.parametrize('new_value', [
        None,
        '',
        'testvalue',
        0,
        1234,
        '1234'
    ])
    def test_create_audit_event_new_value(self, new_value):
        """Test create audit event, testing new_value field"""
        self.create_audit_event_test(new_value=new_value)

    @pytest.mark.parametrize('username', [
        'testusername',
        'Built-in admin',
        '',
        None
    ])
    def test_create_audit_event_username(self, username):
        """Test create audit event, testing username field"""
        self.create_audit_event_test(username=username)
