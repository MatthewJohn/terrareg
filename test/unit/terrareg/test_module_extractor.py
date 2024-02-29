
import os
import shutil
import subprocess
import tempfile
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

    @setup_test_data()
    def test__get_git_commit_sha(self, mock_models):
        """Test _get_git_commit_sha method"""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name='gitcommithash')
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

        module_extractor = GitModuleExtractor(module_version=module_version)
        with module_extractor as me, \
                tempfile.TemporaryDirectory() as temp_dir:

            # Create git repo and commit file
            subprocess.check_output(["git", "init"], cwd=temp_dir)
            with open(os.path.join(temp_dir, "test_file"), "w") as fh:
                pass
            subprocess.check_output(["git", "add", "test_file"], cwd=temp_dir)
            commit_output = subprocess.check_output(["git", "commit", "-m", "commit"], cwd=temp_dir).decode('utf-8')
            # Get partial commit ID from commit output
            git_sha_compare = commit_output.split(' ')[2][:-1]

            # Run _get_git_commit_sha and ensure the hash is the correct
            # length and it starts with the hash returned from the commit
            git_commit_output = me._get_git_commit_sha(temp_dir)
            assert isinstance(git_commit_output, str)
            assert len(git_commit_output) == 40
            assert git_commit_output[:7] == git_sha_compare

    @setup_test_data()
    def test__get_git_commit_sha_mock(self, mock_models):
        """Test _get_git_commit_sha method"""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name='gitcommithash')
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

        check_call_mock = unittest.mock.MagicMock(return_value=b"d94485323894790b52c16558b3fe1650542038bd")
        module_extractor = GitModuleExtractor(module_version=module_version)
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                assert me._get_git_commit_sha("/tmp/some-dir/of/module") == "d94485323894790b52c16558b3fe1650542038bd"

        check_call_mock.assert_called_with([
            'git', 'rev-parse', 'HEAD'],
            cwd="/tmp/some-dir/of/module"
        )

    @setup_test_data()
    def test__get_git_commit_sha_failure(self, mock_models):
        """Test _get_git_commit_sha method"""
        namespace = terrareg.models.Namespace(name='moduleextraction')
        module = terrareg.models.Module(namespace=namespace, name='gitextraction')
        module_provider = terrareg.models.ModuleProvider(module=module, name='gitcommithash')
        module_version = terrareg.models.ModuleVersion(module_provider=module_provider, version='4.3.2')

        test_error = subprocess.CalledProcessError(returncode=1, cmd=[])
        test_error.output = 'Preceeding line\nnot a recognised output\nend of output'.encode('ascii')
        check_call_mock = unittest.mock.MagicMock(side_effect=test_error)
        module_extractor = GitModuleExtractor(module_version=module_version)
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output', check_call_mock):
            with module_extractor as me:
                with pytest.raises(terrareg.errors.UnableToProcessTerraformError):
                    me._get_git_commit_sha("/tmp/some-dir/of/module")

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
        """Test running terraform init"""
        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_call', unittest.mock.MagicMock()) as check_output_mock, \
                unittest.mock.patch("terrareg.module_extractor.ModuleExtractor._create_terraform_rc_file", unittest.mock.MagicMock()) as mock_create_terraform_rc_file:

            module_extractor = GitModuleExtractor(module_version=None)

            assert module_extractor._run_tf_init(module_path='/tmp/mock-patch/to/module') is True

            check_output_mock.assert_called_once_with(
                [os.path.join(os.getcwd(), 'bin', 'terraform'), 'init'],
                cwd='/tmp/mock-patch/to/module'
            )
            mock_create_terraform_rc_file.assert_called_once_with()

    def test_run_tf_init_error(self):
        """Test running terraform init with error returned"""

        def raise_exception(*args, **kwargs):
            raise subprocess.CalledProcessError(cmd="test", returncode=2)

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_call', unittest.mock.MagicMock(side_effect=raise_exception)) as mock_check_call, \
                unittest.mock.patch("terrareg.module_extractor.ModuleExtractor._create_terraform_rc_file", unittest.mock.MagicMock()) as mock_create_terraform_rc_file:

            module_extractor = GitModuleExtractor(module_version=None)

            assert module_extractor._run_tf_init(module_path='/tmp/mock-patch/to/module') is False

            mock_check_call.assert_called_once_with(
                [os.path.join(os.getcwd(), 'bin', 'terraform'), 'init'],
                cwd='/tmp/mock-patch/to/module'
            )
            mock_create_terraform_rc_file.assert_called_once_with()

    @pytest.mark.parametrize('file_contents,expected_backend_file', [
        (
            {
                "state.tf": """
terraform {
  backend "s3" {
    bucket  = "does-not-exist"
    key     = "path/to/my/key"
    region  = "us-east-1"
    profile = "thisdoesnotexistforterrareg"
  }
}
"""
            },
            "state_override.tf"
        ),
        (
            {
                # Files named B, S, V, so that the backend is stored in
                # a file that will not be found first by glob
                "bucket.tf": """
resource "aws_s3_bucket" "test_bucket" {
  name = var.name
}
""",
                # More complex terraform block, with comment and required providers
                "state.tf": """
# With content before the terraform block
terraform {
  # Multiple terraform backend configurations
  required_providers {
    aws = {
      version = ">= 2.7.0"
      source = "hashicorp/aws"
    }
  }

  backend "s3" {
    bucket  = "does-not-exist"
    key     = "path/to/my/key"
    region  = "us-east-1"
    profile = "thisdoesnotexistforterrareg"
  }
}
""",
                "variables.tf": """
variable "name" {
  description = "Bucket name"
  type        = string
}
"""
            },
            "state_override.tf"
        ),
        # Without backend config
        (
            {
                # Files named B, S, V, so that the backend is stored in
                # a file that will not be found first by glob
                "bucket.tf": """
resource "aws_s3_bucket" "test_bucket" {
  name = var.name
}
""",
                "variables.tf": """
variable "name" {
  description = "Bucket name"
  type        = string
}
"""
            },
            None
        )
    ])
    def test_override_tf_backend(self, file_contents, expected_backend_file):
        """Test _override_tf_backend method with backend in terraform files"""
        temp_dir = tempfile.mkdtemp()
        try:
            for file_name in file_contents:
                with open(os.path.join(temp_dir, file_name), "w") as fh:
                    fh.write(file_contents[file_name])

            module_extractor = GitModuleExtractor(module_version=None)
            backend_file = module_extractor._override_tf_backend(module_path=temp_dir)

            if not expected_backend_file:
                assert backend_file is None

            else:
                assert backend_file == f"{temp_dir}/{expected_backend_file}"

                with open(backend_file, "r") as backend_fh:
                    assert backend_fh.read().strip() == """
terraform {
  backend "local" {
    path = "./.local-state"
  }
}
""".strip()
        finally:
            if temp_dir:
                shutil.rmtree(temp_dir)


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
        """Test switching terraform versions whilst Terraform locked is already acquired."""
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
            raise subprocess.CalledProcessError(cmd="the command", returncode=2, output=b"")

        with unittest.mock.patch('terrareg.module_extractor.subprocess.check_output',
                                 unittest.mock.MagicMock(side_effect=raise_error)) as mock_check_output:

            assert module_extractor._get_graph_data(module_path='/tmp/mock-patch/to/module') is None

            mock_check_output.assert_called_once_with(
                [os.getcwd() + '/bin/terraform', 'graph'],
                cwd='/tmp/mock-patch/to/module'
            )

    @pytest.mark.parametrize("public_url,manage_terraform_rc_file,should_create_file,should_contain_credentials_block", [
        ("", False, False, False),
        ("", True, True, False),
        ("https://unittest-example-domain.com", False, False, False),
        ("https://unittest-example-domain.com", True, True, True)
    ])
    def test_create_terraform_rc_file(self, public_url, manage_terraform_rc_file, should_create_file, should_contain_credentials_block):
        """Test terraform RC file"""
        # Create temporary file and remove
        temp_file = tempfile.mktemp()
        # os.unlink(temp_file)

        with unittest.mock.patch("terrareg.module_extractor.ModuleExtractor.terraform_rc_file", temp_file), \
                unittest.mock.patch("os.makedirs") as mock_makedirs, \
                unittest.mock.patch("terrareg.config.Config.MANAGE_TERRAFORM_RC_FILE", manage_terraform_rc_file), \
                unittest.mock.patch("terrareg.config.Config.PUBLIC_URL", public_url):

            module_extractor = GitModuleExtractor(module_version=None)
            module_extractor._create_terraform_rc_file()

            if should_create_file:
                assert os.path.isfile(temp_file)

                mock_makedirs.assert_called_once_with(f"{os.path.expanduser('~')}/.terraform.d/plugin-cache", exist_ok=True)

                with open(temp_file, "r") as temp_file_fh:
                    if should_contain_credentials_block:
                        assert "".join(temp_file_fh.readlines()) == f"""
# Cache plugins
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true


credentials "unittest-example-domain.com" {{
  token = "internal-terrareg-analytics-token"
}}
"""

                    else:
                        assert "".join(temp_file_fh.readlines()) == f"""
# Cache plugins
plugin_cache_dir   = "$HOME/.terraform.d/plugin-cache"
disable_checkpoint = true

"""

            else:
                assert not os.path.isfile(temp_file)


    def test_extract_example_files(self):
        """Test _extract_example_files method"""

        test_module_dir = '/tmp/extraction_test'

        tested_file_extensions = []
        opened_files = []

        pattern_responses = {
            '*.tf': [
                'subdirectory/main.tf',
                'subdirectory/output.tf'
            ],
            '*.ext2': {},
            '*.ext3': [
                'subdirectory/blah.ext3'
            ]
        }
        file_contents = {
            'subdirectory/main.tf': 'test_content_main.tf',
            'subdirectory/output.tf': 'output file content',
            'subdirectory/blah.ext3': 'some ext3 content'
        }

        def mock_safe_iglob_effect(base_dir, pattern, recursive, is_file):
            """Create mock iglob method to return list of files that match pattern
            and mark the pattern as having been tested.
            """
            # Ensure base directory and other attributes are correct
            assert base_dir == '/tmp/extraction_test/subdirectory'
            assert recursive is False
            assert is_file is True
            # Ensure an expected pattern was provided
            tested_file_extensions.append(pattern)
            assert pattern in pattern_responses

            # Return list of matching files
            return [
                f'{test_module_dir}/{filename}'
                for filename in pattern_responses[pattern]
            ]

        mock_safe_iglob = unittest.mock.MagicMock(side_effect=mock_safe_iglob_effect)

        def mock_open_file_effect(path, mode):
            """Create mock side effect for open(), returning mock context manager.
            The mock context manager with return a mock FH with mocked readlines() method
            """
            assert mode == 'r'
            opened_files.append(path)
            mock_open_context = unittest.mock.MagicMock()
            mock_fh = unittest.mock.MagicMock()
            mock_open_context.__enter__ = unittest.mock.MagicMock(return_value=mock_fh)
            mock_fh.readlines = unittest.mock.MagicMock(
                # Return file contents, removing the base directory from the path provided to open()
                side_effect=lambda: file_contents[path.replace(f'{test_module_dir}/', '')]
            )
            return mock_open_context

        mock_open_file = unittest.mock.MagicMock(side_effect=mock_open_file_effect)

        mock_example = unittest.mock.MagicMock()
        mock_example.path = './subdirectory'

        created_example_files = {}

        def mock_create_example_file(example, path):
            """Mock ExampleFile.create to return mock instance of ExampleFile"""
            assert example is mock_example
            mock_example_file = unittest.mock.MagicMock()
            created_example_files[path] = mock_example_file
            return mock_example_file

        # Create mock for ExampleFile and mock .create
        mock_example_file = unittest.mock.MagicMock()
        mock_example_file.create = unittest.mock.MagicMock(side_effect=mock_create_example_file)

        # Create module version object with mocked git path,
        # to allow mock.patch to read the previous property value
        # during the mocking of GitModuleExtractor.module_directory
        mock_module_version = unittest.mock.MagicMock()
        mock_module_version.git_path = ''

        with unittest.mock.patch('terrareg.module_extractor.Config.EXAMPLE_FILE_EXTENSIONS', ["tf", "ext2", "ext3"]), \
                unittest.mock.patch('terrareg.module_extractor.open', mock_open_file), \
                unittest.mock.patch('terrareg.module_extractor.safe_iglob', mock_safe_iglob), \
                unittest.mock.patch('terrareg.module_extractor.GitModuleExtractor.module_directory', '/tmp/extraction_test'), \
                unittest.mock.patch('terrareg.models.ExampleFile', mock_example_file):

            module_extractor = GitModuleExtractor(module_version=mock_module_version)

            module_extractor._extract_example_files(example=mock_example)

        # Ensure each of the extensions was globbed for
        assert tested_file_extensions == ['*.tf',  '*.ext2', '*.ext3']

        # Ensure open() was called against each example file
        assert opened_files == [
            '/tmp/extraction_test/subdirectory/main.tf',
            '/tmp/extraction_test/subdirectory/output.tf',
            '/tmp/extraction_test/subdirectory/blah.ext3'
        ]

        # Ensure each returned file had an example created for
        assert list(created_example_files.keys()) == [
            'subdirectory/main.tf',
            'subdirectory/output.tf',
            'subdirectory/blah.ext3'
        ]

        # Ensure each example file was updated with correct content of file
        for example_path, mock_example_file_instance in created_example_files.items():
            mock_example_file_instance.update_attributes.assert_called_once_with(
                content=file_contents[example_path]
            )