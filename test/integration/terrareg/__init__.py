
import os

import pytest

from terrareg.database import Database
from terrareg.server import Server


class TerraregIntegrationTest:

    def setup_method(self, method):
        """Setup database"""
        Database.SQLITE_DB_PATH = 'temp-integration.db'

        # Remove any pre-existing database files
        if os.path.isfile(Database.SQLITE_DB_PATH):
            os.unlink(Database.SQLITE_DB_PATH)

        Database.reset()
        self.SERVER = Server()

        self.SERVER._app.config['TESTING'] = True
