
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
        cls._config_openid_connect_button_text = mock.patch('terrareg.config.Config.OPENID_CONNECT_LOGIN_TEXT', '')
        cls._config_saml_button_text = mock.patch('terrareg.config.Config.SAML2_LOGIN_TEXT', '')
        cls._config_admin_authentication_token = mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', '')
        cls._config_enable_access_controls = mock.patch('terrareg.config.Config.ENABLE_ACCESS_CONTROLS', False)

        cls.register_patch(cls._config_openid_connect_button_text)
        cls.register_patch(cls._config_saml_button_text)
        cls.register_patch(cls._config_admin_authentication_token)
        cls.register_patch(cls._config_enable_access_controls)
        super(TestLogin, cls).setup_class()

    def teardown_method(self, method):
        """Clear down any cookes from the trst."""
        self.selenium_instance.delete_all_cookies()
        super(TestLogin, self).teardown_method(method)

    def _wait_for_login_form_ready(self):
        """Wait for login form to be rendered"""
        # Wait for login title
        self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'login-title').is_displayed(), True)

    @pytest.mark.parametrize('admin_token,openid_enabled,saml_enabled,warning_shown', [
        # Check warning is shown when no methods are available
        ('', False, False, True),
        # Cases where authentication method is enabled
        ('pass', False, False, False),
        ('', True, False, False),
        ('', False, True, False)
    ])
    def test_no_authentication_methods_warning(self, admin_token, openid_enabled, saml_enabled, warning_shown):
        """Test warning is shown when no authentication methods are available."""
        with self.update_multiple_mocks(
                (self._config_admin_authentication_token, 'new', admin_token),
                (self._mock_openid_connect_is_enabled, 'return_value', openid_enabled),
                (self._mock_saml2_is_enabled, 'return_value', saml_enabled)):
            self.selenium_instance.get(self.get_url('/login'))
            self._wait_for_login_form_ready()

            # Ensure warning is displayed
            warning = self.selenium_instance.find_element(By.ID, 'no-authentication-methods-warning')
            assert warning.is_displayed() == warning_shown
            if warning_shown:
                assert warning.text == 'Login is not available as there are no authentication methods configured'

    def test_ensure_admin_authentication_not_shown(self):
        """Ensure admin login form is not shown when admin password is not configured"""
        with self.update_mock(self._config_admin_authentication_token, 'new', ''):
            self.selenium_instance.get(self.get_url('/login'))
            self._wait_for_login_form_ready()

            # Ensure admin login is not displayed
            assert self.selenium_instance.find_element(By.ID, 'admin-login').is_displayed() == False

    def test_ensure_openid_connect_login_not_shown(self):
        """Ensure OpenID connect login button is not shown when OpenId connect login is not available"""
        with self.update_mock(self._mock_openid_connect_is_enabled, 'return_value', False):
            self.selenium_instance.get(self.get_url('/login'))
            self._wait_for_login_form_ready()

            # Ensure OpenID Connect login is not displayed
            assert self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed() == False

    @pytest.mark.parametrize('test_password', [
        '',
        'incorrectpassword'
    ])
    def test_admin_password_login_invalid_password(self, test_password):
        """Test admin authentication using incorrect password"""
        with self.update_multiple_mocks((self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                (self._config_admin_authentication_token, 'new', 'correct-password')):

            self.selenium_instance.get(self.get_url('/login'))
            # Wait for admin login form to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'admin-login').is_displayed(), True)

            # Fill out admin password
            self.selenium_instance.find_element(By.ID, 'admin_token_input').send_keys(test_password)
            self.selenium_instance.find_element(By.ID, 'login-button').click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/login'))

            # Ensure user is logged in
            assert self.selenium_instance.find_element(By.ID, 'navbar_login_span').text == 'Login'

            # Ensure error is displayed about incorrect password
            error_div = self.selenium_instance.find_element(By.ID, 'login_error')
            assert error_div.is_displayed() == True
            assert error_div.text == 'Incorrect admin token'

    def test_admin_password_login(self):
        """Test admin authentication using password"""
        with self.update_multiple_mocks((self._mock_saml2_is_enabled, 'return_value', True), \
                (self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                (self._config_admin_authentication_token, 'new', 'testloginpassword')):

            self.selenium_instance.get(self.get_url('/login'))
            # Wait for admin login form to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'admin-login').is_displayed(), True)

            # Fill out admin password
            self.selenium_instance.find_element(By.ID, 'admin_token_input').send_keys('testloginpassword')
            self.selenium_instance.find_element(By.ID, 'login-button').click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))

            # Ensure user is logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')

            # Ensure 'settings' drop-down is shown, depending on whether
            # user is a site admin
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), True)

            # Ensure 'create' drop-down is shown, depending on whether
            # user has permissions to a namespace
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), True)

    @pytest.mark.parametrize('enable_access_controls,group_memberships,has_site_admin,can_create_module', [
        (True, ['nopermissions'], False, False),
        (True, ['siteadmin'], True, True),
        (True, ['moduledetailsfull'], False, True),
        (True, [], False, False),
        (False, ['nopermissions'], True, True),
        (False, ['siteadmin'], True, True),
        (False, ['moduledetailsfull'], True, True),
        (False, [], True, True),
    ])
    def test_valid_openid_connect_login(self, enable_access_controls, group_memberships, has_site_admin, can_create_module):
        """Ensure OpenID Connect login works"""
        with self.update_multiple_mocks((self._mock_openid_connect_is_enabled, 'return_value', True),
                (self._config_enable_access_controls, 'new', enable_access_controls),
                (self._mock_openid_connect_get_authorize_redirect_url, 'return_value',
                                 ('/openid/callback?code=abcdefg&state=unitteststate', 'unitteststate')),
                (self._mock_openid_connect_fetch_access_token, 'return_value',
                                 {'access_token': 'unittestaccesstoken', 'id_token': 'unittestidtoken', 'expires_in': 6000}), \
                (self._mock_openid_connect_get_user_info, 'return_value',
                                 {'groups': group_memberships}), \
                (self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                (self._config_openid_connect_button_text, 'new', 'Unittest OpenID Connect Login Button'), \
                (self._mock_openid_connect_validate_session_token, 'return_value', True)):
            self.selenium_instance.get(self.get_url('/login'))
            # Wait for SSO login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed(), True)

            openid_connect_login_button = self.selenium_instance.find_element(By.ID, 'openid-connect-login')

            assert openid_connect_login_button.text == 'Unittest OpenID Connect Login Button'
            openid_connect_login_button.click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))

            # Ensure user is logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')

            # Ensure 'settings' drop-down is shown, depending on whether
            # user is a site admin
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), has_site_admin)

            # Ensure 'create' drop-down is shown, depending on whether
            # user has permissions to a namespace
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), can_create_module)

            self._mock_openid_connect_validate_session_token.assert_called_with('unittestidtoken')

    def test_invalid_openid_connect_response(self):
        """Test handling of invalid OpenID connect authentication error"""
        def raise_exception():
            raise Exception('Unittest exception')
        with self.update_multiple_mocks((self._mock_openid_connect_is_enabled, 'return_value', True), \
                (self._mock_openid_connect_get_authorize_redirect_url, 'return_value',
                                 ('/openid/callback?code=abcdefg&state=unitteststate', 'unitteststate')), \
                (self._mock_openid_connect_fetch_access_token, 'side_effect',
                                 raise_exception), \
                (self._config_secret_key_mock, 'new', 'abcdefabcdef')):

            self.selenium_instance.get(self.get_url('/login'))
            # Wait for OpenID connect login button to be displayed and click
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed(), True)
            self.selenium_instance.find_element(By.ID, 'openid-connect-login').click()

            # Ensure still on callback URL and error is displayed
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/openid/callback?code=abcdefg&state=unitteststate'))
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-title').text, 'Login error')
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-content').text, 'Invalid response from SSO')

    def test_ensure_saml_login_not_shown(self):
        """Ensure SAML login button is not shown when SAML login is not available"""
        with self.update_mock(self._mock_saml2_is_enabled, 'return_value', False):
            self.selenium_instance.get(self.get_url('/login'))
            self._wait_for_login_form_ready()

            # Ensure SAML login is not displayed
            assert self.selenium_instance.find_element(By.ID, 'saml-login').is_displayed() == False

    @pytest.mark.parametrize('enable_access_controls,group_memberships,has_site_admin,can_create_module', [
        (True, ['nopermissions'], False, False),
        (True, ['siteadmin'], True, True),
        (True, ['moduledetailsfull'], False, True),
        (True, [], False, False),
        (False, ['nopermissions'], True, True),
        (False, ['siteadmin'], True, True),
        (False, ['moduledetailsfull'], True, True),
        (False, [], True, True),
    ])
    def test_valid_saml_login(self, enable_access_controls, group_memberships, has_site_admin, can_create_module):
        """Ensure SAML login works"""

        mock_auth_object = mock.MagicMock()

        # Functions for initial login call
        mock_auth_object.login = mock.MagicMock(return_value='/saml/login?acs')
        mock_auth_object.get_last_request_id = mock.MagicMock(return_value='unittestAuthRequestId')

        # Mothods for ACS redirect 
        mock_auth_object.process_response = mock.MagicMock()
        mock_auth_object.get_errors = mock.MagicMock(return_value=[])
        mock_auth_object.is_authenticated = mock.MagicMock(return_value=True)
        mock_auth_object.get_attributes = mock.MagicMock(return_value={'Login': ['testuser@localhost.com'], 'groups': group_memberships})
        mock_auth_object.get_nameid = mock.MagicMock(return_value='unittestSamlNamId')
        mock_auth_object.get_nameid_format = mock.MagicMock(return_value='unittestSamlNamIdFormat')
        mock_auth_object.get_nameid_nq = mock.MagicMock(return_value='unittestSamlNamIdNq')
        mock_auth_object.get_nameid_spnq = mock.MagicMock(return_value='unittestSamlNamIdSPNQ')
        mock_auth_object.get_session_index = mock.MagicMock(return_value='unittestSamlSessionIndex')

        with self.update_multiple_mocks((self._mock_saml2_is_enabled, 'return_value', True), \
                (self._config_enable_access_controls, 'new', enable_access_controls), \
                (self._mock_saml2_initialise_request_auth_object, 'return_value',
                                 mock_auth_object), \
                (self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                (self._config_saml_button_text, 'new', 'Unittest SAML Login Button')):

            self.selenium_instance.get(self.get_url('/login'))
            # Wait for SSO login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'saml-login').is_displayed(), True)

            saml_login_button = self.selenium_instance.find_element(By.ID, 'saml-login')

            assert saml_login_button.text == 'Unittest SAML Login Button'
            saml_login_button.click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))

            # Ensure user is logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')

            # Ensure 'settings' drop-down is shown, depending on whether
            # user is a site admin
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarSettingsDropdown').is_displayed(), has_site_admin)

            # Ensure 'create' drop-down is shown, depending on whether
            # user has permissions to a namespace
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbarCreateDropdown').is_displayed(), can_create_module)

            mock_auth_object.process_response.assert_called_with(request_id='unittestAuthRequestId')
            mock_auth_object.get_errors.assert_called()

    def test_invalid_saml_response(self):
        """Test handling of invalid SAML authentication error"""
        mock_auth_object = mock.MagicMock()

        # Functions for initial login call
        mock_auth_object.login = mock.MagicMock(return_value='/saml/login?acs')
        mock_auth_object.get_last_request_id = mock.MagicMock(return_value='unittestAuthRequestId')

        # Mothods for ACS redirect 
        mock_auth_object.process_response = mock.MagicMock()
        mock_auth_object.get_errors = mock.MagicMock(return_value=['This is an error'])
        mock_auth_object.is_authenticated = mock.MagicMock(return_value=True)
        mock_auth_object.get_attributes = mock.MagicMock(return_value='unittestSamlAttributes')

        with self.update_multiple_mocks((self._mock_saml2_is_enabled, 'return_value', True), \
                (self._mock_saml2_initialise_request_auth_object, 'return_value',
                                 mock_auth_object), \
                (self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                (self._config_saml_button_text, 'new', 'Unittest SAML Login Button')):

            self.selenium_instance.get(self.get_url('/login'))
            # Wait for SSO login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'saml-login').is_displayed(), True)

            self.selenium_instance.find_element(By.ID, 'saml-login').click()

            # Ensure still on callback URL and error is displayed
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/saml/login?acs'))
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-title').text, 'Login error')
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'error-content').text,
                               'An error occured whilst processing SAML login request')

            # Ensure user is not logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Login')

            mock_auth_object.get_attributes.assert_not_called()
