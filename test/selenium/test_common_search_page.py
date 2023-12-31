

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
        cls._config_trusted_namespaces_mock = mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['modulesearch-trusted', 'relevancysearch'])
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        cls.register_patch(cls._config_trusted_namespaces_mock)
        super(TestModuleSearch, cls).setup_class()

    @pytest.mark.parametrize('search_string', [
        # Test string that will match modules and providers
        (''),
        ('mixed'),
    ])
    def test_search_from_homepage_common_search(self, search_string):
        """Check search functionality from homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Enter text into search input
        self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)

        search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
        assert search_button.text == 'Search'
        search_button.click()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(f'/search?q={search_string}'))
        assert self.selenium_instance.title == 'Search - Terrareg'


    @pytest.mark.parametrize('search_string, expected_url, expected_title', [
        ('fullypopulated', '/search/modules?q=fullypopulated', 'Module Search - Terrareg'),
        ('initial-providers', '/search/providers?q=initial-providers', 'Provider Search - Terrareg'),
    ])
    def test_search_from_homepage_redirect_type_search(self, search_string, expected_url, expected_title):
        """Check search functionality from homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Enter text into search input
        self.selenium_instance.find_element(By.ID, 'navBarSearchInput').send_keys(search_string)

        search_button = self.selenium_instance.find_element(By.ID, 'navBarSearchButton')
        assert search_button.text == 'Search'
        search_button.click()

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(expected_url))
        self.assert_equals(lambda: self.selenium_instance.title, expected_title)

        # Ensure search filter has been populated with correct search string
        assert self.selenium_instance.find_element(By.ID, 'search-query-string').get_attribute('value') == search_string


    def test_result_cards(self):
        """Check result cards in common search page."""
        self.selenium_instance.get(self.get_url('/search?q=mixed'))

        self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")

        # Check provider cards
        provider_cards = [
            {"link": "/providers/contributed-providersearch/mixedsearch-result", "text": "contributed-providersearch / mixedsearch-result"},
            {"link": "/providers/contributed-providersearch/mixedsearch-result-multiversion", "text": "contributed-providersearch / mixedsearch-result-multiversion"},
            {"link": "/providers/providersearch-trusted/mixedsearch-trusted-result", "text": "providersearch-trusted / mixedsearch-trusted-result"},
            {"link": "/providers/providersearch-trusted/mixedsearch-trusted-second-result", "text": "providersearch-trusted / mixedsearch-trusted-second-result"},
            {"link": "/providers/providersearch-trusted/mixedsearch-trusted-result-multiversion", "text": "providersearch-trusted / mixedsearch-trusted-result-multiversion"}
        ]
        for card in self.selenium_instance.find_element(By.ID, "results-providers-content").find_elements(By.CLASS_NAME, "result-box"):
            card_details = provider_cards.pop(0)
            for link in card.find_elements(By.TAG_NAME, "a"):
                assert link.get_attribute("href") == self.get_url(card_details["link"])
            assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]

        # Check module cards
        module_cards = [
            {"link": "/modules/modulesearch-contributed/mixedsearch-result/aws", "text": "modulesearch-contributed / mixedsearch-result"},
            {"link": "/modules/modulesearch-contributed/mixedsearch-result-multiversion/aws", "text": "modulesearch-contributed / mixedsearch-result-multiversion"},
            {"link": "/modules/modulesearch-trusted/mixedsearch-trusted-result/aws", "text": "modulesearch-trusted / mixedsearch-trusted-result"},
            {"link": "/modules/modulesearch-trusted/mixedsearch-trusted-second-result/datadog", "text": "modulesearch-trusted / mixedsearch-trusted-second-result"},
            {"link": "/modules/modulesearch-trusted/mixedsearch-trusted-result-multiversion/null", "text": "modulesearch-trusted / mixedsearch-trusted-result-multiversion"},
            {"link": "/modules/modulesearch-trusted/mixedsearch-trusted-result-verified/gcp", "text": "modulesearch-trusted / mixedsearch-trusted-result-verified"}
        ]
        for card in self.selenium_instance.find_element(By.ID, "results-modules-content").find_elements(By.CLASS_NAME, "result-box"):
            card_details = module_cards.pop(0)
            for link in card.find_elements(By.TAG_NAME, "a"):
                if "provider-logo-link" not in link.get_attribute("class"):
                    assert link.get_attribute("href") == self.get_url(card_details["link"])
            assert card.find_element(By.CLASS_NAME, "module-card-title").text == card_details["text"]


    def test_provider_results_button(self):
        """Check link to provider results"""
        self.selenium_instance.get(self.get_url('/search?q=mixed'))

        self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
        button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all provider results']")
        button.click()
        assert self.selenium_instance.current_url == self.get_url('/search/providers?q=mixed')

    def test_module_results_button(self):
        """Check link to provider results"""
        self.selenium_instance.get(self.get_url('/search?q=mixed'))

        self.wait_for_element(By.ID, "contributed-providersearch.mixedsearch-result.1.0.0")
        button = self.selenium_instance.find_element(By.XPATH, ".//button[text()='View all module results']")
        button.click()
        assert self.selenium_instance.current_url == self.get_url('/search/modules?q=mixed')
