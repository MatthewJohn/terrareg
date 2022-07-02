
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

        # Get content section
        content = self.wait_for_element(By.ID, 'namespace-list-content')

        # Check title
        assert content.find_element(By.TAG_NAME, 'h1').text == 'Namespaces'

        expected_namespaces = [
            'moduledetails', 'moduleextraction', 'modulesearch', 'modulesearch-contributed',
            'modulesearch-trusted', 'mostrecent', 'mostrecentunpublished', 'onlybeta', 'onlyunpublished',
            'real_providers', 'repo_url_tests', 'searchbynamespace', 'testnamespace', 'trustednamespace'
        ]

        # Check namespaces
        table_body = content.find_element(By.ID, 'namespaces-table-data')
        for namespace_tr in table_body.find_elements(By.TAG_NAME, 'tr'):
            expected_namespace = expected_namespaces.pop(0)

            link = namespace_tr.find_element(By.TAG_NAME, 'a')
            assert link.text == expected_namespace
            assert link.get_attribute('href') == self.get_url(f'/modules/{expected_namespace}')


class TestNamespaceListSingleNamespace(SeleniumTest):
    """Test namespace list page with single namespace"""

    _TEST_DATA = one_namespace_test_data

    def test_namespace_list_page_redirect(self):
        """Test namespace list page with one namespace."""
        self.selenium_instance.get(self.get_url('/modules'))

        # Ensure page is redirected to namespace page
        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/modules/testnamespace'))


class TestNamespaceListNoNamespaces(SeleniumTest):
    """Test namespace list page with no namespaces"""

    _TEST_DATA = {}

    def test_namespace_list_page_warning(self):
        """Test namespace list page with no namespaces."""
        self.selenium_instance.get(self.get_url('/modules'))

        # Ensure warning is added to table
        table_content = self.wait_for_element(By.ID, 'namespaces-table-data')
        warning_row = self.wait_for_element(By.TAG_NAME, 'tr', parent=table_content)
        warning_td = warning_row.find_element(By.TAG_NAME, 'td')

        assert warning_td.text == 'There are no namespaces with published modules.'
