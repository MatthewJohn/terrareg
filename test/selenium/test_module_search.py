
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

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted']))
        super(TestModuleSearch, cls).setup_class()

    def test_search_from_homepage(self):
        """Check search functionality from homepage."""
        self.selenium_instance.get(self.get_url('/'))

        assert self.selenium_instance.title == 'Search - Terrareg'

        # Enter text into search input
        self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys('modulesearch')

        search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
        assert search_button.text == 'Search'
        search_button.click()

        assert self.selenium_instance.current_url == self.get_url('/modules/search?q=modulesearch')

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)

    def test_result_cards(self):
        """Check the result cards."""

        self.selenium_instance.get(self.get_url('/modules/search?q=modulesearch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')

        expected_card_headings = [
            'modulesearch-trusted / mixedsearch-trusted-result',
            'modulesearch-trusted / mixedsearch-trusted-result-multiversion',
            'modulesearch-trusted / mixedsearch-trusted-result-verified',
            'modulesearch-trusted / mixedsearch-trusted-second-result'
        ]
        expected_card_links = [
            '/modules/modulesearch-trusted/mixedsearch-trusted-result/aws/1.0.0',
            '/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null/2.0.0',
            '/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp/2.0.0',
            '/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog/5.2.1'
        ]
        expected_card_provider_text = [
            'Provider: aws',
            'Provider: null',
            'Provider: gcp',
            'Provider: datadog'
        ]
        for card in result_cards:
            heading = card.find_element(By.CLASS_NAME, 'module-card-title')

            # Check heading
            assert heading.text == expected_card_headings.pop(0)

            # Check link
            assert heading.get_attribute('href') == self.get_url(expected_card_links.pop(0))

            # Check provider
            assert card.find_element(By.CLASS_NAME, 'module-provider-card-provider-text').text == expected_card_provider_text.pop(0)

    def test_search_filters(self):
        """Check value of search filters."""

        self.selenium_instance.get(self.get_url('/modules/search?q=modulesearch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)

        # Check counts of filters
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-verified-count').text, '3')
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces-count').text, '4')
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed-count').text, '9')

        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-verified').is_selected(), False)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces').is_selected(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed').is_selected(), False)

        # Click verified label
        self.selenium_instance.find_element(By.ID, 'search-verified').click()

        # Ensure that the number of results has changed
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 1)
        for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
            assert card.find_element(By.CLASS_NAME, 'module-card-title').text == 'modulesearch-trusted / mixedsearch-trusted-result-verified'

        # Click contributed label
        self.selenium_instance.find_element(By.ID, 'search-contributed').click()

        # Ensure that the number of results has changed
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 3)

    def test_next_prev_buttons(self):
        """Check next and previous buttons."""
        self.selenium_instance.get(self.get_url('/modules/search?q=modulesearch'))

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
        assert len(result_cards) == 4

        # Ensure both buttons are disabled
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == False
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

        # Search for contributed modules
        self.selenium_instance.find_element(By.ID, 'search-contributed').click()

        # Ensure NextButton is active
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == True
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

        # Check number of results, which will also implicitly wait
        # for results to update.
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 10)

        # Get list of all cards
        first_page_cards = []
        for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
            first_page_cards.append(card.find_element(By.CLASS_NAME, 'module-card-title').text)

        # Select next page
        self.selenium_instance.find_element(By.ID, 'nextButton').click()

        # Ensure next button is disabled and prev button is enabled
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled(), False)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled(), True)

        # Wait again for first card to update
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'results').find_elements(
                    By.CLASS_NAME, 'card')[0].find_element(
                        By.CLASS_NAME, 'module-card-title').text,
            'modulesearch-trusted / mixedsearch-trusted-result-multiversion'
        )

        # Ensure that all cards have been updated
        for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
            assert card.find_element(By.CLASS_NAME, 'module-card-title').text not in first_page_cards

        # Select previous page
        self.selenium_instance.find_element(By.ID, 'prevButton').click()

        # Ensure prev button is disabled and next button is enabled
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled(), False)

        # Wait again for first card to update
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'results').find_elements(
                    By.CLASS_NAME, 'card')[0].find_element(
                        By.CLASS_NAME, 'module-card-title').text,
            'modulesearch / contributedmodule-differentprovider'
        )

        # Ensure that all of the original cards are displayed
        for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
            card_title = card.find_element(By.CLASS_NAME, 'module-card-title').text
            assert card_title in first_page_cards
            first_page_cards.remove(card_title)

        assert len(first_page_cards) == 0
