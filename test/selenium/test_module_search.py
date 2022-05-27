
from datetime import datetime
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestModuleSearch(SeleniumTest):
    """Test homepage."""

    def setup_class(self):
        """Setup required mocks."""
        self._mocks = []
        self._mocks.append(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        self._mocks.append(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        self._mocks.append(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        self._mocks.append(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted']))
        for mock_ in self._mocks:
            mock_.start()
        super(TestModuleSearch, self).setup_class(self)

    def teardown_class(self):
        """Setup required mocks."""
        for mock_ in self._mocks:
            mock_.stop()

    def test_search_from_homepage(self):
        """Check search functionality from homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Enter text into search input
        self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys('modulesearch')

        search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
        assert search_button.text == 'Search'
        search_button.click()

        assert self.selenium_instance.current_url == self.get_url('/modules/search?q=modulesearch')

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
        assert len(result_cards) == 3

    def test_result_cards(self):
        """Check the result cards."""

        self.selenium_instance.get(self.get_url('/modules/search?q=modulesearch'))

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
        assert len(result_cards) == 3

        expected_card_headings = [
            'modulesearch-trusted / mixedsearch-trusted-result',
            'modulesearch-trusted / mixedsearch-trusted-result-multiversion',
            'modulesearch-trusted / mixedsearch-trusted-second-result'
        ]
        expected_card_links = [
            '/modules/modulesearch-trusted/mixedsearch-trusted-result/aws/1.0.0',
            '/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/aws/2.0.0',
            '/modules/modulesearch-trusted/mixedsearch-trusted-second-result/aws/5.2.1'
        ]
        for card in result_cards:
            heading = card.find_element(By.CLASS_NAME, 'module-card-title')

            # Check heading
            assert heading.text == expected_card_headings.pop(0)

            # Check link
            assert heading.get_attribute('href') == self.get_url(expected_card_links.pop(0))

    def test_search_filters(self):
        """Check value of search filters."""

        self.selenium_instance.get(self.get_url('/modules/search?q=modulesearch'))

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
        assert len(result_cards) == 3

        # Check counts of filters
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-verified-count').text, '3')
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces-count').text, '3')
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed-count').text, '10')
