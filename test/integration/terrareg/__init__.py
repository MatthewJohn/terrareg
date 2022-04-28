
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

        self.SERVER._app.config['TESTING'] = True
