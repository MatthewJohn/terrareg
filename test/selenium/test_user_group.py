
from datetime import datetime
import re
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import Select
import selenium.webdriver
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider, UserGroup, UserGroupNamespacePermission

from test.selenium.test_data import two_empty_namespaces


class TestUserGroup(SeleniumTest):
    """Test User Group page."""

    _TEST_DATA = two_empty_namespaces
    _USER_GROUP_DATA = None
    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.config.Config.ENABLE_ACCESS_CONTROLS', True))
        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))
        super(TestUserGroup, cls).setup_class()

    def _delete_all_user_groups(self):
        """Delete all user groups"""
        # Delete any pre-existing user groups
        for user_group in UserGroup.get_all_user_groups():
            user_group.delete()

    def setup_method(self):
        r = super().setup_method()
        self._delete_all_user_groups()
        return r

    def teardown_method(self, method):
        self._delete_all_user_groups()
        return super().teardown_method(method)

    def _fill_out_field_by_label(self, form, label, input=None, check=None):
        """Find input field by label and fill out input."""
        input_field = form.find_element(By.XPATH, ".//label[text()='{label}']/parent::*//input".format(label=label))
        if check is not None:
            if (not input_field.get_attribute('checked') and check) or (input_field.get_attribute('checked') and not check):
                input_field.click()
        else:
            input_field.send_keys(input)

    def _find_element_by_text(self, parent, text):
        """Find element in element by content text"""
        return parent.find_element(By.XPATH, f".//*[text()='{text}']")

    def test_navigation_from_homepage(self):
        """Test navigation user navbar to user group page"""
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

        user_groups_button = drop_down.find_element(By.LINK_TEXT, 'User Groups')
        assert user_groups_button.text == 'User Groups'
        assert user_groups_button.is_displayed() == True

        user_groups_button.click()

        assert self.selenium_instance.current_url == self.get_url('/user-groups')

    @pytest.mark.parametrize('is_site_admin', [
        True,
        False
    ])
    def test_add_user_group(self, is_site_admin):
        """Test adding user group."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/user-groups'))

        # Fill out form for new group
        form = self.selenium_instance.find_element(By.ID, 'create-user-group-form')
        self._fill_out_field_by_label(form, 'SSO Group Name', 'UnittestUserGroup')
        self._fill_out_field_by_label(form, 'Site Admin', check=is_site_admin)
        # Click create button
        self._find_element_by_text(form, 'Create').click()

        # Ensure user group exists in database
        user_group = UserGroup.get_by_group_name('UnittestUserGroup')
        assert user_group.name == 'UnittestUserGroup'
        assert user_group.site_admin == is_site_admin

        # Ensure user group is displayed on page
        user_group_table = self.selenium_instance.find_element(By.ID, 'user-group-table')
        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        assert f'UnittestUserGroup (Site admin: {"Yes" if is_site_admin else "No"})' in [r.text for r in user_table_rows]

    def test_add_user_group_permission(self):
        """Test adding user group permission"""
        user_group = UserGroup.create(name='AddPermissionUserGroup', site_admin=False)

        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/user-groups'))

        # Wait for datatable to load :(
        sleep(1)

        user_group_table = self.wait_for_element(By.ID, 'user-group-table')

        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        create_user_permission_row = None
        for itx, row in enumerate(user_table_rows):
            if row.text == 'AddPermissionUserGroup (Site admin: No)':
                create_user_permission_row = user_table_rows[itx + 1]
                break
        else:
            raise Exception('Could not find user group table row')

        # Fill out form
        namespace_select = Select(
            create_user_permission_row.find_element(
                By.ID, 'createUserGroupPermission-Namespace-AddPermissionUserGroup'))

        # Ensure all namespaces are listed
        assert ['firstnamespace', 'second-namespace'] == [option.text for option in namespace_select.options]
        namespace_select.select_by_visible_text('second-namespace')

        permission_select = Select(
            create_user_permission_row.find_element(
                By.ID, 'createUserGroupPermission-Permission-AddPermissionUserGroup'))
        assert ['Full', 'Modify'] == [option.text for option in permission_select.options]

        permission_select.select_by_visible_text('Modify')

        # Create user permission
        self._find_element_by_text(create_user_permission_row, 'Create').click()

        # User user permission was added to DB
        permissions = UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
        assert len(permissions) == 1
        assert permissions[0].namespace.name == 'second-namespace'
        assert permissions[0].permission_type is UserGroupNamespacePermissionType.MODIFY

        ### Add permission for namespace
        user_group_table = self.wait_for_element(By.ID, 'user-group-table')

        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        create_user_permission_row = None
        for itx, row in enumerate(user_table_rows):
            if row.text == 'AddPermissionUserGroup (Site admin: No)':
                # First row is the first created permision
                assert user_table_rows[itx + 1].text == 'second-namespace MODIFY\nDelete'
                create_user_permission_row = user_table_rows[itx + 2]
                break
        else:
            raise Exception('Could not find user group table row')

        # Fill out form
        namespace_select = Select(
            create_user_permission_row.find_element(
                By.ID, 'createUserGroupPermission-Namespace-AddPermissionUserGroup'))

        # Ensure all namespaces are listed
        assert ['firstnamespace'] == [option.text for option in namespace_select.options]
        namespace_select.select_by_visible_text('firstnamespace')

        permission_select = Select(
            create_user_permission_row.find_element(
                By.ID, 'createUserGroupPermission-Permission-AddPermissionUserGroup'))
        assert ['Full', 'Modify'] == [option.text for option in permission_select.options]

        permission_select.select_by_visible_text('Full')

        # Create user permission
        self._find_element_by_text(create_user_permission_row, 'Create').click()

        # User user permission was added to DB
        permissions = UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
        assert len(permissions) == 2
        permissions.sort(key=lambda x: x.namespace.name)
        assert permissions[0].namespace.name == 'firstnamespace'
        assert permissions[0].permission_type is UserGroupNamespacePermissionType.FULL
        assert permissions[1].namespace.name == 'second-namespace'
        assert permissions[1].permission_type is UserGroupNamespacePermissionType.MODIFY

    def test_delete_namespace_permission(self):
        """Test deleting namespace permission"""
        user_group = UserGroup.create('UnittestGroupToDeletePerm', site_admin=False)
        UserGroupNamespacePermission.create(
            user_group=user_group,
            namespace=Namespace.get('firstnamespace'),
            permission_type=UserGroupNamespacePermissionType.FULL
        )

        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/user-groups'))

        # Wait for datatable to load :(
        sleep(1)

        user_group_table = self.wait_for_element(By.ID, 'user-group-table')

        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        user_permission_row = None
        for itx, row in enumerate(user_table_rows):
            if row.text == 'UnittestGroupToDeletePerm (Site admin: No)':
                # Rows - Group title, first permission
                user_permission_row = user_table_rows[itx + 1]
                break
        else:
            raise Exception('Could not find user group table row')

        self._find_element_by_text(user_permission_row, 'Delete').click()

        # User user permission was added to DB
        permissions = UserGroupNamespacePermission.get_permissions_by_user_group(user_group)
        assert len(permissions) == 0

        # Ensure user group permission is no longer displayed on page
        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        user_permission_create_row = None
        for itx, row in enumerate(user_table_rows):
            if row.text == 'UnittestGroupToDeletePerm (Site admin: No)':
                # Rows - Group title, first permission
                user_permission_create_row = user_table_rows[itx + 1]
                break
        else:
            raise Exception('Could not find user group table row')

        # Ensure row looks like a create row with both namespaces
        assert user_permission_create_row.text == """
firstnamespace
second-namespace
Full
Modify
Create
""".strip()

    def test_delete_user_gruop(self):
        """Test deleting user group"""
        UserGroup.create('UnittestGroupToDelete', site_admin=False)

        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/user-groups'))

        sleep(1)

        user_group_table = self.wait_for_element(By.ID, 'user-group-table')

        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        user_group_delete_row = None
        for itx, row in enumerate(user_table_rows):
            if row.text == 'UnittestGroupToDelete (Site admin: No)':
                # Rows - Group title, create permission, delete user group
                user_group_delete_row = user_table_rows[itx + 2]
                break
        else:
            raise Exception('Could not find user group table row')

        self._find_element_by_text(user_group_delete_row, 'Delete user group').click()

        # Ensure user group has been removed
        assert len(UserGroup.get_all_user_groups()) == 0

        # Ensure user group is no longer in table

        user_table_rows = user_group_table.find_elements(By.TAG_NAME, 'tr')
        for itx, row in enumerate(user_table_rows):
            assert row.text != 'UnittestGroupToDelete (Site admin: No)'
