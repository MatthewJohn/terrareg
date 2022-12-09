
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import Select
import selenium

from terrareg.database import Database
from test.selenium import SeleniumTest
from terrareg.audit_action import AuditAction
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider


class TestAuditHistory(SeleniumTest):
    """Test audit_history page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def _setup_test_audit_data(cls):
        """Setup test audit data."""
        db = Database.get()
        with db.get_connection() as conn:
            for data in [
                ['test-event-admin', AuditAction.NAMESPACE_CREATE, 'Namespace', 'test-namespace', None, None,
                 datetime(year=2022, month=11, day=27, hour=19, minute=14, second=10)],
                ['test-event-admin', AuditAction.MODULE_PROVIDER_CREATE, 'Module', 'test-namespace/test-module', None, None,
                 datetime(year=2022, month=11, day=27, hour=19, minute=15, second=10)]
            ]:
                conn.execute(
                    db.audit_history.insert().values(
                        username=data[0],
                        action=data[1],
                        object_type=data[2],
                        object_id=data[3],
                        old_value=data[4],
                        new_value=data[5],
                        timestamp=data[6]
                    )
                )


    @classmethod
    def setup_class(cls):
        """Setup test audit data."""
        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))

        super(TestAuditHistory, cls).setup_class()

        cls._setup_test_audit_data()

    def test_navigation_from_homepage(self):
        """Test navigation user navbar to audit history page"""
        self.selenium_instance.get(self.get_url('/'))
        # Ensure Setting drop-down is not shown
        drop_down = self.wait_for_element(By.ID, 'navbarSettingsDropdown', ensure_displayed=False)
        assert drop_down.is_displayed() == False

        self.perform_admin_authentication('unittest-password')
        # Ensure Setting drop-down is shown
        drop_down = self.wait_for_element(By.ID, 'navbarSettingsDropdown')
        assert drop_down.is_displayed() == True
        # Move mouse to settings drop-down
        selenium.webdriver.ActionChains(self.selenium_instance).move_to_element(drop_down).perform()

        user_groups_button = drop_down.find_element(By.LINK_TEXT, 'Audit History')
        assert user_groups_button.text == 'Audit History'
        assert user_groups_button.is_displayed() == True

        user_groups_button.click()

        assert self.selenium_instance.current_url == self.get_url('/audit-history')

    def _ensure_audit_row_is_like(self, row, timestamp, username, action, object_id, old_value, new_value):
        """Ensure row matches expected data"""
        assert [timestamp, username, action, object_id, old_value, new_value] == [
            td.text
            for td in row.find_elements(By.TAG_NAME, 'td')
        ]


    def test_basic_view(self):
        """Ensure page shows basic audit history."""
        self.perform_admin_authentication('unittest-password')

        # Load homepage, waiting for drop-down to be rendered and navigate to audit history page
        self.selenium_instance.get(self.get_url('/audit-history'))

        assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Audit History'

        # Ensure table is shown and fields are present
        audit_table = self.selenium_instance.find_element(By.ID, 'audit-history-table')

        # Re-order table by timestamp desc
        #audit_table.find_element(By.XPATH, ".//th[text()='Timestamp']").click()

        rows = [r for r in audit_table.find_elements(By.TAG_NAME, 'tr')]
        print(rows)
        # Ignore first two audit events caused by login
        self._ensure_audit_row_is_like(
            rows[1],
            timestamp='2022-11-27T19:15:10', username='test-event-admin',
            action='NAMESPACE_CREATE', object_id='test-namespace/test-module', old_value='', new_value='')

        self._ensure_audit_row_is_like(
            rows[2],
            timestamp='2022-11-27T19:15:10', username='test-event-admin',
            action='MODULE_PROVIDER_CREATE', object_id='test-namespace/test-module', old_value='', new_value='')

    def test_pagination(self):
        """Test pagination for audit history."""
        pass

    def test_column_ordering(self):
        """Test ordering data by column."""
        pass

    def test_result_filtering(self):
        """Test filtering results using query string."""
        pass
