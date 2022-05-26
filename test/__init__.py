
from datetime import datetime
import functools
import os

import pytest

from terrareg.models import (
    Namespace, Module, ModuleProvider,
    ModuleVersion, GitProvider
)
from terrareg.database import Database
from terrareg.server import Server
import terrareg.config


class BaseTest:

    _TEST_DATA = None
    _GIT_PROVIDER_DATA = None

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        raise NotImplementedError

    def setup_class(self):
        """Setup database"""
        database_url = os.environ.get('INTEGRATION_DATABASE_URL', 'sqlite:///{}'.format(self._get_database_path()))
        terrareg.config.Config.DATABASE_URL = database_url

        # Remove any pre-existing database files
        if os.path.isfile(self._get_database_path()):
            os.unlink(self._get_database_path())

        Database.reset()
        self.SERVER = Server()

        # Create DB tables
        Database.get().get_meta().create_all(Database.get().get_engine())

        self._setup_test_data()

        self.SERVER._app.config['TESTING'] = True

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

            # Iterate through modules
            for module_name in namespace_data:
                module_data = namespace_data[module_name]

                # Iterate through providers
                for provider_name in import_data[namespace_name][module_name]:
                    module_provider_test_data = module_data[provider_name]

                    # Update provided test attributes
                    module_provider_attributes = {
                        'namespace': namespace_name,
                        'module': module_name,
                        'provider': provider_name,
                        'internal': False
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
                    for module_version in (
                            module_provider_test_data['versions']
                            if 'versions' in module_provider_test_data else
                            []):
                        data = {
                            'module_provider_id': module_provider_attributes['id'],
                            'version': module_version,
                            # Default beta flag to false
                            'beta': False,
                            'published_at': datetime.now()
                        }
                        # Update column values from test data
                        data.update(module_provider_test_data['versions'][module_version])

                        insert = Database.get().module_version.insert().values(
                            **data
                        )
                        with Database.get_engine().connect() as conn:
                            conn.execute(insert)
