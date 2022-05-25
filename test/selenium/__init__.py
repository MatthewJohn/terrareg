
import functools
import multiprocessing
import os
import random
import threading
from time import sleep
from unittest.mock import patch
from flask import request


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


def stop_server():
    """Shutdown flask server."""
    request.environ.get('werkzeug.server.shutdown')()

class SeleniumTestServer:

    def __init__(self, test_instance):
        """Capture test_instance."""
        self.test_instance = test_instance

    def __enter__(self) -> webdriver.Firefox:
        """Setup flask server."""
        return self.test_instance.selenium_instance

    def __exit__(self, *args, **kwargs):
        """Teardown test server."""
        pass


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

    def run_server(self) -> SeleniumTestServer:
        """Return instance of SeleniumTestServer"""
        return SeleniumTestServer(test_instance=self)

    def get_url(self, path):
        """Return full URL to perform selenium request."""
        return 'http://localhost:{port}{path}'.format(port=self.SERVER.port, path=path)

    def setup_class(self):
        """Setup host/port to host server."""
        super(SeleniumTest, self).setup_class(self)

        self.SERVER.host = '127.0.0.1'

        self.display_instance = Display(visible=0, size=SeleniumTest.DEFAULT_RESOLUTION)
        self.display_instance.start()
        self.selenium_instance = webdriver.Firefox()
        self.selenium_instance.delete_all_cookies()
        self.selenium_instance.implicitly_wait(1)

        self.SERVER._app.route('/SHUTDOWN')(stop_server)
        self.SERVER.port = 5001
        self._server_thread = threading.Thread(
            target=self.SERVER.run,
            kwargs={'debug': False},
            daemon=True
        )
        self._server_thread.start()

    def teardown_class(self):
        """Teardown display instance."""
        self.display_instance.stop()
        # Shutdown server
        self.SERVER._app.test_client().get('/SHUTDOWN')
        self._server_thread.join()
