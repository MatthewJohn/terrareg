
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestModuleProvider(SeleniumTest):
    """Test homepage."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._mocks = [
            mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'),
            mock.patch('terrareg.config.Config.SECRET_KEY', '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe')
        ]
        for mock_ in cls._mocks:
            mock_.start()
        print('MOCKED THE STUFF')
        super(TestModuleProvider, cls).setup_class()

    @classmethod
    def teardown_class(cls):
        """Setup required mocks."""
        for mock_ in cls._mocks:
            mock_.stop()
        super(TestModuleProvider, cls).teardown_class()

    def test_delete_module_version(self):
        """Check provider logos are displayed correctly."""

        self.perform_admin_authentication(password='unittest-password')

        # Create test module version
        namespace = Namespace(name='moduledetails')
        module = Module(namespace=namespace, name='fullypopulated')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        # Create test module version
        module_version = ModuleVersion(module_provider=module_provider, version='2.5.5')
        module_version.prepare_module()
        module_version.publish()

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/2.5.5'))

        # Click on settings tab
        self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab = self.selenium_instance.find_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Ensure the verification text is not shown

        # Click on the delete module version button
        delete_button = self.selenium_instance.find_element(By.ID, 'module-version-delete-button')
        delete_button.click()

        # Ensure the verification text is shown
        verification_div = self.selenium_instance.find_element(By.ID, 'confirm-delete-module-version-div')
        assert 'Confirm deletion of module version 2.5.5:' in verification_div.text

        # Click checkbox for verifying deletion
        delete_checkbox = verification_div.find_element(By.ID, 'confirm-delete-module-version')
        delete_checkbox.click()

        # Click delete module version button again
        delete_button.click()

        # Ensure user is redirected to module page
        assert self.selenium_instance.current_url == self.get_url('/modules/moduledetails/fullypopulated/testprovider')

        # Ensure module version no longer exists
        assert ModuleVersion.get(module_provider=module_provider, version='2.5.5') is None
