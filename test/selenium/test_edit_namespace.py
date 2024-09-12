
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
import terrareg.models


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
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/edit-namespace/test-deletion"))

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

    def test_delete_namespace_with_providers(self):
        """Test attempt to delete namespace with providers present"""
        self.perform_admin_authentication(password="unittest-password")

        self.selenium_instance.get(self.get_url("/edit-namespace/initial-providers"))

        # Ensure deletion error is not shown
        assert self.selenium_instance.find_element(By.ID, "delete-error").is_displayed() is False

        delete_button = self.selenium_instance.find_element(By.ID, "deleteNamespaceButton")
        assert delete_button.is_displayed() is True

        delete_button.click()

        # Ensure user is still on namespace edit page
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/edit-namespace/initial-providers"))

        # Ensure error is correctly shown
        error = self.selenium_instance.find_element(By.ID, "delete-error")
        assert error.is_displayed() is True
        assert error.text == "Namespace cannot be deleted as it contains providers"

        # Ensure namespace is not deleted
        assert Namespace.get("initial-providers") is not None

    def test_add_delete_gpg_key(self):
        """Test add and deleting GPG key"""
        self.perform_admin_authentication(password="unittest-password")
        self.selenium_instance.get(self.get_url("/edit-namespace/second-provider-namespace"))

        gpg_key_table = self.selenium_instance.find_element(By.ID, "gpg-key-table-data")
        assert [row.text for row in gpg_key_table.find_elements(By.TAG_NAME, "tr")] == [
            "E42600BAB40EE715\nDelete"
        ]

        gpg_input = self.selenium_instance.find_element(By.ID, "create-gpg-key-ascii-armor")
        assert gpg_input.get_attribute("value") == ""
        gpg_input.send_keys("""
-----BEGIN PGP PUBLIC KEY BLOCK-----

mI0EZXvmywEEAL9R5rql33+BP0A1ioqoZuiK9WJqCagtAYqURVlk0OQtw05CSLl3
GzkGzwa+b8sJu2e0Q1WvHxe05qFZXmWlhql47fKoHdn5rp4UYy+qt0/347stT1GS
xukGLpVLHutScsZE16jAWxCO00SthMDpRr4n3hkFshb+rSRMARuNLndvABEBAAG0
JEF1dG9nZW5lcmF0ZWQgS2V5IDxtYXR0aGV3QGxhcHRvcDIxPojOBBMBCgA4FiEE
R/TsPhFrKYk6/XuAu1qLOJMNyu8FAmV75ssCGy8FCwkIBwIGFQoJCAsCBBYCAwEC
HgECF4AACgkQu1qLOJMNyu/c9wP/cic9RRl83lyfM+U7GfGmzegQnEU+qoLyB6H4
ldT5r1sVHeKIYxgKBAPFnasPEqFfXhOS9wsbJZNC1tq+i1TQla0PectWTlBrBjDJ
n9wkhjrvcVuqfzvFSX6JA+BmRuQdXmDll3gPSzfXUtrIEcIy8S40liVXsnQaoJ6C
2INHHhk=
=uOWN
-----END PGP PUBLIC KEY BLOCK-----
""")
        self.selenium_instance.find_element(By.ID, "create-gpg-key-form").find_element(By.XPATH, ".//button[text()='Add GPG Key']").click()

        # 47F4EC3E116B29893AFD7B80BB5A8B38930DCAEF
        self.assert_equals(lambda: terrareg.models.GpgKey.get_by_fingerprint("47F4EC3E116B29893AFD7B80BB5A8B38930DCAEF") is not None, True)

        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, "gpg-key-table-data").find_element(By.XPATH, ".//td[text()='BB5A8B38930DCAEF']").is_displayed(), True)
        gpg_key_row = self.selenium_instance.find_element(By.ID, "gpg-key-table-data").find_element(By.XPATH, ".//td[text()='BB5A8B38930DCAEF']/..")
        gpg_key_row.find_element(By.XPATH, ".//button[text()='Delete']").click()

        self.assert_equals(lambda: terrareg.models.GpgKey.get_by_fingerprint("47F4EC3E116B29893AFD7B80BB5A8B38930DCAEF") is None, True)

        gpg_key_table = self.selenium_instance.find_element(By.ID, "gpg-key-table-data")
        assert [row.text for row in gpg_key_table.find_elements(By.TAG_NAME, "tr")] == [
            "E42600BAB40EE715\nDelete"
        ]

    def test_add_invalid_gpg_key(self):
        """Test adding invalid GPG key"""
        self.perform_admin_authentication(password="unittest-password")
        self.selenium_instance.get(self.get_url("/edit-namespace/second-provider-namespace"))

        gpg_input = self.selenium_instance.find_element(By.ID, "create-gpg-key-ascii-armor")
        gpg_input.send_keys("blah blah")
        self.selenium_instance.find_element(By.ID, "create-gpg-key-form").find_element(By.XPATH, ".//button[text()='Add GPG Key']").click()

        error = self.wait_for_element(By.ID, "create-gpg-key-error")
        assert error.is_displayed() is True
        assert error.text == "GPG key provided is invalid or could not be read"