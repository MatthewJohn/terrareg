
from datetime import datetime
from unittest import mock

import pytest
from selenium.webdriver.common.by import By
import selenium

from test.selenium import SeleniumTest

class TestLogin(SeleniumTest):
    """Test homepage."""

    @classmethod
    def setup_class(cls):
        """Setup required mocks."""
        cls._mock_openid_connect_is_enabled = mock.MagicMock(return_value=False)
        cls._mock_openid_connect_get_authorize_redirect_url = mock.MagicMock(return_value=(None, None))
        cls._mock_openid_connect_fetch_access_token = mock.MagicMock(return_value=None)
        cls._config_secret_key_mock = mock.patch('terrareg.config.Config.SECRET_KEY', '')

        cls.register_patch(mock.patch('terrareg.openid_connect.OpenidConnect.is_enabled', cls._mock_openid_connect_is_enabled))
        cls.register_patch(mock.patch('terrareg.openid_connect.OpenidConnect.get_authorize_redirect_url', cls._mock_openid_connect_get_authorize_redirect_url))
        cls.register_patch(mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', cls._mock_openid_connect_fetch_access_token))
        cls.register_patch(cls._config_secret_key_mock)
        super(TestLogin, cls).setup_class()

    def test_ensure_sso_login_not_shown(self):
        """Ensure SSO login button is not shown when SSO login is not available"""
        with self.update_mock(self._mock_openid_connect_is_enabled, 'return_value', False):
            self.selenium_instance.get(self.get_url('/login'))
            # Wait for normal login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'login-button').is_displayed(), True)
            # Ensure SSO login is not displayed
            assert self.selenium_instance.find_element(By.ID, 'sso-login').is_displayed() == False

    def test_valid_sso_login(self):
        """Ensure SSO login works"""
        with self.update_mock(self._mock_openid_connect_is_enabled, 'return_value', True), \
                self.update_mock(self._mock_openid_connect_get_authorize_redirect_url, 'return_value',
                                 ('/openid/callback?code=abcdefg&state=unitteststate', 'unitteststate')), \
                self.update_mock(self._mock_openid_connect_fetch_access_token, 'return_value',
                                 {'access_token': 'unittestaccesstoken', 'id_token': 'unittestidtoken', 'expires_in': 6000}), \
                self.update_mock(self._config_secret_key_mock, 'new', 'abcdefabcdef'):
            self.selenium_instance.get(self.get_url('/login'))
            # Wait for SSO login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'sso-login').is_displayed(), True)

            self.selenium_instance.find_element(By.ID, 'sso-login').click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))

            # Ensure user is logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')
