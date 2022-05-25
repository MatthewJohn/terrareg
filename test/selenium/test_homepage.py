
import pytest

from selenium.webdriver.common.by import By

from test.selenium import SeleniumTest

class TestHomepage(SeleniumTest):
    """Test homepage."""

    def test_homepage(self):
        """Check homepage."""
        selenium = self.get_selenium_instance()
        selenium.get(self.get_url('/'))
        assert selenium.find_element(By.ID, 'title').text == 'Terrareg'
