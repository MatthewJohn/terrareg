
import os
import subprocess
from unittest.main import MODULE_EXAMPLES
import unittest.mock

import pytest

import terrareg.errors

from test.unit.terrareg import (
    TerraregUnitTest, setup_test_data,
    mock_models
)
from terrareg.module_extractor import GitModuleExtractor, ModuleExtractor
import terrareg.models


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
    def test__clone_repository(self, module_provider_name, expected_git_url, expected_git_tag, mock_models):
        """Test _clone_repository method"""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name=module_provider_name)
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

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
            env=unittest.mock.ANY,
            timeout=300)
        assert check_call_mock.call_args.kwargs['env']['GIT_SSH_COMMAND'] == 'ssh -o StrictHostKeyChecking=accept-new'

    @setup_test_data()
    def test_known_git_error(self, mock_models):
        """Test error thrown by git with expected format of error."""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name='staticrepourl')
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

        module_extractor = GitModuleExtractor(module_version=module_version)

        check_call_mock = unittest.mock.MagicMock()
        test_error = subprocess.CalledProcessError(returncode=1, cmd=[])
        test_error.output = 'Preceeding line\nfatal: unittest error here\nend of output'.encode('ascii')
        check_call_mock.side_effect = test_error

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                with pytest.raises(terrareg.errors.GitCloneError) as error:
                    me._clone_repository()

                assert str(error.value) == 'Error occurred during git clone: fatal: unittest error here'

    @setup_test_data()
    def test_unknown_git_error(self, mock_models):
        """Test error thrown by git with expected format of error."""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name='staticrepourl')
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

        module_extractor = GitModuleExtractor(module_version=module_version)

        check_call_mock = unittest.mock.MagicMock()
        test_error = subprocess.CalledProcessError(returncode=1, cmd=[])
        test_error.output = 'Preceeding line\nnot a recognised output\nend of output'.encode('ascii')
        check_call_mock.side_effect = test_error

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                with pytest.raises(terrareg.errors.GitCloneError) as error:
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
    def test_description_extraction(self, readme, expected_description, mock_models):
        """Test description extraction from README."""
        module_extractor = GitModuleExtractor(module_version=None)
        assert module_extractor._extract_description(readme) == expected_description

    def test_description_extraction_when_disabled(self, mock_models):
        """Test that description extraction returns None when description extraction is disabled in the config"""
        test_text = "This is a perfectly valid description"
        module_extractor = GitModuleExtractor(module_version=None)
        assert module_extractor._extract_description(test_text) == test_text

        with unittest.mock.patch('terrareg.config.Config.AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION', False):
            assert module_extractor._extract_description(test_text) == None

    def test_run_tf_init(self):
        pass

    def test_switch_terraform_versions(self):
        """Test switching terraform versions."""
        module_extractor = GitModuleExtractor(module_version=None)
        mock_lock = unittest.mock.MagicMock()
        mock_lock.acquire.return_value = True

        with unittest.mock.patch('terrareg.module_extractor.ModuleExtractor.TERRAFORM_LOCK', mock_lock), \
                unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', unittest.mock.MagicMock()) as check_output_mock, \
                unittest.mock.patch('terrareg.config.Config.DEFAULT_TERRAFORM_VERSION', 'unittest-tf-version'), \
                unittest.mock.patch('terrareg.config.Config.TERRAFORM_ARCHIVE_MIRROR', 'https://localhost-archive/mirror/terraform'):

            with module_extractor._switch_terraform_versions(module_path='/tmp/mock-patch/to/module'):
                pass

            mock_lock.acquire.assert_called_once_with(blocking=True, timeout=60)
            expected_env = os.environ.copy()
            expected_env['TF_VERSION'] = "unittest-tf-version"
            check_output_mock.assert_called_once_with(
                ["tfswitch", "--mirror", "https://localhost-archive/mirror/terraform", "--bin", f"{os.getcwd()}/bin/terraform"],
                env=expected_env,
                cwd="/tmp/mock-patch/to/module"
            )

    def test_switch_terraform_versions_error(self):
        """Test running switch_terraform_version with erorr in tfswitch"""
        module_extractor = GitModuleExtractor(module_version=None)
        mock_lock = unittest.mock.MagicMock()
        mock_lock.acquire.return_value = True
        def raise_exception(*args, **kwargs):
            raise subprocess.CalledProcessError(cmd="test", returncode=2)

        with unittest.mock.patch('terrareg.module_extractor.ModuleExtractor.TERRAFORM_LOCK', mock_lock), \
                unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', unittest.mock.MagicMock(side_effect=raise_exception)) as check_output_mock, \
                unittest.mock.patch('terrareg.config.Config.DEFAULT_TERRAFORM_VERSION', 'unittest-tf-version'):

            with pytest.raises(terrareg.errors.TerraformVersionSwitchError):
                with module_extractor._switch_terraform_versions(module_path='/tmp/mock-patch/to/module'):
                    pass

    def test_switch_terraform_versions_with_lock(self):
        """Test switching terraform versions whilst Terraform locked is already aquired."""
        module_extractor = GitModuleExtractor(module_version=None)
        mock_lock = unittest.mock.MagicMock()
        mock_lock.acquire.return_value = False

        with unittest.mock.patch('terrareg.module_extractor.ModuleExtractor.TERRAFORM_LOCK', mock_lock), \
                unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', unittest.mock.MagicMock()) as check_output_mock:
            
            with pytest.raises(terrareg.errors.UnableToGetGlobalTerraformLockError):
                with module_extractor._switch_terraform_versions(module_path='test'):
                    pass

            mock_lock.acquire.assert_called_once_with(blocking=True, timeout=60)
            check_output_mock.assert_not_called()

    def test_get_graph_data(self):
        """Test call to terraform graph to generate graph output"""
        module_extractor = GitModuleExtractor(module_version=None)

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output',
                                 unittest.mock.MagicMock(return_value="Output graph data".encode("utf-8"))) as mock_check_output:

            module_extractor._get_graph_data(module_path='/tmp/mock-patch/to/module')

            mock_check_output.assert_called_once_with(
                [os.getcwd() + '/bin/terraform', 'graph'],
                cwd='/tmp/mock-patch/to/module'
            )

    def test_get_graph_data_terraform_error(self):
        """Test call to terraform graph with error thrown"""
        module_extractor = GitModuleExtractor(module_version=None)

        def raise_error(*args, **kwargs):
            raise subprocess.CalledProcessError(cmd="the command", returncode=2)

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output',
                                 unittest.mock.MagicMock(side_effect=raise_error)) as mock_check_output:

            with pytest.raises(terrareg.errors.UnableToProcessTerraformError) as exc:
                module_extractor._get_graph_data(module_path='/tmp/mock-patch/to/module')

            mock_check_output.assert_called_once_with(
                [os.getcwd() + '/bin/terraform', 'graph'],
                cwd='/tmp/mock-patch/to/module'
            )
