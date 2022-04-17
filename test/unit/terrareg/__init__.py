
from unittest.mock import MagicMock

import pytest

from terrareg.server import Server


SERVER = Server()

@pytest.fixture
def client():
    """Configures the app for testing

    Sets app config variable ``TESTING`` to ``True``

    :return: App for testing
    """

    SERVER._app.config['TESTING'] = True
    client = SERVER._app.test_client()

    yield client
