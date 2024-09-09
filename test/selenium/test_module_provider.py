
import base64
import inspect
import json
import os
import re
from time import sleep
from io import BytesIO, StringIO
from unittest import mock

import requests
import pytest
from selenium.webdriver.common.by import By
from selenium.webdriver.support.select import Select
import selenium.common
from PIL import Image
import imagehash

import terrareg.config
from terrareg.database import Database
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test import mock_create_audit_event, skipif_unless_ci
from test.selenium import SeleniumTest
from terrareg.models import (
    GitProvider, ModuleVersion, Namespace, Module,
    ModuleProvider, ProviderLogo,
    UserGroup, UserGroupNamespacePermission
)
import terrareg.analytics


class TestModuleProvider(SeleniumTest):
    """Test module provider page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._api_version_create_mock = mock.Mock(return_value={'status': 'Success'})
        cls._api_version_publish_mock = mock.Mock(return_value={'status': 'Success'})
        cls._config_publish_api_keys_mock = mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', [])
        cls._config_allow_custom_repo_urls_module_provider = mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', True)
        cls._config_allow_custom_repo_urls_module_version = mock.patch('terrareg.config.Config.ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', True)
        cls._config_enable_access_controls = mock.patch('terrareg.config.Config.ENABLE_ACCESS_CONTROLS', False)
        cls._config_module_links = mock.patch('terrareg.config.Config.MODULE_LINKS', '[]')
        cls._config_terraform_example_version_template = mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', '>= {major}.{minor}.{patch}, < {major_plus_one}.0.0, unittest')
        cls._config_disable_analytics = mock.patch('terrareg.config.Config.DISABLE_ANALYTICS', False)
        cls._config_allow_forceful_module_provider_redirect_deletion = mock.patch('terrareg.config.Config.ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION', True)
        cls._config_default_ui_details_view = mock.patch('terrareg.config.Config.DEFAULT_UI_DETAILS_VIEW', terrareg.config.DefaultUiInputOutputView.TABLE)

        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))
        cls.register_patch(mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS', '[["License", ["first-file", "LICENSE", "second-file"]], ["Changelog", ["CHANGELOG.md"]], ["doesnotexist", ["DOES_NOT_EXIST"]]]'))
        cls.register_patch(mock.patch('terrareg.server.api.ApiModuleVersionCreate._post', cls._api_version_create_mock))
        cls.register_patch(mock.patch('terrareg.server.api.ApiTerraregModuleVersionPublish._post', cls._api_version_publish_mock))
        cls.register_patch(cls._config_publish_api_keys_mock)
        cls.register_patch(cls._config_allow_custom_repo_urls_module_provider)
        cls.register_patch(cls._config_allow_custom_repo_urls_module_version)
        cls.register_patch(cls._config_enable_access_controls)
        cls.register_patch(cls._config_module_links)
        cls.register_patch(cls._config_terraform_example_version_template)
        cls.register_patch(cls._config_disable_analytics)
        cls.register_patch(cls._config_allow_forceful_module_provider_redirect_deletion)
        cls.register_patch(cls._config_default_ui_details_view)

        super(TestModuleProvider, cls).setup_class()

    @pytest.mark.parametrize('url,expected_title', [
        ('/modules/moduledetails/noversion/testprovider', 'moduledetails/noversion/testprovider - Terrareg'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0', 'moduledetails/fullypopulated/testprovider - Terrareg'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example', 'moduledetails/fullypopulated/testprovider/examples/test-example - Terrareg'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1', 'moduledetails/fullypopulated/testprovider/modules/example-submodule1 - Terrareg'),
    ])
    def test_page_titles(self, url, expected_title):
        """Check page titles on pages."""
        self.selenium_instance.get(self.get_url(url))
        self.assert_equals(lambda: self.selenium_instance.title, expected_title)

    @pytest.mark.parametrize('url,expected_breadcrumb', [
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0',
         'Modules\nmoduledetails\nfullypopulated\ntestprovider\n1.5.0'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
         'Modules\nmoduledetails\nfullypopulated\ntestprovider\n1.5.0\nexamples/test-example'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
         'Modules\nmoduledetails\nfullypopulated\ntestprovider\n1.5.0\nmodules/example-submodule1'),
    ])
    def test_breadcrumbs(self, url, expected_breadcrumb):
        """Test breadcrumb displayed on page"""
        self.selenium_instance.get(self.get_url(url))
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'breadcrumb-ul').text, expected_breadcrumb)

    def _get_settings_field_by_label(self, label, form="settings-form", type_="input"):
        """Return input element by label."""
        form = self.wait_for_element(By.ID, form)
        return form.find_element(By.XPATH, ".//label[text()='{label}']/parent::*//{type_}".format(label=label, type_=type_))

    def _fill_out_settings_field_by_label(self, label, input):
        """Find input field by label and fill out input."""
        input_field = self._get_settings_field_by_label(label)
        input_field.send_keys(input)

    def _confirm_move(self):
        """Click confirm move checkbox"""
        confirm = self._get_settings_field_by_label("Confirm", form="settings-move-form")
        confirm.click()

    def _click_save_move(self):
        """Click save move button on settings tab."""
        self.selenium_instance.find_element(By.ID, 'settings-move-submit').click()

    def _click_save_settings(self):
        """Click save button on settings tab."""
        self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()

    def test_module_without_versions(self):
        """Test page functionality on a module without published versions."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

        # Ensure integrations tab link is display and tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-integrations')
        self.wait_for_element(By.ID, 'module-tab-integrations')

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

        # Ensure warning about no available version
        no_versions_div = self.wait_for_element(By.ID, 'no-version-available')
        assert no_versions_div.text == 'There are no versions of this module'
        assert no_versions_div.is_displayed() == True

        # Ensure none of the following elements are displayed
        for element_id in ['module-title', 'module-provider', 'module-description', 'published-at',
                           'module-owner', 'source-url', 'submodule-back-to-parent',
                           'submodule-select-container', 'example-select-container',
                           'module-download-stats-container', 'usage-example-container']:
            assert self.selenium_instance.find_element(By.ID, element_id).is_displayed() == False

    @pytest.mark.parametrize('attribute_to_remove,related_element,expect_displayed,expected_display_value', [
        # Without any modified fields
        (None, None, None, None),

        # Remove description
        ('description', 'module-description', False, ''),

        # Remove owner
        ('owner', 'module-owner', False, ''),

        # Remove source URL
        ('repo_base_url_template', 'source-url', False, '')
    ])
    def test_module_with_versions(self, attribute_to_remove, related_element, expect_displayed, expected_display_value):
        """Test page functionality on a module with versions."""
        # Remove custom browse/base URL from module provider
        module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
        original_provider_repo_base_url_template = module_provider._get_db_row()['repo_base_url_template']
        original_provider_repo_browse_url_template = module_provider._get_db_row()['repo_browse_url_template']
        module_provider.update_attributes(repo_base_url_template=None, repo_browse_url_template=None)

        if attribute_to_remove:
            module_version = ModuleVersion.get(module_provider, '1.5.0')
            original_value = module_version._get_db_row()[attribute_to_remove]
            module_version.update_attributes(**{attribute_to_remove: None})

        try:
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider'))

            # Ensure readme link is displayed
            self.assert_equals(lambda: self.wait_for_element(By.ID, f'module-tab-link-readme').is_displayed(), True)

            # Ensure tab content is displayed
            assert self.selenium_instance.find_element(By.ID, f'module-tab-readme').is_displayed() == True

            # Ensure all other tabs aren't shown
            for tab_name in ['inputs', 'outputs', 'providers', 'resources', 'analytics', 'usage-builder', 'integrations']:
                # Ensure tab links are displayed
                self.assert_equals(lambda: self.wait_for_element(By.ID, f'module-tab-link-{tab_name}').is_displayed(), True)

                # Ensure tab content isn't displayed
                assert self.selenium_instance.find_element(By.ID, f'module-tab-{tab_name}').is_displayed() == False

            # Ensure exaple files tab link isn't displayed
            assert self.selenium_instance.find_element(By.ID, 'module-tab-link-example-files').is_displayed() == False

            # Ensure security issues aren't displayed
            assert self.selenium_instance.find_element(By.ID, 'security-issues').is_displayed() == False

            # Ensure cost isn't displayed
            assert self.selenium_instance.find_element(By.ID, 'yearly-cost').is_displayed() == False

            # Check basic details of module
            expected_element_details = {
                'module-title': 'fullypopulated',
                'module-labels': 'Contributed',
                'module-provider': 'Provider: testprovider',
                'module-description': 'This is a test module version for tests.',
                'published-at': 'Published January 05, 2022 by moduledetails',
                'module-owner': 'Module managed by This is the owner of the module',
                'source-url': 'Source code: https://link-to.com/source-code-here'
            }
            for element_name in expected_element_details:
                element = self.selenium_instance.find_element(By.ID, element_name)

                if element_name == related_element:
                    assert element.is_displayed() == expect_displayed
                    assert element.text == expected_display_value
                else:
                    assert element.is_displayed() == True
                    assert element.text == expected_element_details[element_name]

            # Login
            self.perform_admin_authentication(password='unittest-password')

            # Ensure settings tab link is displayed
            self.selenium_instance.get(self.get_url('/modules/moduledetails/noversion/testprovider'))

            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-link-settings').is_displayed(), True)
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, f'module-tab-settings').is_displayed(), False)
            self.wait_for_element(By.ID, 'no-version-available', ensure_displayed=False).is_displayed == False

        finally:
            module_provider.update_attributes(
                repo_base_url_template=original_provider_repo_base_url_template,
                repo_browse_url_template=original_provider_repo_browse_url_template)
            if attribute_to_remove:
                module_version.update_attributes(**{attribute_to_remove: original_value})

    @pytest.mark.parametrize('url,expected_label_displayed,expected_critical,expected_high,expected_medium_low', [
        ('/modules/moduledetails/withsecurityissues/testprovider', False, 0, 0, 0),
        ('/modules/moduledetails/withsecurityissues/testprovider/1.2.0/submodule/modules/withanotherissue', True, 0, 0, 1),
        ('/modules/moduledetails/withsecurityissues/testprovider/1.1.0/example/examples/withsecissue', True, 0, 1, 2),
        ('/modules/moduledetails/withsecurityissues/testprovider/1.0.0', True, 1, 3, 2),
    ])
    def test_module_with_security_issues(self, url, expected_label_displayed, expected_critical, expected_high, expected_medium_low):
        """Test module with security issues."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for inputs tab label
        self.wait_for_element(By.ID, 'module-tab-link-inputs')

        # Ensure security issues is disaplyed as expected
        security_issues = self.wait_for_element(By.ID, 'security-issues', ensure_displayed=False)
        assert security_issues.is_displayed() == expected_label_displayed

        # If the label should not be displayed, return early
        if not expected_label_displayed:
            return

        # Check label text
        expected_text = 'Security Issues'
        if expected_critical:
            expected_text += f'\n{expected_critical} Critical'
        if expected_high:
            expected_text += f'\n{expected_high} High'
        if expected_medium_low:
            expected_text += f'\n{expected_medium_low} Medium/Low'
        assert security_issues.text == expected_text

        # Check each of the count labels
        expected_base_color = None
        for label_id, label_name, color, expected_count in [
                ('result-card-label-security-issues-critical-count', 'Critical', 'danger', expected_critical),
                ('result-card-label-security-issues-high-count', 'High', 'warning', expected_high),
                ('result-card-label-security-issues-low-count', 'Medium/Low', 'info', expected_medium_low)
                ]:
            label = security_issues.find_element(By.ID, label_id)
            if expected_count:
                assert label.is_displayed() == True
                assert label.text == f'{expected_count} {label_name}'
                # Check only color associated to label is the expected one
                assert [
                    class_name
                    for class_name in label.get_attribute('class').split(' ')
                    if class_name.startswith('is-') and class_name != 'is-light'
                ] == [f'is-{color}']
                # Set expected parent color to this label color,
                # if not already set
                if expected_base_color is None:
                    expected_base_color = color

            else:
                assert label.is_displayed() == False

        # Ensure base color of security tag is correct
        icon_label = security_issues.find_element(By.ID, 'result-card-label-security-issues-icon')
        assert [
            class_name
            for class_name in icon_label.get_attribute('class').split(' ')
            if class_name.startswith('is-') and class_name != 'is-light'
        ] == [f'is-{expected_base_color}']

    @pytest.mark.parametrize('url,cost', [
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example', '2373.60'),
        ('/modules/moduledetails/infracost/testprovider/1.0.0/example/examples/with-cost', '150.15'),
        ('/modules/moduledetails/infracost/testprovider/1.0.0/example/examples/free', '0.00'),
        ('/modules/moduledetails/infracost/testprovider/1.0.0/example/examples/no-infracost-data', None),
    ])
    def test_example_with_cost_analysis(self, url, cost):
        """Test module with cost analysis."""
        self.selenium_instance.get(self.get_url(url))

        # Ensure yearly cost is displayed
        cost_text = self.wait_for_element(By.ID, 'yearly-cost', ensure_displayed=False)
        if cost is None:
            self.assert_equals(lambda: cost_text.is_displayed(), False)
        else:
            self.assert_equals(lambda: cost_text.is_displayed(), True)
            assert cost_text.text == f'Estimated yearly cost:\n${cost}'

        # Ensure cost label is displayed
        cost_label = self.wait_for_element(By.ID, 'yearly-cost-label', ensure_displayed=False)
        if cost is None:
            self.assert_equals(lambda: cost_label.is_displayed(), False)
        else:
            self.assert_equals(lambda: cost_label.is_displayed(), True)
            assert cost_label.text == f'${cost}/yr'

    @pytest.mark.parametrize('base_url, drop_down_type, drop_down_text, expected_url, expected_version_string, expected_submodule_title, expected_module_title, expected_provider', [
        # Test sub-module
        ('/modules/moduledetails/fullypopulated/testprovider',
         'submodule-select', 'modules/example-submodule1',
         '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
         'Version: 1.5.0', 'Submodule: modules/example-submodule1',
         'fullypopulated', 'Provider: testprovider'),
        # Test example
        ('/modules/moduledetails/fullypopulated/testprovider',
         'example-select', 'examples/test-example',
         '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
         'Version: 1.5.0', 'Example: examples/test-example',
         'fullypopulated', 'Provider: testprovider'),
        # Test submodule using 'latest'
        ('/modules/moduledetails/fullypopulated/testprovider/latest/submodule/modules/example-submodule1',
         None, None,
         '/modules/moduledetails/fullypopulated/testprovider/latest/submodule/modules/example-submodule1',
         'Version: 1.5.0', 'Submodule: modules/example-submodule1',
         'fullypopulated', 'Provider: testprovider'),
        # Test example using 'latest
        ('/modules/moduledetails/fullypopulated/testprovider/latest/example/examples/test-example',
         None, None,
         '/modules/moduledetails/fullypopulated/testprovider/latest/example/examples/test-example',
         'Version: 1.5.0', 'Example: examples/test-example',
         'fullypopulated', 'Provider: testprovider'),
    ])
    def test_submodule_example_basic_details(self, base_url, drop_down_type, drop_down_text, expected_url, expected_version_string, expected_submodule_title, expected_module_title, expected_provider):
        """Test basic details shown on submodule/example page."""
        self.selenium_instance.get(self.get_url(base_url))

        # If a drop-down type/value is provided, select from the dropdown
        if drop_down_type:
            # Select from dropdown
            select = Select(self.wait_for_element(By.ID, drop_down_type))
            select.select_by_visible_text(drop_down_text)

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(expected_url))

        # Check title, version, module title, provider
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'version-text').text, expected_version_string)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'current-submodule').text, expected_submodule_title)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'module-title').text, expected_module_title)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'module-provider').text, expected_provider)

    @pytest.mark.parametrize('url,git_provider_id,module_provider_browse_url_template,module_provider_base_url_template,module_version_browse_url_template,module_version_base_url_template,allow_custom_git_urls_module_provider,allow_custom_git_urls_module_version,expected_source', [
        # Test with all URLs configured and all custom URLs allowed
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/suffix'
        ),
        # - non-latest version, defaults to using module provider, as the module version
        # has been configured in the latest version
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.2.0',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.2.0/suffix'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/examples/test-examplesuffix'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/modules/example-submodule1suffix'
        ),
        # Test with all URLs configured and custom module version URLs disabled
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/suffix'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/examples/test-examplesuffix'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/modules/example-submodule1suffix'
        ),
        # Test with all URLs configured and all custom module URLs disabled
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/examples/test-example'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            'https://module-version.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/modules/example-submodule1'
        ),
        # Test with all URLs configured except module version browse and all custom URLs enabled
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/suffix'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/examples/test-examplesuffix'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            1,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse/{tag}/{path}suffix',
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/modules/example-submodule1suffix'
        ),
        # Test with all URLs configured except module version/provider browse and all custom URLs enabled
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            1,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            1,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/examples/test-example'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            1,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://localhost.com/moduledetails/fullypopulated-testprovider/browse/1.5.0/modules/example-submodule1'
        ),
        # Test with no browse URLs configured and all custom URLs enabled, testing revert to base URL
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            True,
            'https://module-version.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # Test with no browse URLs configured and module version custom URLs disabled, testing revert to base URL
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            True,
            False,
            'https://module-provider.com/moduledetails/fullypopulated-testprovider/browse'
        ),
        # Test with no browse URLs configured and all custom URLs disabled, testing revert to base URL
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://base-url.com/moduledetails/fullypopulated-testprovider'
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://base-url.com/moduledetails/fullypopulated-testprovider'
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            4,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            'https://base-url.com/moduledetails/fullypopulated-testprovider'
        ),
        # Test with no browse URLs configured and all custom URLs disabled and not git provider
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            None,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            None
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            None,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            None
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            None,
            None,
            'https://module-provider.com/{namespace}/{module}-{provider}/browse',
            None,
            'https://module-version.com/{namespace}/{module}-{provider}/browse',
            False,
            False,
            None
        ),
        # Test with no URLs configured and all custom URLs enabled and not git provider
        # - base URL
        (
            '/modules/moduledetails/fullypopulated/testprovider',
            None,
            None,
            None,
            None,
            None,
            True,
            True,
            None
        ),
        # - example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            None,
            None,
            None,
            None,
            None,
            True,
            True,
            None
        ),
        # - submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            None,
            None,
            None,
            None,
            None,
            True,
            True,
            None
        ),
    ])
    def test_source_code_urls(self, url, git_provider_id,
                              module_provider_browse_url_template, module_provider_base_url_template,
                              module_version_browse_url_template, module_version_base_url_template,
                              allow_custom_git_urls_module_provider, allow_custom_git_urls_module_version,
                              expected_source):
        """Test source code URL shown."""

        module_provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
        module_version = ModuleVersion.get(module_provider, '1.5.0')
        module_provider_row = module_provider._get_db_row()
        module_version_row = module_version._get_db_row()

        original_git_provider_id = module_provider_row['git_provider_id']
        original_module_provider_browse_url_template = module_provider_row['repo_browse_url_template']
        original_module_provider_base_url_template = module_provider_row['repo_base_url_template']
        original_module_version_browse_url_template = module_version_row['repo_browse_url_template']
        original_module_version_base_url_template = module_version_row['repo_base_url_template']


        try:
            module_version.update_attributes(
                repo_browse_url_template=module_version_browse_url_template,
                repo_base_url_template=module_version_base_url_template)
            module_provider.update_attributes(
                git_provider_id=git_provider_id,
                repo_browse_url_template=module_provider_browse_url_template,
                repo_base_url_template=module_provider_base_url_template
            )

            with self.update_multiple_mocks((self._config_allow_custom_repo_urls_module_provider, 'new', allow_custom_git_urls_module_provider), \
                    (self._config_allow_custom_repo_urls_module_version, 'new', allow_custom_git_urls_module_version)):

                self.selenium_instance.get(self.get_url(url))

                # Ensure source code URL is correct
                source_url = self.wait_for_element(By.ID, 'source-url', ensure_displayed=False)
                if expected_source:
                    self.assert_equals(lambda: source_url.is_displayed(), True)
                    assert source_url.text == f'Source code: {expected_source}'
                else:
                    # Wait for inputs tab
                    self.wait_for_element(By.ID, 'module-tab-link-inputs')
                    assert source_url.is_displayed() == False

        finally:
            module_version.update_attributes(
                repo_browse_url_template=original_module_version_browse_url_template,
                repo_base_url_template=original_module_version_base_url_template)
            module_provider.update_attributes(
                git_provider_id=original_git_provider_id,
                repo_browse_url_template=original_module_provider_browse_url_template,
                repo_base_url_template=original_module_provider_base_url_template
            )

    @pytest.mark.parametrize('url,expected_readme_content', [
        # Root module
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0', """
This is an example README!
Following this example module call:
module "test_example_call" {
  source  = "localhost/my-tf-application__moduledetails/fullypopulated/testprovider"
  version = ">= 1.5.0, < 2.0.0, unittest"

  name = "example-name"
}
This should work with all versions > 5.2.0 and <= 6.0.0
module "text_ternal_call" {
  source  = "a-public/module"
  version = "> 5.2.0, <= 6.0.0"

  another = "example-external"
}
""".strip()),
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

        # Ensure that README content is no longer visible
        assert self.selenium_instance.find_element(By.ID, 'module-tab-readme').is_displayed() == False

        # Click on README tab again
        self.selenium_instance.find_element(By.ID, 'module-tab-link-readme').click()

        # Ensure README content is visible again and content is correct
        readme_content = self.selenium_instance.find_element(By.ID, 'module-tab-readme')
        assert readme_content.is_displayed() == True
        assert readme_content.text == expected_readme_content

    def test_additional_links(self):
        """Test additions links in module provider page."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for input tab to be present
        self.wait_for_element(By.ID, 'module-tab-link-inputs')

        # Ensure no links are present
        links = self.selenium_instance.find_element(By.ID, 'custom-links')
        assert len([el for el in links.find_elements(By.CLASS_NAME, 'custom-link')]) == 0

        with self.update_mock(self._config_module_links, 'new', json.dumps([
                    {"text": "Placeholders in text module:{module} provider:{provider} ns:{namespace}",
                     "url": "https://example.com/placeholders-in-link/{namespace}/{module}-{provider}/end"},
                    {"text": "Link that does not apply",
                     "url": "https://mydomain.example.com/",
                     "namespaces": ["not-the-namespace", "another-namespace"]},
                    {"text": "Link that applies to this namespace",
                     "url": "https://applies-to-this-module.com",
                     "namespaces": ["another-namespace", "moduledetails", "another-another-one"]}
                ])):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Ensure all links are present
            self.assert_equals(lambda: [
                [link.text, link.get_attribute("href")]
                for link in self.selenium_instance.find_element(By.ID, "custom-links").find_elements(By.CLASS_NAME, "custom-link")
            ], [
                ['Placeholders in text module:fullypopulated provider:testprovider ns:moduledetails',
                 'https://example.com/placeholders-in-link/moduledetails/fullypopulated-testprovider/end'],
                ['Link that applies to this namespace',
                 'https://applies-to-this-module.com/']
            ])

    @pytest.mark.parametrize('url,expected_inputs', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['name_of_application', '<p>Enter the application name\nThis should be a real name</p>\n<p>Double line break</p>', 'string', 'Required'],
                ['string_with_default_value', '<p>Override the default string</p>', 'string', '"this is the default"'],
                ['example_boolean_input', '<p>Override the truthful boolean</p>', 'bool', 'true'],
                ['example_list_input', '<p>Override the stringy list</p>', 'list', '["value 1","value 2"]']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['input_for_example', 'Enter the example name', 'string', 'Required']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['input_for_submodule', 'Enter the submodule name', 'string', 'Required']]
        )
    ])
    def test_inputs_tab(self, url, expected_inputs):
        """Ensure inputs tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for input tab button to be visible
        input_tab_button = self.wait_for_element(By.ID, 'module-tab-link-inputs')

        # Ensure the inputs tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-inputs', ensure_displayed=False).is_displayed() == False

        # Click on inputs tab
        input_tab_button.click()

        # Obtain tab content
        inputs_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-inputs')

        # Ensure tab is displayed
        self.assert_equals(lambda: inputs_tab_content.is_displayed(), True)

        inputs_table = inputs_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = inputs_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_inputs) + 1)

        for row_itx, expected_row in enumerate([['Name', 'Description', 'Type', 'Default value']] + expected_inputs):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 4

            # Check columns of row match expected text
            row_text = [col.get_attribute("innerHTML") for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('url,expected_outputs', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['generated_name', '<p>Name with randomness\nThis random will not change.</p>\n<p>Double line break</p>'],
                ['no_desc_output', '']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['example_output', 'Example name with randomness']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['submodule_output', 'Submodule name with randomness']]
        )
    ])
    def test_outputs_tab(self, url, expected_outputs):
        """Ensure outputs tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for outputs tab button to be visible
        outputs_tab_button = self.wait_for_element(By.ID, 'module-tab-link-outputs')

        # Ensure the outputs tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-outputs', ensure_displayed=False).is_displayed() == False

        # Click on outputs tab
        outputs_tab_button.click()

        # Obtain tab content
        outputs_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-outputs')

        # Ensure tab is displayed
        self.assert_equals(lambda: outputs_tab_content.is_displayed(), True)

        outputs_table = outputs_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = outputs_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_outputs) + 1)

        for row_itx, expected_row in enumerate([['Name', 'Description']] + expected_outputs):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 2 columns
            assert len(row_columns) == 2

            # Check columns of row match expected text
            row_text = [col.get_attribute('innerHTML') for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('tab', [
        'input',
        'output',
    ])
    @pytest.mark.parametrize('default_config_view', [
        terrareg.config.DefaultUiInputOutputView.TABLE,
        terrareg.config.DefaultUiInputOutputView.EXPANDED,
    ])
    def test_switch_view_types(self, tab, default_config_view):
        """Test switching view types"""
        # Open page for inputs

        def assert_looks_like_table_view():
            """Check both inputs and outputs tab look correct"""
            for tab_to_check in ["inputs", "outputs"]:
                tab_button = self.wait_for_element(By.ID, f'module-tab-link-{tab_to_check}')
                tab_button.click()

                content = self.selenium_instance.find_element(By.ID, f'module-tab-{tab_to_check}-content')
                first_row = list(content.find_element(By.TAG_NAME, "table").find_element(By.TAG_NAME, "tbody").find_elements(By.TAG_NAME, "tr"))[0]
                # Ensure first row of inputs is shown (excluding heading)
                assert first_row.find_elements(By.TAG_NAME, "td")[0].text == ("name_of_application" if tab_to_check == "inputs" else "generated_name")

            # Return to main tab
            tab_button = self.selenium_instance.find_element(By.ID, f'module-tab-link-{tab}s')
            tab_button.click()

        def assert_looks_like_detailed_view():
            for tab_to_check in ["inputs", "outputs"]:
                tab_button = self.wait_for_element(By.ID, f'module-tab-link-{tab_to_check}')
                tab_button.click()

                content = self.selenium_instance.find_element(By.ID, f'module-tab-{tab_to_check}-content')
                # Check heading
                assert content.find_elements(By.TAG_NAME, "h4")[0].text == ("name_of_application" if tab_to_check == "inputs" else "generated_name")
                if tab_to_check == "inputs":
                    assert "Type: string" in content.text
                else:
                    assert "Description\nName with randomness" in content.text

            # Return to main tab
            tab_button = self.selenium_instance.find_element(By.ID, f'module-tab-link-{tab}s')
            tab_button.click()

        with self.update_mock(self._config_default_ui_details_view, 'new', default_config_view):
            try:
                self.delete_cookies_and_local_storage()
                self.selenium_instance.get(self.get_url("/modules/moduledetails/fullypopulated/testprovider/1.5.0"))

                # Wait for tab button to be visible
                tab_button = self.wait_for_element(By.ID, f'module-tab-link-{tab}s')

                # Ensure the tab content is not visible
                assert self.wait_for_element(By.ID, f'module-tab-{tab}s-left', ensure_displayed=False).is_displayed() == False

                # Click on tab link
                tab_button.click()

                # Obtain tab content
                tab_content = self.selenium_instance.find_element(By.ID, f'module-tab-{tab}s')

                # Find select box for view type
                select = Select(tab_content.find_element(By.TAG_NAME, "select"))

                if default_config_view is terrareg.config.DefaultUiInputOutputView.TABLE:
                    # Ensure selected type is default
                    assert select.first_selected_option.text == "Table View"

                    # Ensure the table view is activated
                    assert_looks_like_table_view()
                else:
                    # Ensure selected type is default
                    assert select.first_selected_option.text == "Expanded View"

                    # Ensure the table view is activated
                    assert_looks_like_detailed_view()

                # Select new view type
                select.select_by_visible_text("Expanded View")

                # Wait for page to re-render

                assert_looks_like_detailed_view()

                # Reload page and assert that the new format is kept
                self.selenium_instance.refresh()

                # Wait for outputs tab button to be visible
                tab_button = self.wait_for_element(By.ID, f'module-tab-link-{tab}s')
                tab_button.click()
                assert_looks_like_detailed_view()

                # Switch back to table view
                tab_content = self.selenium_instance.find_element(By.ID, f'module-tab-{tab}s')
                select = Select(tab_content.find_element(By.TAG_NAME, "select"))
                select.select_by_visible_text("Table View")

                assert_looks_like_table_view()

                # Refresh and re-check
                self.selenium_instance.refresh()
                assert_looks_like_table_view()

            finally:
                self.delete_cookies_and_local_storage()

    @pytest.mark.parametrize('url,expected_resources', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['string', 'random_suffix', 'random', 'hashicorp/random', 'managed', 'latest', '']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['string', 'example_random_suffix', 'example_random', 'hashicorp/example_random', 'managed', 'latest', '']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['string', 'submodule_random_suffix', 'submodule_random', 'hashicorp/submodule_random', 'managed', 'latest', '']]
        )
    ])
    def test_resources_tab(self, url, expected_resources):
        """Ensure resources tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for resources tab button to be visible
        resources_tab_button = self.wait_for_element(By.ID, 'module-tab-link-resources')

        # Ensure the resources tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-resources', ensure_displayed=False).is_displayed() == False

        # Click on resources tab
        resources_tab_button.click()

        # Obtain tab content
        resources_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-resources')

        # Ensure tab is displayed
        self.assert_equals(lambda: resources_tab_content.is_displayed(), True)

        resources_table = resources_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = resources_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_resources) + 1)

        for row_itx, expected_row in enumerate(
                [['Type', 'Name', 'Provider', 'Source', 'Mode', 'Version', 'Description']] +
                expected_resources):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 7 columns
            assert len(row_columns) == 7

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    @pytest.mark.parametrize('url,expected_providers', [
        # Root module
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            [
                ['random', 'hashicorp', '', '>= 5.2.1, < 6.0.0'],
                ['unsafe', 'someothercompany', '', '2.0.0']
            ]
        ),
        # Module example
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example',
            [['example_random', 'hashicorp', '', '']]
        ),
        # Submodule
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
            [['submodule_random', 'hashicorp', '', '']]
        )
    ])
    def test_providers_tab(self, url, expected_providers):
        """Ensure providers tab is displayed correctly."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for providers tab button to be visible
        providers_tab_button = self.wait_for_element(By.ID, 'module-tab-link-providers')

        # Ensure the providers tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-providers', ensure_displayed=False).is_displayed() == False

        # Click on providers tab
        providers_tab_button.click()

        # Obtain tab content
        providers_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-providers')

        # Ensure tab is displayed
        self.assert_equals(lambda: providers_tab_content.is_displayed(), True)

        providers_table = providers_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = providers_table.find_elements(By.TAG_NAME, 'tr')

        # Ensure table has 1 heading and 1 row per expected variable
        assert len(table_rows) == (len(expected_providers) + 1)

        for row_itx, expected_row in enumerate(
                [['Name', 'Namespace', 'Source', 'Version']] +
                expected_providers):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'th' if row_itx == 0 else 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 4

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    def test_integrations_tab(self):
        """Ensure integrations tab is displayed correctly."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure tab is displayed
        self.assert_equals(lambda: integrations_tab_content.is_displayed(), True)

        integrations_table = integrations_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = integrations_table.find_elements(By.TAG_NAME, 'tr')

        expected_integrations = [
            [
                'Create module version using source archive',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/upload")}\n' +
                'Source ZIP file must be provided as data.'
            ],
            [
                'Trigger module version import',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/import")}'
            ],
            [
                'Bitbucket hook trigger',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/bitbucket")}'
            ],
            [
                'Github hook trigger',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/github")}\n' + 
                'Only accepts `Releases` events, all other events will return an error.'
            ],
            [
                'Gitlab hook trigger (Coming soon)',
                f'{self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/hooks/gitlab")}'
            ],
            [
                'Mark module version as published',
                f'POST {self.get_url("/v1/terrareg/modules/moduledetails/fullypopulated/testprovider/${version}/publish")}'
            ]
        ]

        # Check number of rows in tab
        assert len(table_rows) == len(expected_integrations)

        for row_itx, expected_row in enumerate(expected_integrations):
            # Find all columns (heading row uses th and subsequent rows use td)
            row_columns = table_rows[row_itx].find_elements(By.TAG_NAME, 'td')

            ## Ensure each table row has 4 columns
            assert len(row_columns) == 2

            # Check columns of row match expected text
            row_text = [col.text for col in row_columns]
            assert row_text == expected_row

    def test_integration_tab_index_version(self):
        """Test indexing a new module version from the integration tab"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure publish button exists and is not checked
        assert integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').is_selected() == False

        # Type version number and submit form
        integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('5.2.1')
        integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

        # Wait for success message to be displayed
        success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
        self.assert_equals(lambda: success_message.is_displayed(), True)
        self.assert_equals(lambda: success_message.text, 'Successfully indexed version')

        # Check error message is not displayed
        error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
        assert error_message.is_displayed() == False

        # Ensure version create endpoint was called and publish was not
        self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')
        self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_and_publish(self):
        """Test indexing and publishing a new module version from the integration tab"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

        # Ensure publish button exists and is not checked
        publish_checkbox = integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish')
        assert publish_checkbox.is_selected() == False

        # Check publish checkbox
        publish_checkbox.click()

        # Type version number and submit form
        integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('5.2.1')
        integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

        # Wait for success message to be displayed
        success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
        self.assert_equals(lambda: success_message.is_displayed(), True)
        self.assert_equals(lambda: success_message.text, 'Successfully indexed and published version.')

        # Check error message is not displayed
        error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
        assert error_message.is_displayed() == False

        # Ensure version create and publish endpoints were called
        self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')
        self._api_version_publish_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='5.2.1')

    def test_settings_module_version_already_published(self):
        """Ensure settings tab does not contain publish button for already published module"""
        self.perform_admin_authentication(password='unittest-password')

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Wait for settings tab button to be visible
        settings_tab_button = self.wait_for_element(By.ID, 'module-tab-link-settings')

        # Click on settings tab
        settings_tab_button.click()

        # Ensure the settings tab content is visible
        settings_tab_content = self.wait_for_element(By.ID, 'module-tab-settings', ensure_displayed=True)

        assert settings_tab_content.find_element(By.ID, "settings-publish-button-container").is_displayed() == False

    def test_settings_module_version_publish_action(self):
        """Test settings tab for unpublished versino and user is able to publish a version"""
        self.perform_admin_authentication(password='unittest-password')

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.6.0'))

        # Wait for settings tab button to be visible
        settings_tab_button = self.wait_for_element(By.ID, 'module-tab-link-settings')

        # Click on settings tab
        settings_tab_button.click()

        # Ensure the settings tab content is visible
        settings_tab_content = self.wait_for_element(By.ID, 'module-tab-settings', ensure_displayed=True)

        # Ensure publish button exists and is not checked
        assert settings_tab_content.find_element(By.ID, "settings-publish-button-container").is_displayed() == True

        # Check publish button
        publish_button = settings_tab_content.find_element(By.ID, "settings-publish-button")
        assert publish_button.text == "Publish"
        publish_button.click()

        # Ensure version publish endpoint was called
        self._api_version_publish_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='1.6.0')

    @pytest.mark.parametrize('user_groups,expected_result', [
        ([], False),
        (['siteadmin'], True),
        (['nopermissions'], False),
        (['moduledetailsmodify'], True),
        (['moduledetailsfull'], True)
    ])
    def test_integration_tab_publish_button_permissions(self, user_groups, expected_result):
        """Test disabling of publish button, logged in with various user groups."""
        with self.update_multiple_mocks((self._config_publish_api_keys_mock, 'new', ['abcdefg']), \
                (self._config_enable_access_controls, 'new', True)):
            # Clear cookies to remove authentication
            self.selenium_instance.delete_all_cookies()

            with self.log_in_with_openid_connect(user_groups):
                self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

                # Wait for integrations tab button to be visible
                integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

                # Ensure the integrations tab content is not visible
                assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

                # Click on integrations tab
                integrations_tab_button.click()

                integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

                # Ensure publish button exists and is not disaplyed
                assert integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').is_displayed() == expected_result

                # Ensure publish button container is not displayed
                assert integrations_tab_content.find_element(By.ID, 'integrations-index-module-version-publish').is_displayed() == expected_result

    def test_integration_tab_index_version_with_publish_disabled(self):
        """Test indexing a new module version from the integration tab whilst publishing is not possible"""
        with self.update_mock(self._config_publish_api_keys_mock, 'new', ['abcdefg']):
            # Clear cookies to remove authentication
            self.selenium_instance.delete_all_cookies()

            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Ensure publish button exists and is not disaplyed
            assert integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').is_displayed() == False

            # Ensure publish button container is not displayed
            assert integrations_tab_content.find_element(By.ID, 'integrations-index-module-version-publish').is_displayed() == False

            # Type version number and submit form
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for success message to be displayed
            success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
            self.assert_equals(lambda: success_message.is_displayed(), True)
            self.assert_equals(lambda: success_message.text, 'Successfully indexed version')

            # Check error message is not displayed
            error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
            assert error_message.is_displayed() == False

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_with_indexing_failure(self):
        """Test indexing a new module version from the integration tab with an indexing failure"""
        # Update indexing mocks to cause indexing failure
        with self.update_mock(
                self._api_version_create_mock,
                'return_value',
                ({'status': 'Error', 'message': 'Unittest error message'}, 500)):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Type version number and submit form
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for error message to be displayed
            error_message = self.wait_for_element(By.ID, 'index-version-error', parent=integrations_tab_content)
            assert error_message.text == 'Unittest error message'

            success_message = integrations_tab_content.find_element(By.ID, 'index-version-success')
            assert success_message.is_displayed() == False

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_not_called()

    def test_integration_tab_index_version_with_publishing_failure(self):
        """Test indexing a new module version from the integration tab with a publishing failure"""
        # Update mock to cause publishing failure
        with self.update_mock(
                self._api_version_publish_mock,
                'return_value',
                ({'status': 'Error', 'message': 'Unittest publish error message'}, 500)):
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Wait for integrations tab button to be visible
            integrations_tab_button = self.wait_for_element(By.ID, 'module-tab-link-integrations')

            # Ensure the integrations tab content is not visible
            assert self.wait_for_element(By.ID, 'module-tab-integrations', ensure_displayed=False).is_displayed() == False

            # Click on integrations tab
            integrations_tab_button.click()

            integrations_tab_content = self.selenium_instance.find_element(By.ID, 'module-tab-integrations')

            # Type version number
            integrations_tab_content.find_element(By.ID, 'indexModuleVersion').send_keys('2.2.2')
            # Check publish button
            integrations_tab_content.find_element(By.ID, 'indexModuleVersionPublish').click()
            # Click indexing button
            integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

            # Wait for error message to be displayed
            error_message = self.wait_for_element(By.ID, 'index-version-error', parent=integrations_tab_content)
            assert error_message.text == 'Unittest publish error message'

            success_message = integrations_tab_content.find_element(By.ID, 'index-version-success')
            assert success_message.is_displayed() == True
            assert success_message.text == 'Successfully indexed version'

            # Ensure version create endpoint was called and publish was not
            self._api_version_create_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')
            self._api_version_publish_mock.assert_called_once_with(namespace='moduledetails', name='fullypopulated', provider='testprovider', version='2.2.2')

    @pytest.mark.parametrize('current_version,expected_versions,expected_selected_version', [
        # On root page without version
        (None, ['1.5.0 (latest)', '1.2.0'], '1.5.0 (latest)'),
        # With 'latest' in URL
        ('latest', ['1.5.0 (latest)', '1.2.0'], '1.5.0 (latest)'),
        # On latest version
        ('1.5.0', ['1.5.0 (latest)', '1.2.0'], '1.5.0 (latest)'),
        # On previous version
        ('1.2.0', ['1.5.0 (latest)', '1.2.0'], '1.2.0'),
        # On beta version
        ('1.6.1-beta', ['1.5.0 (latest)', '1.2.0', '1.6.1-beta (beta)'], '1.6.1-beta (beta)'),
        # On unpublished version
        ('1.6.0', ['1.5.0 (latest)', '1.2.0', '1.6.0 (unpublished)'], '1.6.0 (unpublished)'),
        # On beta unpublished version
        ('1.0.0-beta', ['1.5.0 (latest)', '1.2.0', '1.0.0-beta (beta) (unpublished)'], '1.0.0-beta (beta) (unpublished)')
    ])
    def test_version_dropdown(self, current_version, expected_versions, expected_selected_version):
        """Test version dropdown contains expected values."""
        # Go to page
        url = self.get_url('/modules/moduledetails/fullypopulated/testprovider')
        if current_version:
            url += f'/{current_version}'

        self.selenium_instance.get(url)

        select_dropdown = self.wait_for_element(By.ID, 'version-select')
        version_options = select_dropdown.find_elements(By.TAG_NAME, 'option')

        # Check the number of items
        #assert len(version_options) == len(expected_versions)

        # Ensure the current selected item is as expected
        assert expected_selected_version == Select(select_dropdown).first_selected_option.text

        # Check each of the select options
        for itx, version_item in enumerate(version_options):
            assert version_item.text == expected_versions[itx]

    @pytest.mark.parametrize('base_url,example_name,expected_file_list,example_root_module_call_name,expected_version_comment,expected_version_string', [
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.5.0',
            'examples/test-example',
            ['main.tf', 'data.tf', 'variables.tf'],
            'root',
            '',
            '>= 1.5.0, < 2.0.0, unittest'
        ),
        # Test old version
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.2.0',
            'examples/old-version-example',
            ['main.tf'],
            'old_version_root_call',
            '# This version of the module is not the latest version.\n  # To use this specific version, it must be pinned in Terraform\n  ',
            '1.2.0'
        ),
        # Test beta version
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.7.0-beta',
            'examples/beta-example',
            ['main.tf'],
            'beta_root_call',
            '# This version of the module is a beta version.\n  # To use this version, it must be pinned in Terraform\n  ',
            '1.7.0-beta'
        ),
        # Test un-published version
        (
            '/modules/moduledetails/fullypopulated/testprovider/1.6.0',
            'examples/unpublished-example',
            ['main.tf'],
            'unpublished_root_call',
            '# This version of this module has not yet been published,\n  # meaning that it cannot yet be used by Terraform\n  ',
            '1.6.0'
        ),
    ])
    def test_example_file_version_string(
            self, base_url, example_name, expected_file_list,
            example_root_module_call_name,
            expected_version_comment, expected_version_string):
        """Test example version string in examples."""
        self.selenium_instance.get(self.get_url(base_url))


        # Select example from dropdown
        select = Select(self.wait_for_element(By.ID, 'example-select'))
        select.select_by_visible_text(example_name)

        # Wait for page to reload and example title to be displayed
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'current-submodule').text,
            f'Example: {example_name}'
        )

        # Ensure example files tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-example-files').click()

        file_tab_content = self.wait_for_element(By.ID, 'module-tab-example-files')

        # Ensure all files are displayed in file list
        file_list = file_tab_content.find_element(By.ID, 'example-file-list-nav').find_elements(By.TAG_NAME, 'a')

        # Ensure files match expected order and name
        assert [file.text for file in file_list] == expected_file_list

        # Ensure contents of main.tf is shown in data
        expected_main_tf_content = f"""
# Call root module
module "{example_root_module_call_name}" {{
  source  = "localhost/my-tf-application__moduledetails/fullypopulated/testprovider"
  {expected_version_comment}version = "{expected_version_string}"
}}""".strip()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == expected_main_tf_content

    def test_example_file_contents(self):
        """Check example files are displayed correctly."""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Select example from dropdown
        select = Select(self.wait_for_element(By.ID, 'example-select'))
        select.select_by_visible_text('examples/test-example')

        # Wait for page to reload and example title to be displayed
        self.assert_equals(
            lambda: self.selenium_instance.find_element(By.ID, 'current-submodule').text,
            'Example: examples/test-example'
        )

        # Ensure example files tab is displayed
        self.wait_for_element(By.ID, 'module-tab-link-example-files').click()

        file_tab_content = self.wait_for_element(By.ID, 'module-tab-example-files')

        # Ensure all files are displayed in file list
        file_list = file_tab_content.find_element(By.ID, 'example-file-list-nav').find_elements(By.TAG_NAME, 'a')

        # Ensure files match expected order and name
        assert [file.text for file in file_list] == ['main.tf', 'data.tf', 'variables.tf']

        # Ensure contents of main.tf is shown in data
        expected_main_tf_content = f"""
# Call root module
module "root" {{
  source  = "localhost/my-tf-application__moduledetails/fullypopulated/testprovider"
  version = ">= 1.5.0, < 2.0.0, unittest"
}}
""".strip()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == expected_main_tf_content

        # Select main.tf file and check content
        file_list[0].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == expected_main_tf_content

        # Select data.tf and check content
        file_list[1].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == '# This contains data objects'

        # Select variables.tf and check content
        file_list[2].click()
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == """
variable "test" {
  description = "test variable"
  type = string
}""".strip()

    def test_example_file_content_heredoc(self):
        """Test example file with heredoc content"""
        self.selenium_instance.get(self.get_url("/modules/javascriptinjection/modulename/testprovider/1.5.0/example/examples/heredoc-tags"))

        file_tab_content = self.wait_for_element(By.ID, 'module-tab-example-files')
        assert file_tab_content.find_element(By.ID, 'example-file-content').text == """
module "test" {
  input = <<EOF
Test heredoc content
EOF
}
""".strip()

    def test_delete_module_version(self, mock_create_audit_event):
        """Test the delete version functionality in settings tab."""

        self.perform_admin_authentication(password='unittest-password')

        namespace = Namespace(name='moduledetails')
        module = Module(namespace=namespace, name='fullypopulated')
        module_provider = ModuleProvider.get(module=module, name='testprovider')

        with mock_create_audit_event:
            # Create test module version
            module_version = ModuleVersion(module_provider=module_provider, version='2.5.5')
            module_version.prepare_module()
            module_version.publish()
            module_version_pk = module_version.pk

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
        assert 'Type the version number of the current version to be deleted (e.g. 1.0.0) and click delete again:' in verification_div.text

        # Provide incorrect version number to confirmation
        verification_input = verification_div.find_element(By.ID, 'confirm-delete-module-version')
        verification_input.send_keys('5.4.4')

        # Click delete module version button again
        delete_button.click()

        # Wait and ensure page has not changed
        sleep(0.2)
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/fullypopulated/testprovider/2.5.5#settings'))

        # Ensure module version still exists
        module_version._cache_db_row = None
        assert module_version.pk == module_version_pk

        # Update input field to correct version
        verification_input.clear()
        verification_input.send_keys('2.5.5')

        # Click delete module version button again
        delete_button.click()

        # Ensure user is redirected to module page
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/fullypopulated/testprovider'))

        # Ensure module version no longer exists
        assert ModuleVersion.get(module_provider=module_provider, version='2.5.5') is None

    def test_git_path_setting(self):
        """Test setting git path in module provider settings."""
        self.perform_admin_authentication(password='unittest-password')

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider#settings'))
        self.wait_for_element(By.ID, 'module-tab-link-settings')

        settings_input = self._get_settings_field_by_label('Module path')
        assert settings_input.get_attribute('value') == ''

        # Enter git path
        settings_input.send_keys('test/sub/directory')
        self._click_save_settings()

        module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
        assert module_provider.git_path == 'test/sub/directory'

        self.selenium_instance.refresh()
        self.wait_for_element(By.ID, 'module-tab-link-settings')
        settings_input = self._get_settings_field_by_label('Module path')
        assert settings_input.get_attribute('value') == 'test/sub/directory'
        settings_input.clear()

        self._click_save_settings()
        module_provider._cache_db_row = None
        assert module_provider.git_path == None

    def test_archive_git_path_setting(self):
        """Test setting archive git path in module provider settings."""
        self.perform_admin_authentication(password='unittest-password')

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider#settings'))
        self.wait_for_element(By.ID, 'module-tab-link-settings')

        settings_input = self._get_settings_field_by_label('Only include module path in archive')
        assert settings_input.get_attribute('checked') == None

        # Enter git path
        settings_input.click()
        self._click_save_settings()

        module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
        assert module_provider.archive_git_path is True

        self.selenium_instance.refresh()
        self.wait_for_element(By.ID, 'module-tab-link-settings')
        settings_input = self._get_settings_field_by_label('Only include module path in archive')
        assert settings_input.get_attribute('checked') == 'true'
        settings_input.click()

        self._click_save_settings()
        module_provider._cache_db_row = None
        assert module_provider.archive_git_path is False

    def test_updating_module_name(self):
        """Test changing module name in module provider settings"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changename#settings"))

        module_name_input = self._get_settings_field_by_label("Module Name", form="settings-move-form")
        assert module_name_input.get_attribute("value") == "testmove"

        # Enter new name, confirm and submit form
        module_name_input.clear()
        module_name_input.send_keys("testmovenew")
        self._confirm_move()
        self._click_save_move()

        # Ensure browser has been redirect to new module
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/moduledetails/testmovenew/changename"))

        # Ensure name of module on page matches new name
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, "module-title").text, "testmovenew")

    def test_updating_module_provider(self):
        """Test changing module provider in module provider settings"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changeprovider#settings"))

        module_provider_input = self._get_settings_field_by_label("Provider", form="settings-move-form")
        assert module_provider_input.get_attribute("value") == "changeprovider"

        # Enter new name, confirm and submit form
        module_provider_input.clear()
        module_provider_input.send_keys("testnewprovider")
        self._confirm_move()
        self._click_save_move()

        # Ensure browser has been redirect to new module
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/moduledetails/testmove/testnewprovider"))

        # Ensure module provider on page matches new name
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, "module-provider").text, "Provider: testnewprovider")

    def test_updating_namespace(self):
        """Test changing namespace in module provider settings"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changenamespace#settings"))

        module_namespace_input = Select(self._get_settings_field_by_label("Namespace", form="settings-move-form", type_="select"))
        assert module_namespace_input.first_selected_option.text == "moduledetails"

        # Update namespace, confirm and submit form
        module_namespace_input.select_by_visible_text("scratchnamespace")
        self._confirm_move()
        self._click_save_move()

        # Ensure browser has been redirect to new module
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/scratchnamespace/testmove/changenamespace"))

        # Ensure namespace in Terraform usage example is correct (namespace is not shown in many places!)
        self.assert_equals(lambda: "scratchnamespace/testmove/changenamespace" in self.selenium_instance.find_element(By.ID, "usage-example-terraform").text, True)

    def test_updating_name_provider_and_namespace(self):
        """Test updating module provider name, provider and namespace"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changeall#settings"))

        # Enter new name
        module_name_input = self._get_settings_field_by_label("Module Name", form="settings-move-form")
        module_name_input.clear()
        module_name_input.send_keys("changeallnew")

        module_provider_input = self._get_settings_field_by_label("Provider", form="settings-move-form")
        module_provider_input.clear()
        module_provider_input.send_keys("changeallnewprovider")

        module_namespace_input = Select(self._get_settings_field_by_label("Namespace", form="settings-move-form", type_="select"))
        module_namespace_input.select_by_visible_text("scratchnamespace")

        self._confirm_move()
        self._click_save_move()

        # Ensure browser has been redirect to new module
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/scratchnamespace/changeallnew/changeallnewprovider"))

        # Ensure namespace in Terraform usage example is correct (namespace is not shown in many places!)
        self.assert_equals(lambda: "scratchnamespace/changeallnew/changeallnewprovider" in self.selenium_instance.find_element(By.ID, "usage-example-terraform").text, True)

        # Attempt to naviage to old page with anchor and query string, ensuring page is redirected
        # to new module provider name
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changeall?test=value&test2=value2#readme"))
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/scratchnamespace/changeallnew/changeallnewprovider?test=value&test2=value2#readme"))

        # Test redirect with version and example
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/changeall/1.0.0/example/examples/test?test=value&test2=value2#readme"))
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/scratchnamespace/changeallnew/changeallnewprovider/1.0.0/example/examples/test?test=value&test2=value2#readme"))

    def test_updating_module_name_to_duplicate(self):
        """Test updating module name to duplicate module name"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/duplicatemovetest#settings"))

        module_provider_input = self._get_settings_field_by_label("Provider", form="settings-move-form")

        # Enter new name, confirm and submit form
        module_provider_input.clear()
        module_provider_input.send_keys("duplicatetest")
        self._confirm_move()
        self._click_save_move()

        # Ensure error is visible and verify text
        error = self.selenium_instance.find_element(By.ID, "settings-move-error")
        assert error.is_displayed() == True
        assert error.text == "A module/provider already exists with the same name in the namespace"

        # Ensure URL has not been changed
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/moduledetails/testmove/duplicatemovetest#settings"))

    def test_updating_module_without_confirmation(self):
        """Test updating module name to new name without confirmation"""
        self.perform_admin_authentication(password="unittest-password")

        # Ensure user is redirected to module page
        self.selenium_instance.get(self.get_url("/modules/moduledetails/testmove/duplicatemovetest#settings"))

        module_provider_input = self._get_settings_field_by_label("Provider", form="settings-move-form")

        # Enter new name, confirm and submit form
        module_provider_input.clear()
        module_provider_input.send_keys("uniquename")
        self._click_save_move()

        # Ensure error is visible and verify text
        error = self.selenium_instance.find_element(By.ID, "settings-move-error")
        assert error.is_displayed() == True
        assert error.text == "The move action must be confirmed, by checking the confirm checkbox."

        # Ensure URL has not been changed
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url("/modules/moduledetails/testmove/duplicatemovetest#settings"))

    def test_delete_module_provider(self, mock_create_audit_event):
        """Test the delete provider functionality in settings tab."""

        self.perform_admin_authentication(password='unittest-password')

        with mock_create_audit_event:
            # Create test module version
            namespace = Namespace(name='moduledetails')
            module = Module(namespace=namespace, name='fullypopulated')
            module_provider = ModuleProvider.get(module=module, name='providertodelete', create=True)
            module_provider_pk = module_provider.pk

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/providertodelete'))

        # Click on settings tab
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Click on the delete module version button
        delete_button = self.selenium_instance.find_element(By.ID, 'module-provider-delete-button')
        delete_button.click()

        # Ensure the verification text is shown
        verification_div = self.selenium_instance.find_element(By.ID, 'confirm-delete-module-provider-div')
        assert 'Type the \'id\' of the module provider (e.g. namespace/module/provider) and click delete again:' in verification_div.text

        # Provide incorrect version number to confirmation
        verification_input = verification_div.find_element(By.ID, 'confirm-delete-module-provider')
        verification_input.send_keys('5.4.4')

        # Click delete module version button again
        delete_button.click()

        # Wait and ensure page has not changed and settings tab is still displayed
        sleep(0.2)
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/fullypopulated/providertodelete#settings'))
        self.wait_for_element(By.ID, 'module-tab-link-settings')

        # Ensure module version still exists
        module_provider._cache_db_row = None
        assert module_provider.pk == module_provider_pk

        # Update input field to correct version
        verification_input.clear()
        verification_input.send_keys('moduledetails/fullypopulated/providertodelete')

        # Click delete module version button again
        delete_button.click()

        # Ensure user is redirected to module page
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/fullypopulated/providertodelete'))

        # Ensure warning about non-existent module provider is displayed
        self.assert_equals(
            lambda: self.wait_for_element(By.ID, 'error-title').text,
            'Module/Provider does not exist')

        self.assert_equals(
            lambda: self.wait_for_element(By.ID, 'error-content').text,
            'The module moduledetails/fullypopulated/providertodelete does not exist')

        # Ensure module version no longer exists
        assert ModuleProvider.get(module=module, name='providertodelete') is None

    def assert_custom_url_input_visibility(self, should_be_shown: bool):
        """Check custom input visibility"""
        for element in self.selenium_instance.find_elements(By.CLASS_NAME, "settings-custom-git-provider-container"):
            self.assert_equals(lambda: element.is_displayed(), should_be_shown)
        for element_id in ["settings-base-url-template", "settings-clone-url-template", "settings-browse-url-template"]:
            assert self.selenium_instance.find_element(By.ID, element_id).is_displayed() is should_be_shown

    @pytest.mark.parametrize('allow_custom_git_url_setting', [True, False])
    def test_git_provider_config(self, allow_custom_git_url_setting):
        """Ensure git provider configuration work as expected."""

        with self.update_mock(self._config_allow_custom_repo_urls_module_provider, 'new', allow_custom_git_url_setting):

            self.perform_admin_authentication(password='unittest-password')

            ModuleProvider(
                Module(
                    Namespace('moduledetails'),
                    'fullypopulated'),
                'testprovider'
            ).update_attributes(
                git_provider_id=None,
                git_path='testoriginal',
                repo_base_url_template="https://sometestbaseurl.com",
                repo_clone_url_template="ssh://sometestcloneurl.com",
                repo_browse_url_template="https://sometestbaseurl.com/{tag}/{path}",
            )

            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Click on settings tab
            tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
            tab.click()

            # Check git provider dropdown
            git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')
            git_provider_select_option_elements = git_provider_select_element.find_elements(By.TAG_NAME, 'option')

            expected_git_providers = [
                'testgitprovider', 'repo_url_tests', 'repo_url_tests_uri_encoded',
                'no_browse_url', 'with_git_path_template'
            ]
            if allow_custom_git_url_setting:
                expected_git_providers.insert(0, 'Custom')

            # Ensure the list of providers in select match the expected
            assert expected_git_providers == [element.text for element in git_provider_select_option_elements]

            for option in git_provider_select_option_elements:
                # Check option name matches expected
                expected_name = expected_git_providers.pop(0)
                assert option.text == expected_name

                # Ensure git provider pk is used for value of option, or
                # custom option has an empty value
                expected_value = ''
                if expected_name != 'Custom':
                    git_provider = GitProvider.get_by_name(expected_name)
                    expected_value = str(git_provider.pk)
                assert option.get_attribute('value') == expected_value

            git_provider_select = Select(git_provider_select_element)

            if allow_custom_git_url_setting:
                # Ensure the currently selected item is custom
                assert git_provider_select_element.get_attribute('value') == ''

            # Ensure git path is empty
            assert self.selenium_instance.find_element(By.ID, "settings-git-path").get_attribute('value') == 'testoriginal'

            # Ensure all custom URL inputs are visible, if custom git provider is allowed,
            # and hidden, if not
            self.assert_custom_url_input_visibility(allow_custom_git_url_setting)

            # Select a different git provider and save
            git_provider_select.select_by_visible_text('with_git_path_template')

            # Ensure custom URL elements have been hidden
            self.assert_custom_url_input_visibility(False)

            # Ensure git path has been set by git provider
            assert self.selenium_instance.find_element(By.ID, "settings-git-path").get_attribute('value') == '/modules/{module}'

            try:
                # Press Update button
                self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()

                self.assert_equals(lambda: self.wait_for_element(By.ID, 'settings-status-success').text, 'Settings Updated')

                module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
                # Ensure git provider has been set
                assert module_provider._get_db_row()['git_provider_id'] == 5

                # Ensure the custom URLs has been cleared
                assert module_provider._get_db_row()['repo_base_url_template'] == None
                assert module_provider._get_db_row()['repo_clone_url_template'] == None
                assert module_provider._get_db_row()['repo_browse_url_template'] == None
                assert module_provider._get_db_row()['git_path'] == "/modules/{module}"

                # Reload page, assert the new git provider has been set
                self.selenium_instance.refresh()
                git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')
                self.assert_equals(lambda: git_provider_select_element.get_attribute('value'), '5')

            finally:
                # Reset git provider for module
                ModuleProvider(
                    Module(
                        Namespace('moduledetails'),
                        'fullypopulated'),
                    'testprovider'
                ).update_attributes(
                    git_provider_id=None,
                    git_path=None,
                    repo_base_url_template=None,
                    repo_clone_url_template=None,
                    repo_browse_url_template=None,
                )

    def test_custom_git_provider_custom_urls(self):
        """Ensure setting git provider to custom and setting URLs works."""

        with self.update_mock(self._config_allow_custom_repo_urls_module_provider, 'new', True):

            # Set git provider ID to custom
            ModuleProvider(
                Module(
                    Namespace('moduledetails'),
                    'fullypopulated'),
                'testprovider'
            ).update_attributes(
                git_provider_id=1,
                git_path='sometest',
                repo_base_url_template=None,
                repo_clone_url_template=None,
                repo_browse_url_template=None,
            )

            self.perform_admin_authentication(password='unittest-password')

            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

            # Click on settings tab
            tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
            tab.click()

            # Check git provider dropdown
            git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')
            git_provider_select = Select(git_provider_select_element)

            assert git_provider_select.first_selected_option.text == 'testgitprovider'

            # Ensure all custom URL inputs are not visible
            self.assert_custom_url_input_visibility(False)

            # Ensure git path has been correctly populated
            assert self.selenium_instance.find_element(By.ID, "settings-git-path").get_attribute('value') == 'sometest'

            # Select a different git provider and save
            git_provider_select.select_by_visible_text('Custom')

            # Ensure custom URL elements have been hidden
            self.assert_custom_url_input_visibility(True)

            # Set custom URLs
            base_url_input = self.selenium_instance.find_element(By.ID, 'settings-base-url-template')
            base_url_input.clear()
            base_url_input.send_keys("https://base-example.com/somenamespace/module")
            clone_url_input = self.selenium_instance.find_element(By.ID, 'settings-clone-url-template')
            clone_url_input.clear()
            clone_url_input.send_keys("ssh://git@clone-example.com/somenamespace/module.git")
            browse_url_input = self.selenium_instance.find_element(By.ID, 'settings-browse-url-template')
            browse_url_input.clear()
            browse_url_input.send_keys("https://browse-example.com/somenamespace/module/{tag}/{path}")

            # Ensure git path has not been changed
            assert self.selenium_instance.find_element(By.ID, "settings-git-path").get_attribute('value') == 'sometest'

            try:
                # Press Update button
                self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()

                self.assert_equals(lambda: self.wait_for_element(By.ID, 'settings-status-success').text, 'Settings Updated')

                module_provider = ModuleProvider(Module(Namespace('moduledetails'), 'fullypopulated'), 'testprovider')
                assert module_provider._get_db_row()['git_provider_id'] == None
                assert module_provider._get_db_row()['repo_base_url_template'] == "https://base-example.com/somenamespace/module"
                assert module_provider._get_db_row()['repo_clone_url_template'] == "ssh://git@clone-example.com/somenamespace/module.git"
                assert module_provider._get_db_row()['repo_browse_url_template'] == "https://browse-example.com/somenamespace/module/{tag}/{path}"
                assert module_provider._get_db_row()['git_path'] == "sometest"

                # Reload page, assert the new git provider has been set
                self.selenium_instance.refresh()
                git_provider_select_element = self.selenium_instance.find_element(By.ID, 'settings-git-provider')
                self.assert_equals(lambda: git_provider_select_element.get_attribute('value'), '')

            finally:
                # Reset git provider for module
                ModuleProvider(
                    Module(
                        Namespace('moduledetails'),
                        'fullypopulated'),
                    'testprovider'
                ).update_attributes(
                    git_provider_id=None,
                    repo_base_url_template=None,
                    repo_clone_url_template=None,
                    repo_browse_url_template=None,
                    git_path=None,
                )

    def test_updating_settings_after_logging_out(self):
        """Test accessing settings tab, logging out and attempting to save the changes in the settings tab."""
        self.perform_admin_authentication(password='unittest-password')

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Click on settings tab
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Delete all sessions from the database
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.session.delete())

        # Click update button
        self.selenium_instance.find_element(By.ID, 'module-provider-settings-update').click()
        error = self.wait_for_element(By.ID, 'settings-status-error')
        assert error.text == ('You must be logged in to perform this action.\n'
                              'If you were previously logged in, please re-authentication and try again.')

    @pytest.mark.parametrize('site_admin, group_permission, should_have_access', [
        # Without site admin access or group permission
        (False, None, False),
        # With site admin access
        (True, None, True),
        # With group modify
        (False, UserGroupNamespacePermissionType.MODIFY, True),
        # With group full
        (False, UserGroupNamespacePermissionType.FULL, True),
    ])
    def test_settings_tab_display_with_group_access(self, site_admin, group_permission, should_have_access):
        """Test whether settings tab is available with various permission types for SSO users."""
        # Enable access controls
        with self.update_mock(self._config_enable_access_controls, 'new', True):

            # Create test user group for authentication
            self.selenium_instance.delete_all_cookies()

            with self._patch_audit_event_creation():
                user_group = UserGroup.create(name='selenium-test-user-group', site_admin=site_admin)

            namespace = Namespace.get(name='moduledetails')

            try:
                # Add group permission, if it exists
                user_group_permission = None
                if group_permission:
                    with self._patch_audit_event_creation():
                        user_group_permission = UserGroupNamespacePermission.create(
                            user_group=user_group,
                            namespace=namespace,
                            permission_type=group_permission
                        )

                with self.log_in_with_openid_connect(user_groups=[user_group.name]):
                    # Access module provider page
                    self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

                    # Wait for README tab
                    self.wait_for_element(By.ID, 'module-tab-link-readme')

                    # Check if settings tab is available
                    settings_tab_link = self.selenium_instance.find_element(By.ID, 'module-tab-link-settings')
                    assert settings_tab_link.is_displayed() is should_have_access
            finally:
                with self._patch_audit_event_creation():
                    if user_group_permission:
                        user_group_permission.delete()

                    # Clear up test user group
                    user_group.delete()

    def test_deleting_module_version_after_logging_out(self):
        """Test accessing settings tab, logging out and attempting to delete module version."""
        self.perform_admin_authentication(password='unittest-password')

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Click on settings tab
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Click delete version button
        delete_button = self.selenium_instance.find_element(By.ID, 'module-version-delete-button')
        delete_button.click()

        # Ensure version number into verify input
        self.selenium_instance.find_element(By.ID, 'confirm-delete-module-version').send_keys('1.5.0')

        # Delete all sessions from the database
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.session.delete())

        # Click delete button again
        delete_button.click()
        error = self.wait_for_element(By.ID, 'settings-status-error')
        assert error.text == ('You must be logged in to perform this action.\n'
                              'If you were previously logged in, please re-authentication and try again.')

    def test_deleting_module_provider_after_logging_out(self):
        """Test accessing settings tab, logging out and attempting to delete module provider."""
        self.perform_admin_authentication(password='unittest-password')

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        # Click on settings tab
        tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
        tab.click()

        # Click delete version button
        delete_button = self.selenium_instance.find_element(By.ID, 'module-provider-delete-button')
        delete_button.click()

        # Ensure version number into verify input
        self.selenium_instance.find_element(By.ID, 'confirm-delete-module-provider').send_keys('moduledetails/fullypopulated/testprovider')

        # Delete all sessions from the database
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.session.delete())

        # Click delete button again
        delete_button.click()
        error = self.wait_for_element(By.ID, 'settings-status-error')
        assert error.text == ('You must be logged in to perform this action.\n'
                              'If you were previously logged in, please re-authentication and try again.')

    @pytest.mark.parametrize('publish_api_keys, is_authenticated, should_show_settings_note', [
        # When no API keys are set, the info should be shown
        ([], False, True),
        ([], True, True),

        # With API keys, the note should be shown if the user is authenticated
        (['apikey'], False, False),
        (['apikey'], True, True),
    ])
    def test_unpublished_settings_note(self, publish_api_keys, is_authenticated, should_show_settings_note):
        """Test whether information about using settings tab for publishing is displayed"""
        self.delete_cookies_and_local_storage()

        with self.update_mock(self._config_publish_api_keys_mock, 'new', publish_api_keys):
            if is_authenticated:
                self.perform_admin_authentication('unittest-password')

            self.selenium_instance.get(self.get_url('/modules/unpublished-beta-version-module-providers/onlyunpublished/testprovider/1.0.0'))
            assert self.wait_for_element(By.ID, 'unpublished-warning').text == (
                'WARNING: This version of the module is not published.\n'
                'It cannot be used in Terraform until it is published.' +
                ("\nSee the 'settings' tab to publish the module." if should_show_settings_note else '')
            )


    def test_unpublished_only_module_provider(self):
        """Test module provider page for a module provider that only has an unpublished version."""
        with self.update_mock(self._config_publish_api_keys_mock, 'new', ['akey']):
            self.delete_cookies_and_local_storage()

            self.selenium_instance.get(self.get_url('/modules/unpublished-beta-version-module-providers/onlyunpublished/testprovider'))

            # Ensure warning about no available version
            no_versions_div = self.wait_for_element(By.ID, 'no-version-available')
            assert no_versions_div.text == 'There are no versions of this module'
            assert no_versions_div.is_displayed() == True

            # Load version page
            self.selenium_instance.get(self.get_url('/modules/unpublished-beta-version-module-providers/onlyunpublished/testprovider/1.0.0'))

            # Check description
            assert self.wait_for_element(By.ID, 'module-description').text == 'Test description'

            # Ensure warning exists for not published
            assert self.wait_for_element(By.ID, 'unpublished-warning').text == (
                'WARNING: This version of the module is not published.\n'
                'It cannot be used in Terraform until it is published.'
            )

            # Ensure no versions available is not displayed
            no_versions_div = self.wait_for_element(By.ID, 'no-version-available', ensure_displayed=False)
            assert no_versions_div.is_displayed() == False

    def test_beta_only_module_provider(self):
        """Test module provider page for a module provider that only has a beta version."""
        self.selenium_instance.get(self.get_url('/modules/unpublished-beta-version-module-providers/onlybeta/testprovider'))

        # Ensure warning about no available version
        no_versions_div = self.wait_for_element(By.ID, 'no-version-available')
        assert no_versions_div.text == 'There are no versions of this module'
        assert no_versions_div.is_displayed() == True

        # Load version page
        self.selenium_instance.get(self.get_url('/modules/unpublished-beta-version-module-providers/onlybeta/testprovider/2.2.4-beta'))

        # Check description
        assert self.wait_for_element(By.ID, 'module-description').text == 'Test description'

        # Ensure warning exists for not published
        assert self.wait_for_element(By.ID, 'beta-warning').text == (
            'WARNING: This is a beta module version.\n'
            'To use this version in Terraform, it must '
            'be specifically pinned.\n'
            'For an example, see the \'Usage\' section.'
        )

        # Ensure no versions available is not displayed
        no_versions_div = self.wait_for_element(By.ID, 'no-version-available', ensure_displayed=False)
        assert no_versions_div.is_displayed() == False

    def test_viewing_non_latest_version(self):
        """Test viewing non-latest version of module."""
        self.selenium_instance.get(self.get_url('/modules/testnamespace/wrongversionorder/testprovider'))
        version_dropdown = self.wait_for_element(By.ID, 'version-select')
        version_select = Select(version_dropdown)
        assert version_select.first_selected_option.text == '10.23.0 (latest)'

        non_latest_warning = self.wait_for_element(By.ID, 'non-latest-version-warning', ensure_displayed=False)
        assert non_latest_warning.is_displayed() == False

        # Select old version
        version_select.select_by_visible_text('2.1.0')

        # Wait for new user to be redirected
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespace/wrongversionorder/testprovider/2.1.0'))

        # Ensure warning is shown
        non_latest_warning = self.wait_for_element(By.ID, 'non-latest-version-warning')
        assert non_latest_warning.text == ('WARNING: This is an outdated version of the module.\n'
                                           'If you wish to view the latest version of the module,\n'
                                           'use the version drop-down above.')

    @pytest.mark.parametrize('url', [
        '/modules/javascriptinjection/modulename/testprovider',
        '/modules/javascriptinjection/modulename/testprovider/1.5.0',
        '/modules/javascriptinjection/modulename/testprovider/1.5.0/submodule/modules/example-submodule1',
        '/modules/javascriptinjection/modulename/testprovider/1.5.0/example/examples/test-example'
    ])
    @pytest.mark.parametrize('local_storage', [
        {"input-output-view": "expanded"},
        {"input-output-view": "table"},
    ])
    def test_injected_html(self, url, local_storage):
        """Check for any injected HTML from module."""
        self.selenium_instance.get(self.get_url("/"))
        for k, v in local_storage.items():
            self.selenium_instance.execute_script("window.localStorage.setItem(arguments[0], arguments[1]);", k, v)

        self.selenium_instance.get(self.get_url(url))

        # Wait for tabs to be displayed
        self.wait_for_element(By.ID, 'module-tab-link-readme')

        for injected_element in [
                'injectedDescription',
                'injectedOwner',
                'injectedReadme',
                'injectedVariableTemplateName',
                'injectedVariableTemplateType',
                'injectedVariableAdditionalHelp',
                'injectedTerraformDocsInputName',
                'injectedTerraformDocsInputType',
                'injectedTerraformDocsInputDescription',
                'injectedTerraformDocsInputDefault',
                'injectedTerraformDocsOutputName',
                'injectedTerraformDocsOutputDescription',
                'injectedTerraformProviderName',
                'injectedTerraformDocsProviderAlias',
                'injectedTerraformDocsProviderVersion',
                'injectedTerraformDocsResourceType',
                'injectedTerraformDocsResourceName',
                'injectedTerraformDocsResourceProvider',
                'injectedTerraformDocsResourceSource',
                'injectedTerraformDocsResourceMode',
                'injectedTerraformDocsResourceVersion',
                'injectedTerraformDocsResourceDescription',
                'injectedExampleFileContent',
                'injectedExampleReadme',
                'injectedSubemoduleFileContent',
                'injectedAdditionalTabsPlainText',
                'injectedAdditionalTabsMarkDown']:

            with pytest.raises(selenium.common.exceptions.NoSuchElementException):
                print(f"Checking for {injected_element}")
                self.selenium_instance.find_element(By.ID, injected_element)

    @pytest.mark.parametrize('enable_beta,enable_unpublished,expected_versions', [
        (False, False, ['1.5.0 (latest)', '1.2.0']),
        (True, False, ['1.7.0-beta (beta)', '1.6.1-beta (beta)', '1.5.0 (latest)', '1.2.0']),
        (False, True, ['1.6.0 (unpublished)', '1.5.0 (latest)', '1.2.0']),
        (True, True, ['1.7.0-beta (beta)', '1.6.1-beta (beta)', '1.6.0 (unpublished)', '1.5.0 (latest)', '1.2.0', '1.0.0-beta (beta) (unpublished)']),
    ])
    def test_user_preferences(self, enable_beta, enable_unpublished, expected_versions):
        """Test user preferences"""

        def get_select_names(element_id):
            select = Select(self.wait_for_element(By.ID, element_id))
            options = select.options
            return [option.text for option in options]

        # Clear local storage
        self.delete_cookies_and_local_storage()

        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))
    
        # Wait for version select
        assert get_select_names('version-select') == ['1.5.0 (latest)', '1.2.0']

        preferences_modal = self.open_user_preferences_modal()

        # Ensure checkboxes are unchecked
        beta_checkbox = preferences_modal.find_element(By.XPATH, "//label[contains(text(),\"Show 'beta' versions\")]//input")
        unpublished_checkbox = preferences_modal.find_element(By.XPATH, "//label[contains(text(),\"Show 'unpublished' versions\")]//input")

        assert beta_checkbox.is_selected() == False
        assert unpublished_checkbox.is_selected() == False

        # Enable preferences
        if enable_beta:
            beta_checkbox.click()
        if enable_unpublished:
            unpublished_checkbox.click()


        # Click save
        self.save_user_preferences_modal()

        # Reload page and ensure beta versions are displayed
        self.selenium_instance.refresh()
        assert get_select_names('version-select') == expected_versions

        # Reset options and ensure versions go back to original
        preferences_modal = self.open_user_preferences_modal()

        # Ensure checkboxes are the same checked state
        beta_checkbox = preferences_modal.find_element(By.XPATH, "//label[contains(text(),\"Show 'beta' versions\")]//input")
        unpublished_checkbox = preferences_modal.find_element(By.XPATH, "//label[contains(text(),\"Show 'unpublished' versions\")]//input")

        assert beta_checkbox.is_selected() == enable_beta
        assert unpublished_checkbox.is_selected() == enable_unpublished

        if enable_beta:
            beta_checkbox.click()
        if enable_unpublished:
            unpublished_checkbox.click()
        
        # Save changes again
        self.save_user_preferences_modal()

        # Reload and check versions go back to no longer including any additional versions
        self.selenium_instance.refresh()
        assert get_select_names('version-select') == ['1.5.0 (latest)', '1.2.0']

        # Clear local storage
        self.selenium_instance.execute_script("window.localStorage.clear();")

    def test_additional_tabs(self):
        """Test additional tabs"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0'))

        self.wait_for_element(By.ID, 'module-tab-link-analytics')

        # Ensure tab for non-existent file isn't displayed
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            self.selenium_instance.find_element(By.ID, 'module-tab-link-custom-doesnotexist')
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            self.selenium_instance.find_element(By.ID, 'module-tab-custom-doesnotexist')

        # Ensure tabs exist
        license_tab_link = self.wait_for_element(By.ID, 'module-tab-link-custom-License')
        assert license_tab_link.text == "License"
        changelog_tab_link = self.wait_for_element(By.ID, 'module-tab-link-custom-Changelog')
        assert changelog_tab_link.text == "Changelog"

        # Ensure tab content is not shown
        assert self.selenium_instance.find_element(By.ID, 'module-tab-custom-License').is_displayed() == False
        assert self.selenium_instance.find_element(By.ID, 'module-tab-custom-Changelog').is_displayed() == False

        # Click license tab and check it's displayed and content is correct
        license_tab_link.click()
        license_content = self.wait_for_element(By.ID, 'module-tab-custom-License')
        # Check license content has been put into pre tags
        assert license_content.get_attribute('innerHTML') == """
<pre>This is a license file
All rights are not reserved for this example file content
This license &gt; tests
various &lt; characters that could be escaped.</pre>
        """.strip()

        # Click license tab and check it's displayed and content is correct
        changelog_tab_link.click()
        changelog_content = self.wait_for_element(By.ID, 'module-tab-custom-Changelog')
        # Check changelog has been converted from markdown to HTML
        assert changelog_content.get_attribute('innerHTML') == """
<h1 id="terrareg-anchor-CHANGELOGmd-changelog" class="subtitle is-3">Changelog</h1>
<h2 id="terrareg-anchor-CHANGELOGmd-100" class="subtitle is-4">1.0.0</h2>
<ul>
<li>This is an initial release</li>
</ul>
<p>This tests &gt; 2 &lt; 3 escapable characters</p>
        """.strip()

    def test_security_issues_tab(self):
        """Check security issues tab"""
        self.selenium_instance.get(self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.0.0'))

        # Security issues tab button is shown
        tab_link = self.wait_for_element(By.ID, 'module-tab-link-security-issues')
        assert tab_link.is_displayed() == True
        assert tab_link.text == 'Security Issues\nTerrareg Exclusive'

        # Ensure tab is not shown
        tab_content = self.wait_for_element(By.ID, "module-tab-security-issues", ensure_displayed=False)
        assert tab_content.is_displayed() == False

        # Click tab link
        tab_link.click()

        assert tab_content.is_displayed() == True

        # Check rows for security issues
        expected_rows = [
            ['', 'Severity', '', 'Description', '', '', '', '', '', '', '', '', ''],
            ['ignored.tf'],
            ['', 'CRITICAL', '', 'Critical code has an issue', '', '', '', '', '', '', '', '', ''],
            ['different.tf'],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['main.tf'],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['', 'LOW', '', 'Secrets Manager should use customer managed keys', '', '', '', '', '', '', '', '', ''],
            ['ignored.tf'],
            ['', 'MEDIUM', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
        ]
        for row in tab_content.find_elements(By.TAG_NAME, "tr"):
            column_data = [td.text for td in row.find_elements(By.TAG_NAME, "th") + row.find_elements(By.TAG_NAME, "td")]
            assert column_data == expected_rows.pop(0)
        assert len(expected_rows) == 0

        # Select third row (first issue) and expand
        tab_content.find_elements(By.TAG_NAME, "tr")[2].find_elements(By.TAG_NAME, "td")[0].click()

        # Ensure row is expanded, showing additional information
        expected_rows = [
            ['', 'Severity', '', 'Description', '', '', '', '', '', '', '', '', ''],
            ['ignored.tf'],
            ['', 'CRITICAL', '', 'Critical code has an issue', '', '', '', '', '', '', '', '', ''],
            [(
                'File ignored.tf\n'
                'ID DDG-ANC-007\n'
                'Provider bad\n'
                'Service code\n'
                'Resource some_data_resource.this\n'
                'Starting Line 6\n'
                'Ending Line 1\n'
                'Impact This is critical\n'
                'Resolution Fix critical issue\n'
                'Resources\n'
                '- https://example.com/issuehere\n'
                '- https://example.com/docshere'
            )],
            ['different.tf'],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['main.tf'],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['', 'HIGH', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
            ['', 'LOW', '', 'Secrets Manager should use customer managed keys', '', '', '', '', '', '', '', '', ''],
            ['ignored.tf'],
            ['', 'MEDIUM', '', 'Dodgy code should be removed', '', '', '', '', '', '', '', '', ''],
        ]
        for row in tab_content.find_elements(By.TAG_NAME, "tr"):
            column_data = [td.text for td in row.find_elements(By.TAG_NAME, "th") + row.find_elements(By.TAG_NAME, "td")]
            assert column_data == expected_rows.pop(0)
        assert len(expected_rows) == 0

        # Go to 1.1.0 version, with no security issues
        Select(self.selenium_instance.find_element(By.ID, 'version-select')).select_by_visible_text('1.1.0')

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.1.0'))

        # Wait for inputs tab, to indicate page has loaded
        self.wait_for_element(By.ID, 'module-tab-link-integrations')

        # Security issues tab button is not shown
        tab_link = self.wait_for_element(By.ID, 'module-tab-link-security-issues', ensure_displayed=False)
        assert tab_link.is_displayed() == False

        # Go to example
        Select(self.selenium_instance.find_element(By.ID, 'example-select')).select_by_visible_text('examples/withsecissue')
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.1.0/example/examples/withsecissue'))

        tab_link = self.wait_for_element(By.ID, 'module-tab-link-security-issues')
        assert tab_link.is_displayed() == True
        assert tab_link.text == 'Security Issues\nTerrareg Exclusive'

        # Click tab link
        tab_link.click()

        tab_content = self.wait_for_element(By.ID, "module-tab-security-issues")

        assert tab_content.is_displayed() == True

        # Check rows for security issues.
        # All data contains invalid data.
        expected_rows = [
            ['', 'Severity', '', 'Description', '', '', '', '', '', '', '', '', ''],
            ['second.tf'],
            ['', 'HIGH', '', 'This type of second issue is High', '', '', '', '', '', '', '', '', ''],
            ['first.tf'],
            ['', 'LOW', '', 'This type of first issue is Low', '', '', '', '', '', '', '', '', ''],
            ['third.tf'],
            ['', 'MEDIUM', '', 'This type of third issue is Medium', '', '', '', '', '', '', '', '', ''],
        ]
        for row in tab_content.find_elements(By.TAG_NAME, "tr"):
            column_data = [td.text for td in row.find_elements(By.TAG_NAME, "th") + row.find_elements(By.TAG_NAME, "td")]
            assert column_data == expected_rows.pop(0)
        assert len(expected_rows) == 0

        # Go back to parent
        self.selenium_instance.find_element(By.ID, 'submodule-back-to-parent').click()
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.1.0'))

        # Go to 1.2.0 version, with no security issues
        Select(self.selenium_instance.find_element(By.ID, 'version-select')).select_by_visible_text('1.2.0 (latest)')
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.2.0'))

        # Security issues tab button is not shown
        tab_link = self.wait_for_element(By.ID, 'module-tab-link-security-issues', ensure_displayed=False)
        assert tab_link.is_displayed() == False

        # Go to submodule
        Select(self.selenium_instance.find_element(By.ID, 'submodule-select')).select_by_visible_text('modules/withanotherissue')
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/moduledetails/withsecurityissues/testprovider/1.2.0/submodule/modules/withanotherissue'))

        # Ensure 3 security issues are shown
        tab_link = self.wait_for_element(By.ID, 'module-tab-link-security-issues')
        assert tab_link.is_displayed() == True
        assert tab_link.text == 'Security Issues\nTerrareg Exclusive'

        # Click tab link
        tab_link.click()

        tab_content = self.wait_for_element(By.ID, "module-tab-security-issues")
        assert tab_content.is_displayed() == True

        # Check rows for security issues
        expected_rows = [
            ['', 'Severity', '', 'Description', '', '', '', '', '', '', '', '', ''],
            ['first.tf'],
            ['', 'MEDIUM', '', 'This type of first issue is Medium', '', '', '', '', '', '', '', '', ''],
        ]
        for row in tab_content.find_elements(By.TAG_NAME, "tr"):
            column_data = [td.text for td in row.find_elements(By.TAG_NAME, "th") + row.find_elements(By.TAG_NAME, "td")]
            assert column_data == expected_rows.pop(0)
        assert len(expected_rows) == 0

    def _compare_canvas(self, compare_filename):
        """Compare current canvas data for graph to expected image"""
        png_url = self.selenium_instance.execute_script("return document.getElementById('cy').getElementsByTagName('canvas')[2].toDataURL('image/png');").replace("data:image/png;base64,", "")
        image_data = base64.decodebytes(png_url.encode("utf-8"))

        # Enable to regenerate expected images
        # sleep(5)
        # with open(compare_filename, "wb") as fh:
        #     fh.write(image_data)

        actual_image = Image.open(BytesIO(image_data), formats=["PNG"])
        actual_image = actual_image.crop(actual_image.getbbox())
        expected_image = Image.open(compare_filename)
        expected_image = expected_image.crop(expected_image.getbbox())

        return imagehash.phash(actual_image) == imagehash.phash(expected_image)

    @skipif_unless_ci(
        not os.environ.get("RUNNING_IN_DOCKER"),
        reason="Canvas image comparison does not work outside of docker"
    )
    @pytest.mark.parametrize("base_url,expected_url,base_filename,", [
        ("/modules/moduledetails/fullypopulated/testprovider/1.5.0",
         "/modules/moduledetails/fullypopulated/testprovider/1.5.0/graph",
         "moduledetails_fullypopulated_testprovider_1.5.0_root_module"),
        ("/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1",
         "/modules/moduledetails/fullypopulated/testprovider/1.5.0/graph/submodule/modules/example-submodule1",
         "moduledetails_fullypopulated_testprovider_1.5.0_submodule"),
        ("/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example",
         "/modules/moduledetails/fullypopulated/testprovider/1.5.0/graph/example/examples/test-example",
         "moduledetails_fullypopulated_testprovider_1.5.0_example")
    ])
    @pytest.mark.parametrize("full_resource_names,full_module_names", [
        (False, False),
        (False, True),
        (True, False),
        (True, True)
    ])
    def test_resource_graph(self, base_url, expected_url, base_filename, full_resource_names, full_module_names):
        """Test resource graph page"""

        self.selenium_instance.get(self.get_url(base_url))
        # Wait for resources tab to load
        resources_link = self.wait_for_element(By.ID, 'module-tab-link-resources')
        resources_link.click()

        # Ensure link to resource graph is displayed
        resource_graph_link = self.selenium_instance.find_element(By.ID, "resourceDependencyGraphLink")
        assert resource_graph_link.text == "View resource dependency graph"
        resource_graph_link.click()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(expected_url))

        if full_resource_names:
            self.selenium_instance.find_element(By.ID, "graphOptionsShowFullResourceNames").click()
        if full_module_names:
            self.selenium_instance.find_element(By.ID, "graphOptionsShowFullModuleNames").click()

        file_name = os.path.join(
            os.path.dirname(inspect.getfile(TestModuleProvider)),
            "test_graph_canvas_images",
            f"{base_filename}{'_full_resources' if full_resource_names else ''}{'_full_modules' if full_module_names else ''}.png"
        )

        # Attempt check canvas data
        self.assert_equals(lambda: self._compare_canvas(file_name), True, sleep_period=1)

    @pytest.mark.parametrize('url,expected_module_name,expected_module_path,expected_comment,expected_module_version_constraint', [
        # Base module
        ('/modules/moduledetails/fullypopulated/testprovider',
         'fullypopulated',
         'moduledetails/fullypopulated/testprovider',
         '',
         '>= 1.5.0, < 2.0.0, unittest'),
        # Explicit version
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0',
         'fullypopulated',
         'moduledetails/fullypopulated/testprovider',
         '',
         '>= 1.5.0, < 2.0.0, unittest'),
        # Submodule
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
         'fullypopulated',
         'moduledetails/fullypopulated/testprovider//modules/example-submodule1',
         '',
         '>= 1.5.0, < 2.0.0, unittest'),
        # Non-latest version
        ('/modules/moduledetails/fullypopulated/testprovider/1.2.0',
         'fullypopulated',
         'moduledetails/fullypopulated/testprovider',
         '\n  # This version of the module is not the latest version.\n  # To use this specific version, it must be pinned in Terraform',
         '1.2.0'),
        # Beta version
        ('/modules/moduledetails/fullypopulated/testprovider/1.7.0-beta',
         'fullypopulated',
         'moduledetails/fullypopulated/testprovider',
         '\n  # This version of the module is a beta version.\n  # To use this version, it must be pinned in Terraform',
         '1.7.0-beta')

    ])
    def test_example_usage(self, url, expected_module_name, expected_module_path, expected_comment, expected_module_version_constraint):
        """Check example usage panel"""
        self.selenium_instance.get(self.get_url(url))

        # Wait for inputs tab to be ready
        self.wait_for_element(By.ID, 'module-tab-link-inputs')

        assert self.selenium_instance.find_element(By.ID, "usage-example-terraform").text == f"""
module "{expected_module_name}" {{
  source  = "localhost/my-tf-application__{expected_module_path}"{expected_comment}
  version = "{expected_module_version_constraint}"

  # Provide variables here
}}
""".strip()

    @pytest.mark.parametrize('url,expected_terraform_version', [
        # Base module
        ('/modules/moduledetails/fullypopulated/testprovider',
         '>= 1.0, < 2.0.0'),
        # Explicit version
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0',
         '>= 1.0, < 2.0.0'),
        # Submodule
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1',
         '>= 2.0.0'),
        # Non-latest version
        ('/modules/moduledetails/fullypopulated/testprovider/1.2.0',
         '>= 2.1.1, < 2.5.4'),
        # Beta version
        ('/modules/moduledetails/fullypopulated/testprovider/1.7.0-beta',
         '>= 5.12, < 21.0.0'),
        # Module without TF constraint
        ('/modules/moduledetails/withsecurityissues/testprovider/1.2.0',
         None)
    ])
    def test_example_usage_terraform_version(self, url, expected_terraform_version):
        """Check example usage panel"""
        self.selenium_instance.get(self.get_url(url))

        # Wait for inputs tab to be ready
        self.wait_for_element(By.ID, 'module-tab-link-inputs')

        version_text = self.selenium_instance.find_element(By.ID, "supported-terraform-versions")
        assert version_text.is_displayed() == (expected_terraform_version is not None)
        if expected_terraform_version is not None:
            assert version_text.text == f"Supported Terraform versions: {expected_terraform_version}"

    @pytest.mark.parametrize("url", [
        # Example
        ("/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example"),
        # Unpublished version
        ("/modules/moduledetails/fullypopulated/testprovider/1.6.0")
    ])
    def test_example_usage_ensure_not_shown(self, url):
        """Ensure example usage section is not displayed in example submodule"""
        self.selenium_instance.get(self.get_url(url))

        # Wait for inputs tab to be ready
        self.wait_for_element(By.ID, 'module-tab-link-inputs')

        version_text = self.selenium_instance.find_element(By.ID, "usage-example-container")
        assert version_text.is_displayed() == False

    @pytest.mark.parametrize('url,expect_warning,wait_for_tab', [
        # No versions should not produce the warning
        ('/modules/moduledetails/noversion/testprovider', False, 'integrations'),
        # Version with outdated extraction data
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0', True, 'resources'),
        # Warning should not be displayed on examples/submodules
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example', False, 'resources'),
        ('/modules/moduledetails/fullypopulated/testprovider/1.5.0/submodule/modules/example-submodule1', False, 'resources'),
        # Version with up-to-date extraction data
        ('/modules/moduledetails/fullypopulated/testprovider/1.2.0', False, 'resources')
    ])
    def test_outdated_extraction_data_warning(self, url, expect_warning, wait_for_tab):
        """Check outdated extraction data warning."""
        self.selenium_instance.get(self.get_url(url))

        # Wait for inputs tab to be ready
        self.wait_for_element(By.ID, f'module-tab-link-{wait_for_tab}')

        # Get warning element
        warning_element = self.selenium_instance.find_element(By.ID, 'outdated-extraction-warning')
        assert warning_element.is_displayed() == expect_warning
        if expect_warning:
            assert warning_element.text == """
This module version was extracted using a previous version of Terrareg meaning that some data maybe not be available.
Consider re-indexing this module version to enable all features.
""".strip()

    @pytest.mark.parametrize('provider', [
        'aws',
        'gcp',
        'null',
        'consul',
        'nomad',
        'vagrant',
        'vault',
    ])
    def test_provider_logos(self, provider):
        """Test provider logos"""
        self.selenium_instance.get(self.get_url(f"/modules/real_providers/test-module/{provider}"))

        # Get image object
        image = self.selenium_instance.find_element(By.ID, "provider-logo-img")
        self.assert_equals(lambda: image.get_attribute("src"), self.get_url(ProviderLogo(provider).source))

        # Ensure image exists
        res = requests.get(self.get_url(ProviderLogo(provider).source))
        assert res.status_code == 200

    def test_analytics_disabled(self):
        """Test module provider page with analytics disabled."""
        with self.update_mock(self._config_disable_analytics, 'new', True):
            self.selenium_instance.get(self.get_url("/modules/moduledetails/fullypopulated/testprovider/1.5.0"))

            # Wait for README tab link
            self.wait_for_element(By.ID, "module-tab-link-readme")

            # Test example in README does not contain analytics token
            assert self.selenium_instance.find_element(By.ID, "module-tab-readme").text == """
This is an example README!
Following this example module call:
module "test_example_call" {
  source  = "localhost/moduledetails/fullypopulated/testprovider"
  version = ">= 1.5.0, < 2.0.0, unittest"

  name = "example-name"
}
This should work with all versions > 5.2.0 and <= 6.0.0
module "text_ternal_call" {
  source  = "a-public/module"
  version = "> 5.2.0, <= 6.0.0"

  another = "example-external"
}
""".strip()

            # Ensure usage example does not contain analytics token
            usage_example = self.wait_for_element(By.ID, "usage-example-container")
            usage_instructions = usage_example.find_element(By.CLASS_NAME, "content")
            assert usage_instructions.text == """
Supported Terraform versions: >= 1.0, < 2.0.0
To use this module:
Add the following example to your Terraform,
Add the required inputs - use the 'Usage Builder' tab for help and 'Inputs' tab for a full list.
""".strip()

            assert usage_example.find_element(By.ID, "usage-example-terraform").text == """
module "fullypopulated" {
  source  = "localhost/moduledetails/fullypopulated/testprovider"
  version = ">= 1.5.0, < 2.0.0, unittest"

  # Provide variables here
}
""".strip()

            # Ensure analytics tab is not shown
            analytics_tab_link = self.selenium_instance.find_element(By.ID, "module-tab-link-analytics")
            assert analytics_tab_link.is_displayed() == False

            # Check example file content
            self.selenium_instance.get(self.get_url('/modules/moduledetails/fullypopulated/testprovider/1.5.0/example/examples/test-example'))
            # Go to example file content
            self.wait_for_element(By.ID, "module-tab-link-example-files").click()
            self.assert_equals(
                lambda: self.selenium_instance.find_element(By.ID, "example-file-content").text,
                """
# Call root module
module "root" {
  source  = "localhost/moduledetails/fullypopulated/testprovider"
  version = ">= 1.5.0, < 2.0.0, unittest"
}
""".strip()
            )


    @pytest.mark.parametrize('terraform_version, expected_compatibility_result, expected_color', [
        ('1.5.2', 'Compatible', 'success'),
        ('0.11.31', 'Incompatible', 'danger'),
    ])
    def test_terraform_compatibility_result(self, terraform_version, expected_compatibility_result, expected_color):
        """Test terraform version compatibility result text"""
        self.delete_cookies_and_local_storage()

        self.selenium_instance.get(self.get_url("/modules/moduledetails/fullypopulated/testprovider/1.5.0"))

        # Wait for terraform constraint in usage example
        self.wait_for_element(By.ID, "supported-terraform-versions")

        # Ensure the compatibility text is not displayed
        assert self.selenium_instance.find_element(By.ID, "supported-terraform-compatible").is_displayed() == False

        # Update user preferences to set Terraform version
        preferences_modal = self.open_user_preferences_modal()
        terraform_constraint_input = preferences_modal.find_element(By.XPATH, "//label[contains(text(),\"Terraform Version for compatibility checks\")]//input")
        terraform_constraint_input.send_keys(terraform_version)
        self.save_user_preferences_modal()

        # Page will reload, check that the version constraint is shown
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, "supported-terraform-compatible").is_displayed(), True)

        # Check text for compatibility
        assert self.selenium_instance.find_element(By.ID, "supported-terraform-compatible").text == f"Terraform {terraform_version} compatibility:\n{expected_compatibility_result}"
        # Check color of label
        assert self.selenium_instance.find_element(By.ID, "supported-terraform-compatible-tag").get_attribute("class") == f"tag is-medium is-light is-{expected_color}"

    def test_delete_module_provider_redirect(self, mock_create_audit_event):
        """Test deletion of a module provider redirect"""
        with mock_create_audit_event:
            namespace = Namespace.get("moduledetails")
            module_provider = ModuleProvider.create(module=Module(namespace, "testredirect"), name="testprovider")
            module_provider = module_provider.update_name(namespace=namespace, module_name="secondredirect", provider_name="testprovider")
            module_provider = module_provider.update_name(namespace=namespace, module_name="newredirectname", provider_name="testprovider")
            version = ModuleVersion(module_provider, "1.0.0")
            version.prepare_module()

            # Add analytics
            terrareg.analytics.AnalyticsEngine.record_module_version_download(
                namespace_name="moduledetails",
                module_name="testredirect",
                provider_name="testprovider",
                module_version=version,
                analytics_token=None, terraform_version="1.0.0",
                user_agent=None, auth_token=None
            )

        try:
            self.delete_cookies_and_local_storage()
            self.perform_admin_authentication(password='unittest-password')

            self.selenium_instance.get(self.get_url("/modules/moduledetails/newredirectname/testprovider"))

            # Click on settings tab
            tab = self.wait_for_element(By.ID, 'module-tab-link-settings')
            tab.click()

            # Ensure redirect card is present
            redirect_card = self.selenium_instance.find_element(By.ID, "settingsRedirectCard")
            assert redirect_card.is_displayed() == True

            # Check rows of table
            table_body = redirect_card.find_element(By.ID, "settingsRedirectTable")
            expected_rows = [
                ["moduledetails", "testredirect", "testprovider", "Delete"],
                ["moduledetails", "secondredirect", "testprovider", "Delete"]
            ]
            first_redirect_row = None
            for row in table_body.find_elements(By.TAG_NAME, "tr"):
                found_row = [td.text for td in row.find_elements(By.TAG_NAME, "td")]
                assert found_row in expected_rows
                expected_rows.remove(found_row)

                if found_row[1] == "secondredirect":
                    first_redirect_row = row

            assert len(expected_rows) == 0

            # Click delete button against analytics
            delete_button = first_redirect_row.find_element(By.TAG_NAME, "button")
            assert delete_button.text == "Delete"
            delete_button.click()

            # Wait for page reload
            sleep(1)

            # Wait for page to reload and ensure redirect card is present
            redirect_card = self.wait_for_element(By.ID, "settingsRedirectCard")
            assert redirect_card.is_displayed() == True

            # Ensure original redirect has been removed
            table_body = redirect_card.find_element(By.ID, "settingsRedirectTable")
            expected_rows = [
                ["moduledetails", "testredirect", "testprovider", "Delete"]
            ]
            found_rows = []
            second_redirect_row = None
            for row in table_body.find_elements(By.TAG_NAME, "tr"):
                found_rows.append([td.text for td in row.find_elements(By.TAG_NAME, "td")])
                if found_rows[0][1] == "testredirect":
                    second_redirect_row = row

            assert expected_rows == found_rows

            # Ensure error is not shown
            assert self.selenium_instance.find_element(By.ID, "settings-redirect-error").is_displayed() == False

            # Attempt to remove second row
            delete_button = second_redirect_row.find_element(By.TAG_NAME, "button")
            assert delete_button.text == "Delete"
            delete_button.click()

            # Ensure error is shown
            error = self.selenium_instance.find_element(By.ID, "settings-redirect-error")
            assert error.is_displayed() == True
            assert error.text == (
                'Module provider redirect is in use, so cannot be deleted without forceful deletion\n'
                'Force Retry'
            )

            # Find force retry button and click
            force_retry_button = error.find_element(By.TAG_NAME, "button")
            assert force_retry_button.text == "Force Retry"
            force_retry_button.click()

            # Wait for page reload
            sleep(1)

            # Wait for reload and settings tab
            tab = self.wait_for_element(By.ID, 'module-tab-link-settings')

            # Ensure redirects card is not shown
            assert self.selenium_instance.find_element(By.ID, "settingsRedirectCard").is_displayed() == False

        finally:
            # Delete module provider
            with mock_create_audit_event:
                module_provider.delete()

