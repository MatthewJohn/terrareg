
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider
from .test_data import one_namespace_test_data

class TestNamespaceList(SeleniumTest):
    """Test namespace list page."""

    def test_namespace_list_page(self):
        """Test namespace list page."""
        self.selenium_instance.get(self.get_url('/modules'))

        assert self.selenium_instance.title == 'Namespaces - Terrareg'

        # Get content section
        content = self.wait_for_element(By.ID, 'namespace-list-content')

        # Check title
        assert content.find_element(By.TAG_NAME, 'h1').text == 'Namespaces'

        expected_namespaces = [
            ['javascriptinjection', 'javascriptinjection'],
            ['moduledetails', 'moduledetails'],
            ['modulesearch', 'modulesearch'],
            ['modulesearch-contributed', 'modulesearch-contributed'],
            ['modulesearch-trusted', 'modulesearch-trusted'],
            ['mostrecent', 'mostrecent'],
            ['real_providers', 'real_providers'],
            ['relevancysearch', 'relevancysearch'],
            ['searchbynamespace', 'searchbynamespace'],
            ['testnamespace', 'testnamespace'],
        ]

        # Check namespaces
        table_body = content.find_element(By.ID, 'namespaces-table-data')
        for namespace_tr in table_body.find_elements(By.TAG_NAME, 'tr'):
            expected_name, expected_id = expected_namespaces.pop(0)

            link = namespace_tr.find_element(By.TAG_NAME, 'a')
            assert link.text == expected_name
            assert link.get_attribute('href') == self.get_url(f'/modules/{expected_id}')

        # Ensure previous link is not active and next link is
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == True
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

        # Click next button
        self.selenium_instance.find_element(By.ID, 'nextButton').click()

        # Ensure new namespace lists are correct
        expected_namespaces = [
            ['trustednamespace', 'trustednamespace'],
            ['unpublished-beta-version-module-providers', 'unpublished-beta-version-module-providers'],
            ['version-constraint-test', 'version-constraint-test'],
        ]
        # Check namespaces
        table_body = content.find_element(By.ID, 'namespaces-table-data')
        for namespace_tr in table_body.find_elements(By.TAG_NAME, 'tr'):
            expected_name, expected_id = expected_namespaces.pop(0)

            link = namespace_tr.find_element(By.TAG_NAME, 'a')
            assert link.text == expected_name
            assert link.get_attribute('href') == self.get_url(f'/modules/{expected_id}')

        # Ensure prev button is enabled and next button is not
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == False
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == True

    def test_namespace_list_page_with_unpublished(self):
        """Test namespace list page with unpublished selected."""
        self.selenium_instance.get(self.get_url('/modules'))

        assert self.selenium_instance.title == 'Namespaces - Terrareg'

        # Get content section
        content = self.wait_for_element(By.ID, 'namespace-list-content')

        # Check title
        assert content.find_element(By.TAG_NAME, 'h1').text == 'Namespaces'

        # Check "show empty namespaces" checkbox
        show_unpublished_checkbox = self.selenium_instance.find_element(By.ID, "show-unpublished")
        assert show_unpublished_checkbox.is_selected() == False
        show_unpublished_checkbox.click()

        expected_namespaces = [
            ['contributed-providersearch', 'contributed-providersearch'],
            ['emptynamespace', 'emptynamespace'],
            ['initial-providers', 'initial-providers'],
            ['javascriptinjection', 'javascriptinjection'],
            ['moduledetails', 'moduledetails'],
            ['moduleextraction', 'moduleextraction'],
            ['modulesearch', 'modulesearch'],
            ['modulesearch-contributed', 'modulesearch-contributed'],
            ['modulesearch-trusted', 'modulesearch-trusted'],
            ['mostrecent', 'mostrecent'],
        ]

        # Check namespaces
        table_body = content.find_element(By.ID, 'namespaces-table-data')
        for namespace_tr in table_body.find_elements(By.TAG_NAME, 'tr'):
            expected_name, expected_id = expected_namespaces.pop(0)

            link = namespace_tr.find_element(By.TAG_NAME, 'a')
            assert link.text == expected_name
            assert link.get_attribute('href') == self.get_url(f'/modules/{expected_id}')

        # Ensure previous link is not active and next link is
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == True
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

        # Click next button
        self.selenium_instance.find_element(By.ID, 'nextButton').click()

        # Ensure new namespace lists are correct
        expected_namespaces = [
            ['mostrecentunpublished', 'mostrecentunpublished'],
            ['onlybeta', 'onlybeta'],
            ['onlyunpublished', 'onlyunpublished'],
            ['providersearch', 'providersearch'],
            ['providersearch-trusted', 'providersearch-trusted'],
            ['real_providers', 'real_providers'],
            ['relevancysearch', 'relevancysearch'],
            ['repo_url_tests', 'repo_url_tests'],
            ['scratchnamespace', 'scratchnamespace'],
            ['searchbynamespace', 'searchbynamespace'],
            ['testmodulecreation', 'testmodulecreation'],
        ]
        # Check namespaces
        table_body = content.find_element(By.ID, 'namespaces-table-data')
        for namespace_tr in table_body.find_elements(By.TAG_NAME, 'tr'):
            expected_name, expected_id = expected_namespaces.pop(0)

            link = namespace_tr.find_element(By.TAG_NAME, 'a')
            assert link.text == expected_name
            assert link.get_attribute('href') == self.get_url(f'/modules/{expected_id}')

        # Ensure prev button is enabled and next button is not
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == True
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == True


class TestNamespaceListSingleNamespace(SeleniumTest):
    """Test namespace list page with single namespace"""

    _TEST_DATA = one_namespace_test_data
    _USER_GROUP_DATA = None

    def test_namespace_list_page_redirect(self):
        """Test namespace list page with one namespace."""
        self.selenium_instance.get(self.get_url('/modules'))

        # Ensure page is redirected to namespace page
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespace'))


class TestNamespaceListNoNamespaces(SeleniumTest):
    """Test namespace list page with no namespaces"""

    _TEST_DATA = {}
    _USER_GROUP_DATA = None

    def test_namespace_list_page_warning(self):
        """Test namespace list page with no namespaces."""
        self.selenium_instance.get(self.get_url('/modules'))

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/initial-setup'))
