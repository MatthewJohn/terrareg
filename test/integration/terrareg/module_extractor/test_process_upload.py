
import json
import os

from unittest import mock
import pytest

import terrareg.errors
from terrareg.models import GitProvider, Module, ModuleProvider, ModuleVersion, Namespace
from test.integration.terrareg import TerraregIntegrationTest
from test import client
from test.integration.terrareg.module_extractor import UploadTestModule


class TestProcessUpload(TerraregIntegrationTest):
    """Test the module extractor process_upload."""

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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

        # Check tfsec returned no results
        assert module_version.module_details.tfsec == {'results': None}

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'description': 'unittestdescription!',
                        'owner': 'unittestowner.',
                        'variable_template': [{'test_variable': {}}]
                    }))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                        main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                    with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                        metadata_fh.writelines(json.dumps({
                            'description': 'unittestdescription!',
                            'owner': 'unittestowner.',
                            'variable_template': [{'test_variable': {}}]
                        }))

                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                        main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                    with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                        metadata_fh.writelines(json.dumps(terrareg_json))

                # Ensure an exception is raised about missing attributes
                with pytest.raises(terrareg.errors.MetadataDoesNotContainRequiredAttributeError):
                    UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)


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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines('This is invalid JSON!')

            # Ensure an exception is raised about invalid metadata JSON
            with pytest.raises(terrareg.errors.InvalidTerraregMetadataFileError):
                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'repo_clone_url': 'ssh://overrideurl_here.com/{namespace}/{module}-{provider}',
                        'repo_base_url': 'https://realoverride.com/blah/{namespace}-{module}-{provider}',
                        'repo_browse_url': 'https://base_url.com/{namespace}-{module}-{provider}-{tag}/{path}'
                    }))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                os.mkdir(os.path.join(upload_directory, 'modules'))

                # Create main.tf in each of the submodules
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'modules', 'testmodule{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                os.mkdir(os.path.join(upload_directory, 'examples'))

                # Create main.tf in each of the examples
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'examples', 'testexample{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                # Create README
                with open(os.path.join(upload_directory, 'README.md'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.TEST_README_CONTENT)

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure README is present in module version
        assert module_version.get_readme_content() == UploadTestModule.TEST_README_CONTENT

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
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

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
                    main_tf_fh.writelines(UploadTestModule.TEST_README_CONTENT)

                os.mkdir(os.path.join(upload_directory, 'modules'))

                # Create main.tf in each of the submodules
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'modules', 'testmodule{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

                os.mkdir(os.path.join(upload_directory, 'examples'))

                # Create main.tf in each of the examples
                for itx in [1, 2]:
                    root_dir = os.path.join(upload_directory, 'examples', 'testexample{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure README is present in module version
        assert module_version.get_readme_content() == UploadTestModule.TEST_README_CONTENT

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
        assert module_version.description == 'Test unittest description'
        assert module_version.owner == 'Test unittest owner'
        assert module_version.variable_template == [{'test_variable': {}}]

    def test_non_root_repo_directory(self):
        """Test uploading a module within a sub-directory of a module."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='git-path')
        module_provider = ModuleProvider.get(module=module, name='test', create=True)

        module_provider.update_git_provider(GitProvider(2))
        module_provider.update_git_path('subdirectory/in/repo')

        module_version = ModuleVersion(module_provider=module_provider, version='1.1.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:

                module_dir = os.path.join(upload_directory, 'subdirectory/in/repo')
                os.makedirs(module_dir)

                # Create main.tf
                with open(os.path.join(module_dir, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

                with open(os.path.join(module_dir, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'description': 'Test unittest description',
                        'owner': 'Test unittest owner'
                    }))

                # Create README
                with open(os.path.join(module_dir, 'README.md'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.TEST_README_CONTENT)

                os.mkdir(os.path.join(module_dir, 'modules'))

                # Create main.tf in each of the submodules
                for itx in [1, 2]:
                    root_dir = os.path.join(module_dir, 'modules', 'testmodule{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

                os.mkdir(os.path.join(module_dir, 'examples'))

                # Create main.tf in each of the examples
                for itx in [1, 2]:
                    root_dir = os.path.join(module_dir, 'examples', 'testexample{itx}'.format(itx=itx))
                    os.mkdir(root_dir)
                    with open(os.path.join(root_dir, 'main.tf'), 'w') as main_tf_fh:
                        main_tf_fh.writelines(UploadTestModule.SUB_MODULE_MAIN_TF.format(itx=itx))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure README is present in module version
        assert module_version.get_readme_content() == UploadTestModule.TEST_README_CONTENT

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
        assert module_version.get_source_browse_url() == 'https://browse-url.com/testprocessupload/git-path-test/browse/1.1.0/subdirectory/in/reposuffix'
        assert module_version.get_source_browse_url(path='subdir') == 'https://browse-url.com/testprocessupload/git-path-test/browse/1.1.0/subdirectory/in/repo/subdirsuffix'

        # Check attributes from terrareg
        assert module_version.description == 'Test unittest description'
        assert module_version.owner == 'Test unittest owner'

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
                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

    def test_uploading_module_with_security_issue(self):
        """Test uploading a module with security issue."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='11.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
                    resource "aws_secretsmanager_secret" "this" {
                        name = "example"
                    }
                    """)

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure tfsec output contains security issue about missing encryption key
        assert module_version.module_details.tfsec == {'results': [
            {
                'description': 'Secret explicitly uses the default key.',
                'impact': 'Using AWS managed keys reduces the flexibility and '
                          'control over the encryption key',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.26.0/checks/aws/ssm/secret-use-customer-key/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/secretsmanager_secret#kms_key_id'
                ],
                'location': {
                    'end_line': 4,
                    'filename': 'main.tf',
                    'start_line': 2
                },
                'long_id': 'aws-ssm-secret-use-customer-key',
                'resolution': 'Use customer managed keys',
                'resource': 'aws_secretsmanager_secret.this',
                'rule_description': 'Secrets Manager should use customer managed '
                                    'keys',
                'rule_id': 'AVD-AWS-0098',
                'rule_provider': 'aws',
                'rule_service': 'ssm',
                'severity': 'LOW',
                'status': 0,
                'warning': False
            }
        ]}

        # Ensure security issue count shows the issue
        assert module_version.get_tfsec_failure_count() == 1

    def test_uploading_module_with_reference_to_inaccessible_remote_module(self):
        """Test uploading a module with reference to inaccessible external module."""
        test_upload = UploadTestModule()

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='12.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
                    module "inaccessible" {
                      source = "http://example.com/not-accessible.zip"
                    }
                    """)

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

    def test_upload_via_server(self, client):
        """Test basic module upload with single depth."""
        test_upload = UploadTestModule()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE)

            client.post(
                '/v1/terrareg/modules/testprocessupload/test-module/uplaodthroughserver/1.5.2/upload',
                data={
                    'file': (zip_file, 'module.zip')
                }
            )

        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='uplaodthroughserver')
        module_version = ModuleVersion(module_provider=module_provider, version='1.5.2')

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

    def test_bad_upload_via_server(self, client):
        """Test invalid module upload, checking that module provider is cleared up in transaction."""
        test_upload = UploadTestModule()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.INVALID_MAIN_TF_FILE)

            res = client.post(
                '/v1/terrareg/modules/testprocessupload/test-module/badupload/2.0.0/upload',
                data={
                    'file': (zip_file, 'module.zip')
                }
            )

            assert res.status_code == 500

        # Ensure that module provider was not created.
        namespace = Namespace(name='testprocessupload')
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='badupload')
        
        assert module_provider is None
