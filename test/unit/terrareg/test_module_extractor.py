
from unittest.main import MODULE_EXAMPLES
import unittest.mock

import pytest

from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import GitModuleExtractor


class TestGitModuleExtractor:
    """Test GitModuleExtractor class."""

    @pytest.mark.parametrize(
        'git_url,sanitised_url',
        [
            ('test://domain.com/example.git', 'test://domain.com/example.git'),
            ('https://domain.com/example.git?ref=this.that', 'https://domain.com/example.git?ref=this.that'),
            ('https:// rm -rf /', 'https://%20rm%20-rf%20/'),
            ('echo this; echo that', 'echo%20this%3B%20echo%20that')
        ])
    def test_git_url(self, git_url, sanitised_url):
        """Test escaping of git URL"""
        module_extractor = GitModuleExtractor(git_url=git_url, tag_name='', module_version=None)
        assert module_extractor._git_url == sanitised_url

    def test__clone_repository(self):
        """Test _clone_repository method"""
        namespace = Namespace(name='testnamespace')
        module = Module(namespace=namespace, name='testmodule')
        module_provider = ModuleProvider(module=module, name='testprovider')
        module_version = ModuleVersion(module_provider=module_provider, version='4.3.2')

        check_call_mock = unittest.mock.MagicMock()
        module_extractor = GitModuleExtractor(git_url='ssh://example.com/repo.git', tag_name='v4.3.2', module_version=module_version)
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_call', check_call_mock):
            with module_extractor as me:
                me._clone_repository()

        check_call_mock.assert_called_with([
            'git', 'clone',
            '--single-branch',
            '--branch', 'v4.3.2',
            'ssh://example.com/repo.git',
            module_extractor.extract_directory],
            env=unittest.mock.ANY)
        assert check_call_mock.call_args.kwargs['env']['GIT_SSH_COMMAND'] == 'ssh -o StrictHostKeyChecking=accept-new'
