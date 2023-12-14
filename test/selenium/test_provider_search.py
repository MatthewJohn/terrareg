
from datetime import datetime
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestProviderSearch(SeleniumTest):
    """Test provider search page."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._config_trusted_namespaces_mock = mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['providersearch-trusted', 'relevancysearch'])
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(cls._config_trusted_namespaces_mock)
        super().setup_class()

    def test_result_cards(self):
        """Check the result cards."""

        self.selenium_instance.get(self.get_url('/search/providers?q=providersearch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 3)

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')

        expected_card_headings = [
            'providersearch-trusted / mixedsearch-trusted-result',
            'providersearch-trusted / mixedsearch-trusted-second-result',
            'providersearch-trusted / mixedsearch-trusted-result-multiversion',
        ]
        expected_card_links = [
            '/providers/providersearch-trusted/mixedsearch-trusted-result',
            '/providers/providersearch-trusted/mixedsearch-trusted-second-result',
            '/providers/providersearch-trusted/mixedsearch-trusted-result-multiversion',
        ]
        expected_descriptions = [
            'Provider: aws',
            'Provider: datadog',
            'Provider: null',
        ]
        expected_sources = [
            'Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-result',
            'Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-second-result',
            'Source: https://github.example.com/providersearch-trusted/terraform-provider-mixedsearch-trusted-result-multiversion',
        ]
        for card in result_cards:
            heading = card.find_element(By.CLASS_NAME, 'module-card-title')

            # Check heading
            assert heading.text == expected_card_headings.pop(0)

            # Check link
            assert heading.get_attribute('href') == self.get_url(expected_card_links.pop(0))

            footer = card.find_element(By.CLASS_NAME, "card-footer")
            assert footer.find_element(By.CLASS_NAME, "card-source-link").text == expected_sources.pop(0)

    def test_search_filters(self):
        """Check value of search filters."""

        self.selenium_instance.get(self.get_url('/search/providers?q=providersearch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 3)

        # Check counts of filters
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces-count').text, '3')
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed-count').text, '4')

        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces').is_selected(), True)
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'search-contributed').is_selected(), False)

        # Click contributed label
        self.selenium_instance.find_element(By.ID, 'search-contributed').click()

        # Ensure that the number of results has changed
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 7)

        # Uncheck trusted contributed label
        self.selenium_instance.find_element(By.ID, 'search-trusted-namespaces').click()

        # Ensure that the number of results has changed
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)


    def test_next_prev_buttons(self):
        """Check next and previous buttons."""
        self.selenium_instance.get(self.get_url('/search/providers?q=providersearch'))

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
        assert len(result_cards) == 3

        # Ensure both buttons are disabled
        self.selenium_instance.find_element(By.ID, 'nextButton').is_enabled() == False
        self.selenium_instance.find_element(By.ID, 'prevButton').is_enabled() == False

        # Search for contributed modules
        self.selenium_instance.get(self.get_url('/search/providers?q='))
        # Wait for results
        self.wait_for_element(By.CLASS_NAME, "module-card-title")
        # Show contributed modules
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
            'providersearch / contributedprovider-multiversion'
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
            'relevancysearch / descriptionmatch'
        )

        # Ensure that all of the original cards are displayed
        for card in self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card'):
            card_title = card.find_element(By.CLASS_NAME, 'module-card-title').text
            assert card_title in first_page_cards
            first_page_cards.remove(card_title)

        assert len(first_page_cards) == 0

    def test_result_counts(self):
        """Check result count text."""
        self.selenium_instance.get(self.get_url('/search/providers?q='))

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 7 of 7')

        # Search for contributed modules
        self.selenium_instance.find_element(By.ID, 'search-contributed').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 16')

        # Select next page
        self.selenium_instance.find_element(By.ID, 'nextButton').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 11 - 16 of 16')

        # Select previous page
        self.selenium_instance.find_element(By.ID, 'prevButton').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 16')

        self.selenium_instance.get(self.get_url('/search/providers?q=doesnotexist'))

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 0 - 0 of 0')

    def test_result_relevancy_ordering(self):
        """Test results are displayed in relevancy order"""
        self.selenium_instance.get(self.get_url('/search/providers?q=namematch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)
        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')

        expected_card_headings = [
            'relevancysearch / namematch',
            'relevancysearch / descriptionmatch',
            'relevancysearch / partialnamematch',
            'relevancysearch / partialdescriptionmatch',
        ]

        for expected_heading in expected_card_headings:
            card = result_cards.pop(0)
            heading = card.find_element(By.CLASS_NAME, 'module-card-title')

            assert heading.text == expected_heading
