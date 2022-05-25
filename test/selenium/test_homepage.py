
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

    def test_title(self):
        """Check homepage."""
        with mock.patch('terrareg.config.Config.APPLICATION_NAME', 'unittest application name'), \
                mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', return_value=2005), \
                mock.patch('terrareg.config.Config.CONTRIBUTED_NAMESPACE_LABEL', 'unittest contributed module'), \
                mock.patch('terrareg.config.Config.TRUSTED_NAMESPACE_LABEL', 'unittest trusted namespace'), \
                mock.patch('terrareg.config.Config.VERIFIED_MODULE_LABEL', 'unittest verified label'), \
                mock.patch('terrareg.config.Config.TRUSTED_NAMESPACES', ['trustednamespace']):
            with self.run_server() as selenium_con:

                selenium_con.implicitly_wait(5)

                selenium_con.get(self.get_url('/'))

                # Ensure title is injected correctly
                assert selenium_con.find_element(By.ID, 'title').text == 'unittest application name'

                # Ensure counts on page are correct
                assert selenium_con.find_element(By.ID, 'namespace-count').text == '11'
                assert selenium_con.find_element(By.ID, 'module-count').text == '45'
                assert selenium_con.find_element(By.ID, 'version-count').text == '59'
                assert selenium_con.find_element(By.ID, 'download-count').text == '2005'

                # Check tabs for most recent uploaded and most downloaded
                most_recent_module_version_title = selenium_con.find_element(
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

                # Make module verified and ensure tag is applied
                namespace = Namespace('mostrecent')
                module = Module(namespace=namespace, name='modulename')
                provider = ModuleProvider(module=module, name='providername')
                provider.update_attributes(verified=True)

                # Reload page and ensure verified tag is present
                selenium_con.get(self.get_url('/'))
                assert selenium_con.find_element(
                    By.ID, 'most-recent-module-version'
                ).find_element(
                    By.CLASS_NAME, 'result-card-label-verified'
                ).text == 'unittest verified label'

                # Update trusted module to latest module
                namespace = Namespace('trustednamespace')
                module = Module(namespace=namespace, name='secondlatestmodule')
                provider = ModuleProvider(module=module, name='aws')
                module_version = ModuleVersion(module_provider=provider, version='4.4.1')
                module_version.update_attributes(published_at=datetime.now())

                # Reload page and ensure updated module is shown and
                # trusted label is shown
                selenium_con.get(self.get_url('/'))
                most_recent_module_version_title = selenium_con.find_element(
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
