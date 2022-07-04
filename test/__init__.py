
from datetime import datetime
import functools
import os
import unittest.mock

import pytest

from terrareg.models import (
    Example, ExampleFile, ModuleDetails, Namespace, Module, ModuleProvider,
    ModuleVersion, GitProvider, Submodule
)
from terrareg.database import Database
from terrareg.server import Server, require_admin_authentication
import terrareg.config


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


class BaseTest:

    _TEST_DATA = None
    _GIT_PROVIDER_DATA = None

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
        cls.SERVER = Server()

        # Create DB tables
        Database.get().get_meta().create_all(Database.get().get_engine())

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
    def _setup_test_data(cls, test_data=None):
        """Setup test data in database"""
        # Delete any pre-existing data
        db = Database.get()
        with Database.get_engine().connect() as conn:
            conn.execute(db.sub_module.delete())
            conn.execute(db.module_version.delete())
            conn.execute(db.module_provider.delete())
            conn.execute(db.git_provider.delete())
            conn.execute(db.analytics.delete())

        # Setup test git providers
        for git_provider_id in cls._GIT_PROVIDER_DATA:
            insert = Database.get().git_provider.insert().values(
                id=git_provider_id,
                **cls._GIT_PROVIDER_DATA[git_provider_id]
            )
            with Database.get_engine().connect() as conn:
                conn.execute(insert)

        # Setup test Namespaces, Modules, ModuleProvider and ModuleVersion
        import_data = cls._TEST_DATA if test_data is None else test_data

        # Iterate through namespaces
        for namespace_name in import_data:
            namespace_data = import_data[namespace_name]
            namespace = Namespace(name=namespace_name)

            # Iterate through modules
            for module_name in namespace_data:
                module_data = namespace_data[module_name]
                module = Module(namespace=namespace, name=module_name)

                # Iterate through providers
                for provider_name in import_data[namespace_name][module_name]:
                    module_provider_test_data = module_data[provider_name]
                    module_provider = ModuleProvider(module=module, name=provider_name)

                    # Update provided test attributes
                    module_provider_attributes = {
                        'namespace': namespace_name,
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
                            terraform_docs=version_data.get('terraform_docs', None)
                        )

                        data = {
                            'module_provider_id': module_provider_attributes['id'],
                            'version': version_number,
                            # Default beta flag to false
                            'beta': False,
                            'published_at': datetime.now(),
                            'internal': False,
                            'module_details_id': module_details.pk
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
                            if attr not in ['examples', 'submodules', 'published', 'readme_content', 'terraform_docs']
                        }
                        if values_to_update:
                            module_version.update_attributes(**values_to_update)

                        # If module version is published, do so
                        if version_data.get('published', False):
                            module_version.publish()

                        # Iterate over submodules and create them
                        for submodule_path in version_data.get('submodules', {}):
                            submodule_config = version_data['submodules'][submodule_path]

                            module_details = ModuleDetails.create()
                            module_details.update_attributes(
                                readme_content=submodule_config.get('readme_content', None),
                                terraform_docs=submodule_config.get('terraform_docs', None)
                            )

                            submodule = Submodule.create(module_version=module_version, module_path=submodule_path)
                            attributes_to_update = {
                                attr: submodule_config[attr]
                                for attr in submodule_config
                                if attr not in ['readme_content', 'terraform_docs']
                            }
                            if attributes_to_update:
                                submodule.update_attributes(
                                    module_details_id=module_details.pk,
                                    **attributes_to_update
                                )

                        # Iterate over examples and create them
                        for example_path in version_data.get('examples', {}):
                            example_config = version_data['examples'][example_path]

                            module_details = ModuleDetails.create()
                            module_details.update_attributes(
                                readme_content=example_config.get('readme_content', None),
                                terraform_docs=example_config.get('terraform_docs', None)
                            )

                            example = Example.create(module_version=module_version, module_path=example_path)
                            attributes_to_update = {
                                attr: example_config[attr]
                                for attr in example_config
                                if attr not in ['example_files', 'readme_content', 'terraform_docs']
                            }
                            if attributes_to_update:
                                example.update_attributes(
                                    module_details_id=module_details.pk,
                                    **attributes_to_update
                                )

                            for example_file_path in example_config.get('example_files', {}):
                                example_file = ExampleFile.create(example=example, path=example_file_path)
                                example_file.update_attributes(content=example_config['example_files'][example_file_path])

