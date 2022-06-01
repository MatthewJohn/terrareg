
import functools
import multiprocessing
import os
import random
import logging
import threading
from time import sleep
from unittest.mock import patch
from flask import request


from pyvirtualdisplay import Display
from selenium import webdriver
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

    def get_url(self, path):
        """Return full URL to perform selenium request."""
        return 'http://localhost:{port}{path}'.format(port=self.SERVER.port, path=path)

    @classmethod
    def setup_class(cls):
        """Setup host/port to host server."""
        super(SeleniumTest, cls).setup_class()

        cls.SERVER.host = '127.0.0.1'

        cls.display_instance = Display(visible=0, size=SeleniumTest.DEFAULT_RESOLUTION)
        cls.display_instance.start()
        cls.selenium_instance = webdriver.Firefox()
        cls.selenium_instance.delete_all_cookies()
        cls.selenium_instance.implicitly_wait(1)

        cls.SERVER.port = random.randint(5000, 6000)

        log = logging.getLogger('werkzeug')
        log.disabled = True

        cls._werzeug_server = werkzeug.serving.make_server(
            "localhost",
            cls.SERVER.port,
            cls.SERVER._app)
        cls._server_thread = threading.Thread(
            target=cls._werzeug_server.serve_forever
        )
        cls._server_thread.start()

    @classmethod
    def teardown_class(cls):
        """Teardown display instance."""
        cls.selenium_instance.quit()
        cls.display_instance.stop()
        # Shutdown server
        cls._werzeug_server.shutdown()
        cls._server_thread.join()
        super(SeleniumTest, cls).teardown_class()

    def assert_equals(self, callback, value):
        """Attempt to verify assertion and retry on failure."""
        max_attempts = 5
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
