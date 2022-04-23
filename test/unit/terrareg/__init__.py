
import datetime
import unittest.mock

import pytest

from terrareg.database import Database
from terrareg.models import Module, ModuleProvider, ModuleVersion
from terrareg.server import Server

Database.SQLITE_DB_PATH = 'temp-unittest.db'
SERVER = Server()
SERVER._app.config['TESTING'] = True


@pytest.fixture
def client():
    """Return test client"""
    client = SERVER._app.test_client()

    yield client

@pytest.fixture
def test_request_context():
    """Return test request context"""
    return SERVER._app.test_request_context()

class MockModule(Module):
    """Mocked module."""

    MOCK_MODULE_PROVIDERS = ['testprovider']

    def get_providers(self):
        """Return list of mocked module providers"""

        # Add custom name for non-existent module
        if self._name == 'unittestdoesnotexist':
            return []

        return [MockModuleProvider(module=self, name=module_provider)
                for module_provider in self.MOCK_MODULE_PROVIDERS]


class MockModuleVersion(ModuleVersion):
    """Mocked module version."""

    def _get_db_row(self):
        """Return mock DB row"""
        # Return no data for non-existent version
        if self._version == '0.1.2':
            return None

        return {
            'id': 1,
            'module_provider_id': 1,
            'version': self._version,
            'owner': 'Mock Owner',
            'description': 'Mock description',
            'source': 'http://mock.example.com/mockmodule',
            'published_at': datetime.datetime(year=2020, month=1, day=1,
                                              hour=23, minute=18, second=12),
            'readme_content': 'Mock module README file',
            'module_details': '{"inputs": [], "outputs": [], "providers": [], "resources": []}',
            'variable_template': '{}',
            'verified': False,
            'artifact_location': None
        }


class MockModuleProvider(ModuleProvider):
    """Mocked module provider."""

    MOCK_LATEST_VERSION_NUMBER = '1.0.0'

    def get_latest_version(self):
        """Return mocked latest version of module"""
        # Handle fake non-existent module
        if self._name == 'unittestproviderdoesnotexist':
            return None

        return MockModuleVersion(module_provider=self, version=self.MOCK_LATEST_VERSION_NUMBER)


def mocked_server_module_version(request):
    """Mock server ModuleVersion class."""
    patch = unittest.mock.patch('terrareg.server.ModuleVersion', MockModuleVersion)

    def cleanup_mock():
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()


@pytest.fixture()
def mocked_server_module_version_fixture(request):
    """Mock module version as fixture."""
    mocked_server_module_version(request)


def mocked_server_module_provider(request):
    """Mock server ModuleProvider class."""
    patch = unittest.mock.patch('terrareg.server.ModuleProvider', MockModuleProvider)

    def cleanup_mock():
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()
    mocked_server_module_version(request)


@pytest.fixture()
def mocked_server_module_provider_fixture(request):
    """Mock module provider as fixture."""
    mocked_server_module_provider(request)


def mocked_server_module(request):
    """Mock server Module class."""
    patch = unittest.mock.patch('terrareg.server.Module', MockModule)

    def cleanup_mock():
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()

    mocked_server_module_provider(request)


@pytest.fixture()
def mocked_server_module_fixture(request):
    """Mock module as fixture."""
    mocked_server_module(request)
