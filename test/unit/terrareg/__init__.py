
import datetime
import unittest.mock

import pytest

from terrareg.models import Module, ModuleProvider, ModuleVersion
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


class MockModule(Module):
    """Mocked module."""

    MOCK_MODULE_PROVIDERS = ['testprovider']

    def get_providers(self):
        """Return list of mocked module providers"""
        return [MockModuleProvider(module=self, name=module_provider)
                for module_provider in self.MOCK_MODULE_PROVIDERS]


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

    MOCK_LATEST_VERSION_NUMBER = '1.0.0'

    def get_latest_version(self):
        """Return mocked latest version of module"""
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
