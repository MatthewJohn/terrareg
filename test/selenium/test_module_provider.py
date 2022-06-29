
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

    def test_module_without_versions(self):
        """Test page functionality on a module without published versions."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

        # Ensure integrations tab link is display and tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-integrations')
        integration_tab = self.wait_for_element(By.ID, 'module-tab-integrations')

        # Ensure all other tabs aren't shown
        for tab_name in ['readme', 'example-files', 'inputs', 'outputs', 'providers', 'resources', 'analytics', 'usage-builder', 'settings']:
            # Ensure tab link isn't displayed
            assert self.selenium_instance.find_element(By.ID, f'module-tab-link-{tab_name}').is_displayed() == False

            # Ensure tab content isn't displayed
            assert self.selenium_instance.find_element(By.ID, f'module-tab-{tab_name}').is_displayed() == False

        # Login
        self.perform_admin_authentication(password='unittest-password')

        # Ensure settings tab link is displayed
        self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-link-settings').is_displayed(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-settings').is_displayed(), False)

    @pytest.mark.parametrize('url,expected_readme_content', [
        # Root module
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0', 'This is an exaple README!'),
        # Module example
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example', 'Example 1 README'),
        # Submodule
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1', 'Submodule 1 README')
    ])
    def test_readme_tab(self, url, expected_readme_content):
        """Ensure README is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))
        
        # Ensure README link and tab is displayed by default
        self.wait_for_element(By.ID, 'module-tab-link-readme')
        readme_content = self.wait_for_element(By.ID, 'module-tab-readme')

        # Check contents of REAMDE
        assert readme_content.text == expected_readme_content

        # Navigate to another tab
        self.selenium_instance.find_element(By.ID, 'module-tab-link-inputs').click()

        # Ensure that README content is not longer visible
        assert self.selenium_instance.find_element(By.ID, 'module-tab-readme').is_displayed() == False

        # Click on README tab again
        self.selenium_instance.find_element(By.ID, 'module-tab-link-readme').click()

        # Ensure README content is visible again and content is correct
        assert self.selenium_instance.find_element(By.ID, 'module-tab-readme').is_displayed() == True
        assert readme_content.text == expected_readme_content

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
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
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
