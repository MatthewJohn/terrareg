
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
from test import BaseTest
from .test_data import integration_test_data, integration_git_providers


class TerraregIntegrationTest(BaseTest):

    _TEST_DATA = integration_test_data
    _GIT_PROVIDER_DATA = integration_git_providers

    @staticmethod
    def _get_database_path():
        """Return path of database file to use."""
        return 'temp-integration.db'
