
import os
import unittest

import pytest

from terrareg.models import Example, ExampleFile, Module, ModuleVersion, Namespace, ModuleProvider
from test.integration.terrareg import TerraregIntegrationTest
from test.integration.terrareg.module_extractor import UploadTestModule

class TestExampleFile(TerraregIntegrationTest):

    def test_example_files(self):
        """Test uploading module with examples."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-example-file')
        module_provider = ModuleProvider.get(module=module, name='testprovider', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='1.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                os.mkdir(os.path.join(upload_directory, 'examples'))

                # Create terraform files in example
                root_dir = os.path.join(upload_directory, 'examples', 'testexample')
                os.mkdir(root_dir)
                for file_name in ['variables.tf', 'data.tf', 'outputs.tf', 'main.tf']:
                    with open(os.path.join(root_dir, file_name), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=1))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        example = module_version.get_examples()[0]

        # Ensure file list is ordered correctly
        file_list = example.get_files()
        file_list = sorted(file_list)
        assert [file.file_name for file in file_list] == ['main.tf', 'data.tf', 'outputs.tf', 'variables.tf']

    @pytest.mark.parametrize('file_content,expected_output', [
        # Test empty file
        ('', ''),

        # Basic call to example module (bit cyclic, so won't happen in real life)
        (
"""
module "test-module" {
    source = "./"
}
""",
"""
module "test-module" {
    source  = "example.com/moduledetails/readme-tests/provider//examples/testreadmeexample"
    version = ">= 1.0.0, < 1.1.0"
}
"""
        ),

        # Basic call to root module
        (
"""
module "test-module" {
    source = "../../"

    some_attribute = "test"
}
""",
"""
module "test-module" {
    source  = "example.com/moduledetails/readme-tests/provider"
    version = ">= 1.0.0, < 1.1.0"

    some_attribute = "test"
}
"""
        )
    ])
    def test_source_replacement_in_file_content(self, file_content, expected_output):
        """Test source replacement in example file content."""

        module_version = ModuleVersion(ModuleProvider(Module(Namespace('moduledetails'), 'readme-tests'), 'provider'), '1.0.0')
        example = Example(module_version, 'examples/testreadmeexample')
        example_file = ExampleFile(example, 'examples/testreadmeexample/main.tf')
        example_file.update_attributes(content=file_content)

        with unittest.mock.patch('terrareg.config.Config.TERRAFORM_EXAMPLE_VERSION_TEMPLATE', '>= {major}.{minor}.{patch}, < {major}.{minor_plus_one}.0'):
            assert example_file.get_content(server_hostname='example.com') == expected_output
