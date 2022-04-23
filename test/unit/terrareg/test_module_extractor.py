
from unittest.main import MODULE_EXAMPLES
import unittest.mock

import pytest

from test.unit.terrareg import (
    MockNamespace, MockModule, MockModuleProvider,
    MockModuleVersion, setup_test_data
)
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import GitModuleExtractor


class TestGitModuleExtractor:
    """Test GitModuleExtractor class."""

    @setup_test_data()
    def test__clone_repository(self):
        """Test _clone_repository method"""
        namespace = MockNamespace(name='moduleextraction')
        module = MockModule(namespace=namespace, name='test-module')
        module_provider = MockModuleProvider(module=module, name='testprovider')
        module_version = MockModuleVersion(module_provider=module_provider, version='4.3.2')

        check_call_mock = unittest.mock.MagicMock()
        module_extractor = GitModuleExtractor(module_version=module_version)
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_call', check_call_mock):
            with module_extractor as me:
                me._clone_repository()

        check_call_mock.assert_called_with([
            'git', 'clone',
            '--single-branch',
            '--branch', '4.3.2',
            'ssh://example.com/repo.git',
            module_extractor.extract_directory],
            env=unittest.mock.ANY)
        assert check_call_mock.call_args.kwargs['env']['GIT_SSH_COMMAND'] == 'ssh -o StrictHostKeyChecking=accept-new'
