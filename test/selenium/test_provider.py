from unittest import mock

from flask import request
import pytest
from selenium.webdriver.common.by import By

from test.selenium import SeleniumTest


class TestProvider(SeleniumTest):
    """Test provider page."""

    _SECRET_KEY = '354867a669ef58d17d0513a0f3d02f4403354915139422a8931661a3dbccdffe'

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        # Capture last JSON data from versions POST request
        cls._last_api_mock_version_publish_body = None
        def version_post_side_effect(*args, **kwargs):
            cls._last_api_mock_version_publish_body = request.json
            return {'status': 'Success'}
        cls._api_version_post_mock = mock.Mock(side_effect=version_post_side_effect)
        cls._config_publish_api_keys_mock = mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', [])
        cls._config_enable_access_controls = mock.patch('terrareg.config.Config.ENABLE_ACCESS_CONTROLS', False)

        cls.register_patch(mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', 'unittest-password'))
        cls.register_patch(mock.patch('terrareg.server.api.ApiProviderVersions._post', cls._api_version_post_mock))
        cls.register_patch(cls._config_publish_api_keys_mock)
        cls.register_patch(cls._config_enable_access_controls)

        super(TestProvider, cls).setup_class()

    @pytest.mark.parametrize('url,expected_title', [
        ('/providers/initial-providers/mv', 'initial-providers/mv/2.0.1 - Terrareg'),
        ('/providers/initial-providers/mv/1.5.0', 'initial-providers/mv/1.5.0 - Terrareg'),
        ('/providers/initial-providers/mv/1.5.0/docs/resources/some_new_resource', 'mv_some_new_resource - Resources - initial-providers/mv/1.5.0 - Terrareg'),
        ('/providers/initial-providers/mv/1.5.0/docs/data-sources/some_thing', 'mv_some_thing - Data Sources - initial-providers/mv/1.5.0 - Terrareg'),
    ])
    def test_page_titles(self, url, expected_title):
        """Check page titles on pages."""
        self.selenium_instance.get(self.get_url(url))
        self.assert_equals(lambda: self.selenium_instance.title, expected_title)

    @pytest.mark.parametrize('url,expected_breadcrumb', [
        ('/providers/initial-providers/mv',
         'Providers\ninitial-providers\nmv'),
        ('/providers/initial-providers/mv/1.5.0',
         'Providers\ninitial-providers\nmv\n1.5.0'),
        ('/providers/initial-providers/mv/1.5.0/docs/resources/some_new_resource',
         'Providers\ninitial-providers\nmv\n1.5.0\nDocs\nResources\nmv_some_new_resource'),
        ('/providers/initial-providers/mv/1.5.0/docs/data-sources/some_thing',
         'Providers\ninitial-providers\nmv\n1.5.0\nDocs\nData Sources\nmv_some_thing'),
    ])
    def test_breadcrumbs(self, url, expected_breadcrumb):
        """Test breadcrumb displayed on page"""
        self.selenium_instance.get(self.get_url(url))
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'breadcrumb-ul').text, expected_breadcrumb)

    def test_provider_with_versions(self):
        """Test page functionality on a provider with version."""
        self.selenium_instance.get(self.get_url('/providers/initial-providers/mv/1.5.0'))

        # Wait for docs tab
        self.wait_for_element(By.ID, 'provider-tab-link-documentation')

        # Check index of docs are shown
        docs = self.selenium_instance.find_element(By.ID, "provider-doc-content")
        assert docs.is_displayed() is True
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, "provider-doc-content").text, "This is an overview of the provider!")

        # Ensure title, description, published date are shown
        title = self.selenium_instance.find_element(By.ID, "provider-title")
        assert title.is_displayed() is True
        assert title.text == "mv"
        title = self.selenium_instance.find_element(By.ID, "provider-description")
        assert title.is_displayed() is True
        assert title.text == "Test Multiple Versions"

        logo = self.selenium_instance.find_element(By.ID, "provider-logo-img")
        assert logo.is_displayed() is True
        assert logo.get_attribute("src") == "https://git.example.com/initalproviders/terraform-provider-test-initial.png"

        published_at = self.selenium_instance.find_element(By.ID, "published-at")
        assert published_at.is_displayed() is True
        assert published_at.text == "Published Mon, 11 Dec 2023 by initial-providers"
        assert published_at.find_element(By.TAG_NAME, "a").get_attribute("href") == self.get_url("/providers/initial-providers")

    def test_doc_urls(self):
        """Check sidebar doc links."""
        self.selenium_instance.get(self.get_url('/providers/initial-providers/mv/1.5.0'))

        # Wait for links to load
        self.wait_for_element(By.ID, 'doclink-data-sources-some_thing')

        doc_sidebar = self.wait_for_element(By.CLASS_NAME, 'provider-doc-menu')

        assert doc_sidebar.text == """
Overview
Resources
mv_thing_new
mv_thing
Data Sources
mv_some_thing
""".strip()

    @pytest.mark.parametrize('link_text, href', [
        ('Overview', '/providers/initial-providers/mv/1.5.0/docs'),
        ('mv_thing_new', '/providers/initial-providers/mv/1.5.0/docs/resources/some_new_resource'),
        ('mv_thing', '/providers/initial-providers/mv/1.5.0/docs/resources/some_resource'),
        ('mv_some_thing', '/providers/initial-providers/mv/1.5.0/docs/data-sources/some_thing')
    ])
    def test_doc_url_links(self, link_text, href):
        """Test documentation link redirection"""
        self.selenium_instance.get(self.get_url('/providers/initial-providers/mv/1.5.0'))

        # Wait for links to load
        self.wait_for_element(By.ID, 'doclink-data-sources-some_thing')

        doc_sidebar = self.wait_for_element(By.CLASS_NAME, 'provider-doc-menu')
        for sidebar_link in doc_sidebar.find_elements(By.TAG_NAME, "a"):
            if sidebar_link.text == link_text:
                sidebar_link.click()
                break

        self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url(href))

    def test_integrations_tab(self):
        """Ensure integrations tab is displayed correctly."""
        self.selenium_instance.get(self.get_url('/providers/initial-providers/mv/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'provider-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'provider-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'provider-tab-integrations')

        # Ensure tab is displayed
        self.assert_equals(lambda: integrations_tab_content.is_displayed(), True)

        integrations_table = integrations_tab_content.find_element(By.TAG_NAME, 'table')
        table_rows = integrations_table.find_elements(By.TAG_NAME, 'tr')

        expected_integrations = [
            [
                'Trigger version import',
                f'POST {self.get_url("/v1/providers/initial-providers/mv/versions")}\n' +
                'Accepts JSON body with "version" key with value of version to be imported'
            ],
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
        self.selenium_instance.get(self.get_url('/providers/initial-providers/mv/1.5.0'))

        # Wait for integrations tab button to be visible
        integrations_tab_button = self.wait_for_element(By.ID, 'provider-tab-link-integrations')

        # Ensure the integrations tab content is not visible
        assert self.wait_for_element(By.ID, 'provider-tab-integrations', ensure_displayed=False).is_displayed() == False

        # Click on integrations tab
        integrations_tab_button.click()

        integrations_tab_content = self.selenium_instance.find_element(By.ID, 'provider-tab-integrations')

        # Type version number and submit form
        integrations_tab_content.find_element(By.ID, 'indexProviderVersion').send_keys('5.2.1')
        integrations_tab_content.find_element(By.ID, 'integration-index-version-button').click()

        # Wait for success message to be displayed
        success_message = self.wait_for_element(By.ID, 'index-version-success', parent=integrations_tab_content)
        self.assert_equals(lambda: success_message.is_displayed(), True)
        self.assert_equals(lambda: success_message.text, 'Successfully indexed version')

        # Check error message is not displayed
        error_message = integrations_tab_content.find_element(By.ID, 'index-version-error')
        assert error_message.is_displayed() == False

        # Ensure version create endpoint was called and publish was not
        self._api_version_post_mock.assert_called_once_with(namespace='initial-providers', provider='mv')
        assert "version" in TestProvider._last_api_mock_version_publish_body
        assert TestProvider._last_api_mock_version_publish_body["version"] == "5.2.1"
