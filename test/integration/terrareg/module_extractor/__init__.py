
import tempfile
from enum import Enum
import os
import subprocess
import re

from terrareg.module_extractor import ApiUploadModuleExtractor


class UploadTestModuleState(Enum):
    NONE = 0
    CREATED_DIRECTORIES = 1
    READY_TO_CREATE_ZIP = 2
    CREATED_ZIP = 3
    DELETED_DIRECTORIES = 4


class UploadTestModule:
    """Provide interface to upload test module."""

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

    INVALID_MAIN_TF_FILE = """
var "test" {

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
            subprocess.call(
                ['zip', '-r', self._zip_file_name, '.'],
                cwd=self._source_directory.name
            )
            self._state = UploadTestModuleState.CREATED_ZIP
        elif self._state is UploadTestModuleState.CREATED_ZIP:
            # On second exit, delete temporary directories
            self._source_directory.__exit__(*args, **kwargs)
            self._zip_file_directory.__exit__(*args, **kwargs)

    @staticmethod
    def upload_module_version(module_version, zip_file):
        """Use ApiUploadModuleExtractor to upload version of module."""
        with ApiUploadModuleExtractor(upload_file=None, module_version=module_version) as me:
            with open(zip_file, 'rb') as zip_file_fh:
                me._source_file = zip_file_fh
                me._extract_archive()
            # Perform base module upload
            super(ApiUploadModuleExtractor, me).process_upload()
