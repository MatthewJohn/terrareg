
import functools
import multiprocessing
import os


from pyvirtualdisplay import Display
from selenium import webdriver
import pytest

from terrareg.models import (
    Namespace, Module, ModuleProvider,
    ModuleVersion, GitProvider
)
from terrareg.database import Database
from terrareg.server import Server
import terrareg.config
from test import BaseTest
from .test_data import integration_test_data, integration_git_providers


class SeleniumTest(BaseTest):

    _TEST_DATA = integration_test_data
    _GIT_PROVIDER_DATA = integration_git_providers
    DISPLAY_INSTANCE = None
    SELENIUM_INSTANCE = None
    RESET_COOKIES = True

    DEFAULT_RESOLUTION = (1280, 720)

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        return 'temp-selenium.db'

    @staticmethod
    def get_selenium_instance():
        """Obtain singleton instance of selenium"""
        if SeleniumTest.SELENIUM_INSTANCE is None:
            SeleniumTest.get_display_instance()
            SeleniumTest.SELENIUM_INSTANCE = webdriver.Firefox()
        elif SeleniumTest.RESET_COOKIES:
            SeleniumTest.SELENIUM_INSTANCE.delete_all_cookies()
            SeleniumTest.RESET_COOKIES = False
        return SeleniumTest.SELENIUM_INSTANCE

    @staticmethod
    def get_display_instance():
        """Obtain singleton instance of display"""
        if SeleniumTest.DISPLAY_INSTANCE is None:
            SeleniumTest.DISPLAY_INSTANCE = Display(visible=0, size=SeleniumTest.DEFAULT_RESOLUTION)
            SeleniumTest.DISPLAY_INSTANCE.start()
        return SeleniumTest.DISPLAY_INSTANCE

    def setup_class(self):
        """Setup test server"""
        super(SeleniumTest, self).setup_class(self)

        self._server_thread = multiprocessing.Process(
            target=self.SERVER.run,
            daemon=True
        )
        self._server_thread.start()

    def teardown_class(self):
        """Stop test server."""
        self._server_thread.terminate()
