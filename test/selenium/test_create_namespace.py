
from datetime import datetime
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test import mock_create_audit_event
from terrareg.database import Database
from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestCreateNamespace(SeleniumTest):
    """Test create_module_provider page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))

        super(TestCreateNamespace, cls).setup_class()

    def teardown_method(self, method):
        """Clear down any cookes from the trst."""
        self.selenium_instance.delete_all_cookies()
        super(TestCreateNamespace, self).teardown_method(method)

    def _fill_out_field_by_label(self, label, input):
        """Find input field by label and fill out input."""
        form = self.selenium_instance.find_element(By.ID, 'create-namespace-form')
        input_field = form.find_element(By.XPATH, ".//label[text()='{label}']/parent::*//input".format(label=label))
        input_field.send_keys(input)

    def _click_create(self):
        """Click create button"""
        self.selenium_instance.find_element(By.XPATH, "//button[text()='Create Namespace']").click()

    def test_page_details(self):
        """Test page contains required information."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-namespace'))

        assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Create Namespace'

        expected_labels = [
            'Name',
            'Display Name'
        ]
        for label in self.selenium_instance.find_element(By.ID, 'create-namespace-form').find_elements(By.TAG_NAME, 'label'):
            assert label.text == expected_labels.pop(0)

    def test_create_basic(self):
        """Test creating module provider with inputs populated."""
        try:
            self.perform_admin_authentication('unittest-password')

            self.selenium_instance.get(self.get_url('/create-namespace'))

            self._fill_out_field_by_label('Name', 'testnamespacecreation')

            self._click_create()

            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespacecreation'))

            # Ensure namespace was created
            namespace = Namespace.get('testnamespacecreation')
            assert namespace is not None
        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="testnamespacecreation"))


    def test_create_with_display_name(self):
        """Test creating module provider with display name."""
        try:
            self.perform_admin_authentication('unittest-password')

            self.selenium_instance.get(self.get_url('/create-namespace'))

            self._fill_out_field_by_label('Name', 'testnamespacedisplayname')

            self._fill_out_field_by_label('Display Name', 'Test namespace Creation')

            self._click_create()

            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespacedisplayname'))

            # self.assert_equals(lambda: )

            # Ensure namespace was created
            namespace = Namespace.get('testnamespacedisplayname')
            assert namespace is not None
            assert namespace.name == 'testnamespacedisplayname'
            assert namespace.display_name == 'Test namespace Creation'

        finally:
            db = Database.get()
            with db.get_connection() as conn:
                conn.execute(db.namespace.delete(db.namespace.c.namespace=="testnamespacedisplayname"))

    def test_unauthenticated(self):
        """Test creating a namespace when not authenticated."""
        self.selenium_instance.get(self.get_url('/create-namespace'))

        self._fill_out_field_by_label('Name', 'testnamespaceunauthenticated')

        self._click_create()

        error = self.wait_for_element(By.ID, 'create-error')
        assert error.text == """You must be logged in to perform this action.
If you were previously logged in, please re-authentication and try again."""

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-namespace'))

        # Ensure namespace was created
        namespace = Namespace.get('testnamespacecreation')
        assert namespace is not None

    def test_duplicate_namespace(self, mock_create_audit_event):
        """Test creating a namespace that already exists."""
        self.perform_admin_authentication('unittest-password')

        with mock_create_audit_event:
            pre_existing_namespace = Namespace.create('duplicate-namespace-create')

        self.selenium_instance.get(self.get_url('/create-namespace'))

        self._fill_out_field_by_label('Name', 'duplicate-namespace-create')

        self._click_create()

        error = self.wait_for_element(By.ID, 'create-error')
        assert error.text == 'A namespace already exists with this name'

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-namespace'))

        # Ensure original namespace is returned
        namespace = Namespace.get('duplicate-namespace-create')
        assert namespace.pk == pre_existing_namespace.pk
