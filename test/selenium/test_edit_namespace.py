
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

class TestEditNamespace(SeleniumTest):
    """Test create_module_provider page."""

    _SECRET_KEY = "354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe"

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch("terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN", "unittest-password"))

        super(TestEditNamespace, cls).setup_class()

    def teardown_method(self, method):
        """Clear down any cookes from the trst."""
        self.selenium_instance.delete_all_cookies()
        super(TestEditNamespace, self).teardown_method(method)

    def test_navigation_from_namespace(self):
        """Test navigation to namespace edit page from namespage module list"""
        # Ensure button is not present when not logged in
        self.selenium_instance.get(self.get_url("/modules/moduledetails"))

        edit_button = self.selenium_instance.find_element(By.ID, "edit-namespace-link")
        assert edit_button.is_displayed() is False

        # Authenticate and check that button appears
        self.perform_admin_authentication(password="unittest-password")

        self.selenium_instance.get(self.get_url("/modules/moduledetails"))

        edit_button = self.selenium_instance.find_element(By.ID, "edit-namespace-link")
        self.assert_equals(lambda: edit_button.is_displayed(), True)

        # Click button and ensure redirect works
        edit_button.click()
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/edit-namespace/moduledetails"))


    def test_delete_namespace(self, mock_create_audit_event):
        """Test successful deletion of namespace"""
        # Create fake namespace
        with mock_create_audit_event:
            namespace = Namespace.create("test-deletion")

        try:
            self.perform_admin_authentication(password="unittest-password")

            self.selenium_instance.get(self.get_url("/edit-namespace/test-deletion"))

            delete_button = self.selenium_instance.find_element(By.ID, "deleteNamespaceButton")
            assert delete_button.is_displayed() is True

            delete_button.click()

            # Ensure user is redirected to homepage
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/"))

            # Ensure namespace is deleted
            assert Namespace.get("test-deletion") is None
        
        finally:
            # Clear up namespace from deletion test
            with mock_create_audit_event:    
                if namespace := Namespace.get("test-deletion"):
                    namespace.delete()


    def test_delete_namespace_with_modules(self, mock_create_audit_event):
        """Test attempt to delete namespace with modules present"""
        # Create fake namespace
        with mock_create_audit_event:
            namespace = Namespace.create("test-deletion")
            module_provider = ModuleProvider.create(Module(namespace, "test"), "test")

        try:
            self.perform_admin_authentication(password="unittest-password")

            self.selenium_instance.get(self.get_url("/edit-namespace/test-deletion"))

            # Ensure deletion error is not shown
            assert self.selenium_instance.find_element(By.ID, "delete-error").is_displayed() is False

            delete_button = self.selenium_instance.find_element(By.ID, "deleteNamespaceButton")
            assert delete_button.is_displayed() is True

            delete_button.click()

            # Ensure user is still on namespace edit page
            assert self.selenium_instance.current_url == self.get_url("/edit-namespace/test-deletion")

            # Ensure error is correctly shown
            error = self.selenium_instance.find_element(By.ID, "delete-error")
            assert error.is_displayed() is True
            assert error.text == "Namespace cannot be deleted as it contains modules"

            # Ensure namespace is not deleted
            assert Namespace.get("test-deletion") is not None
        
        finally:
            with mock_create_audit_event:
                module_provider.delete()
                namespace.delete()
