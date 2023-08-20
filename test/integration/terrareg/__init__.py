
import unittest.mock

from terrareg.auth import AdminApiKeyAuthMethod

from test import BaseTest
import terrareg.database
from .test_data import integration_test_data, integration_git_providers


class TerraregIntegrationTest(BaseTest):

    _TEST_DATA = integration_test_data
    _GIT_PROVIDER_DATA = integration_git_providers

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        return 'temp-integration.db'

    @classmethod
    def setup_class(cls):
        """Setup class method"""
        super(TerraregIntegrationTest, cls).setup_class()

        # Mock get_current_auth_method, which is used when
        # creating audit events.
        cls._get_current_auth_method_mock = unittest.mock.patch(
            'terrareg.auth.AuthFactory.get_current_auth_method',
            return_value=AdminApiKeyAuthMethod())
        cls._get_current_auth_method_mock.start()

    @classmethod
    def teardown_class(cls):
        """Teardown class"""
        cls._get_current_auth_method_mock.stop()

        super(TerraregIntegrationTest, cls).teardown_class()

    def _delete_audit_events(self):
        """Delete all audit events"""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.audit_history.delete())
