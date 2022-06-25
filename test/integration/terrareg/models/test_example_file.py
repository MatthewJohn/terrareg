
import os

from terrareg.models import Module, ModuleVersion, Namespace, ModuleProvider
import terrareg.errors
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
