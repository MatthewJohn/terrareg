
import datetime
import functools
import unittest.mock

import pytest

from terrareg.database import Database
from terrareg.models import GitProvider, Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.server import Server
import terrareg.config
from .test_data import test_data_full, test_git_providers


class TerraregUnitTest:

    INSTANCE_ = None

    @classmethod
    def get(cls):
        """Get instance of class"""
        return cls.INSTANCE_

    def setup_method(self, method):
        """Setup database"""
        TerraregUnitTest.INSTANCE_ = self
        Database.reset()
        self.SERVER = Server()
        terrareg.config.DATABASE_URL = 'sqlite:///temp-unittest.db'
        self.SERVER._app.config['TESTING'] = True


@pytest.fixture
def client():
    """Return test client"""
    client = TerraregUnitTest.get().SERVER._app.test_client()

    yield client

@pytest.fixture
def test_request_context():
    """Return test request context"""
    return TerraregUnitTest.get().SERVER._app.test_request_context()

@pytest.fixture
def app_context():
    """Return test request context"""
    return TerraregUnitTest.get().SERVER._app.app_context()

TEST_MODULE_DATA = {}
TEST_GIT_PROVIDER_DATA = {}

def setup_test_data(test_data=None):
    """Provide decorator to setup test data to be used for mocked objects."""
    def deco(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            global TEST_MODULE_DATA
            TEST_MODULE_DATA = dict(test_data if test_data else test_data_full)
            global TEST_GIT_PROVIDER_DATA
            TEST_GIT_PROVIDER_DATA = dict(test_git_providers)
            res = func(*args, **kwargs)
            TEST_MODULE_DATA = {}
            TEST_GIT_PROVIDER_DATA = {}
            return res
        return wrapper
    return deco


class MockGitProvider(GitProvider):
    """Mocked GitProvider."""

    def _get_db_row(self):
        """Return mocked data for git provider."""
        return TEST_GIT_PROVIDER_DATA.get(self._id, None)

class MockModule(Module):
    """Mocked module."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._namespace._unittest_data[self._name] if self._name in self._namespace._unittest_data else {}

    def get_providers(self):
        """Return list of mocked module providers"""
        return [MockModuleProvider(module=self, name=module_provider)
                for module_provider in self._unittest_data]


class MockModuleVersion(ModuleVersion):
    """Mocked module version."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return (
            self._module_provider._unittest_data['versions'][self._version]
            if ('versions' in self._module_provider._unittest_data and
                self._version in self._module_provider._unittest_data['versions']) else
            None
        )

    def update_attributes(self, **kwargs):
        """Mock updating module version attributes"""
        self._unittest_data.update(kwargs)

    def _get_db_row(self):
        """Return mock DB row"""
        if self._unittest_data is None:
            return None
        return {
            'id': self._unittest_data.get('id'),
            'module_provider_id': self._module_provider._unittest_data['id'],
            'version': self._version,
            'owner': self._unittest_data.get('owner', 'Mock Owner'),
            'description': self._unittest_data.get('description', 'Mock description'),
            'repo_base_url_template': self._unittest_data.get('repo_base_url_template', None),
            'repo_clone_url_template': self._unittest_data.get('repo_clone_url_template', None),
            'repo_browse_url_template': self._unittest_data.get('repo_clone_url_template', None),
            'published_at': self._unittest_data.get(
                'published_at',
                datetime.datetime(year=2020, month=1, day=1,
                                  hour=23, minute=18, second=12)
            ),
            'readme_content': self._unittest_data.get('readme_content', 'Mock module README file'),
            'module_details': self._unittest_data.get(
                'module_details',
                '{"inputs": [], "outputs": [], "providers": [], "resources": []}'
            ),
            'variable_template': self._unittest_data.get('variable_template', '{}'),
            'artifact_location': self._unittest_data.get('artifact_location', None)
        }


class MockModuleProvider(ModuleProvider):
    """Mocked module provider."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._module._unittest_data[self._name] if self._name in self._module._unittest_data else {}

    @classmethod
    def _create(cls, module, name):
        """Mock version of upstream mock object"""
        if not module._namespace.name in TEST_MODULE_DATA:
            TEST_MODULE_DATA[module._namespace.name] = {}
        if module.name not in TEST_MODULE_DATA[module._namespace.name]:
            TEST_MODULE_DATA[module._namespace.name][module.name] = {}
        if name not in TEST_MODULE_DATA[module._namespace.name][module.name]:
            TEST_MODULE_DATA[module._namespace.name][module.name][name] = {
                'id': 99,
                'latest_version': None,
                'versions': {},
                'repo_base_url_template': None,
                'repo_clone_url_template': None,
                'repo_browse_url_template': None
            }

    def get_git_provider(self):
        """Return Mocked git provider"""
        if self._get_db_row()['git_provider_id']:
            return MockGitProvider.get(self._get_db_row()['git_provider_id'])
        return None

    def _get_db_row(self):
        """Return fake data in DB row."""
        if self._name not in self._module._unittest_data:
            return None
        return {
            'id': self._unittest_data.get('id'),
            'namespace': self._module._namespace.name,
            'module': self._module.name,
            'provider': self.name,
            'verified': self._unittest_data.get('verified', False),
            'repo_base_url_template': self._unittest_data.get('repo_base_url_template', None),
            'repo_clone_url_template': self._unittest_data.get('repo_clone_url_template', None),
            'repo_browse_url_template': self._unittest_data.get('repo_clone_url_template', None),
            'git_provider_id': self._unittest_data.get('git_provider_id', None),
            'git_tag_format': self._unittest_data.get('git_tag_format', None)
        }

    def get_latest_version(self):
        """Return mocked latest version of module"""
        if 'latest_version' in self._unittest_data:
            return MockModuleVersion.get(module_provider=self, version=self._unittest_data['latest_version'])
        return None

    def get_versions(self):
        """Return all MockModuleVersion objects for ModuleProvider."""
        return [MockModuleVersion(module_provider=self, version=version)
                for version in self._unittest_data['versions']]

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
        return TEST_MODULE_DATA[self._name] if self._name in TEST_MODULE_DATA else {}


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

