
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
        input_field.clear()
        input_field.send_keys(input)

    def _click_create(self):
        """Click create button"""
        self.selenium_instance.find_element(By.XPATH, "//button[text()='Create']").click()

    def test_page_details(self):
        """Test page contains required information."""

        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        assert self.selenium_instance.find_element(By.CLASS_NAME, 'breadcrumb').text == 'Create Module'

        expected_fields = [
            # label, placeholder value, default value
            ('Namespace', None, 'javascriptinjection'),
            ('Module Name', 'my-module', ''),
            ('Provider', 'e.g. aws, null...', ''),
            ('Git Repository Provider', None, 'Custom'),
            ('Custom Repository base URL', 'https://github.com/my-team/my-module-provider', ''),
            ('Custom Repository Clone URL', 'ssh://git@github.com:my-team/my-module-provider.git', ''),
            ('Custom Repository source browse URL', 'https://github.com/my-team/my-module-provider/tree/{tag}/{path}', ''),
            ('Git tag format', 'v{version}', 'v{version}'),
            ('Git path', '/', '')
        ]
        for label in self.selenium_instance.find_element(By.ID, 'create-module-form').find_elements(By.TAG_NAME, 'label'):
            expected_values = expected_fields.pop(0)
            assert label.text == expected_values[0]

            # Find parent field div
            parent = label.find_element(By.XPATH, '..')

            # Find form input element
            input = parent.find_element(By.XPATH, ".//*[local-name() = 'select' or local-name() = 'input']")

            # For select elements, find the selected option
            if input.tag_name == 'select':
                input = Select(input).first_selected_option
                assert input.text == expected_values[2]
            else:
                # Check placeholder and value
                if expected_values[1]:
                    assert input.get_attribute('placeholder') == expected_values[1]
                assert input.get_attribute('value') == expected_values[2]

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
            # Test displaying a module with a display name
            'A Display Name'
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
        with self._patch_audit_event_creation():
            module_provider.delete()

    def test_create_against_namespace_with_display_name(self):
        """Test creating module provider with inputs populated."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('A Display Name')
        self._fill_out_field_by_label('Module Name', 'minimal-module')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._fill_out_field_by_label('Git tag format', 'vunit{version}test')

        self._click_create()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/withdisplayname/minimal-module/testprovider'))

        # Ensure module was created
        module_provider = ModuleProvider.get(Module(Namespace('withdisplayname'), 'minimal-module'), 'testprovider')
        assert module_provider is not None
        assert module_provider.git_tag_format == 'vunit{version}test'
        with self._patch_audit_event_creation():
            module_provider.delete()

    def test_with_git_path(self):
        """Test creating module provider with inputs populated."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('testmodulecreation')
        self._fill_out_field_by_label('Module Name', 'with-git-path')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._fill_out_field_by_label('Git path', './testmodulesubdir')

        self._click_create()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/with-git-path/testprovider'))

        # Ensure module was created
        module_provider = ModuleProvider.get(Module(Namespace('testmodulecreation'), 'with-git-path'), 'testprovider')
        assert module_provider is not None
        assert module_provider._get_db_row()['git_path'] == './testmodulesubdir'
        with self._patch_audit_event_creation():
            module_provider.delete()

    @pytest.mark.parametrize('git_tag_format,should_show_form_validation_error,should_error,expected_git_tag_format', [
        # Leave default value
        (None, False, False, 'v{version}'),
        # Test empty value
        ('', True, False, None),

        # Tag format without template placeholder
        ('testgittag', False, True, None),

        ('unittestvalue{version}', False, False, 'unittestvalue{version}')
    ])
    def test_with_git_tag_format(self, git_tag_format, should_show_form_validation_error, should_error, expected_git_tag_format):
        """Test creating module provider with inputs populated."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('testmodulecreation')
        self._fill_out_field_by_label('Module Name', 'with-git-path')
        self._fill_out_field_by_label('Provider', 'testprovider')

        if git_tag_format is not None:
            self._fill_out_field_by_label('Git tag format', git_tag_format)

        try:
            self._click_create()

            # Check if form validation is shown
            if should_show_form_validation_error:
                self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-module-git-tag-format').get_attribute('validationMessage'), 'Please fill out this field.')
                self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-module'))

            # Check if error is returned
            elif should_error:
                self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
                self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').text, "Invalid git tag format. Must contain one placeholder: {version}.")
                self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-module'))

            # Otherwise, check that it was created correctly
            else:
                self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testmodulecreation/with-git-path/testprovider'))

                # Ensure module was created
                module_provider = ModuleProvider.get(Module(Namespace('testmodulecreation'), 'with-git-path'), 'testprovider')
                assert module_provider is not None
                assert module_provider._get_db_row()['git_tag_format'] == expected_git_tag_format
        finally:
            module_provider = ModuleProvider.get(Module(Namespace('testmodulecreation'), 'with-git-path'), 'testprovider')
            if module_provider:
                with self._patch_audit_event_creation():
                    module_provider.delete()

    def test_unauthenticated(self):
        """Test creating a module when not authenticated."""
        self.selenium_instance.delete_all_cookies()

        self.selenium_instance.get(self.get_url('/create-module'))

        self._fill_out_field_by_label('Module Name', 'with-git-path')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._click_create()
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'create-error').text,
            "You must be logged in to perform this action.\nIf you were previously logged in, please re-authentication and try again.")
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-module'))


    def test_duplicate_module(self):
        """Test creating a module that already exists."""
        self.perform_admin_authentication('unittest-password')

        self.selenium_instance.get(self.get_url('/create-module'))

        Select(self.selenium_instance.find_element(By.ID, 'create-module-namespace')).select_by_visible_text('moduledetails')
        self._fill_out_field_by_label('Module Name', 'fullypopulated')
        self._fill_out_field_by_label('Provider', 'testprovider')

        self._click_create()
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').is_displayed(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'create-error').text, "Module provider already exists")
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/create-module'))

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
