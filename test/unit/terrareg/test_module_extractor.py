
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

    @pytest.mark.parametrize('readme,expected_description', [
        # Expect empty README produces None description
        (None, None),
        ('', None),

        # Check with empty lines
        ('\n  \n\t\n     \n\n', None),

        # Test with invalid sentences
        ('This line is too longThis line is too longThis line is too longThis line is '
         'This line is too longThis line is too longThis line is too longThis line is ',
         None),
        ('ContainsOneWord', None),
        ('Contains too few words', None),
        # Not enough characters
        ('N E C F T A G', None),

        # Test invalid words
        ('This is a good description from https://example.com check it', None),
        ('This is a good description from http://example.com check it', None),
        ('Do not forget to email me at test@example.com', None),

        # Test skipping short scentence and picking valid one
        ('Too short\nThis is a good description for a module.', 'This is a good description for a module.'),

        # Test stripping of characters
        ('    A terraform module to create a cluster!   ', 'A terraform module to create a cluster!'),

        # Test combining first scentences
        ('This is short. This is a little bit longer.', 'This is short. This is a little bit longer.'),
        ('This is short. This is a bit longer. This is way waywaywaywaywaywaywaywaywaywaywaywaywaywaywaywaywaywaywayway too long',
         'This is short. This is a bit longer'),

        # Complete tests
        ('WSA EQW Cluster\n\nDownload from https://example.com/mymodule.git\nTerraform module to create an WSA EQW Cluster. This module is written by ExampleCorp.\n',
        'Terraform module to create an WSA EQW Cluster'),
        ('WSA DJW Terraform module\n\nTerraform module which creates DJW resources on WSA.\n\nUsage\nResource...',
        'Terraform module which creates DJW resources on WSA.'),
        ('Terraform module designed to generate consistent helps and blanks for resources. Use this module to implement a test blank helps.\n\nThere are 6 inputs considered "helps" or "tests" (because the tests are used to construct the ID):',
        'Terraform module designed to generate consistent helps and blanks for resources')
    ])
    def test_description_extraction(self, readme, expected_description):
        """Test description extraction from README."""
        module_extractor = GitModuleExtractor(module_version=None)
        assert module_extractor._extract_description(readme) == expected_description

    def test_description_extraction_when_disabled(self):
        """Test that description extraction returns None when description extraction is disabled in the config"""
        test_text = "This is a perfectly valid description"
        module_extractor = GitModuleExtractor(module_version=None)
        assert module_extractor._extract_description(test_text) == test_text

        with unittest.mock.patch('terrareg.config.Config.AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION', False):
            assert module_extractor._extract_description(test_text) == None
