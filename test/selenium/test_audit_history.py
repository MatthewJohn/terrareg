
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from terrareg.database import Database
from test.selenium import SeleniumTest
from terrareg.audit_action import AuditAction


class TestAuditHistory(SeleniumTest):
    """Test audit_history page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    _AUDIT_DATA = {
        'user_group_delete': ['useradmin', AuditAction.USER_GROUP_DELETE, 'UserGroup', 'test-user-group', None, None,
                              datetime(year=2099, month=1, day=2, hour=9, minute=2, second=0)],
        'user_group_create': ['useradmin', AuditAction.USER_GROUP_CREATE, 'UserGroup', 'test-user-group', None, None,
                              datetime(year=2099, month=1, day=2, hour=9, minute=1, second=0)],
        'namespace_create': [
            'test-event-admin', AuditAction.NAMESPACE_CREATE, 'Namespace', 'test-namespace', None, None,
            datetime(year=2020, month=11, day=27,
                     hour=19, minute=14, second=10)
        ],
        'module_provider_create': [
            'test-event-admin', AuditAction.NAMESPACE_CREATE, 'Namespace', 'test-namespace', None, None,
            datetime(year=2020, month=11, day=27, hour=19, minute=14, second=10)],
        'module_version_index_1': [
            'namespaceowner', AuditAction.MODULE_VERSION_INDEX, 'Module', 'test-namespace/test-module/2.0.1', None, None,
            datetime(year=2021, month=12, day=28, hour=19, minute=16, second=23)],
        'module_version_index_2': [
            'namespaceowner', AuditAction.MODULE_VERSION_INDEX, 'Module', 'test-namespace/test-module/2.0.1', None, None,
            datetime(year=2021, month=12, day=28, hour=19, minute=15, second=10)],
        'module_version_publish': [
            'namespaceowner', AuditAction.MODULE_VERSION_PUBLISH, 'Module', 'test-namespace/test-module/2.0.1', None, None,
            datetime(year=2021, month=12, day=29, hour=19, minute=23, second=31)],
        'module_version_delete': [
            'namespaceowner', AuditAction.MODULE_VERSION_DELETE, 'Module', 'test-namespace/test-module/2.0.1', None, None,
            datetime(year=2021, month=12, day=29, hour=20, minute=12, second=23)],
        'user_login_1': [
            'testuser1', AuditAction.USER_LOGIN, 'User', 'testuser1', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=49, second=20)],
        'user_login_2': [
            'testuser2', AuditAction.USER_LOGIN, 'User', 'testuser2', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=50, second=20)],
        'user_login_3': [
            'testuser3', AuditAction.USER_LOGIN, 'User', 'testuser4', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=52, second=20)],
        'user_login_4': [
            'testuser4', AuditAction.USER_LOGIN, 'User', 'testuser4', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=53, second=20)],
        'user_login_5': [
            'testuser5', AuditAction.USER_LOGIN, 'User', 'testuser5', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=54, second=20)],
        'user_login_6': [
            'testuser6', AuditAction.USER_LOGIN, 'User', 'testuser6', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=55, second=20)],
        'user_login_7': [
            'testuser7', AuditAction.USER_LOGIN, 'User', 'testuser7', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=56, second=20)],
        'user_login_8': [
            'testuser8', AuditAction.USER_LOGIN, 'User', 'testuser8', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=57, second=20)],
        'user_login_9': [
            'testuser9', AuditAction.USER_LOGIN, 'User', 'testuser9', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=58, second=20)],
        'user_login_10': [
            'testuser10', AuditAction.USER_LOGIN, 'User', 'testuser10', None, None,
            datetime(year=2020, month=1, day=2, hour=9, minute=59, second=20)]

    }

    @classmethod
    def _setup_test_audit_data(cls):
        """Setup test audit data."""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.audit_history.delete())

            for data in cls._AUDIT_DATA.values():
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
        cls.register_patch(mock.patch(
            'terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))

        super(TestAuditHistory, cls).setup_class()

    def setup_method(self, *args, **kwargs):
        """Setup test audit data."""
        super().setup_method(*args, **kwargs)
        self._setup_test_audit_data()

    def test_navigation_from_homepage(self):
        """Test navigation user navbar to audit history page"""
        self.selenium_instance.get(self.get_url('/'))
        # Ensure Setting drop-down is not shown
        drop_down = self.wait_for_element(
            By.ID, 'navbarSettingsDropdown', ensure_displayed=False)
        assert drop_down.is_displayed() is False

        self.perform_admin_authentication('unittest-password')
        # Ensure Setting drop-down is shown
        drop_down = self.wait_for_element(By.ID, 'navbarSettingsDropdown')
        assert drop_down.is_displayed() is True

        # Move mouse to settings drop-down
        selenium.webdriver.ActionChains(
            self.selenium_instance).move_to_element(drop_down).perform()

        user_groups_button = drop_down.find_element(
            By.LINK_TEXT, 'Audit History')
        assert user_groups_button.text == 'Audit History'
        assert user_groups_button.is_displayed() is True

        user_groups_button.click()

        assert self.selenium_instance.current_url == self.get_url(
            '/audit-history')

    @staticmethod
    def _ensure_audit_row_is_like(row, audit_item):
        """Ensure row matches expected data"""
        assert [
            td.text
            for td in row.find_elements(By.TAG_NAME, 'td')
        ] == [
            audit_item[6].isoformat(),
            audit_item[0],
            audit_item[1].name,
            audit_item[3],
            audit_item[4] if audit_item[4] else '',
            audit_item[5] if audit_item[5] else ''
        ]

    @staticmethod
    def _get_audit_rows(parent):
        """Return rows from table"""
        return [r for r in parent.find_elements(By.TAG_NAME, 'tr')]

    def test_basic_view(self):
        """Ensure page shows basic audit history."""
        self.perform_admin_authentication('unittest-password')

        # Load homepage, waiting for drop-down to be rendered and navigate to
        # audit history page
        self.selenium_instance.get(self.get_url('/audit-history'))

        assert self.selenium_instance.find_element(
            By.CLASS_NAME, 'breadcrumb').text == 'Audit History'

        # Ensure table is shown and fields are present
        audit_table = self.selenium_instance.find_element(
            By.ID, 'audit-history-table')

        # Re-order table by timestamp desc
        audit_table.find_element(By.XPATH, ".//th[text()='Timestamp']").click()

        rows = self._get_audit_rows(audit_table)

        # Ignore header row and first audit event caused by login
        self._ensure_audit_row_is_like(
            rows[1],
            self._AUDIT_DATA['user_group_delete'])

        self._ensure_audit_row_is_like(
            rows[2],
            self._AUDIT_DATA['user_group_create'])

    def test_pagination(self):
        """Test pagination for audit history."""
        self.perform_admin_authentication('unittest-password')

        # Load homepage, waiting for drop-down to be rendered and navigate to
        # audit history page
        self.selenium_instance.get(self.get_url('/audit-history'))

        # Ensure table is shown and fields are present
        audit_table = self.selenium_instance.find_element(
            By.ID, 'audit-history-table')

        # Ensure only 10 rows are shown in table (plus heading)
        self.assert_equals(lambda: len(self._get_audit_rows(audit_table)), 11)

        rows = self._get_audit_rows(audit_table)

        # Check content of rows
        self._ensure_audit_row_is_like(
            rows[1],
            self._AUDIT_DATA['user_login_1'])

        self._ensure_audit_row_is_like(
            rows[2],
            self._AUDIT_DATA['user_login_2'])

        # Ensure previous is disabled and next is available
        # Temporarily disabled due to is_enabled returning True
        #assert self.selenium_instance.find_element(By.ID, 'audit-history-table_previous').find_element(By.CLASS_NAME, 'pagination-link').is_enabled() == False
        #assert self.selenium_instance.find_element(By.ID, 'audit-history-table_next').find_element(By.CLASS_NAME, 'pagination-link').is_enabled() == True

        # Ensure total pages is 2, with prev + next buttons
        page_links = [link for link in self.selenium_instance.find_elements(
            By.CLASS_NAME, 'pagination-link')]
        assert len(page_links) == 4

        # Click next page
        self.selenium_instance.find_element(
            By.ID, 'audit-history-table_next').find_element(By.TAG_NAME, 'a').click()

        # Ensure that there are 9 rows plus header
        self.assert_equals(lambda: len(self._get_audit_rows(audit_table)), 10)

        rows = self._get_audit_rows(audit_table)

        # Check content of rows
        self._ensure_audit_row_is_like(
            rows[1],
            self._AUDIT_DATA['namespace_create'])

        self._ensure_audit_row_is_like(
            rows[1],
            self._AUDIT_DATA['module_provider_create'])

    def test_column_ordering(self):
        """Test ordering data by column."""
        self.perform_admin_authentication('unittest-password')

        # Load homepage, waiting for drop-down to be rendered and navigate to
        # audit history page
        self.selenium_instance.get(self.get_url('/audit-history'))

        # Ensure table is shown and fields are present
        audit_table = self.selenium_instance.find_element(
            By.ID, 'audit-history-table')

        # Ensure only 10 rows are shown in table (plus heading)
        self.assert_equals(lambda: len(self._get_audit_rows(audit_table)), 11)

        rows = self._get_audit_rows(audit_table)

        column_headers = [
            r for r in audit_table.find_elements(By.TAG_NAME, 'th')]
        for column_itx, expected_rows in [
            # Sort by timestamp, reversed
            (0, [self._AUDIT_DATA['user_group_delete'],
                 self._AUDIT_DATA['user_group_create']]),

            # Sort by username
            (1, [None, self._AUDIT_DATA['module_version_index_1'], None]),
            # Sort by username reversed
            (1, [self._AUDIT_DATA['user_group_delete'],
                 self._AUDIT_DATA['user_group_create']]),

            # Sort by action
            (2, [None, self._AUDIT_DATA['module_version_index_1']]),
            # Sort by action reversed
            (2, [self._AUDIT_DATA['user_login_1'],
                 self._AUDIT_DATA['user_login_2']]),

            # Sort by object
            (3, [None, self._AUDIT_DATA['namespace_create']]),
            # Sort by object reversed
            (3, [self._AUDIT_DATA['user_login_9'],
                 self._AUDIT_DATA['user_login_8']])
        ]:
            print(f'Testing column sorting: {column_itx}')

            # Click header
            column_headers[column_itx].click()
            rows = self._get_audit_rows(audit_table)
            # Check rows
            for itx, row_test in enumerate(expected_rows):
                if row_test is not None:
                    self._ensure_audit_row_is_like(
                        rows[itx + 1],
                        row_test
                    )

    @pytest.mark.parametrize(
        'search_string,result_count,expected_rows', [
            ('testuser', 10, [_AUDIT_DATA['user_login_1'],
                              _AUDIT_DATA['user_login_2'], _AUDIT_DATA['user_login_3']]),
            ('namespaceowner', 4, [_AUDIT_DATA['module_version_index_2'], _AUDIT_DATA['module_version_index_1'],
                                   _AUDIT_DATA['module_version_publish'], _AUDIT_DATA['module_version_delete']]),
            ('MODULE_VERSION_INDEX', 2, [
             _AUDIT_DATA['module_version_index_2'], _AUDIT_DATA['module_version_index_1']]),
            ('test-namespace', 6, [_AUDIT_DATA['namespace_create'], _AUDIT_DATA['module_provider_create'],
                                   _AUDIT_DATA['module_version_index_2'], _AUDIT_DATA['module_version_index_1'],
                                   _AUDIT_DATA['module_version_publish'], _AUDIT_DATA['module_version_delete']])
        ]
    )
    def test_result_filtering(
            self, search_string, result_count, expected_rows):
        """Test filtering results using query string."""
        self.perform_admin_authentication('unittest-password')

        # Load homepage, waiting for drop-down to be rendered and navigate to
        # audit history page
        self.selenium_instance.get(self.get_url('/audit-history'))

        # Ensure table is shown and fields are present
        audit_table = self.selenium_instance.find_element(
            By.ID, 'audit-history-table')

        self.selenium_instance.find_element(
            By.ID, 'audit-history-table_filter').find_element(By.TAG_NAME, 'input').send_keys(search_string)

        # Ensure only 10 rows are shown in table (plus heading)
        self.assert_equals(lambda: len(
            self._get_audit_rows(audit_table)), result_count + 1)

        rows = self._get_audit_rows(audit_table)

        rows = self._get_audit_rows(audit_table)
        # Check rows
        for itx, row_test in enumerate(expected_rows):
            if row_test is not None:
                self._ensure_audit_row_is_like(
                    rows[itx + 1],
                    row_test
                )
