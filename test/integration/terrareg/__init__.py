
import os

import pytest

from terrareg.database import Database
from terrareg.server import Server
import terrareg.config


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
