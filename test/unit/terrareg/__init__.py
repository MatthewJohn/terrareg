
from copy import deepcopy
import datetime
import functools
import secrets
import unittest.mock

import pytest

from terrareg.database import Database
from terrareg.errors import DuplicateNamespaceDisplayNameError, NamespaceAlreadyExistsError
import terrareg.models
from terrareg.server import Server
import terrareg.config
from test import BaseTest
from .test_data import test_data_full, test_git_providers, test_user_group_data_full
from terrareg.constants import EXTRACTION_VERSION


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
TEST_NAMESPACE_REDIRECTS = {}
TEST_MODULE_PROVIDER_REDIRECTS = {}
TEST_MODULE_DETAILS = {}
TEST_MODULE_DETAILS_ITX = 0
USER_GROUP_CONFIG = {}

def setup_test_data(test_data=None, user_group_data=None, namespace_redirects=None, module_provider_redirects=None):
    """Provide decorator to setup test data to be used for mocked objects."""
    def deco(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            global TEST_MODULE_DETAILS
            global TEST_MODULE_DETAILS_ITX
            global TEST_MODULE_DATA
            global USER_GROUP_CONFIG
            global TEST_NAMESPACE_REDIRECTS
            global TEST_MODULE_PROVIDER_REDIRECTS
            TEST_MODULE_DATA = deepcopy(test_data if test_data else test_data_full)
            TEST_MODULE_DETAILS = {}
            USER_GROUP_CONFIG = deepcopy(user_group_data if user_group_data else test_user_group_data_full)
            TEST_NAMESPACE_REDIRECTS = deepcopy(namespace_redirects if namespace_redirects else {})
            TEST_MODULE_PROVIDER_REDIRECTS = deepcopy(module_provider_redirects if module_provider_redirects else {})

            # Replace all ModuleDetails in test data with IDs and move contents to
            # TEST_MODULE_DETAILS
            default_readme = 'Mock module README file'
            default_terraform_docs = '{"inputs": [], "outputs": [], "providers": [], "resources": []}'
            default_tfsec = '{"results": null}'
            default_terraform_modules = ''
            default_terraform_version = '{"terraform_version": "1.4.6", "platform": "linux_amd64", "provider_selections": {"registry.terraform.io/hashicorp/random": "3.5.1"}, "terraform_outdated": false}'
            for namespace in TEST_MODULE_DATA:
                # Generate ID, as necessary
                if 'id' not in TEST_MODULE_DATA[namespace]:
                    max_id = max(*[TEST_MODULE_DATA[ns_itx].get('id', 0) for ns_itx in TEST_MODULE_DATA])
                    TEST_MODULE_DATA[namespace]['id'] = max_id + 1

                for module in TEST_MODULE_DATA[namespace].get('modules', {}):
                    for provider in TEST_MODULE_DATA[namespace]['modules'][module]:
                        for version in TEST_MODULE_DATA[namespace]['modules'][module][provider].get('versions', {}):
                            version_config = TEST_MODULE_DATA[namespace]['modules'][module][provider]['versions'][version]
                            TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
                                'readme_content': Database.encode_blob(version_config.get('readme_content', default_readme)),
                                'terraform_docs': Database.encode_blob(version_config.get('terraform_docs', default_terraform_docs)),
                                'tfsec': Database.encode_blob(version_config.get('tfsec', default_tfsec)),
                                'terraform_modules': Database.encode_blob(version_config.get('terraform_modules', default_terraform_modules)),
                                'terraform_version': Database.encode_blob(version_config.get('terraform_version', default_terraform_version)),
                            }
                            version_config['module_details_id'] = TEST_MODULE_DETAILS_ITX

                            TEST_MODULE_DETAILS_ITX += 1

                            for type_ in ['examples', 'submodules']:
                                for submodule_name in version_config.get(type_, {}):
                                    config = version_config[type_][submodule_name]
                                    TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
                                        'readme_content': Database.encode_blob(config.get('readme_content', default_readme)),
                                        'terraform_docs': Database.encode_blob(config.get('terraform_docs', default_terraform_docs)),
                                        'tfsec': Database.encode_blob(config.get('tfsec', default_tfsec)),
                                        'terraform_modules': Database.encode_blob(config.get('terraform_modules', default_terraform_modules)),
                                        'terraform_version': Database.encode_blob(config.get('terraform_version', default_terraform_version)),
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
            TEST_NAMESPACE_REDIRECTS = {}
            TEST_MODULE_PROVIDER_REDIRECTS = {}
            return res
        return wrapper
    return deco


def mock_method(request, path, mocked_method):
    """Patch a method."""
    patch = unittest.mock.patch(
        path,
        mocked_method
    )
    request.addfinalizer(lambda: patch.stop())
    patch.start()


def mock_git_provider(request):
    """Mock GitProvider class"""
    def get_all():
        """Return all mocked git provider."""
        global TEST_GIT_PROVIDER_DATA
        return [
            terrareg.models.GitProvider(git_provider_id)
            for git_provider_id in TEST_GIT_PROVIDER_DATA
        ]
    mock_method(request, 'terrareg.models.GitProvider.get_all', get_all)

    def _get_db_row(self):
        """Return mocked data for git provider."""
        global TEST_GIT_PROVIDER_DATA
        data = TEST_GIT_PROVIDER_DATA.get(self._id, None)
        data['id'] = self._id
        return data
    mock_method(request, 'terrareg.models.GitProvider._get_db_row', _get_db_row)


def get_namespace_mock_data(namespace):
    global TEST_MODULE_DATA
    return TEST_MODULE_DATA[namespace._name] if namespace._name in TEST_MODULE_DATA else None

def get_module_mock_data(module):
    return get_namespace_mock_data(module._namespace)['modules'][module._name] if module._name in get_namespace_mock_data(module._namespace)['modules'] else {}

def get_module_provider_mock_data(module_provider):
    return get_module_mock_data(module_provider._module)[module_provider._name] if module_provider._name in get_module_mock_data(module_provider._module) else {}

def get_module_version_mock_data(module_version):
    """Return unit test data structure for namespace."""
    module_provider_data = get_module_provider_mock_data(module_version._module_provider)
    return (
        module_provider_data['versions'][module_version._version]
        if ('versions' in module_provider_data and
            module_version._version in module_provider_data['versions']) else
        None
    )

def mock_module(request):
    """Mock Module class"""
    def mock_get_providers(self):
        """Return all mocked git provider."""
        return [terrareg.models.ModuleProvider(module=self, name=module_provider)
                for module_provider in get_module_mock_data(self)]

    mock_method(request, 'terrareg.models.Module.get_providers', mock_get_providers)


def mock_module_details(request):
    def create(cls):
        """Mock create method"""
        global TEST_MODULE_DETAILS_ITX
        TEST_MODULE_DETAILS[str(TEST_MODULE_DETAILS_ITX)] = {
            'readme_content': None,
            'terraform_docs': None,
            'terraform_version': None,
            'terraform_modules': None
        }

        module_details = terrareg.models.ModuleDetails(TEST_MODULE_DETAILS_ITX)
        TEST_MODULE_DETAILS_ITX += 1
        return module_details
    mock_method(request, 'terrareg.models.ModuleDetails.create', create)

    def update_attributes(self, **kwargs):
        TEST_MODULE_DETAILS[str(self._id)].update(**kwargs)
    mock_method(request, 'terrareg.models.ModuleDetails.update_attributes', update_attributes)

    def _get_db_row(self):
        return dict(TEST_MODULE_DETAILS[str(self._id)])
    mock_method(request, 'terrareg.models.ModuleDetails._get_db_row', _get_db_row)


def mock_module_version(request):
    @property
    def module_details(self):
        return terrareg.models.ModuleDetails(self._get_db_row()['module_details_id'])
    mock_method(request, 'terrareg.models.ModuleVersion.module_details', module_details)

    @property
    def module_version_files(self):
        """Return list of mocked module version files"""
        return [
            terrareg.models.ModuleVersionFile(module_version=self, path=path)
            for path in get_module_version_mock_data(self).get('files', {})
        ]
    mock_method(request, 'terrareg.models.ModuleVersion.module_version_files', module_version_files)

    def update_attributes(self, **kwargs):
        """Mock updating module version attributes"""
        get_module_version_mock_data(self).update(kwargs)
    mock_method(request, 'terrareg.models.ModuleVersion.update_attributes', update_attributes)

    def _create_db_row(self):
        """Mock create DB row"""

        module_provider_data = get_module_provider_mock_data(self._module_provider)
        previous_published = False
        if 'versions' not in module_provider_data:
            module_provider_data['versions'] = {}
        if self._version in module_provider_data['versions']:
            previous_published = module_provider_data['versions'][self._version].get('published', False)
            del module_provider_data['versions'][self._version]
        module_provider_data['versions'][self._version] = {
            'beta': False,
            'internal': False,
            'published': False
        }
        
        return previous_published
    mock_method(request, 'terrareg.models.ModuleVersion._create_db_row', _create_db_row)

    def _get_db_row(self):
        """Return mock DB row"""
        unittest_data = get_module_version_mock_data(self)
        if unittest_data is None:
            return None
        return {
            'id': unittest_data.get('id'),
            'module_provider_id': get_module_provider_mock_data(self._module_provider),
            'version': self._version,
            'owner': unittest_data.get('owner', 'Mock Owner'),
            'description': unittest_data.get('description', 'Mock description'),
            'repo_base_url_template': unittest_data.get('repo_base_url_template', None),
            'repo_clone_url_template': unittest_data.get('repo_clone_url_template', None),
            'repo_browse_url_template': unittest_data.get('repo_browse_url_template', None),
            'published_at': unittest_data.get(
                'published_at',
                datetime.datetime(year=2020, month=1, day=1,
                                  hour=23, minute=18, second=12)
            ),
            'variable_template': Database.encode_blob(unittest_data.get('variable_template', '{}')),
            'internal': unittest_data.get('internal', False),
            'published': unittest_data.get('published', False),
            'beta': unittest_data.get('beta', False),
            'module_details_id': unittest_data.get('module_details_id', None),
            'extraction_version': unittest_data.get('extraction_version', EXTRACTION_VERSION)
        }
    mock_method(request, 'terrareg.models.ModuleVersion._get_db_row', _get_db_row)


def mock_module_version_file(request):

    def update_attributes(self, *args, **kwargs):
        raise Exception("update_attributes has not been implemented")
    mock_method(request, 'terrareg.models.ModuleVersionFile.update_attributes', update_attributes)

    def _get_db_row(self):
        data = get_module_version_mock_data(self._module_version).get('files', {}).get(self._path, None)
        if data is None:
            return None
        return {
            "content": Database.encode_blob(data),
            "path": self._path
        }
    mock_method(request, "terrareg.models.ModuleVersionFile._get_db_row", _get_db_row)


def mock_module_provider_redirect(request):
    """Mock ModuleProviderRedirect class"""

    @classmethod
    def create(cls, module_provider, original_namespace, original_name, original_provider):
        """Create instance of object in database."""
        global TEST_MODULE_PROVIDER_REDIRECTS
        key_ = (original_namespace.pk, original_name, original_provider)
        if key_ in TEST_MODULE_PROVIDER_REDIRECTS:
            raise Exception('Namespace redirect already exists')

        TEST_NAMESPACE_REDIRECTS[key_] = {
            'module_provider_id': module_provider.pk
        }
    mock_method(request, 'terrareg.models.ModuleProviderRedirect.create', create)

    @classmethod
    def get_module_provider_by_original_details(cls, namespace, module, provider, case_insensitive=False):
        global TEST_MODULE_PROVIDER_REDIRECTS
        global TEST_MODULE_DATA
        key_ = (namespace.pk, module, provider)
        if key_ not in TEST_MODULE_PROVIDER_REDIRECTS:
            return None

        module_provider_id = TEST_MODULE_PROVIDER_REDIRECTS[key_].get('module_provider_id')
        if not module_provider_id:
            raise Exception('Unittest error: module_provider_id not associated with ModuleProviderRedirect')

        for namespace_name in TEST_MODULE_DATA:
            for module_name in TEST_MODULE_DATA[namespace_name].get('modules', {}):
                for provider_name in TEST_MODULE_DATA[namespace_name]['modules'][module_name]:
                    module_provider_id_itx = TEST_MODULE_DATA[namespace_name]['modules'][module_name][provider_name].get('id')
                    if module_provider_id == module_provider_id_itx:
                        return terrareg.models.ModuleProvider.get(
                            terrareg.models.Module(
                                terrareg.models.Namespace.get(name=namespace_name),
                                module_name
                            ),
                            provider_name
                        )
        return None
    mock_method(request, 'terrareg.models.ModuleProviderRedirect.get_module_provider_by_original_details', get_module_provider_by_original_details)


def mock_module_provider(request):

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

    mock_method(request, 'terrareg.models.ModuleProvider.create', create)

    def get_git_provider(self):
        """Return Mocked git provider"""
        if self._get_db_row()['git_provider_id']:
            return terrareg.models.GitProvider.get(self._get_db_row()['git_provider_id'])
        return None
    mock_method(request, 'terrareg.models.ModuleProvider.get_git_provider', get_git_provider)

    def _get_db_row(self):
        """Return fake data in DB row."""
        if self._name not in get_module_mock_data(self._module):
            return None
        data = get_module_provider_mock_data(self)
        return {
            'id': data.get('id'),
            'namespace': self._module._namespace.name,
            'module': self._module.name,
            'provider': self.name,
            'verified': data.get('verified', False),
            'repo_base_url_template': data.get('repo_base_url_template', None),
            'repo_clone_url_template': data.get('repo_clone_url_template', None),
            'repo_browse_url_template': data.get('repo_browse_url_template', None),
            'git_provider_id': data.get('git_provider_id', None),
            'git_tag_format': data.get('git_tag_format', None),
            'git_path': data.get('git_path', None)
        }
    mock_method(request, 'terrareg.models.ModuleProvider._get_db_row', _get_db_row)

    def get_latest_version(self):
        """Return mocked latest version of module"""
        data = get_module_provider_mock_data(self)
        if 'latest_version' in data and data['latest_version']:
            return terrareg.models.ModuleVersion.get(module_provider=self, version=data['latest_version'])
        return None
    mock_method(request, 'terrareg.models.ModuleProvider.get_latest_version', get_latest_version)

    def get_versions(self, include_beta=True, include_unpublished=False):
        """Return all ModuleVersion objects for ModuleProvider."""
        versions = []
        for version in get_module_provider_mock_data(self).get('versions', {}):
            version_obj = terrareg.models.ModuleVersion(module_provider=self, version=version)
            if version_obj.beta and not include_beta:
                continue
            if not version_obj.published and not include_unpublished:
                continue
            versions.append(version_obj)
        return versions
    mock_method(request, 'terrareg.models.ModuleProvider.get_versions', get_versions)

    def update_attributes(self, **kwargs):
        """Update mock data attributes"""
        get_module_provider_mock_data(self).update(kwargs)
    mock_method(request, 'terrareg.models.ModuleProvider.update_attributes', update_attributes)

def mock_namespace_redirect(request):
    """Mock NamespaceRedirect class"""

    @classmethod
    def create(cls, namespace, name):
        """Create instance of object in database."""
        global TEST_NAMESPACE_REDIRECTS
        global TEST_MODULE_DATA
        if name in TEST_NAMESPACE_REDIRECTS:
            raise Exception('Namespace redirect already exists')

        TEST_NAMESPACE_REDIRECTS[name] = {
            'namespace_id': namespace.pk
        }
    mock_method(request, 'terrareg.models.NamespaceRedirect.create', create)

    @classmethod
    def get_namespace_by_name(cls, name, case_insensitive=False):
        global TEST_NAMESPACE_REDIRECTS
        if name not in TEST_NAMESPACE_REDIRECTS:
            return None

        namespace_id = TEST_NAMESPACE_REDIRECTS[name].get('namespace_id')
        if not namespace_id:
            raise Exception('Unittest error: namespace_id not associated with NamespaceRedirect')
        for namespace_name in TEST_MODULE_DATA:
            if TEST_MODULE_DATA[namespace_name]['id'] == namespace_id:
                return terrareg.models.Namespace.get(name=namespace_name)
    mock_method(request, 'terrareg.models.NamespaceRedirect.get_namespace_by_name', get_namespace_by_name)

def mock_namespace(request):

    @classmethod
    def insert_into_database(cls, name, display_name):
        """Create namespace"""
        global TEST_MODULE_DATA
        TEST_MODULE_DATA[name] = {
            'id': len(TEST_MODULE_DATA) + 1,
            'display_name': display_name,
            'modules': {}
        }
    mock_method(request, 'terrareg.models.Namespace.insert_into_database', insert_into_database)

    @classmethod
    def get_by_case_insensitive_name(cls, name, include_redirect=True):
        """Get namespace by case-insensitive name match."""
        global TEST_MODULE_DATA
        for namespace_name_itx in TEST_MODULE_DATA:
            if namespace_name_itx.lower() == name.lower():
                return cls(name)

        # Check for redirect
        if include_redirect and (redirect_namespace := terrareg.models.NamespaceRedirect.get_namespace_by_name(
                name=name, case_insensitive=True)):
            return redirect_namespace

        return None
    mock_method(request, 'terrareg.models.Namespace.get_by_case_insensitive_name', get_by_case_insensitive_name)

    @classmethod
    def get_by_display_name(cls, display_name):
        global TEST_MODULE_DATA
        if not display_name:
            return None
        for namespace_name in TEST_MODULE_DATA:
            if TEST_MODULE_DATA[namespace_name].get('display_name') == display_name:
                return cls(namespace_name)
        return None
    mock_method(request, 'terrareg.models.Namespace.get_by_display_name', get_by_display_name)

    def update_attributes(self, **kwargs):
        """Mock update_attributes method of Namespace"""
        global TEST_MODULE_DATA
        if 'namespace' in kwargs:
            new_name = kwargs['namespace']
            TEST_MODULE_DATA[new_name] = TEST_MODULE_DATA[self._name]
            del TEST_MODULE_DATA[self._name]
            del kwargs['namespace']
            self._name = new_name

        TEST_MODULE_DATA[self._name].update(**kwargs)

        self._cache_db_row = None

    mock_method(request, 'terrareg.models.Namespace.update_attributes', update_attributes)

    def _get_db_row(self):
        mock_namespace_data = get_namespace_mock_data(self)
        if mock_namespace_data is None:
            return None
        return {
            'namespace': self._name,
            'id': mock_namespace_data['id'],
            'display_name': mock_namespace_data.get('display_name')
        }
    mock_method(request, 'terrareg.models.Namespace._get_db_row', _get_db_row)

    def get_total_count():
        """Get total number of namespaces."""
        return len(TEST_MODULE_DATA)
    mock_method(request, 'terrareg.models.Namespace.get_total_count', get_total_count)

    def get_all(only_published=False):
        """Return all namespaces."""
        valid_namespaces = []
        if only_published:
            # Iterate through all module versions of each namespace
            # to determine if the namespace has a published version
            for namespace_name in TEST_MODULE_DATA.keys():
                namespace = terrareg.models.Namespace(namespace_name)
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
            terrareg.models.Namespace(namespace)
            for namespace in valid_namespaces
        ]
    mock_method(request, 'terrareg.models.Namespace.get_all', get_all)

    def get_all_modules(self):
        """Return all modules for namespace."""
        return [
            terrareg.models.Module(namespace=self, name=n)
            for n in (TEST_MODULE_DATA[self._name]['modules'].keys()
                      if self._name in TEST_MODULE_DATA else
                      {})
        ]
    mock_method(request, 'terrareg.models.Namespace.get_all_modules', get_all_modules)

MOCK_SESSIONS = {}

def mock_session(request):
    global MOCK_SESSIONS
    # Reset mock sessions on each fixture execution
    MOCK_SESSIONS = {}

    @classmethod
    def create_session(cls):
        """Create new session object."""
        global MOCK_SESSIONS
        session_id = secrets.token_urlsafe(terrareg.models.Session.SESSION_ID_LENGTH)
        MOCK_SESSIONS[session_id] = (datetime.datetime.now() + datetime.timedelta(minutes=terrareg.config.Config().ADMIN_SESSION_EXPIRY_MINS))
        return cls(session_id=session_id)
    mock_method(request, 'terrareg.models.Session.create_session', create_session)

    @classmethod
    def cleanup_old_sessions(cls):
        """Mock cleanup old sessions"""
        pass
    mock_method(request, 'terrareg.models.Session.cleanup_old_sessions', cleanup_old_sessions)

    @classmethod
    def check_session(cls, session_id):
        """Get session object."""
        # Check session ID is not empty
        global MOCK_SESSIONS
        if not session_id:
            return None

        if MOCK_SESSIONS.get(session_id, None) and MOCK_SESSIONS[session_id] >= datetime.datetime.now():
            return cls(session_id)

        return None
    mock_method(request, 'terrareg.models.Session.check_session', check_session)

    def delete(self):
        """Delete session from database"""
        global MOCK_SESSIONS
        if self.id in MOCK_SESSIONS:
            del MOCK_SESSIONS[self.id]
    mock_method(request, 'terrareg.models.Session.delete', delete)


def mock_user_group(request):

    @classmethod
    def get_by_group_name(cls, name):
        """Obtain group by name."""
        global USER_GROUP_CONFIG
        if name in USER_GROUP_CONFIG:
            return cls(name)
        return None
    mock_method(request, 'terrareg.models.UserGroup.get_by_group_name', get_by_group_name)

    @classmethod
    def _insert_into_database(cls, name, site_admin):
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
    mock_method(request, 'terrareg.models.UserGroup._insert_into_database', _insert_into_database)

    @classmethod
    def get_all_user_groups(cls):
        """Obtain all user groups."""
        global USER_GROUP_CONFIG
        return [
            cls(user_group_name)
            for user_group_name in USER_GROUP_CONFIG
        ]
    mock_method(request, 'terrareg.models.UserGroup.get_all_user_groups', get_all_user_groups)

    def _get_db_row(self):
        """Return DB row for user group."""
        global USER_GROUP_CONFIG
        if self._name in USER_GROUP_CONFIG:
            return {
                'id': USER_GROUP_CONFIG[self._name].get('id', 100),
                'name': self._name,
                'site_admin': USER_GROUP_CONFIG[self._name].get('site_admin', False)
            }
    mock_method(request, 'terrareg.models.UserGroup._get_db_row', _get_db_row)

    def _delete_from_database(self):
        """Delete user group"""
        global USER_GROUP_CONFIG
        del USER_GROUP_CONFIG[self._name]
    mock_method(request, 'terrareg.models.UserGroup._delete_from_database', _delete_from_database)


def mock_user_group_namespace_permission(request):

    @classmethod
    def get_permissions_by_user_group(cls, user_group):
        """Return permissions by user group"""
        global USER_GROUP_CONFIG
        return [
            cls(user_group=user_group, namespace=terrareg.models.Namespace.get(name=namespace))
            for namespace in USER_GROUP_CONFIG[user_group.name].get('namespace_permissions', {})
        ]
    mock_method(request, 'terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_group', get_permissions_by_user_group)

    @classmethod
    def get_permissions_by_user_groups_and_namespace(cls, user_groups, namespace):
        """Obtain user permission by multiple user groups for a single namespace"""
        global USER_GROUP_CONFIG
        permissions = []
        for user_group in user_groups:
            if user_group.name in USER_GROUP_CONFIG and namespace.name in USER_GROUP_CONFIG[user_group.name].get('namespace_permissions', {}):
                permissions.append(terrareg.models.UserGroupNamespacePermission(user_group, namespace))

        return permissions
    mock_method(request, 'terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace', get_permissions_by_user_groups_and_namespace)

    @classmethod
    def _insert_into_database(cls, user_group, namespace, permission_type):
        """Insert user group namespace permission into DB"""
        global USER_GROUP_CONFIG
        if 'namespace_permissions' not in USER_GROUP_CONFIG[user_group.name]:
            USER_GROUP_CONFIG[user_group.name]['namespace_permissions'] = {}
        if namespace.name in USER_GROUP_CONFIG[user_group.name]['namespace_permissions']:
            raise Exception('MOCK - namepsace_permission for namespace already exists')
        USER_GROUP_CONFIG[user_group.name]['namespace_permissions'][namespace.name] = permission_type
    mock_method(request, 'terrareg.models.UserGroupNamespacePermission._insert_into_database', _insert_into_database)

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
    mock_method(request, 'terrareg.models.UserGroupNamespacePermission._get_db_row', _get_db_row)

    def _delete_from_database(self):
        """Delete user group namespace permission."""
        global USER_GROUP_CONFIG
        del USER_GROUP_CONFIG[self.user_group.name]['namespace_permissions'][self.namespace.name]
    mock_method(request, 'terrareg.models.UserGroupNamespacePermission._delete_from_database', _delete_from_database)

@pytest.fixture()
def mock_models(request):
    mock_git_provider(request)
    mock_namespace_redirect(request)
    mock_namespace(request)
    mock_module_provider(request)
    mock_module_provider_redirect(request)
    mock_module(request)
    mock_module_details(request)
    mock_module_version(request)
    mock_module_version_file(request)
    mock_session(request)
    mock_user_group(request)
    mock_user_group_namespace_permission(request)