
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest
from terrareg.models import ModuleVersion, Namespace, Module, ModuleProvider

class TestHomepage(SeleniumTest):
    """Test homepage."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls.register_patch(mock.patch('terrareg.config.Config.APPLICATION_NAME', 'unittest application name')),
        cls.register_patch(mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', return_value=2005))
        cls.register_patch(mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'))
        cls.register_patch(mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'))
        cls.register_patch(mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['trustednamespace']))
        super(TestHomepage, cls).setup_class()

    def test_title(self):
        """Check homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Ensure title is injected correctly
        assert self.selenium_instance.find_element(By.ID, 'title').text == 'unittest application name'

        assert self.selenium_instance.title == 'Home - unittest application name'

    @pytest.mark.parametrize('element,count', [
        ('namespace', 18),
        ('module', 63),
        ('version', 93),
        ('download', 2005)
    ])
    def test_counts(self, element, count):
        """Check counters on homepage."""
        self.selenium_instance.get(self.get_url('/'))

        # Ensure counts on page are correct
        assert self.selenium_instance.find_element(By.ID, f'{element}-count').text == str(count)

    def test_latest_module_version(self):
        """Check tabs for most recent uploaded."""
        # Update latest module publish date to now
        module_version = ModuleVersion(
            module_provider=ModuleProvider(
                module=Module(
                    namespace=Namespace(name='mostrecent'),
                    name='modulename'
                ),
                name='providername'
            ),
            version='1.2.3'
        )
        module_version.update_attributes(published_at=datetime.now())

        self.selenium_instance.get(self.get_url('/'))

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
        module_version = ModuleVersion(
            module_provider=ModuleProvider(
                module=Module(
                    namespace=Namespace(name='mostrecent'),
                    name='modulename'
                ),
                name='providername'
            ),
            version='1.2.3'
        )
        module_version.update_attributes(published_at=datetime.now())

        self.selenium_instance.get(self.get_url('/'))

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
