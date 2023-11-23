
from datetime import datetime
import functools
import json
import os
import unittest.mock

import pytest

import terrareg.models
from terrareg.models import (
    Example, ExampleFile, ModuleDetails, ModuleVersionFile, Namespace, Module, ModuleProvider,
    ModuleVersion, GitProvider, Submodule, UserGroup, UserGroupNamespacePermission
)
from terrareg.database import Database
from terrareg.server import Server
import terrareg.config
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.constants import EXTRACTION_VERSION
import terrareg.provider_category_model
import terrareg.provider_source.factory
import terrareg.repository_model
import terrareg.provider_version_binary_model
import terrareg.provider_version_documentation_model
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.provider_tier


@pytest.fixture
def client():
    """Return test client"""
    client = BaseTest.get().SERVER._app.test_client()

    yield client

@pytest.fixture
def test_request_context():
    """Return test request context"""
    return BaseTest.get().SERVER._app.test_request_context()

@pytest.fixture
def app_context():
    """Return test request context"""
    return BaseTest.get().SERVER._app.app_context()

@pytest.fixture
def mock_create_audit_event():
    """Mock create audit event when modifying objects outside of request context"""
    return unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event')


class BaseTest:

    _TEST_DATA = None
    _GIT_PROVIDER_DATA = None
    _USER_GROUP_DATA = None
    _PROVIDER_SOURCES = []
    _PROVIDER_CATEGORIES = []

    INSTANCE_ = None

    @staticmethod
    def get():
        """Get current test class."""
        return BaseTest.INSTANCE_

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        raise NotImplementedError

    @classmethod
    def setup_class(cls):
        """Setup database"""
        # Setup current test object as
        # property of base class
        BaseTest.INSTANCE_ = cls

        database_url = os.environ.get('INTEGRATION_DATABASE_URL', 'sqlite:///{}'.format(cls._get_database_path()))
        cls.database_config_url_mock = unittest.mock.patch('terrareg.config.Config.DATABASE_URL', database_url)
        cls.database_config_url_mock.start()

        # Remove any pre-existing database files
        if os.path.isfile(cls._get_database_path()):
            os.unlink(cls._get_database_path())

        Database.reset()

        Database.get().initialise()

        # Create DB tables
        Database.get().get_meta().create_all(Database.get().get_engine())

        Database.reset()

        cls.SERVER = Server()

        cls._setup_test_data()

        cls.SERVER._app.config['TESTING'] = True

    @classmethod
    def teardown_class(cls):
        """Empty method for inheritting classes to call super method."""
        cls.SERVER = None
        cls.database_config_url_mock.stop()

    def setup_method(self, method):
        """Empty method for inheritting classes to call super method."""
        pass

    def teardown_method(self, method):
        """Empty method for inheritting classes to call super method."""
        pass

    @classmethod
    def _patch_audit_event_creation(cls):
        """Return context manager for ignoring event creation"""
        return unittest.mock.patch('terrareg.audit.AuditEvent.create_audit_event')

    @classmethod
    def _setup_test_data(cls, test_data=None):
        """Setup test data in database"""
        # Delete any pre-existing data
        db = Database.get()
        with Database.get_engine().connect() as conn:
            conn.execute(db.audit_history.delete())
            conn.execute(db.user_group_namespace_permission.delete())
            conn.execute(db.user_group.delete())
            conn.execute(db.sub_module.delete())
            conn.execute(db.module_version_file.delete())
            conn.execute(db.module_version.delete())
            conn.execute(db.module_provider.delete())
            conn.execute(db.example_file.delete())
            conn.execute(db.module_details.delete())
            conn.execute(db.git_provider.delete())
            conn.execute(db.analytics.delete())
            conn.execute(db.provider_analytics.delete())
            conn.execute(db.provider_version_binary.delete())
            conn.execute(db.provider_version_documentation.delete())
            conn.execute(db.provider_version.delete())
            conn.execute(db.provider.delete())
            conn.execute(db.provider_source.delete())
            conn.execute(db.provider_category.delete())
            conn.execute(db.session.delete())
            conn.execute(db.namespace.delete())

        with cls._patch_audit_event_creation():

            # Setup test git providers
            for git_provider_id in cls._GIT_PROVIDER_DATA:
                insert = Database.get().git_provider.insert().values(
                    id=git_provider_id,
                    **cls._GIT_PROVIDER_DATA[git_provider_id]
                )
                with Database.get_engine().connect() as conn:
                    conn.execute(insert)

            with unittest.mock.patch('terrareg.config.Config.PROVIDER_CATEGORIES', json.dumps(cls._PROVIDER_CATEGORIES)):
                terrareg.provider_category_model.ProviderCategoryFactory.get().initialise_from_config()

            with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES', json.dumps(cls._PROVIDER_SOURCES)):
                terrareg.provider_source.factory.ProviderSourceFactory.get().initialise_from_config()

            # Setup test Namespaces, Modules, ModuleProvider and ModuleVersion
            import_data = cls._TEST_DATA if test_data is None else test_data

            # Iterate through namespaces
            for namespace_name in import_data:
                namespace_data = import_data[namespace_name]
                display_name = import_data[namespace_name].get("display_name")
                namespace = Namespace.create(name=namespace_name, display_name=display_name)

                # Iterate through modules
                for module_name, module_data in namespace_data.get("modules", {}).items():
                    module = Module(namespace=namespace, name=module_name)

                    # Iterate through providers
                    for provider_name in module_data:
                        module_provider_test_data = module_data[provider_name]
                        module_provider = ModuleProvider(module=module, name=provider_name)

                        # Update provided test attributes
                        module_provider_attributes = {
                            'namespace_id': namespace.pk,
                            'module': module_name,
                            'provider': provider_name
                        }
                        for attr in module_provider_test_data:
                            if attr not in ['latest_version', 'versions']:
                                module_provider_attributes[attr] = module_provider_test_data[attr]

                        insert = Database.get().module_provider.insert().values(
                            **module_provider_attributes
                        )
                        with Database.get_engine().connect() as conn:
                            res = conn.execute(insert)

                        # Insert module versions
                        for version_number in (
                                module_provider_test_data['versions']
                                if 'versions' in module_provider_test_data else
                                []):
                            version_data = module_provider_test_data['versions'][version_number]

                            module_details = ModuleDetails.create()
                            module_details.update_attributes(
                                readme_content=version_data.get('readme_content', None),
                                terraform_docs=version_data.get('terraform_docs', None),
                                tfsec=version_data.get('tfsec'),
                                terraform_graph=version_data.get("terraform_graph", None)
                            )

                            data = {
                                'module_provider_id': module_provider_attributes['id'],
                                'version': version_number,
                                # Default beta flag to false
                                'beta': False,
                                'published_at': datetime.now(),
                                'internal': False,
                                'module_details_id': module_details.pk,
                                'extraction_version': version_data.get('extraction_version', EXTRACTION_VERSION)
                            }

                            insert = Database.get().module_version.insert().values(
                                **data
                            )
                            with Database.get_engine().connect() as conn:
                                conn.execute(insert)

                            module_version = ModuleVersion(module_provider=module_provider, version=version_number)

                            values_to_update = {
                                attr: version_data[attr]
                                for attr in version_data
                                if attr not in ['examples', 'submodules', 'published',
                                                'readme_content', 'terraform_docs', 'tfsec',
                                                'infracost', 'files', 'terraform_graph']
                            }
                            if values_to_update:
                                module_version.update_attributes(**values_to_update)

                            # If module version is published, do so
                            if version_data.get('published', False):
                                module_version.publish()

                            # Iterate over module version files
                            for file_path, content in version_data.get('files', {}).items():
                                module_version_file = ModuleVersionFile.create(module_version=module_version, path=file_path)
                                module_version_file.update_attributes(content=content)

                            # Iterate over submodules and create them
                            for submodule_path in version_data.get('submodules', {}):
                                submodule_config = version_data['submodules'][submodule_path]

                                module_details = ModuleDetails.create()
                                module_details.update_attributes(
                                    readme_content=submodule_config.get('readme_content', None),
                                    terraform_docs=submodule_config.get('terraform_docs', None),
                                    tfsec=submodule_config.get('tfsec', None),
                                    terraform_graph=submodule_config.get("terraform_graph", None)
                                )

                                submodule = Submodule.create(module_version=module_version, module_path=submodule_path)
                                attributes_to_update = {
                                    attr: submodule_config[attr]
                                    for attr in submodule_config
                                    if attr not in ['readme_content', 'terraform_docs', 'tfsec', 'terraform_graph']
                                }
                                attributes_to_update['module_details_id'] = module_details.pk
                                submodule.update_attributes(
                                    **attributes_to_update
                                )

                            # Iterate over examples and create them
                            for example_path in version_data.get('examples', {}):
                                example_config = version_data['examples'][example_path]

                                module_details = ModuleDetails.create()
                                module_details.update_attributes(
                                    readme_content=example_config.get('readme_content', None),
                                    terraform_docs=example_config.get('terraform_docs', None),
                                    tfsec=example_config.get('tfsec', None),
                                    infracost=example_config.get('infracost', None),
                                    terraform_graph=example_config.get("terraform_graph", None)
                                )

                                example = Example.create(module_version=module_version, module_path=example_path)
                                attributes_to_update = {
                                    attr: example_config[attr]
                                    for attr in example_config
                                    if attr not in ['example_files', 'readme_content', 'terraform_docs',
                                                    'tfsec', 'infracost', 'terraform_graph']
                                }
                                attributes_to_update['module_details_id'] = module_details.pk
                                example.update_attributes(
                                    **attributes_to_update
                                )

                                for example_file_path in example_config.get('example_files', {}):
                                    example_file = ExampleFile.create(example=example, path=example_file_path)
                                    example_file.update_attributes(content=example_config['example_files'][example_file_path])

                # Iterate through GPG keys
                for gpg_key in namespace_data.get("gpg_keys", []):
                    terrareg.models.GpgKey.create(namespace=namespace, ascii_armor=gpg_key.get("ascii_armor"))

                # Iterate through providers
                for provider_name, provider_data in namespace_data.get("providers", {}).items():
                    repository_data = provider_data.get("repository")
                    repository = terrareg.repository_model.Repository.create(
                        provider_source=terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(repository_data["provider_source"]),
                        provider_id=repository_data.get("provider_id"),
                        name=repository_data.get("name"),
                        description=repository_data.get("description"),
                        owner=repository_data.get("owner"),
                        clone_url=repository_data.get("clone_url"),
                        logo_url=repository_data.get("logo_url"),
                    )

                    provider = terrareg.provider_model.Provider.create(
                        repository=repository,
                        provider_category=terrareg.provider_category_model.ProviderCategoryFactory.get().get_provider_category_by_slug(provider_data.get("category_slug")),
                        use_default_provider_source_auth=provider_data.get("use_default_provider_source_auth", True),
                        tier=terrareg.provider_tier.ProviderTier(provider_data.get("tier", "community"))
                    )

                    for version, version_data in provider_data.get("versions").items():
                        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider, version=version)
                        with version_obj.create_extraction_wrapper(
                                git_tag=version_data.get("git_tag"),
                                gpg_key=terrareg.models.GpgKey.get_by_fingerprint(fingerprint=version_data.get("gpg_key_fingerprint"))):
                            pass

                        # Update any custom attributes from test data
                        update_kwargs = {
                            attr: version_data[attr]
                            for attr in ["published_at"]
                            if attr in version_data
                        }
                        if update_kwargs:
                            version_obj.update_attributes(**update_kwargs)

                        # Import binaries
                        for binary_name, binary_data in version_data.get("binaries", {}).items():
                            terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                                provider_version=version_obj,
                                name=binary_name,
                                checksum=binary_data.get("checksum"),
                                content=binary_data.get("content")
                            )

                        # Import documentation
                        for documentation_id, documentation_data in version_data.get("documentation", {}).items():
                            provider_documentation = terrareg.provider_version_documentation_model.ProviderVersionDocumentation.create(
                                provider_version=version_obj,
                                documentation_type=documentation_data.get("type"),
                                name=documentation_id[0],
                                title=documentation_data.get("title", ""),
                                description=documentation_data.get("description", ""),
                                filename=documentation_data.get("filename"),
                                language=documentation_id[1],
                                subcategory=documentation_data.get("subcategory"),
                                content=documentation_data.get("content")
                            )
                            attributes_to_update = {
                                k: v
                                for k, v in documentation_data.items()
                                if k in ["id"]
                            }
                            if attributes_to_update:
                                with db.get_connection() as conn:
                                    conn.execute(db.provider_version_documentation.update().where(db.provider_version_documentation.c.id==provider_documentation.pk).values(**attributes_to_update))


            if cls._USER_GROUP_DATA:
                for group_name in cls._USER_GROUP_DATA:
                    user_group = UserGroup.create(name=group_name, site_admin=cls._USER_GROUP_DATA[group_name].get('site_admin', False))
                    for namespace_name, permission_type in cls._USER_GROUP_DATA[group_name].get('namespace_permissions', {}).items():
                        namespace = Namespace.get(namespace_name)
                        UserGroupNamespacePermission.create(
                            user_group=user_group,
                            namespace=namespace,
                            permission_type=UserGroupNamespacePermissionType(permission_type))

    def _test_unauthenticated_read_api_endpoint_test(self, request_callback):
        """Check unauthenticated read API endpoint access"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_access_read_api = unittest.mock.MagicMock(return_value=False)
        mock_auth_method.get_username.return_value = 'unauthenticated user'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = request_callback()
            assert res.status_code == 403
            assert res.json == {
                'message': "You don't have the permission to access the requested resource. It is either read-protected or not readable by the server."
            }

            mock_auth_method.can_access_read_api.assert_called_once_with()

    def _test_unauthenticated_terraform_api_endpoint_test(self, request_callback):
        """Check unauthenticated Terraform API endpoint access"""
        mock_auth_method = unittest.mock.MagicMock()
        mock_auth_method.can_access_terraform_api = unittest.mock.MagicMock(return_value=False)
        mock_auth_method.get_username.return_value = 'unauthenticated user'
        mock_get_current_auth_method = unittest.mock.MagicMock(return_value=mock_auth_method)

        with unittest.mock.patch('terrareg.auth.AuthFactory.get_current_auth_method', mock_get_current_auth_method):
            res = request_callback()
            assert res.status_code == 403
            assert res.json == {
                'message': "You don't have the permission to access the requested resource. It is either read-protected or not readable by the server."
            }

            mock_auth_method.can_access_terraform_api.assert_called_once_with()
