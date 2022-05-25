
import time
from unittest import mock
import pytest

from selenium.webdriver.common.by import By

from test.selenium import SeleniumTest

class TestHomepage(SeleniumTest):
    """Test homepage."""

    def test_title(self):
        """Check homepage."""
        with mock.patch('terrareg.config.Config.APPLICATION_NAME', 'unittest application name'), \
               mock.patch('terrareg.analytics.AnalyticsEngine.get_total_downloads', 2003):
            with self.run_server() as selenium:
                selenium.get(self.get_url('/'))
                time.sleep(5)

                # Ensure title is injected correctly
                assert selenium.find_element(By.ID, 'title').text == 'unittest application name'

                # Ensure counts on page are correct
                #assert selenium.find_element(By.ID, 'namespace-count').text == 5
                #assert selenium.find_element(By.ID, 'module-count').text == 5
                #assert selenium.find_element(By.ID, 'version-count').text == 5
                assert selenium.find_element(By.ID, 'download-count').text == 5
