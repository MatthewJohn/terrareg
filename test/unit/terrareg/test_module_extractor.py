
import subprocess
from unittest.main import MODULE_EXAMPLES
import unittest.mock

import pytest

from test.unit.terrareg import (
    MockNamespace, MockModule, MockModuleProvider,
    MockModuleVersion, TerraregUnitTest, setup_test_data
)
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import GitModuleExtractor


class TestGitModuleExtractor(TerraregUnitTest):
    """Test GitModuleExtractor class."""

    @pytest.mark.parametrize('module_provider_name,expected_git_url,expected_git_tag', [
        ('staticrepourl', 'git@localhost:7999/bla/test-module.git', 'v4.3.2'),
        ('placeholdercloneurl', 'git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git', 'v4.3.2'),
        ('usesgitprovider', 'localhost.com/moduleextraction/gitextraction-usesgitprovider', 'v4.3.2'),
        ('nogittagformat', 'localhost.com/moduleextraction/gitextraction-nogittagformat', '4.3.2'),
        ('complexgittagformat', 'localhost.com/moduleextraction/gitextraction-complexgittagformat', 'unittest4.3.2value')
    ])
    @setup_test_data()
    def test__clone_repository(self, module_provider_name, expected_git_url, expected_git_tag):
        """Test _clone_repository method"""
        namespace = MockNamespace(name='moduleextraction')
        module = MockModule(namespace=namespace, name='gitextraction')
        module_provider = MockModuleProvider(module=module, name=module_provider_name)
        module_version = MockModuleVersion(module_provider=module_provider, version='4.3.2')

        check_call_mock = unittest.mock.MagicMock()
        module_extractor = GitModuleExtractor(module_version=module_version)
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                me._clone_repository()

        check_call_mock.assert_called_with([
            'git', 'clone',
            '--single-branch',
            '--branch', expected_git_tag,
            expected_git_url,
            module_extractor.extract_directory],
            stderr=subprocess.STDOUT,
            env=unittest.mock.ANY)
        assert check_call_mock.call_args.kwargs['env']['GIT_SSH_COMMAND'] == 'ssh -o StrictHostKeyChecking=accept-new'
