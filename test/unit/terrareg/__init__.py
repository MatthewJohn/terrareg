
import datetime
import functools
import unittest.mock

import pytest

from terrareg.database import Database
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
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

@pytest.fixture
def app_context():
    """Return test request context"""
    return SERVER._app.app_context()

TEST_MODULE_DATA = {
    'testnamespace': {
        'test_module': {
            'test_provider': {
                'id': 1,
                'repository_url': '',
                'latest_version': '',
                'versions': {
                    '1.2.3': {
                        'id': 1
                    }
                }
            }
        }
    }
}

def setup_test_data(test_data):
    """Provide decorator to setup test data to be used for mocked objects."""
    def deco(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            TEST_MODULE_DATA = test_data
            res = func(*args, **kwargs)
            TEST_MODULE_DATA = {}
            return res
        return wrapper
    return deco


class MockModule(Module):
    """Mocked module."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._namespace._unittest_data[self._name]

    def get_providers(self):
        """Return list of mocked module providers"""
        return [MockModuleProvider(module=self, name=module_provider)
                for module_provider in self._unittest_data]


class MockModuleVersion(ModuleVersion):
    """Mocked module version."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._module_provider._unittest_data['versions'][self._version]

    def _get_db_row(self):
        """Return mock DB row"""
        return {
            'id': self._unittest_data.get('id'),
            'module_provider_id': self._module_provider.pk,
            'version': self._version,
            'owner': self._unittest_data.get('owner', None),
            'description': self._unittest_data.get('description', None),
            'source': self._unittest_data.get('source', None),
            'published_at': self._unittest_data.get(
                'published_at',
                datetime.datetime(year=2020, month=1, day=1,
                                  hour=23, minute=18, second=12)
            ),
            'readme_content': self._unittest_data.get('readme_content', None),
            'module_details': self._unittest_data.get(
                'module_details',
                '{"inputs": [], "outputs": [], "providers": [], "resources": []}'
            ),
            'variable_template': self._unittest_data.get('variable_template', '{}'),
            'verified': self._unittest_data.get('verified', False),
            'artifact_location': self._unittest_data.get('artifact_location', None)
        }


class MockModuleProvider(ModuleProvider):
    """Mocked module provider."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._module._unittest_data[self._name]

    def _get_db_row(self):
        """Return fake data in DB row."""
        if self._name not in self._module._unittest_data:
            return None
        return {
            'id': self._unittest_data.get('id'),
            'namespace': self._module._namespace.name,
            'module': self._module.name,
            'provider': self.name,
            'repository_url': self._unittest_data.get('repository_url', None)
        }

    def get_latest_version(self):
        """Return mocked latest version of module"""
        if 'latest_version' in self._unittest_data:
            return MockModuleVersion.get(module_provider=self, version=self._unittest_data['latest_version'])
        return None

class MockNamespace(Namespace):
    """Mocked namespace."""

    @staticmethod
    def get_total_count():
        """Get total number of namespaces."""
        return len(TEST_MODULE_DATA)

    @staticmethod
    def get_all():
        """Return all namespaces."""
        return TEST_MODULE_DATA.keys()

    def get_all_modules(self):
        """Return all modules for namespace."""
        return [
            MockModule(namespace=self, name=n)
            for n in TEST_MODULE_DATA[self._name].keys()
        ]

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return TEST_MODULE_DATA[self._name]


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


def mocked_server_namespace(request):
    """Mock server Module class."""
    patch = unittest.mock.patch('terrareg.server.Namespace', MockNamespace)

    def cleanup_mock():
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()

    mocked_server_module(request)


@pytest.fixture()
def mocked_server_namespace_fixture(request):
    """Mock namespace as fixture."""
    mocked_server_namespace(request)

