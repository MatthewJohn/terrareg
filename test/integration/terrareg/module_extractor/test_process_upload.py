
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

output "test_input" {
    description = "test output"
    value       = var.test_input
}
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

    def test_terrareg_metadata(self):
        """Test module upload with terrareg metadata file."""
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
            module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')
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
            module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')
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
        pass

    def test_override_repo_urls_with_metadata(self):
        """Test module upload with repo urls in metadata file."""
        pass

    def test_sub_modules(self):
        """Test uploading module with submodules."""
        pass

    def test_examples(self):
        """Test uploading module with examples."""
        pass

    def test_upload_with_readme(self):
        """Test uploading a module with a README."""
        pass

    def test_all_features(self):
        """Test uploading a module with multiple features."""
        pass

    def test_uploading_module_with_invalid_terraform(self):
        """Test uploading a module with invalid terraform."""
        pass
