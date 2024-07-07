
from datetime import datetime
import unittest.mock
import pytest
import sqlalchemy
from terrareg.analytics import AnalyticsEngine
from terrareg.database import Database

from terrareg.models import Example, ExampleFile, Module, Namespace, ModuleProvider, ModuleVersion
import terrareg.errors
from test.integration.terrareg import TerraregIntegrationTest

class CommonBaseSubmodule(TerraregIntegrationTest):

    SUBMODULE_CLASS = None

    @pytest.mark.parametrize('provider_git_path,module_path,expected_path', [
        # Various methods of provider git path being root
        (None, 'path/to/module', 'path/to/module'),
        ('', 'path/to/module', 'path/to/module'),
        ('/', 'path/to/module', 'path/to/module'),
        ('.', 'path/to/module', 'path/to/module'),
        ('./', 'path/to/module', 'path/to/module'),

        ('subsubdir', 'path/to/module', 'subsubdir/path/to/module'),
        ('./subsubdir', 'path/to/module', 'subsubdir/path/to/module'),
        ('./subsubdir/', 'path/to/module', 'subsubdir/path/to/module'),
        ('subsubdir/', 'path/to/module', 'subsubdir/path/to/module'),

        ('multiple/directories/in/', 'path/to/module', 'multiple/directories/in/path/to/module'),
    ])
    def test_git_path(self, provider_git_path, module_path, expected_path):
        provider = ModuleProvider.get(Module(Namespace('moduledetails'), 'git-path'), 'provider')
        version = ModuleVersion.get(provider, '1.0.0')
        version.update_attributes(git_path=provider_git_path)
        submodule = self.SUBMODULE_CLASS.create(version, module_path)

        try:
            assert submodule.git_path == expected_path
        finally:
            submodule.delete()
