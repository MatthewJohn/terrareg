
from copy import deepcopy
import datetime
import functools
import secrets
import unittest.mock

import pytest

from terrareg.database import Database
from terrareg.errors import NamespaceAlreadyExistsError
from terrareg.models import (
    GitProvider, Module, ModuleDetails,
    ModuleProvider, ModuleVersion, ModuleVersionFile, Namespace, Session, UserGroup, UserGroupNamespacePermission
)
from terrareg.server import Server
import terrareg.config
from test import BaseTest
from .test_data import test_data_full, test_git_providers, test_user_group_data_full


class TerraregUnitTest(BaseTest):

    @classmethod
    def _get_database_path(cls):
        return 'temp-unittest.db'

    @classmethod
    def _setup_test_data(cls):
        """Override setup test data method to disable any setup."""
        pass

    def setup_method(self, method):
        """Setup database"""
        # Call super method
        super(TerraregUnitTest, self).setup_method(method)

        BaseTest.INSTANCE_ = self
        terrareg.config.Config.DATABASE_URL = 'sqlite:///temp-unittest.db'

        # Create DB tables
        Database.get().get_meta().create_all(Database.get().get_engine())


TEST_MODULE_DATA = {}
TEST_GIT_PROVIDER_DATA = {}
TEST_MODULE_DETAILS = {}
TEST_MODULE_DETAILS_ITX = 0
USER_GROUP_CONFIG = {}

def setup_test_data(test_data=None, user_group_data=None):
    """Provide decorator to setup test data to be used for mocked objects."""
    def deco(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            global TEST_MODULE_DETAILS
            global TEST_MODULE_DETAILS_ITX
            global TEST_MODULE_DATA
            global USER_GROUP_CONFIG
            TEST_MODULE_DATA = dict(test_data if test_data else test_data_full)
            TEST_MODULE_DETAILS = {}
            USER_GROUP_CONFIG = deepcopy(user_group_data if user_group_data else test_user_group_data_full)

            # Replace all ModuleDetails in test data with IDs and move contents to
            # TEST_MODULE_DETAILS
            default_readme = 'Mock module README file'
            default_terraform_docs = '{"inputs": [], "outputs": [], "providers": [], "resources": []}'
            default_tfsec = '{"results": null}'
            for namespace in TEST_MODULE_DATA:
                for module in TEST_MODULE_DATA[namespace].get('modules', {}):
                    for provider in TEST_MODULE_DATA[namespace]['modules'][module]:
                        for version in TEST_MODULE_DATA[namespace]['modules'][module][provider].get('versions', {}):
                            version_config = TEST_MODULE_DATA[namespace]['modules'][module][provider]['versions'][version]
                            TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
                                'readme_content': Database.encode_blob(version_config.get('readme_content', default_readme)),
                                'terraform_docs': Database.encode_blob(version_config.get('terraform_docs', default_terraform_docs)),
                                'tfsec': Database.encode_blob(version_config.get('tfsec', default_tfsec))
                            }
                            version_config['module_details_id'] = TEST_MODULE_DETAILS_ITX

                            TEST_MODULE_DETAILS_ITX += 1

                            for type_ in ['examples', 'submodules']:
                                for submodule_name in version_config.get(type_, {}):
                                    config = version_config[type_][submodule_name]
                                    TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
                                        'readme_content': Database.encode_blob(config.get('readme_content', default_readme)),
                                        'terraform_docs': Database.encode_blob(config.get('terraform_docs', default_terraform_docs)),
                                        'tfsec': Database.encode_blob(config.get('tfsec', default_tfsec))
                                    }
                                    config['module_details_id'] = TEST_MODULE_DETAILS_ITX

                                    TEST_MODULE_DETAILS_ITX += 1

            global TEST_GIT_PROVIDER_DATA
            TEST_GIT_PROVIDER_DATA = dict(test_git_providers)
            res = func(*args, **kwargs)
            TEST_MODULE_DATA = {}
            TEST_GIT_PROVIDER_DATA = {}
            TEST_MODULE_DETAILS = {}
            TEST_MODULE_DETAILS_ITX = 0
            return res
        return wrapper
    return deco


class MockGitProvider(GitProvider):
    """Mocked GitProvider."""

    @staticmethod
    def get_all():
        """Return all mocked git provider."""
        return [
            MockGitProvider(git_provider_id)
            for git_provider_id in TEST_GIT_PROVIDER_DATA
        ]

    def _get_db_row(self):
        """Return mocked data for git provider."""
        data = TEST_GIT_PROVIDER_DATA.get(self._id, None)
        data['id'] = self._id
        return data

class MockModule(Module):
    """Mocked module."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._namespace._unittest_data['modules'][self._name] if self._name in self._namespace._unittest_data['modules'] else {}

    def get_providers(self):
        """Return list of mocked module providers"""
        return [MockModuleProvider(module=self, name=module_provider)
                for module_provider in self._unittest_data]


class MockModuleDetails(ModuleDetails):

    @classmethod
    def create(cls):
        """Mock create method"""
        global TEST_MODULE_DETAILS_ITX
        TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
            'readme_content': None,
            'terraform_docs': None
        }

        module_details = MockModuleDetails(TEST_MODULE_DETAILS_ITX)
        TEST_MODULE_DETAILS_ITX += 1
        return module_details

    def update_attributes(self, **kwargs):
        TEST_MODULE_DETAILS[str(self._id)].update(**kwargs)

    def _get_db_row(self):
        return dict(TEST_MODULE_DETAILS[str(self._id)])


class MockModuleVersion(ModuleVersion):
    """Mocked module version."""

    @property
    def module_details(self):
        return MockModuleDetails(self._get_db_row()['module_details_id'])

    @property
    def module_version_files(self):
        """Return list of mocked module version files"""
        return [
            MockModuleVersionFile(module_version=self, path=path)
            for path in self._unittest_data.get('files', {})
        ]

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
            'repo_browse_url_template': self._unittest_data.get('repo_browse_url_template', None),
            'published_at': self._unittest_data.get(
                'published_at',
                datetime.datetime(year=2020, month=1, day=1,
                                  hour=23, minute=18, second=12)
            ),
            'variable_template': Database.encode_blob(self._unittest_data.get('variable_template', '{}')),
            'internal': self._unittest_data.get('internal', False),
            'published': self._unittest_data.get('published', False),
            'beta': self._unittest_data.get('beta', False),
            'module_details_id': self._unittest_data.get('module_details_id', None)
        }


class MockModuleVersionFile(ModuleVersionFile):
    """Mocked module version file."""

    def __init__(self, *args, **kwargs):
        """Setup unittest data"""
        super(MockModuleVersionFile, self).__init__(*args, **kwargs)
        content = self._module_version._unittest_data.get('files', {}).get(self._path, None)
        self._unittest_data = None if content is None else {
            'content': Database.encode_blob(content),
            'path': self._path
        }

    def update_attributes(self, **kwargs):
        """Mock updating module version attributes"""
        self._unittest_data.update(kwargs)

    def _get_db_row(self):
        """Return mock DB row"""
        if self._unittest_data is None:
            return None
        return self._unittest_data


class MockModuleProvider(ModuleProvider):
    """Mocked module provider."""

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return self._module._unittest_data[self._name] if self._name in self._module._unittest_data else {}

    @classmethod
    def create(cls, module, name):
        """Mock version of upstream mock object"""
        global TEST_MODULE_DATA
        if not module._namespace.name in TEST_MODULE_DATA:
            raise Exception('Namespace does not exist')
        if module.name not in TEST_MODULE_DATA[module._namespace.name]['modules']:
            TEST_MODULE_DATA[module._namespace.name]['modules'][module.name] = {}
        if name not in TEST_MODULE_DATA[module._namespace.name]['modules'][module.name]:
            TEST_MODULE_DATA[module._namespace.name]['modules'][module.name][name] = {
                'id': 99,
                'latest_version': None,
                'versions': {},
                'repo_base_url_template': None,
                'repo_clone_url_template': None,
                'repo_browse_url_template': None,
                'internal': False
            }
        return cls(module=module, name=name)

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
            'repo_browse_url_template': self._unittest_data.get('repo_browse_url_template', None),
            'git_provider_id': self._unittest_data.get('git_provider_id', None),
            'git_tag_format': self._unittest_data.get('git_tag_format', None),
            'git_path': self._unittest_data.get('git_path', None)
        }

    def get_latest_version(self):
        """Return mocked latest version of module"""
        if 'latest_version' in self._unittest_data and self._unittest_data['latest_version']:
            return MockModuleVersion.get(module_provider=self, version=self._unittest_data['latest_version'])
        return None

    def get_versions(self, include_beta=True, include_unpublished=False):
        """Return all MockModuleVersion objects for ModuleProvider."""
        versions = []
        for version in self._unittest_data.get('versions', {}):
            version_obj = MockModuleVersion(module_provider=self, version=version)
            if version_obj.beta and not include_beta:
                continue
            if not version_obj.published and not include_unpublished:
                continue
            versions.append(version_obj)
        return versions

class MockNamespace(Namespace):
    """Mocked namespace."""

    @classmethod
    def create(cls, name):
        """Create namespace"""
        global TEST_MODULE_DATA
        if name in TEST_MODULE_DATA:
            raise NamespaceAlreadyExistsError('Unittest namespace already exists')
        TEST_MODULE_DATA[name] = {
            'id': len(TEST_MODULE_DATA) + 1,
            'modules': {}
        }
        return cls(name)

    @classmethod
    def get(cls, name, create=False):
        global TEST_MODULE_DATA
        if name in TEST_MODULE_DATA:
            return cls(name)
        elif create:
            return cls.create(name)
        else:
            return None

    def _get_db_row(self):
        return {
            'namespace': self._name,
            'id': self._unittest_data['id']
        }

    @staticmethod
    def get_total_count():
        """Get total number of namespaces."""
        return len(TEST_MODULE_DATA)

    @staticmethod
    def get_all(only_published=False):
        """Return all namespaces."""
        valid_namespaces = []
        if only_published:
            # Iterate through all module versions of each namespace
            # to determine if the namespace has a published version
            for namespace_name in TEST_MODULE_DATA.keys():
                namespace = MockNamespace(namespace_name)
                for module in namespace.get_all_modules():
                    for provider in module.get_providers():
                        for version in provider.get_versions():
                            if (namespace_name not in valid_namespaces and
                                    version.published and
                                    version.beta == False):
                                valid_namespaces.append(namespace_name)
        else:
            valid_namespaces = TEST_MODULE_DATA.keys()

        return [
            MockNamespace(namespace)
            for namespace in valid_namespaces
        ]

    def get_all_modules(self):
        """Return all modules for namespace."""
        return [
            MockModule(namespace=self, name=n)
            for n in (TEST_MODULE_DATA[self._name]['modules'].keys()
                      if self._name in TEST_MODULE_DATA else
                      {})
        ]

    @property
    def _unittest_data(self):
        """Return unit test data structure for namespace."""
        return TEST_MODULE_DATA[self._name] if self._name in TEST_MODULE_DATA else {}


class MockSession(Session):

    MOCK_SESSIONS = {}

    @classmethod
    def create_session(cls):
        """Create new session object."""
        session_id = secrets.token_urlsafe(Session.SESSION_ID_LENGTH)
        cls.MOCK_SESSIONS[session_id] = (datetime.datetime.now() + datetime.timedelta(minutes=terrareg.config.Config().ADMIN_SESSION_EXPIRY_MINS))
        return cls(session_id=session_id)

    @classmethod
    def cleanup_old_sessions(cls):
        """Mock cleanup old sessions"""
        pass

    @classmethod
    def check_session(cls, session_id):
        """Get session object."""
        # Check session ID is not empty
        if not session_id:
            return None

        if cls.MOCK_SESSIONS.get(session_id, None) and cls.MOCK_SESSIONS[session_id] >= datetime.datetime.now():
            return cls(session_id)

        return None

    def delete(self):
        """Delete session from database"""
        if self.id in MockSession.MOCK_SESSIONS:
            del MockSession.MOCK_SESSIONS[self.id]


class MockUserGroup(UserGroup):

    @classmethod
    def get_by_group_name(cls, name):
        """Obtain group by name."""
        global USER_GROUP_CONFIG
        if name in USER_GROUP_CONFIG:
            return MockUserGroup(name)
        return None

    @classmethod
    def _insert_into_db(cls, name, site_admin):
        """Insert usergroup into DB"""
        global USER_GROUP_CONFIG
        if name in USER_GROUP_CONFIG:
            # Should not hit this exception
            raise Exception('MOCK USER GROUP ALREAY EXISTS')
        USER_GROUP_CONFIG[name] = {
            'id': 200,
            'site_admin': site_admin,
            'namespace_permissions': {}
        }

    @classmethod
    def get_all_user_groups(cls):
        """Obtain all user groups."""
        global USER_GROUP_CONFIG
        return [
            cls(user_group_name)
            for user_group_name in USER_GROUP_CONFIG
        ]

    def _get_db_row(self):
        """Return DB row for user group."""
        global USER_GROUP_CONFIG
        if self._name in USER_GROUP_CONFIG:
            return {
                'id': USER_GROUP_CONFIG[self._name].get('id', 100),
                'name': self._name,
                'site_admin': USER_GROUP_CONFIG[self._name].get('site_admin', False)
            }

    def delete(self):
        """Delete user group"""
        global USER_GROUP_CONFIG
        del USER_GROUP_CONFIG[self._name]


class MockUserGroupNamespacePermission(UserGroupNamespacePermission):

    @classmethod
    def get_permissions_by_user_group(cls, user_group):
        """Return permissions by user group"""
        global USER_GROUP_CONFIG
        return [
            cls(user_group=user_group, namespace=MockNamespace.get(name=namespace))
            for namespace in USER_GROUP_CONFIG[user_group.name].get('namespace_permissions', {})
        ]

    @classmethod
    def get_permissions_by_user_groups_and_namespace(cls, user_groups, namespace):
        """Obtain user permission by multiple user groups for a single namespace"""
        global USER_GROUP_CONFIG
        permissions = []
        for user_group in user_groups:
            if user_group.name in USER_GROUP_CONFIG and namespace.name in USER_GROUP_CONFIG[user_group.name].get('namespace_permissions', {}):
                permissions.append(MockUserGroupNamespacePermission(user_group, namespace))

        return permissions

    @classmethod
    def _insert_into_database(cls, user_group, namespace, permission_type):
        """Insert user group namespace permission into DB"""
        global USER_GROUP_CONFIG
        if 'namespace_permissions' not in USER_GROUP_CONFIG[user_group.name]:
            USER_GROUP_CONFIG[user_group.name]['namespace_permissions'] = {}
        if namespace.name in USER_GROUP_CONFIG[user_group.name]['namespace_permissions']:
            raise Exception('MOCK - namepsace_permission for namespace already exists')
        USER_GROUP_CONFIG[user_group.name]['namespace_permissions'][namespace.name] = permission_type

    def _get_db_row(self):
        """Return DB row for user group."""
        global USER_GROUP_CONFIG
        if self._user_group.name in USER_GROUP_CONFIG and self._namespace.name in USER_GROUP_CONFIG[self._user_group.name].get('namespace_permissions', {}):
            return {
                'namespace_id': self._namespace.pk,
                'user_group_id': self._user_group.pk,
                'permission_type': USER_GROUP_CONFIG[self._user_group.name]['namespace_permissions'][self._namespace.name]
            }
        return None

    def delete(self):
        """Delete user group namespace permission."""
        raise NotImplementedError


def mock_server_user_groups(request):
    """Mock UserGroup and UserGroupNamespacePermission classes"""
    user_group_patch = unittest.mock.patch('terrareg.server.UserGroup', MockUserGroup)
    user_group_namespace_permission_patch = unittest.mock.patch(
        'terrareg.server.UserGroupNamespacePermission',
        MockUserGroupNamespacePermission)

    def cleanup_mock():
        user_group_namespace_permission_patch.stop()
        user_group_patch.stop()
    request.addfinalizer(cleanup_mock)
    user_group_patch.start()
    user_group_namespace_permission_patch.start()

@pytest.fixture()
def mock_server_user_groups_fixture(request):
    """Mock UserGroup and UserGroupNamespacePermission as fixture"""
    mock_server_user_groups(request)


def mocked_server_module_version(request):
    """Mock server ModuleVersion class."""
    patch = unittest.mock.patch('terrareg.server.ModuleVersion', MockModuleVersion)
    module_version_file_patch = unittest.mock.patch('terrareg.server.ModuleVersionFile', MockModuleVersionFile)

    def cleanup_mock():
        module_version_file_patch.stop()
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()
    module_version_file_patch.start()


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


def mocked_server_session(request):
    """Mock Session model class in server module."""
    patch = unittest.mock.patch('terrareg.server.Session', MockSession)

    def cleanup_mock():
        patch.stop()
    request.addfinalizer(cleanup_mock)
    patch.start()

    mocked_server_module(request)


@pytest.fixture()
def mocked_server_session_fixture(request):
    """Mock namespace as fixture."""
    mocked_server_session(request)
