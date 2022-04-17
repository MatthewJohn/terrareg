
import datetime

import pytest

from terrareg.models import ModuleProvider, ModuleVersion
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


class MockModuleVersion(ModuleVersion):
    """Mocked module version."""

    def _get_db_row(self):
        """Return mock DB row"""
        return {
            'id': 1,
            'module_provider_id': 1,
            'version': self._version,
            'owner': 'Mock Owner',
            'description': 'Mock description',
            'source': 'http://mock.example.com/mockmodule',
            'published_at': datetime.datetime(year=2020, month=1, day=1, hour=23, minute=18, second=12),
            'readme_contents': 'Mock module README file',
            'module_details': '{}',
            'variable_template': '{}',
            'verified': False,
            'artifact_location': None
        }


class MockModuleProvider(ModuleProvider):
    """Mocked module provider."""

    MOCK_LATEST_VERSION_NUMBER = None

    def get_latest_version(self):
        """Return mocked latest version of module"""
        return MockModuleVersion(module_provider=self, version=self.MOCK_LATEST_VERSION_NUMBER)
