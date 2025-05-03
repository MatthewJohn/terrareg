
from datetime import datetime
from time import sleep
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
import terrareg.database
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestModuleSearch(SeleniumTest):
    """Test homepage."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._config_trusted_namespaces_mock = mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted', 'relevancysearch'])
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        cls.register_patch(cls._config_trusted_namespaces_mock)
        super(TestModuleSearch, cls).setup_class()

    def test_result_cards(self):
        """Check the result cards."""

        self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)

        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')

        expected_card_headings = [
            'modulesearch-trusted / mixedsearch-trusted-result',
            'modulesearch-trusted / mixedsearch-trusted-result-multiversion',
            'modulesearch-trusted / mixedsearch-trusted-result-verified',
            'modulesearch-trusted / mixedsearch-trusted-second-result',
        ]
        expected_card_links = [
            '/modules/modulesearch-trusted/mixedsearch-trusted-result/aws',
            '/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null',
            '/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp',
            '/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog',
        ]
        expected_card_provider_text = [
            'Provider: aws',
            'Provider: null',
            'Provider: gcp',
            'Provider: datadog',
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

        self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))

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
        self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))

        # Ensure 4 results are found
        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 4)

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
        with terrareg.database.Database.get().get_connection() as conn:
            db_dialect = conn.engine.name
        self.assert_equals(
            lambda: self.selenium_instance.find_element(
                By.ID, 'results').find_elements(
                    By.CLASS_NAME, 'card')[0].find_element(
                        By.CLASS_NAME, 'module-card-title').text,
            # Handle differences in module search result ordering for Postgres vs other engines
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

    def test_result_counts(self):
        """Check result count text."""
        self.selenium_instance.get(self.get_url('/search/modules?q=modulesearch'))

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 4 of 4')

        # Search for contributed modules
        self.selenium_instance.find_element(By.ID, 'search-contributed').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 13')

        # Select next page
        self.selenium_instance.find_element(By.ID, 'nextButton').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 11 - 13 of 13')

        # Select previous page
        self.selenium_instance.find_element(By.ID, 'prevButton').click()

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 1 - 10 of 13')

        self.selenium_instance.get(self.get_url('/search/modules?q=doesnotexist'))

        # Check total count
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'result-count').text, 'Showing results 0 - 0 of 0')

    def test_result_relevancy_ordering(self):
        """Test results are displayed in relevancy order"""
        self.selenium_instance.get(self.get_url('/search/modules?q=namematch'))

        self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 8)
        result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')

        expected_card_headings = [
            ('relevancysearch / namematch', 'namematch'),
            ('relevancysearch / namematch', 'partialprovidernamematch'),
            ('relevancysearch / partialmodulenamematch', 'namematch'),
            ('relevancysearch / descriptionmatch', 'testprovider'),
            ('relevancysearch / ownermatch', 'testprovider'),
            ('relevancysearch / partialmodulenamematch', 'partialprovidernamematch'),
            ('relevancysearch / partialdescriptionmatch', 'testprovider'),
            ('relevancysearch / partialownermatch', 'testprovider')
        ]

        for expected_heading, expected_provider in expected_card_headings:
            card = result_cards.pop(0)
            heading = card.find_element(By.CLASS_NAME, 'module-card-title')

            assert heading.text == expected_heading
            assert card.find_element(By.CLASS_NAME, 'module-provider-card-provider-text').text == f"Provider: {expected_provider}"

    @pytest.mark.parametrize('input_terraform_version, expected_compatibility_text', [
        ('2.5.0', [
            # First item has no entry, due to invalid version constraint
            None,
            'Compatible',
            'Implicitly compatible',
            'No version constraint defined',
            'Compatible',
        ]),

        ('0.5.0', [
            # First item has no entry, due to invalid version constraint
            None,
            'Incompatible',
            'Incompatible',
            'No version constraint defined',
            'Incompatible',
        ])
    ])
    def test_terraform_version_compatibility(self, input_terraform_version, expected_compatibility_text):
        """Test terraform compatiblity input and display"""
        # Delete cookies/local storage to remove pre-set terraform version
        self.selenium_instance.delete_all_cookies()

        with self.update_mock(self._config_trusted_namespaces_mock, 'new', ['version-constraint-test']):
            # Search for modules
            self.selenium_instance.get(self.get_url('/search/modules?q=version-constraint-test'))

            self.assert_equals(lambda: len(self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')), 5)

            # Ensure non of the result cards contain version constraints
            assert self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card-terraform-version-compatibility') == []

            # Inspect and enter Terraform version into target Terraform input and update
            terraform_input = self.selenium_instance.find_element(By.ID, 'search-terraform-version')
            assert terraform_input.get_attribute("value") == ""
            terraform_input.send_keys(input_terraform_version)

            # Click update
            self.selenium_instance.find_element(By.ID, 'search-options-update-button').click()

            result_cards = self.selenium_instance.find_element(By.ID, 'results').find_elements(By.CLASS_NAME, 'card')
            def get_result_card_compatibility_text(result_card):
                try:
                    return result_card.find_element(By.CLASS_NAME, 'card-terraform-version-compatibility').text
                except:
                    pass
                return None
            actual_result_card_compatibility_text = [
                get_result_card_compatibility_text(result_card)
                for result_card in result_cards
            ]
            assert actual_result_card_compatibility_text == expected_compatibility_text

    def test_terraform_version_compatibility_retains_state(self):
        """Test terraform compatiblity input is retained between page loads"""
        self.selenium_instance.delete_all_cookies()

        # Search for modules
        self.selenium_instance.get(self.get_url('/search/modules?q='))

        # Inspect and enter Terraform version into target Terraform input and update
        terraform_input = self.selenium_instance.find_element(By.ID, 'search-terraform-version')
        assert terraform_input.get_attribute("value") == ""
        terraform_input.send_keys("5.2.6-unittest")

        # Click update
        self.selenium_instance.find_element(By.ID, 'search-options-update-button').click()

        # Wait for page to show version constraints on results
        self.wait_for_element(By.CLASS_NAME, "card-terraform-version-compatibility")

        # Reload page
        self.selenium_instance.get(self.get_url('/search/modules?q='))
        # Check terraform version is still present
        terraform_version_constraint = self.wait_for_element(By.ID, 'search-terraform-version')
        assert terraform_version_constraint.get_attribute("value") == "5.2.6-unittest"

        # Open user preferences and check terraform version
        self.selenium_instance.find_element(By.ID, 'navbar-user-preferences-link').click()
        # Check user input for terraform version constraint
        self.assert_equals(lambda: self.wait_for_element(By.ID, 'user-preferences-terraform-compatibility-version').get_attribute("value"), "5.2.6-unittest")
