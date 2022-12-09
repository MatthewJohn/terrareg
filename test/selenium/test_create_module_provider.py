
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import Select
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestCreateModuleProvider(SeleniumTest):
    """Test create_module_provider page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))

        super(TestCreateModuleProvider, cls).setup_class()

    def _fill_out_field_by_label(self, label, input):
        """Find input field by label and fill out input."""
        form = self.selenium_instance.find_element(By.ID, 'create-module-form')
        input_field = form.find_element(By.XPATH, ".//label[text()='{label}']/parent::*//input".format(label=label))
        input_field.send_keys(input)

    def _click_create(self):
        """Click create button"""
        self.selenium_instance.find_element(By.XPATH, "//button[text()='Create']").click()

    def test_page_details(self):
        """Test page contains required information."""

        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Create Module'

        expected_labels = [
            'Namespace', 'Module Name', 'Provider',
            'Git Repository Provider', 'Custom Repository base URL', 'Custom Repository Clone URL',
            'Custom Repository source browse URL', 'Git tag format', 'Git path'
        ]
        for label in self.selenium_instance.find_element(By.ID, 'create-module-form').find_elements(By.TAG_NAME, 'label'):
            assert label.text == expected_labels.pop(0)

        assert [
            option.text
            for option in Select(
                self.selenium_instance.find_element(By.ID, 'create-module-namespace')
            ).options
        ] == [
            'javascriptinjection',
            'moduledetails',
            'moduleextraction',
            'modulesearch',
            'modulesearch-contributed',
            'modulesearch-trusted',
            'mostrecent',
            'mostrecentunpublished',
            'onlybeta',
            'onlyunpublished',
            'real_providers',
            'relevancysearch',
            'repo_url_tests',
            'searchbynamespace',
            'testmodulecreation',
            'testnamespace',
            'trustednamespace',
            'unpublished-beta-version-module-providers',
        ]


    def test_create_basic(self):
        """Test creating module provider with inputs populated."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('testmodulecreation')
        self._fill_out_field_by_label('Module Name', 'minimal-module')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._fill_out_field_by_label('Git tag format', 'vunit{version}test')

        self._click_create()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/minimal-module/testprovider'))

        # Ensure module was created
        module_provider = ModuleProvider.get(Module(Namespace('testmodulecreation'), 'minimal-module'), 'testprovider')
        assert module_provider is not None
        assert module_provider.git_tag_format == 'vunit{version}test'

    def test_with_git_path(self):
        """Test creating module provider with inputs populated."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('testmodulecreation')
        self._fill_out_field_by_label('Module Name', 'with-git-path')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._fill_out_field_by_label('Git tag format', 'v{version}')

        self._fill_out_field_by_label('Git path', './testmodulesubdir')

        self._click_create()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/with-git-path/testprovider'))

        # Ensure module was created
        module_provider = ModuleProvider.get(Module(Namespace('testmodulecreation'), 'with-git-path'), 'testprovider')
        assert module_provider is not None
        assert module_provider._get_db_row()['git_path'] == './testmodulesubdir'

    @pytest.mark.skip(reason="Not implemented")
    def test_unauthenticated(self):
        """Test creating a module when not authenticated."""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_duplicate_module(self):
        """Test creating a module that already exists."""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_git_urls(self):
        """Create a module with git URLs"""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_invalid_git_urls(self):
        """Create a module with invalid git URLs"""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_invalid_git_tag(self):
        """Create a module with invalid git tag format"""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_invalid_namespace(self):
        """Attempt to create a module with an invalid namespace."""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_invalid_module_name(self):
        """Attempt to create a module with an invalid module name."""
        pass

    @pytest.mark.skip(reason="Not implemented")
    def test_creating_with_invalid_provider(self):
        """Attempt to create a module with an invalid provider name."""
        pass
