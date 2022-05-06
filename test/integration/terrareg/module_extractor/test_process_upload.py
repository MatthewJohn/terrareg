
from distutils.command.upload import upload
from enum import Enum
import json
import os
import re
import shutil
import tempfile
from unittest import mock
import pytest

import terrareg.errors
from terrareg.models import Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import ApiUploadModuleExtractor
from test.integration.terrareg import TerraregIntegrationTest

class UploadTestModuleState(Enum):
    NONE = 0
    CREATED_DIRECTORIES = 1
    READY_TO_CREATE_ZIP = 2
    CREATED_ZIP = 3
    DELETED_DIRECTORIES = 4

class UploadTestModule:
    """Provide interface to upload test module."""

    def __init__(self):
        """Generate temporary directories"""
        self._source_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._zip_file_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._state = UploadTestModuleState.NONE
        self._zip_file_name = None

    def __enter__(self):
        """Create temporary directories."""
        if self._state is UploadTestModuleState.NONE:
            # On first entry, create temporary directories
            self._source_directory.__enter__()
            self._zip_file_directory.__enter__()
            self._zip_file_name = os.path.join(self._zip_file_directory.name, 'upload.zip')
            self._state = UploadTestModuleState.CREATED_DIRECTORIES
            # Return name of output zip that will be created.
            return self._zip_file_name

        elif self._state is UploadTestModuleState.CREATED_DIRECTORIES:
            # On second entry, just change state to show that this has been entered
            self._state = UploadTestModuleState.READY_TO_CREATE_ZIP
            # Return directory to upload files to
            return self._source_directory.name
        else:
            raise Exception('Called too many times')
    
    def __exit__(self, *args, **kwargs):
        """Zip file and perform tidy up"""
        if self._state is UploadTestModuleState.READY_TO_CREATE_ZIP:
            # Zip source directory
            shutil.make_archive(
                re.sub(r'\.zip$', '', self._zip_file_name),
                'zip',
                self._source_directory.name)
            self._state = UploadTestModuleState.CREATED_ZIP
        elif self._state is UploadTestModuleState.CREATED_ZIP:
            # On second exit, delete temporary directories
            self._source_directory.__exit__(*args, **kwargs)
            self._zip_file_directory.__exit__(*args, **kwargs)


class TestProcessUpload(TerraregIntegrationTest):
    """Test the module extractor process_upload."""

    TEST_README_CONTENT = """
# Test terraform module

## This is for a test module.
"""

    VALID_MAIN_TF_FILE = """
variable "test_input" {
    type        = string
    description = "This is a test input"
    default     = "test_default_val"
}

output "test_output" {
    description = "test output"
    value       = var.test_input
}
"""
    SUB_MODULE_MAIN_TF = """
variable "submodule_test_input_{itx}" {{
    type        = string
    description = "This is a test input in a submodule"
    default     = "test_default_val"
}}

output "submodule_test_output_{itx}" {{
    description = "test output in a submodule"
    value       = var.test_input
}}
"""

    def _upload_module_version(self, module_version, zip_file):
        """Use ApiUploadModuleExtractor to upload version of module."""
        with ApiUploadModuleExtractor(upload_file=None, module_version=module_version) as me:
            with open(zip_file, 'rb') as zip_file_fh:
                me._source_file = zip_file_fh
                me._extract_archive()
            # Perform base module upload
            super(ApiUploadModuleExtractor, me).process_upload()

    def test_basic_module(self):
        """Test basic module upload with single depth."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure terraform docs output contains variable and output
        assert module_version.get_terraform_inputs() == [
            {
                'default': 'test_default_val',
                'description': 'This is a test input',
                'name': 'test_input',
                'required': False,
                'type': 'string'
            }
        ]
        assert module_version.get_terraform_outputs() == [
            {
                'description': 'test output',
                'name': 'test_output'
            }
        ]

    def test_terrareg_metadata(self):
        """Test module upload with terrareg metadata file."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='2.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'description': 'unittestdescription!',
                        'owner': 'unittestowner.',
                        'variable_template': [{'test_variable': {}}]
                    }))

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        assert module_version.description == 'unittestdescription!'
        assert module_version.owner == 'unittestowner.'
        assert module_version.variable_template == [{'test_variable': {}}]


    def test_terrareg_metadata_required_attributes(self):
        """Test module upload with terrareg metadata file with required attributes."""
        with mock.patch('terrareg.config.Config.REQUIRED_MODULE_METADATA_ATTRIBUTES', ['description', 'owner']):
            test_upload = UploadTestModule()

            namespace = Namespace(name='testprocessupload')
            module = Module(namespace=namespace, name='test-module')
            module_provider = ModuleProvider.get(module=module, name='aws', create=True)
            module_version = ModuleVersion(module_provider=module_provider, version='3.0.0')
            module_version.prepare_module()

            with test_upload as zip_file:
                with test_upload as upload_directory:
                    # Create main.tf
                    with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                    with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                        metadata_fh.writelines(json.dumps({
                            'description': 'unittestdescription!',
                            'owner': 'unittestowner.',
                            'variable_template': [{'test_variable': {}}]
                        }))

                self._upload_module_version(module_version=module_version, zip_file=zip_file)

            assert module_version.description == 'unittestdescription!'
            assert module_version.owner == 'unittestowner.'
            assert module_version.variable_template == [{'test_variable': {}}]

    @pytest.mark.parametrize('terrareg_json', [
        {},
        {'description': 'unittest'},
        {'owner': 'testowner'},
        {'owner': 'testowner', 'variable_template': [{}]}
    ])
    def test_terrareg_metadata_missing_required_attributes(self, terrareg_json):
        """Test module upload with missing required terrareg metadata attributes."""
        with mock.patch('terrareg.config.Config.REQUIRED_MODULE_METADATA_ATTRIBUTES', ['description', 'owner']):
            test_upload = UploadTestModule()

            namespace = Namespace(name='testprocessupload')
            module = Module(namespace=namespace, name='test-module')
            module_provider = ModuleProvider.get(module=module, name='aws', create=True)
            module_version = ModuleVersion(module_provider=module_provider, version='4.0.0')
            module_version.prepare_module()

            with test_upload as zip_file:
                with test_upload as upload_directory:
                    # Create main.tf
                    with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                    with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                        metadata_fh.writelines(json.dumps(terrareg_json))

                # Ensure an exception is raised about missing attributes
                with pytest.raises(terrareg.errors.MetadataDoesNotContainRequiredAttributeError):
                    self._upload_module_version(module_version=module_version, zip_file=zip_file)


    def test_invalid_terrareg_metadata_file(self):
        """Test module upload with an invaid terrareg metadata file."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='5.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines('This is invalid JSON!')

            # Ensure an exception is raised about invalid metadata JSON
            with pytest.raises(terrareg.errors.InvalidTerraregMetadataFileError):
                self._upload_module_version(module_version=module_version, zip_file=zip_file)

    def test_override_repo_urls_with_metadata(self):
        """Test module upload with repo urls in metadata file."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name='module-provider-override-git-provider')
        module_provider = ModuleProvider.get(module=module, name='test')

        # Ensure that module provider is setup with git proider and overriden repo URLs
        assert module_provider is not None
        assert module_provider._get_db_row()['repo_base_url_template']
        assert module_provider._get_db_row()['repo_clone_url_template']
        assert module_provider._get_db_row()['repo_browse_url_template']
        assert module_provider._get_db_row()['git_provider_id']

        module_version = ModuleVersion(module_provider=module_provider, version='1.5.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'repo_clone_url': 'ssh://overrideurl_here.com/{namespace}/{module}-{provider}',
                        'repo_base_url': 'https://realoverride.com/blah/{namespace}-{module}-{provider}',
                        'repo_browse_url': 'https://base_url.com/{namespace}-{module}-{provider}-{tag}/{path}'
                    }))

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

            assert module_version.get_source_base_url() == 'https://realoverride.com/blah/repo_url_tests-module-provider-override-git-provider-test'
            assert module_version.get_git_clone_url() == 'ssh://overrideurl_here.com/repo_url_tests/module-provider-override-git-provider-test'
            assert module_version.get_source_browse_url() == 'https://base_url.com/repo_url_tests-module-provider-override-git-provider-test-1.5.0/'
            assert module_version.get_source_browse_url(path='subdir') == 'https://base_url.com/repo_url_tests-module-provider-override-git-provider-test-1.5.0/subdir'

    def test_sub_modules(self):
        """Test uploading module with submodules."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='6.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                os.mkdir(os.path.join(upload_directory, 'modules'))

                # Create main.tf in each of the submodules
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'modules', 'testmodule{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.SUB_MODULE_MAIN_TF.format(itx=itx))

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        submodules = module_version.get_submodules()
        # Order submodules by path
        submodules.sort(key=lambda x: x.path)
        assert len(submodules) == 2
        assert [sm.path for sm in submodules] == ['modules/testmodule1', 'modules/testmodule2']

        for itx, submodule in enumerate(submodules):
            # Ensure terraform docs output contains variable and output
            assert submodule.get_terraform_inputs() == [
                {
                    'default': 'test_default_val',
                    'description': 'This is a test input in a submodule',
                    'name': 'submodule_test_input_{itx}'.format(itx=(itx + 1)),
                    'required': False,
                    'type': 'string'
                }
            ]
            assert submodule.get_terraform_outputs() == [
                {
                    'description': 'test output in a submodule',
                    'name': 'submodule_test_output_{itx}'.format(itx=(itx + 1))
                }
            ]
        assert len(module_version.get_examples()) == 0

    def test_examples(self):
        """Test uploading module with examples."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='7.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                os.mkdir(os.path.join(upload_directory, 'examples'))

                # Create main.tf in each of the examples
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'examples', 'testexample{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.SUB_MODULE_MAIN_TF.format(itx=itx))

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        examples = module_version.get_examples()
        # Order submodules by path
        examples.sort(key=lambda x: x.path)
        assert len(examples) == 2
        assert [example.path for example in examples] == ['examples/testexample1', 'examples/testexample2']

        for itx, example in enumerate(examples):
            # Ensure terraform docs output contains variable and output
            assert example.get_terraform_inputs() == [
                {
                    'default': 'test_default_val',
                    'description': 'This is a test input in a submodule',
                    'name': 'submodule_test_input_{itx}'.format(itx=(itx + 1)),
                    'required': False,
                    'type': 'string'
                }
            ]
            assert example.get_terraform_outputs() == [
                {
                    'description': 'test output in a submodule',
                    'name': 'submodule_test_output_{itx}'.format(itx=(itx + 1))
                }
            ]
        assert len(module_version.get_submodules()) == 0

    def test_upload_with_readme(self):
        """Test uploading a module with a README."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='8.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                # Create README
                with open(os.path.join(upload_directory, 'README.md'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.TEST_README_CONTENT)

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure README is present in module version
        assert module_version.get_readme_content() == self.TEST_README_CONTENT

    def test_all_features(self):
        """Test uploading a module with multiple features."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='repo_url_tests')
        module = Module(namespace=namespace, name='module-provider-override-git-provider')
        module_provider = ModuleProvider.get(module=module, name='test')
        module_version = ModuleVersion(module_provider=module_provider, version='9.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'description': 'Test unittest description',
                        'owner': 'Test unittest owner',
                        'variable_template': [{'test_variable': {}}],
                        'repo_clone_url': 'ssh://overrideurl_here.com/{namespace}/{module}-{provider}',
                        'repo_base_url': 'https://realoverride.com/blah/{namespace}-{module}-{provider}',
                        'repo_browse_url': 'https://base_url.com/{namespace}-{module}-{provider}-{tag}/{path}'
                    }))

                # Create README
                with open(os.path.join(upload_directory, 'README.md'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(self.TEST_README_CONTENT)

                os.mkdir(os.path.join(upload_directory, 'modules'))

                # Create main.tf in each of the submodules
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'modules', 'testmodule{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.SUB_MODULE_MAIN_TF.format(itx=itx))

                os.mkdir(os.path.join(upload_directory, 'examples'))

                # Create main.tf in each of the examples
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'examples', 'testexample{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(self.SUB_MODULE_MAIN_TF.format(itx=itx))

            self._upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure README is present in module version
        assert module_version.get_readme_content() == self.TEST_README_CONTENT

        # Check submodules
        submodules = module_version.get_submodules()
        submodules.sort(key=lambda x: x.path)
        assert len(submodules) == 2
        assert [sm.path for sm in submodules] == ['modules/testmodule1', 'modules/testmodule2']

        # Check examples
        examples = module_version.get_examples()
        examples.sort(key=lambda x: x.path)
        assert len(examples) == 2
        assert [example.path for example in examples] == ['examples/testexample1', 'examples/testexample2']

        # Check repo URLs
        assert module_version.get_source_base_url() == 'https://realoverride.com/blah/repo_url_tests-module-provider-override-git-provider-test'
        assert module_version.get_git_clone_url() == 'ssh://overrideurl_here.com/repo_url_tests/module-provider-override-git-provider-test'
        assert module_version.get_source_browse_url() == 'https://base_url.com/repo_url_tests-module-provider-override-git-provider-test-9.0.0/'
        assert module_version.get_source_browse_url(path='subdir') == 'https://base_url.com/repo_url_tests-module-provider-override-git-provider-test-9.0.0/subdir'

        # Check attributes from terrareg
        assert module_version.description == 'unittestdescription!'
        assert module_version.owner == 'unittestowner.'
        assert module_version.variable_template == [{'test_variable': {}}]

    def test_uploading_module_with_invalid_terraform(self):
        """Test uploading a module with invalid terraform."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='10.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
                    this is { not_Really } valid "terraform"
                    """)

            with pytest.raises(terrareg.errors.UnableToProcessTerraformError):
                self._upload_module_version(module_version=module_version, zip_file=zip_file)

