
import subprocess
from unittest.main import MODULE_EXAMPLES
import unittest.mock

import pytest
from terrareg.errors import GitCloneError

from test.unit.terrareg import (
    MockNamespace, MockModule, MockModuleProvider,
    MockModuleVersion, TerraregUnitTest, setup_test_data
)
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import GitModuleExtractor


class TestGitModuleExtractor(TerraregUnitTest):
    """Test GitModuleExtractor class."""

    @pytest.mark.parametrize('module_provider_name,expected_git_url,expected_git_tag', [
        ('staticrepourl', 'ssh://git@localhost:7999/bla/test-module.git', 'v4.3.2'),
        ('placeholdercloneurl', 'ssh://git@localhost:7999/moduleextraction/gitextraction-placeholdercloneurl.git', 'v4.3.2'),
        ('usesgitprovider', 'ssh://localhost.com/moduleextraction/gitextraction-usesgitprovider', 'v4.3.2'),
        ('nogittagformat', 'ssh://localhost.com/moduleextraction/gitextraction-nogittagformat', '4.3.2'),
        ('complexgittagformat', 'ssh://localhost.com/moduleextraction/gitextraction-complexgittagformat', 'unittest4.3.2value')
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

    @setup_test_data()
    def test_known_git_error(self):
        """Test error thrown by git with expected format of error."""
        namespace = MockNamespace(name='moduleextraction')
        module = MockModule(namespace=namespace, name='gitextraction')
        module_provider = MockModuleProvider(module=module, name='staticrepourl')
        module_version = MockModuleVersion(module_provider=module_provider, version='4.3.2')

        module_extractor = GitModuleExtractor(module_version=module_version)

        check_call_mock = unittest.mock.MagicMock()
        test_error = subprocess.CalledProcessError(returncode=1, cmd=[])
        test_error.output = 'Preceeding line\nfatal: unittest error here\nend of output'.encode('ascii')
        check_call_mock.side_effect = test_error

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                with pytest.raises(GitCloneError) as error:
                    me._clone_repository()

                assert str(error.value) == 'Error occurred during git clone: fatal: unittest error here'

    @setup_test_data()
    def test_unknown_git_error(self):
        """Test error thrown by git with expected format of error."""
        namespace = MockNamespace(name='moduleextraction')
        module = MockModule(namespace=namespace, name='gitextraction')
        module_provider = MockModuleProvider(module=module, name='staticrepourl')
        module_version = MockModuleVersion(module_provider=module_provider, version='4.3.2')

        module_extractor = GitModuleExtractor(module_version=module_version)

        check_call_mock = unittest.mock.MagicMock()
        test_error = subprocess.CalledProcessError(returncode=1, cmd=[])
        test_error.output = 'Preceeding line\nnot a recognised output\nend of output'.encode('ascii')
        check_call_mock.side_effect = test_error

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                with pytest.raises(GitCloneError) as error:
                    me._clone_repository()

                assert str(error.value) == 'Unknown error occurred during git clone'
