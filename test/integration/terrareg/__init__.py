
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
from .test_data import integration_test_data, integration_git_providers


class TerraregIntegrationTest:

    def setup_method(self, method):
        """Setup database"""
        terrareg.config.DATABASE_URL = 'sqlite:///temp-integration.db'

        # Remove any pre-existing database files
        if os.path.isfile('temp-integration.db'):
            os.unlink('temp-integration.db')

        Database.reset()
        self.SERVER = Server()

        # Create DB tables
        Database.get().get_meta().create_all(Database.get().get_engine())

        self.SERVER._app.config['TESTING'] = True

def setup_test_data(test_data=None):
    """Provide decorator to setup test data to be used by tests."""
    def deco(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):

            # Setup test git providers
            for git_provider_id in integration_git_providers:
                insert = Database.get().git_provider.insert().values(
                    id=git_provider_id,
                    **integration_git_providers[git_provider_id]
                )
                with Database.get_engine().connect() as conn:
                    conn.execute(insert)

            # Setup test Namespaces, Modules, ModuleProvider and ModuleVersion
            import_data = integration_test_data if test_data is None else test_data

            # Iterate through namespaces
            for namespace_name in import_data:
                namespace = Namespace(name=namespace_name)
                namespace_data = import_data[namespace_name]

                # Iterate through modules
                for module_name in namespace_data:
                    module = Module(namespace=namespace, name=module_name)
                    module_data = namespace_data[module_name]

                    # Iterate through providers
                    for provider_name in import_data[namespace_name][module_name]:
                        module_provider_test_data = module_data[provider_name]

                        # Update provided test attributes
                        module_provider_attributes = {}
                        for attr in module_provider_test_data:
                            if attr not in ['latest_version', 'versions']:
                                module_provider_attributes[attr] = module_provider_test_data[attr]

                        insert = Database.get().module_provider.insert().values(**module_provider_attributes)
                        with Database.get_engine().connect() as conn:
                            conn.execute(insert)

            res = func(*args, **kwargs)
            return res
        return wrapper
    return deco
