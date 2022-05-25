
from datetime import datetime
from time import sleep
from unittest import mock
import pytest

from selenium.webdriver.common.by import By
import selenium
from terrareg.database import Database

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestHomepage(SeleniumTest):
    """Test homepage."""

    def setup_class(self):
        """Setup required mocks."""
        self._mocks = []
        self._mocks.append(mock.patch('terrareg.config.Config.APPLICATION_NAME', 'unittest application name')),
        self._mocks.append(mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', return_value=2005))
        self._mocks.append(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        self._mocks.append(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        self._mocks.append(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        self._mocks.append(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['trustednamespace']))
        for mock_ in self._mocks:
            mock_.start()
        super(TestHomepage, self).setup_class(self)

    def teardown_class(self):
        """Setup required mocks."""
        for mock_ in self._mocks:
            mock_.stop()

    def test_title(self):
        """Check homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Ensure title is injected correctly
        assert self.selenium_instance.find_element(By.ID, 'title').text == 'unittest application name'

    def test_counts(self):
        """Check counters on homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Ensure counts on page are correct
        assert self.selenium_instance.find_element(By.ID, 'namespace-count').text == '11'
        assert self.selenium_instance.find_element(By.ID, 'module-count').text == '45'
        assert self.selenium_instance.find_element(By.ID, 'version-count').text == '59'
        assert self.selenium_instance.find_element(By.ID, 'download-count').text == '2005'

    def test_latest_module_version(self):
        """Check tabs for most recent uploaded."""
        most_recent_module_version_title = self.selenium_instance.find_element(
            By.ID, 'most-recent-module-version'
        ).find_element(
            By.CLASS_NAME, 'card-header-title'
        )

        # Check title of card
        assert most_recent_module_version_title.find_element(
                By.CLASS_NAME, 'module-card-title'
            ).text == 'mostrecent / modulename'
        # Check contributed tag is applied
        assert most_recent_module_version_title.find_element(
            By.CLASS_NAME,
            'result-card-label-contributed'
        ).text == 'unittest contributed module'

        # Ensure no other tags are applied
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            most_recent_module_version_title.find_element(
                By.CLASS_NAME,
                'result-card-label-trusted'
            )
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            most_recent_module_version_title.find_element(
                By.CLASS_NAME,
                'result-card-label-verified'
            )

    def test_verified_module_label(self):
        """Check verified module is shown with label"""
        # Make module verified and ensure tag is applied
        namespace = Namespace('mostrecent')
        module = Module(namespace=namespace, name='modulename')
        provider = ModuleProvider(module=module, name='providername')
        provider.update_attributes(verified=True)

        # Reload page and ensure verified tag is present
        self.selenium_instance.get(self.get_url('/'))
        assert self.selenium_instance.find_element(
            By.ID, 'most-recent-module-version'
        ).find_element(
            By.CLASS_NAME, 'result-card-label-verified'
        ).text == 'unittest verified label'

    def test_updated_trusted_module(self):
        """Update trusted module and ensure displayed."""
        # Update trusted module to latest module
        namespace = Namespace('trustednamespace')
        module = Module(namespace=namespace, name='secondlatestmodule')
        provider = ModuleProvider(module=module, name='aws')
        module_version = ModuleVersion(module_provider=provider, version='4.4.1')
        module_version.update_attributes(published_at=datetime.now())

        # Reload page and ensure updated module is shown and
        # trusted label is shown
        self.selenium_instance.get(self.get_url('/'))
        most_recent_module_version_title = self.selenium_instance.find_element(
            By.ID, 'most-recent-module-version'
        ).find_element(
            By.CLASS_NAME, 'card-header-title'
        )

        # Check title of card
        assert most_recent_module_version_title.find_element(
                By.CLASS_NAME, 'module-card-title'
            ).text == 'trustednamespace / secondlatestmodule'
        # Check contributed tag is applied
        assert most_recent_module_version_title.find_element(
            By.CLASS_NAME,
            'result-card-label-trusted'
        ).text == 'unittest trusted namespace'

        # Ensure no other tags are applied
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            most_recent_module_version_title.find_element(
                By.CLASS_NAME,
                'result-card-label-contributed'
            )
        with pytest.raises(selenium.common.exceptions.NoSuchElementException):
            most_recent_module_version_title.find_element(
                By.CLASS_NAME,
                'result-card-label-verified'
            )
