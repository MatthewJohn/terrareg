
from contextlib import contextmanager
import functools
import multiprocessing
import os
import random
import logging
import threading
from time import sleep
import unittest.mock
from attr import attrib
from flask import request


from pyvirtualdisplay import Display
import selenium
from selenium.webdriver.common.by import By
import pytest
import werkzeug

from terrareg.models import (
    Namespace, Module, ModuleProvider,
    ModuleVersion, GitProvider
)
from terrareg.database import Database
from terrareg.server import Server
import terrareg.config
from test import BaseTest
from .test_data import integration_test_data, integration_git_providers, selenium_user_group_data


class SeleniumTest(BaseTest):

    _TEST_DATA = integration_test_data
    _GIT_PROVIDER_DATA = integration_git_providers
    _USER_GROUP_DATA = selenium_user_group_data
    _MOCK_PATCHES = []
    _SECRET_KEY = ''

    DISPLAY_INSTANCE = None
    SELENIUM_INSTANCE = None
    RESET_COOKIES = True

    RUN_INTERACTIVELY = os.environ.get('RUN_INTERACTIVELY', False)

    DEFAULT_RESOLUTION = (1280, 720)

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        return 'temp-selenium.db'

    def get_url(self, path, https=False):
        """Return full URL to perform selenium request."""
        return f'http{"s" if https else ""}://localhost:{self.SERVER.port}{path}'

    @classmethod
    def register_patch(cls, patch):
        """Register mock patch in test"""
        assert patch not in cls._MOCK_PATCHES
        cls._MOCK_PATCHES.append(patch)

    @classmethod
    def _setup_auth_mocks(cls):
        """Setup mocks used to perform saml/openid connect authentication"""
        cls._mock_openid_connect_is_enabled = unittest.mock.MagicMock(return_value=False)
        cls._mock_openid_connect_get_authorize_redirect_url = unittest.mock.MagicMock(return_value=(None, None))
        cls._mock_openid_connect_fetch_access_token = unittest.mock.MagicMock(return_value=None)
        cls._mock_openid_connect_validate_session_token = unittest.mock.MagicMock(return_value=False)
        cls._mock_openid_connect_get_user_info = unittest.mock.MagicMock(return_value=None)
        cls._mock_saml2_is_enabled = unittest.mock.MagicMock(return_value=False)
        cls._mock_saml2_initialise_request_auth_object = unittest.mock.MagicMock()
        cls._config_secret_key_mock = unittest.mock.patch('terrareg.config.Config.SECRET_KEY', cls._SECRET_KEY)

        cls.register_patch(unittest.mock.patch('terrareg.openid_connect.OpenidConnect.is_enabled', cls._mock_openid_connect_is_enabled))
        cls.register_patch(unittest.mock.patch('terrareg.openid_connect.OpenidConnect.get_authorize_redirect_url', cls._mock_openid_connect_get_authorize_redirect_url))
        cls.register_patch(unittest.mock.patch('terrareg.openid_connect.OpenidConnect.fetch_access_token', cls._mock_openid_connect_fetch_access_token))
        cls.register_patch(unittest.mock.patch('terrareg.openid_connect.OpenidConnect.validate_session_token', cls._mock_openid_connect_validate_session_token))
        cls.register_patch(unittest.mock.patch('terrareg.openid_connect.OpenidConnect.get_user_info', cls._mock_openid_connect_get_user_info))
        cls.register_patch(unittest.mock.patch('terrareg.saml.Saml2.is_enabled', cls._mock_saml2_is_enabled))
        cls.register_patch(unittest.mock.patch('terrareg.saml.Saml2.initialise_request_auth_object', cls._mock_saml2_initialise_request_auth_object))
        cls.register_patch(cls._config_secret_key_mock)

    @classmethod
    def setup_class(cls):
        """Setup host/port to host server."""

        cls._setup_auth_mocks()

        super(SeleniumTest, cls).setup_class()

        # Start all mock patches
        for patch_ in cls._MOCK_PATCHES:
            patch_.start()

        cls.SERVER.host = '127.0.0.1'

        if not cls.RUN_INTERACTIVELY:
            cls.display_instance = Display(visible=0, size=SeleniumTest.DEFAULT_RESOLUTION)
            cls.display_instance.start()
        cls.selenium_instance = selenium.webdriver.Firefox()
        cls.selenium_instance.delete_all_cookies()
        cls.selenium_instance.implicitly_wait(1)

        log = logging.getLogger('werkzeug')
        if not cls.RUN_INTERACTIVELY:
            log.disabled = True

        cls._setup_server()

    @classmethod
    def teardown_class(cls):
        """Teardown display instance."""
        cls.selenium_instance.quit()
        if not cls.RUN_INTERACTIVELY:
            cls.display_instance.stop()
        # Shutdown server
        cls._teardown_server()

        # Stop all mock patches
        for patch_ in list(cls._MOCK_PATCHES):
            patch_.stop()
            # Remove reference to patch
            cls._MOCK_PATCHES.remove(patch_)

        super(SeleniumTest, cls).teardown_class()

    def setup_method(self):
        """Reset mock call histories."""
        for patch_ in self._MOCK_PATCHES:
            # If patch target is a Mock, reset it
            if isinstance(patch_.new, unittest.mock.Mock):
                patch_.new.reset_mock()

    @classmethod
    def restart_server(cls):
        """Restart server. This can be used when mocks are modified."""
        cls._teardown_server()
        cls._setup_server()

    @classmethod
    def _setup_server(cls):
        """Setup web server."""
        # Replicate APP key setting from Server.run
        cls.SERVER._app.secret_key = terrareg.config.Config().SECRET_KEY

        while True:
            cls.SERVER.port = random.randint(20000, 21000)
            try:
                cls._werzeug_server = werkzeug.serving.make_server(
                    "localhost",
                    cls.SERVER.port,
                    cls.SERVER._app)
                break
            except OSError as exc:
                if '[Errno 98] Address already in use' not in str(exc):
                    raise
                print('Selected port already in use')

        cls._server_thread = threading.Thread(
            target=cls._werzeug_server.serve_forever
        )
        cls._server_thread.start()

    @classmethod
    def _teardown_server(cls):
        """Stop web server."""
        cls._werzeug_server.shutdown()
        cls._server_thread.join()

    def assert_equals(self, callback, value):
        """Attempt to verify assertion and retry on failure."""
        max_attempts = 20
        for itx in range(max_attempts):
            try:
                # Attempt to call callback and assert value against expected result
                actual = callback()
                assert actual == value
                # Break once assertion has completed
                break
            except AssertionError:
                # If it fails the assertion,
                # sleep and retry until last attmept
                # and then re-raise
                if itx < (max_attempts - 1):
                    sleep(0.5)
                else:
                    print('Failed asserting that {} == {}'.format(actual, value))
                    raise

    def wait_for_element(self, by, val, parent=None, ensure_displayed=True):
        """Attempt to find element and wait, if it does not exist yet."""
        if parent is None:
            parent = self.selenium_instance

        max_attempts = 5
        for itx in range(max_attempts):
            try:
                # Attempt to find element
                element = parent.find_element(by, val)
                if ensure_displayed and not element.is_displayed():
                    raise selenium.common.exceptions.NoSuchElementException('Element is not displayed')
                return element
            except selenium.common.exceptions.NoSuchElementException:
                # If it fails the assertion,
                # sleep and retry until last attmept
                # and then re-raise
                if itx < (max_attempts - 1):
                    sleep(0.5)
                else:
                    print('Failed to find element')
                    raise

    def perform_admin_authentication(self, password):
        """Go to admin page and authenticate as admin"""
        self.selenium_instance.get(self.get_url('/login'))
        token_input_field = self.selenium_instance.find_element(By.ID, 'admin_token_input')
        token_input_field.send_keys(password)
        login_button = self.selenium_instance.find_element(By.ID, 'login-button')
        login_button.click()

        # Wait for homepage to load
        self.wait_for_element(By.ID, 'title')

    @contextmanager
    def log_in_with_openid_connect(self, user_groups):
        """Login with OpenID connect"""
        with self.update_mock(self._mock_openid_connect_is_enabled, 'return_value', True), \
                self.update_mock(self._mock_openid_connect_get_authorize_redirect_url, 'return_value',
                                 ('/openid/callback?code=abcdefg&state=unitteststate', 'unitteststate')), \
                self.update_mock(self._mock_openid_connect_fetch_access_token, 'return_value',
                                 {'access_token': 'unittestaccesstoken', 'id_token': 'unittestidtoken', 'expires_in': 6000}), \
                self.update_mock(self._mock_openid_connect_get_user_info, 'return_value',
                                 {'groups': user_groups}), \
                self.update_mock(self._config_secret_key_mock, 'new', 'abcdefabcdef'), \
                self.update_mock(self._mock_openid_connect_validate_session_token, 'return_value', True):
            self.selenium_instance.get(self.get_url('/login'))
            # Wait for SSO login button to be displayed
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'openid-connect-login').is_displayed(), True)

            openid_connect_login_button = self.selenium_instance.find_element(By.ID, 'openid-connect-login')
            openid_connect_login_button.click()

            # Ensure redirected to login
            self.assert_equals(lambda: self.selenium_instance.current_url, self.get_url('/'))

            # Ensure user is logged in
            self.assert_equals(lambda: self.selenium_instance.find_element(By.ID, 'navbar_login_span').text, 'Logout')

            yield

    def update_mock(self, *args, **kwargs):
        """Return context-manager instance for handling updating of mock attributes during selenium test."""
        return SeleniumMockUpdater(self, *args, **kwargs)


class SeleniumMockUpdater:
    """"Provide context-manager to update mock within selenium test"""

    def __init__(self, test: SeleniumTest, mock: object, attribute: str, new_value):
        """Store member variables for updating mock"""
        self._test = test
        self._mock = mock
        self._attribute = attribute
        self._new_value = new_value
        self._original_value = None

    def __enter__(self):
        """On enter, store current mock value, update mock and restart server"""
        self._original_value = getattr(self._mock, self._attribute)
        setattr(self._mock, self._attribute, self._new_value)
        # Stop/start patch, this is required when performing 
        # something like mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', []),
        # as the new value is pushed straight to the target, so there is no direct
        # reference to the new target value in the mock.
        self._restart_mock()
        self._test.restart_server()

    def __exit__(self, *args, **kwargs):
        """On exit, set original mock value and restart server"""
        setattr(self._mock, self._attribute, self._original_value)
        self._restart_mock()
        self._test.restart_server()

    def _restart_mock(self):
        """Restar the mock."""
        self._mock.stop()
        self._mock.start()
