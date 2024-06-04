
import shutil
import subprocess
import json
import os
import tarfile
from tempfile import TemporaryDirectory, mkdtemp
import tempfile
import platform

from unittest import mock
import zipfile
import pytest

import terrareg.config
import terrareg.errors
from terrareg.models import GitProvider, Module, ModuleProvider, ModuleVersion, Namespace
from terrareg.module_extractor import ApiUploadModuleExtractor
from test.integration.terrareg import TerraregIntegrationTest
from test import client, skipif_unless_ci
from test.integration.terrareg.module_extractor import UploadTestModule
import terrareg.utils
import terrareg.file_storage


class TestProcessUpload(TerraregIntegrationTest):
    """Test the module extractor process_upload."""

    def test_basic_module(self):
        """Test basic module upload with single depth."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
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
        assert module_version.module_details.tfsec == {'results': []}

        # Check Infracost returned no results
        assert module_version.module_details.infracost == {}

    def test_terrareg_metadata(self):
        """Test module upload with terrareg metadata file."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
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
                        'variable_template': [
                            {"name": "test_variable","type": "number", "quote_value": True, 'required': False},
                            {"name": "test_defaults"}
                        ],
                    }))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        assert module_version.description == 'unittestdescription!'
        assert module_version.owner == 'unittestowner.'
        assert module_version.variable_template == [
            {
                'name': 'test_variable',
                'quote_value': True,
                'type': 'number',
                'required': False,
                'default_value': None,
                'additional_help': ''
            },
            {
                'name': 'test_defaults',
                'quote_value': True,
                'type': 'text',
                'required': True,
                'default_value': None,
                'additional_help': ''
            },
            {
                'additional_help': 'This is a test input',
                'default_value': 'test_default_val',
                'name': 'test_input',
                'quote_value': True,
                'required': False,
                'type': 'text'
            }
        ]

    def test_terrareg_metadata_override_autogenerate(self):
        """Test module upload with terrareg metadata file overriding autogenrated variable."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
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
                        'variable_template': [{"name": "test_input","type": "text","quote_value": True,'required': False}],
                    }))

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        assert module_version.description == 'unittestdescription!'
        assert module_version.owner == 'unittestowner.'
        assert module_version.variable_template == [
            {
                'name': 'test_input',
                'quote_value': True,
                'default_value': None,
                'type': 'text',
                'required': False,
                'additional_help': ''
            },
        ]


    def test_terrareg_metadata_required_attributes(self):
        """Test module upload with terrareg metadata file with required attributes."""
        with mock.patch('terrareg.config.Config.REQUIRED_MODULE_METADATA_ATTRIBUTES', ['description', 'owner']):
            test_upload = UploadTestModule()

            namespace = Namespace.get(name='testprocessupload', create=True)
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
                            'variable_template': [{"name": "test_variable","type": "text","quote_value": True,'required': False}],
                        }))

                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

            assert module_version.description == 'unittestdescription!'
            assert module_version.owner == 'unittestowner.'
            assert module_version.variable_template == [
            {
                'name': 'test_variable',
                'quote_value': True,
                'type': 'text',
                'required': False,
                'additional_help': '',
                'default_value': None
            },
            {
                'additional_help': 'This is a test input',
                'default_value': 'test_default_val',
                'name': 'test_input',
                'quote_value': True,
                'required': False,
                'type': 'text'
            }
        ]

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

            namespace = Namespace.get(name='testprocessupload', create=True)
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

        namespace = Namespace.get(name='testprocessupload', create=True)
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

        # Ensure that module provider is setup with git proider and overridden repo URLs
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

        namespace = Namespace.get(name='testprocessupload', create=True)
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

        namespace = Namespace.get(name='testprocessupload', create=True)
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

        namespace = Namespace.get(name='testprocessupload', create=True)
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
                    main_tf_fh.writelines(UploadTestModule.MAIN_TF_WITH_HTML_MARKDOWN)

                with open(os.path.join(upload_directory, 'terrareg.json'), 'w') as metadata_fh:
                    metadata_fh.writelines(json.dumps({
                        'description': 'Test unittest description',
                        'owner': 'Test unittest owner',
                        'variable_template': [
                            {"name": "test_variable","type": "text","quote_value": True,'required': False},
                            {"name": "with_html_content","type": "text","quote_value": True,'required': False,
                             "additional_help": "<b>Some <a href='#'>link</a></b>"},
                        ],
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
        assert module_version.get_variable_template() == [
            {
                'additional_help': '',
                'default_value': None,
                'name': 'test_variable',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
            {
                'additional_help': '<b>Some <a href="#">link</a></b>',
                'default_value': None,
                'name': 'with_html_content',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
            {
                'additional_help': '<h1>Title</h1>\n'
                                   '<h2>Sub title</h2>\n'
                                   '\n'
                                   '<b>Some bold text</b>\n'
                                   '\n'
                                   '<code>Code block</code>\n'
                                   '\n'
                                   '<a href="https://duckduckgo.com">Duck Duck Go</a>\n',
                'default_value': '<b>Some bold Text</b>\n',
                'name': 'test_input_html',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
            {
                'additional_help': '# This is markdown title\n'
                                   '\n'
                                   '## Sub title\n'
                                   '\n'
                                   '**Some bold text**\n'
                                   '\n'
                                   '`code block`\n'
                                   '\n'
                                   '```\n'
                                   'Multi-line\n'
                                   'code block\n'
                                   '```\n'
                                   '\n'
                                   '[Duck Duck Go](https://duckduckgo.com)\n'
                                   '\n',
                'default_value': '**Some bold text**\n',
                'name': 'test_input_markdown',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
        ]

        assert module_version.get_variable_template(html=True) == [
            {
                'additional_help': '',
                'default_value': None,
                'name': 'test_variable',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
            {
                'additional_help': '<b>Some <a href="#">link</a></b>',
                'default_value': None,
                'name': 'with_html_content',
                'quote_value': True,
                'required': False,
                'type': 'text'
                },
            {
                'additional_help': '<h1>Title</h1>\n'
                                   '<h2>Sub title</h2>\n'
                                   '\n'
                                   '<p><b>Some bold text</b></p>\n'
                                   '<p><code>Code block</code></p>\n'
                                   '<p><a href="https://duckduckgo.com">Duck Duck Go</a></p>',
                'default_value': '<b>Some bold Text</b>\n',
                'name': 'test_input_html',
                'quote_value': True,
                'required': False,
                'type': 'text'
                },
            {
                'additional_help': '<h1 id="terrareg-anchor--this-is-markdown-title">This is markdown title</h1>\n'
                                   '<h2 id="terrareg-anchor--sub-title">Sub title</h2>\n'
                                   '<p><strong>Some bold text</strong></p>\n'
                                   '<p><code>code block</code></p>\n'
                                   '<pre><code>Multi-line\ncode block\n</code></pre>\n'
                                   '<p><a href="https://duckduckgo.com">Duck Duck Go</a></p>',
                'default_value': '**Some bold text**\n',
                'name': 'test_input_markdown',
                'quote_value': True,
                'required': False,
                'type': 'text'
            },
        ]

    def test_non_root_repo_directory(self):
        """Test uploading a module within a sub-directory of a module."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
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

        # Check attributes from terrareg
        assert module_version.description == 'Test unittest description'
        assert module_version.owner == 'Test unittest owner'

    def test_uploading_module_with_invalid_terraform(self):
        """Test uploading a module with invalid terraform."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
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

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='11.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
                    resource "aws_s3_bucket" "mybucket" {
                      bucket = "mybucket"

                      versioning {
                        enabled = true
                      }
                    }

                    resource "aws_s3_bucket_public_access_block" "mybucket" {
                      bucket = aws_s3_bucket.mybucket.id

                      block_public_acls       = true
                      block_public_policy     = true
                      ignore_public_acls      = true
                      restrict_public_buckets = true
                    }
                    """)

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure tfsec output contains security issue about missing encryption key
        assert module_version.module_details.tfsec == {'results': [
            {
                'rule_id': 'AVD-AWS-0086',
                'long_id': 'aws-s3-block-public-acls',
                'rule_description': 'S3 Access block should block public ACL',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'PUT calls with public ACLs specified can make objects public',
                'resolution': 'Enable blocking any PUT calls with a public ACL specified',
                'links': ['https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/block-public-acls/',
                'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block#block_public_acls'],
                'description': '',
                'severity': 'HIGH',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0087',
                'long_id': 'aws-s3-block-public-policy',
                'rule_description': 'S3 Access block should block public policy',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Users could put a policy that allows public access',
                'resolution': 'Prevent policies that allow public access being PUT',
                'links': ['https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/block-public-policy/',
                'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block#block_public_policy'],
                'description': '',
                'severity': 'HIGH',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0088',
                'long_id': 'aws-s3-enable-bucket-encryption',
                'rule_description': 'Unencrypted S3 bucket.',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'The bucket objects could be read if compromised',
                'resolution': 'Configure bucket encryption',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/enable-bucket-encryption/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket#enable-default-server-side-encryption'
                ],
                'description': 'Bucket does not have encryption enabled',
                'severity': 'HIGH',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0089',
                'long_id': 'aws-s3-enable-bucket-logging',
                'rule_description': 'S3 Bucket does not have logging enabled.',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'There is no way to determine the access to this bucket',
                'resolution': 'Add a logging block to the resource to enable access logging',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/enable-bucket-logging/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket'
                ],
                'description': 'Bucket does not have logging enabled',
                'severity': 'MEDIUM',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0090',
                'long_id': 'aws-s3-enable-versioning',
                'rule_description': 'S3 Data should be versioned',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Deleted or modified data would not be recoverable',
                'resolution': 'Enable versioning to protect against accidental/malicious removal or modification',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/enable-versioning/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket#versioning'
                ],
                'description': '',
                'severity': 'MEDIUM',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
                {'rule_id': 'AVD-AWS-0132',
                'long_id': 'aws-s3-encryption-customer-key',
                'rule_description': 'S3 encryption should use Customer Managed Keys',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Using AWS managed keys does not allow for fine grained control',
                'resolution': 'Enable encryption using customer managed keys',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/encryption-customer-key/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket#enable-default-server-side-encryption'
                ],
                'description': 'Bucket does not encrypt data with a customer managed key.',
                'severity': 'HIGH',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0091',
                'long_id': 'aws-s3-ignore-public-acls',
                'rule_description': 'S3 Access Block should Ignore Public Acl',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'PUT calls with public ACLs specified can make objects public',
                'resolution': 'Enable ignoring the application of public ACLs in PUT calls',
                'links': ['https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/ignore-public-acls/',
                'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block#ignore_public_acls'],
                'description': '',
                'severity': 'HIGH',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0092',
                'long_id': 'aws-s3-no-public-access-with-acl',
                'rule_description': 'S3 Buckets not publicly accessible through ACL.',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Public access to the bucket can lead to data leakage',
                'resolution': "Don't use canned ACLs or switch to private acl",
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/no-public-access-with-acl/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket'
                ],
                'description': '',
                'severity': 'HIGH',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0093',
                'long_id': 'aws-s3-no-public-buckets',
                'rule_description': 'S3 Access block should restrict public bucket to limit access',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Public buckets can be accessed by anyone',
                'resolution': 'Limit the access to public buckets to only the owner or AWS Services (eg; CloudFront)',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/no-public-buckets/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block#restrict_public_buckets¡'
                ],
                'description': '',
                'severity': 'HIGH',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0094',
                'long_id': 'aws-s3-specify-public-access-block',
                'rule_description': 'S3 buckets should each define an aws_s3_bucket_public_access_block',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Public access policies may be applied to sensitive data buckets',
                'resolution': 'Define a aws_s3_bucket_public_access_block for the given bucket to control public access policies',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/specify-public-access-block/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket_public_access_block#bucket'
                ],
                'description': '',
                'severity': 'LOW',
                'warning': False,
                'status': 1,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            }
        ]}


        # Ensure security issue count shows the issue
        assert module_version.get_tfsec_failures() == [
            {
                'rule_id': 'AVD-AWS-0088',
                'long_id': 'aws-s3-enable-bucket-encryption',
                'rule_description': 'Unencrypted S3 bucket.',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'The bucket objects could be read if compromised',
                'resolution': 'Configure bucket encryption',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/enable-bucket-encryption/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket#enable-default-server-side-encryption'
                ],
                'description': 'Bucket does not have encryption enabled',
                'severity': 'HIGH',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0089',
                'long_id': 'aws-s3-enable-bucket-logging',
                'rule_description': 'S3 Bucket does not have logging enabled.',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'There is no way to determine the access to this bucket',
                'resolution': 'Add a logging block to the resource to enable access logging',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/enable-bucket-logging/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket'
                ],
                'description': 'Bucket does not have logging enabled',
                'severity': 'MEDIUM',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            },
            {
                'rule_id': 'AVD-AWS-0132',
                'long_id': 'aws-s3-encryption-customer-key',
                'rule_description': 'S3 encryption should use Customer Managed Keys',
                'rule_provider': 'aws',
                'rule_service': 's3',
                'impact': 'Using AWS managed keys does not allow for fine grained control',
                'resolution': 'Enable encryption using customer managed keys',
                'links': [
                    'https://aquasecurity.github.io/tfsec/v1.28.4/checks/aws/s3/encryption-customer-key/',
                    'https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/s3_bucket#enable-default-server-side-encryption'
                ],
                'description': 'Bucket does not encrypt data with a customer managed key.',
                'severity': 'HIGH',
                'warning': False,
                'status': 0,
                'resource': 'aws_s3_bucket.mybucket',
                'location': {'filename': 'main.tf', 'start_line': 2, 'end_line': 8}
            }
        ]

    def test_graph_data(self):
        """Test graph data for uploaded module."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='20.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf in root module, example and submodule
                for identifier, directory_structure in [
                        ("root_module", []),
                        ("example", ["examples", "first-example"]),
                        ("submodule", ["modules", "first-submodule"])]:
                    # Generate parent directories for example/submodules
                    parent_paths = [upload_directory]
                    for directory in directory_structure:
                        parent_paths.append(directory)
                        os.mkdir(os.path.join(*parent_paths))

                    with open(os.path.join(*parent_paths, 'main.tf'), 'w') as fh:
                        fh.write(f"""
resource "aws_s3_bucket" "test_bucket" {{
  name = var.name
}}

resource "aws_s3_object" "test_obj_{identifier}" {{
  key     = "test/s3/bucket/object/key"
  bucket  = aws_s3_bucket.test_bucket.id
  content = "Important Content"
}}

variable "name" {{
  type    = string
  default = "Test value"
}}

output "name" {{
  value = aws_s3_bucket.test.name
}}

""")

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure Infracost output contains monthly cost
        assert module_version.module_details.terraform_graph.strip() == """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
		"[root] aws_s3_object.test_obj_root_module (expand)" [label = "aws_s3_object.test_obj_root_module", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] var.name" [label = "var.name", shape = "note"]
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] aws_s3_object.test_obj_root_module (expand)" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_object.test_obj_root_module (expand)"
		"[root] root" -> "[root] output.name (expand)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
		"[root] root" -> "[root] var.name"
	}
}
""".strip()

        assert module_version.get_examples()[0].module_details.terraform_graph.strip() == """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
		"[root] aws_s3_object.test_obj_example (expand)" [label = "aws_s3_object.test_obj_example", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] var.name" [label = "var.name", shape = "note"]
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] aws_s3_object.test_obj_example (expand)" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_object.test_obj_example (expand)"
		"[root] root" -> "[root] output.name (expand)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
		"[root] root" -> "[root] var.name"
	}
}
""".strip()

        assert module_version.get_submodules()[0].module_details.terraform_graph.strip() == """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
		"[root] aws_s3_object.test_obj_submodule (expand)" [label = "aws_s3_object.test_obj_submodule", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] var.name" [label = "var.name", shape = "note"]
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] aws_s3_object.test_obj_submodule (expand)" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_object.test_obj_submodule (expand)"
		"[root] root" -> "[root] output.name (expand)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
		"[root] root" -> "[root] var.name"
	}
}
""".strip()


    def test_graph_data_with_s3_backend(self):
        """Test extraction with s3 backend in terraform, ensuring it is overridden with local state"""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='18.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf in root
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as fh:
                    fh.write("""
terraform {
  backend "s3" {
    bucket  = "does-not-exist"
    key     = "path/to/my/key"
    region  = "us-east-1"
    profile = "thisdoesnotexistforterrareg"
  }
}

resource "aws_s3_bucket" "test_bucket" {
  name = "bucketname"
}
""")

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure Infracost output contains monthly cost
        assert module_version.module_details.terraform_graph.strip() == """
digraph {
	compound = "true"
	newrank = "true"
	subgraph "root" {
		"[root] aws_s3_bucket.test_bucket (expand)" [label = "aws_s3_bucket.test_bucket", shape = "box"]
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]" [label = "provider[\\"registry.terraform.io/hashicorp/aws\\"]", shape = "diamond"]
		"[root] aws_s3_bucket.test_bucket (expand)" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"]"
		"[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)" -> "[root] aws_s3_bucket.test_bucket (expand)"
		"[root] root" -> "[root] provider[\\"registry.terraform.io/hashicorp/aws\\"] (close)"
	}
}
""".strip()

    @skipif_unless_ci(terrareg.config.Config().INFRACOST_API_KEY == None, reason="Requires valid infracost API key")
    def test_uploading_module_with_infracost(self):
        """Test uploading a module with real Infracost API key."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='12.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                os.mkdir(os.path.join(upload_directory, 'examples'))
                os.mkdir(os.path.join(upload_directory, 'examples', 'test_example'))
                with open(os.path.join(upload_directory, 'examples', 'test_example', 'main.tf'), 'w') as fh:
                    fh.write("""
resource "aws_s3_bucket" "test" {
  name = "test-s3-bucket"
}
""")

            UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure Infracost output contains monthly cost
        assert 'totalMonthlyCost' in module_version.get_examples()[0].module_details.infracost

    def test_uploading_module_without_infracost_api_key(self):
        """Test uploading a module without Infracost API key."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='19.0.0')
        module_version.prepare_module()

        check_output = subprocess.check_output
        infracost_called = False
        def mock_check_ouput(command, *args, **kwargs):
            if command[0] == 'infracost':
                global infracost_called
                infracost_called = True
            return check_output(command, *args, **kwargs)

        # Mock subprocess.check_output to mock call to Infracost
        with mock.patch('terrareg.module_extractor.subprocess.check_output', mock_check_ouput) as mocked_check_output, \
                mock.patch('terrareg.config.Config.INFRACOST_API_KEY', None):
            with test_upload as zip_file:
                with test_upload as upload_directory:
                    os.mkdir(os.path.join(upload_directory, 'examples'))
                    os.mkdir(os.path.join(upload_directory, 'examples', 'test_example'))
                    with open(os.path.join(upload_directory, 'examples', 'test_example', 'main.tf'), 'w') as fh:
                        fh.write('#example file')

                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure tfsec output contains security issue about missing encryption key
        assert module_version.get_examples()[0].module_details.infracost == {}

        assert infracost_called == False

    def test_uploading_module_with_infracost_run_error(self):
        """Test uploading a module with Infracost throwing an error."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='13.0.0')
        module_version.prepare_module()

        check_output = subprocess.check_output
        infracost_called = False
        def mock_check_ouput(command, *args, **kwargs):
            if command[0] == 'infracost':
                raise subprocess.CalledProcessError(cmd='Unit test error', returncode=1)
            return check_output(command, *args, **kwargs)

        # Mock subprocess.check_output to mock call to Infracost
        with mock.patch('terrareg.module_extractor.subprocess.check_output', mock_check_ouput) as mocked_check_output, \
                mock.patch('terrareg.config.Config.INFRACOST_API_KEY', 'some-api-key'):
            with test_upload as zip_file:
                with test_upload as upload_directory:
                    os.mkdir(os.path.join(upload_directory, 'examples'))
                    os.mkdir(os.path.join(upload_directory, 'examples', 'test_example'))
                    with open(os.path.join(upload_directory, 'examples', 'test_example', 'main.tf'), 'w') as fh:
                        fh.write('#example file')

                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure tfsec output contains security issue about missing encryption key
        assert module_version.get_examples()[0].module_details.infracost == {}

    def test_uploading_module_with_infracost_mocked(self):
        """Test uploading a module with Infracost."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='14.0.0')
        module_version.prepare_module()

        check_output = subprocess.check_output
        def mock_check_ouput(command, *args, **kwargs):
            if command[0] == 'infracost':
                output_file_name = None
                for itx, arg in enumerate(command):
                    if arg == '--out-file':
                        output_file_name = command[itx + 1]
                if not output_file_name:
                    raise Exception('No out-file argument found in infracost command')
                # Write example output to filename
                with open(output_file_name, 'w') as output_file_fh:
                    output_file_fh.write("""
{
    "version": "0.2",
    "metadata": {
        "infracostCommand": "breakdown",
        "branch": "226-investigate-showing-costs-of-each-module-examples",
        "commit": "4822f3af904200b26ff0a3399750c76d20007f6b",
        "commitAuthorName": "Matthew John",
        "commitAuthorEmail": "matthew@dockstudios.co.uk",
        "commitTimestamp": "2022-08-17T06:58:57Z",
        "commitMessage": "Add screenshot of example page to README",
        "vcsRepoUrl": "https://gitlab.dockstudios.co.uk:2222/pub/terrareg.git"
    },
    "currency": "USD",
    "projects": [
        {
            "name": "2222/pub/terrareg/example/cost_example",
            "metadata": {
                "path": ".",
                "type": "terraform_dir",
                "vcsSubPath": "example/cost_example"
            },
            "pastBreakdown": {
                "resources": [],
                "totalHourlyCost": "0",
                "totalMonthlyCost": "0"
            },
            "breakdown": {
                "resources": [
                    {
                        "name": "aws_instance.test",
                        "metadata": {
                            "calls": [
                                {
                                    "blockName": "aws_instance.test",
                                    "filename": "main.tf"
                                }
                            ],
                            "filename": "main.tf"
                        },
                        "hourlyCost": "0.0842958904109589",
                        "monthlyCost": "61.536",
                        "costComponents": [
                            {
                                "name": "Instance usage (Linux/UNIX, on-demand, t3.large)",
                                "unit": "hours",
                                "hourlyQuantity": "1",
                                "monthlyQuantity": "730",
                                "price": "0.0832",
                                "hourlyCost": "0.0832",
                                "monthlyCost": "60.736"
                            },
                            {
                                "name": "CPU credits",
                                "unit": "vCPU-hours",
                                "hourlyQuantity": "0",
                                "monthlyQuantity": "0",
                                "price": "0.05",
                                "hourlyCost": "0",
                                "monthlyCost": "0"
                            }
                        ],
                        "subresources": [
                            {
                                "name": "root_block_device",
                                "metadata": {},
                                "hourlyCost": "0.0010958904109589",
                                "monthlyCost": "0.8",
                                "costComponents": [
                                    {
                                        "name": "Storage (general purpose SSD, gp2)",
                                        "unit": "GB",
                                        "hourlyQuantity": "0.010958904109589",
                                        "monthlyQuantity": "8",
                                        "price": "0.1",
                                        "hourlyCost": "0.0010958904109589",
                                        "monthlyCost": "0.8"
                                    }
                                ]
                            }
                        ]
                    }
                ],
                "totalHourlyCost": "0.0842958904109589",
                "totalMonthlyCost": "61.536"
            },
            "diff": {
                "resources": [
                    {
                        "name": "aws_instance.test",
                        "metadata": {},
                        "hourlyCost": "0.0842958904109589",
                        "monthlyCost": "61.536",
                        "costComponents": [
                            {
                                "name": "Instance usage (Linux/UNIX, on-demand, t3.large)",
                                "unit": "hours",
                                "hourlyQuantity": "1",
                                "monthlyQuantity": "730",
                                "price": "0.0832",
                                "hourlyCost": "0.0832",
                                "monthlyCost": "60.736"
                            },
                            {
                                "name": "CPU credits",
                                "unit": "vCPU-hours",
                                "hourlyQuantity": "0",
                                "monthlyQuantity": "0",
                                "price": "0.05",
                                "hourlyCost": "0",
                                "monthlyCost": "0"
                            }
                        ],
                        "subresources": [
                            {
                                "name": "root_block_device",
                                "metadata": {},
                                "hourlyCost": "0.0010958904109589",
                                "monthlyCost": "0.8",
                                "costComponents": [
                                    {
                                        "name": "Storage (general purpose SSD, gp2)",
                                        "unit": "GB",
                                        "hourlyQuantity": "0.010958904109589",
                                        "monthlyQuantity": "8",
                                        "price": "0.1",
                                        "hourlyCost": "0.0010958904109589",
                                        "monthlyCost": "0.8"
                                    }
                                ]
                            }
                        ]
                    }
                ],
                "totalHourlyCost": "0.0842958904109589",
                "totalMonthlyCost": "61.536"
            },
            "summary": {
                "totalDetectedResources": 1,
                "totalSupportedResources": 1,
                "totalUnsupportedResources": 0,
                "totalUsageBasedResources": 1,
                "totalNoPriceResources": 0,
                "unsupportedResourceCounts": {},
                "noPriceResourceCounts": {}
            }
        }
    ],
    "totalHourlyCost": "0.0842958904109589",
    "totalMonthlyCost": "61.536",
    "pastTotalHourlyCost": "0",
    "pastTotalMonthlyCost": "0",
    "diffTotalHourlyCost": "0.0842958904109589",
    "diffTotalMonthlyCost": "61.536",
    "timeGenerated": "2022-08-17T18:39:55.964808023Z",
    "summary": {
        "totalDetectedResources": 1,
        "totalSupportedResources": 1,
        "totalUnsupportedResources": 0,
        "totalUsageBasedResources": 1,
        "totalNoPriceResources": 0,
        "unsupportedResourceCounts": {},
        "noPriceResourceCounts": {}
    }
}
""")
                return None
            return check_output(command, *args, **kwargs)

        # Mock subprocess.check_output to mock call to Infracost
        with mock.patch('terrareg.module_extractor.subprocess.check_output', mock_check_ouput) as mocked_check_output, \
                mock.patch('terrareg.config.Config.INFRACOST_API_KEY', 'test-infracost-api-key'):
            with test_upload as zip_file:
                with test_upload as upload_directory:
                    os.mkdir(os.path.join(upload_directory, 'examples'))
                    os.mkdir(os.path.join(upload_directory, 'examples', 'test_example'))
                    with open(os.path.join(upload_directory, 'examples', 'test_example', 'main.tf'), 'w') as fh:
                        fh.write('#example file')

                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure tfsec output contains security issue about missing encryption key
        assert module_version.get_examples()[0].module_details.infracost == {
            'currency': 'USD',
            'diffTotalHourlyCost': '0.0842958904109589',
            'diffTotalMonthlyCost': '61.536',
            'metadata': {'branch': '226-investigate-showing-costs-of-each-module-examples',
                        'commit': '4822f3af904200b26ff0a3399750c76d20007f6b',
                        'commitAuthorEmail': 'matthew@dockstudios.co.uk',
                        'commitAuthorName': 'Matthew John',
                        'commitMessage': 'Add screenshot of example page to README',
                        'commitTimestamp': '2022-08-17T06:58:57Z',
                        'infracostCommand': 'breakdown',
                        'vcsRepoUrl': 'https://gitlab.dockstudios.co.uk:2222/pub/terrareg.git'},
            'pastTotalHourlyCost': '0',
            'pastTotalMonthlyCost': '0',
            'projects': [{'breakdown': {'resources': [{'costComponents': [{'hourlyCost': '0.0832',
                                                                            'hourlyQuantity': '1',
                                                                            'monthlyCost': '60.736',
                                                                            'monthlyQuantity': '730',
                                                                            'name': 'Instance '
                                                                                    'usage '
                                                                                    '(Linux/UNIX, '
                                                                                    'on-demand, '
                                                                                    't3.large)',
                                                                            'price': '0.0832',
                                                                            'unit': 'hours'},
                                                                            {'hourlyCost': '0',
                                                                            'hourlyQuantity': '0',
                                                                            'monthlyCost': '0',
                                                                            'monthlyQuantity': '0',
                                                                            'name': 'CPU '
                                                                                    'credits',
                                                                            'price': '0.05',
                                                                            'unit': 'vCPU-hours'}],
                                                        'hourlyCost': '0.0842958904109589',
                                                        'metadata': {'calls': [{'blockName': 'aws_instance.test',
                                                                                'filename': 'main.tf'}],
                                                                    'filename': 'main.tf'},
                                                        'monthlyCost': '61.536',
                                                        'name': 'aws_instance.test',
                                                        'subresources': [{'costComponents': [{'hourlyCost': '0.0010958904109589',
                                                                                            'hourlyQuantity': '0.010958904109589',
                                                                                            'monthlyCost': '0.8',
                                                                                            'monthlyQuantity': '8',
                                                                                            'name': 'Storage '
                                                                                                    '(general '
                                                                                                    'purpose '
                                                                                                    'SSD, '
                                                                                                    'gp2)',
                                                                                            'price': '0.1',
                                                                                            'unit': 'GB'}],
                                                                        'hourlyCost': '0.0010958904109589',
                                                                        'metadata': {},
                                                                        'monthlyCost': '0.8',
                                                                        'name': 'root_block_device'}]}],
                                        'totalHourlyCost': '0.0842958904109589',
                                        'totalMonthlyCost': '61.536'},
                            'diff': {'resources': [{'costComponents': [{'hourlyCost': '0.0832',
                                                                        'hourlyQuantity': '1',
                                                                        'monthlyCost': '60.736',
                                                                        'monthlyQuantity': '730',
                                                                        'name': 'Instance '
                                                                                'usage '
                                                                                '(Linux/UNIX, '
                                                                                'on-demand, '
                                                                                't3.large)',
                                                                        'price': '0.0832',
                                                                        'unit': 'hours'},
                                                                    {'hourlyCost': '0',
                                                                        'hourlyQuantity': '0',
                                                                        'monthlyCost': '0',
                                                                        'monthlyQuantity': '0',
                                                                        'name': 'CPU '
                                                                                'credits',
                                                                        'price': '0.05',
                                                                        'unit': 'vCPU-hours'}],
                                                    'hourlyCost': '0.0842958904109589',
                                                    'metadata': {},
                                                    'monthlyCost': '61.536',
                                                    'name': 'aws_instance.test',
                                                    'subresources': [{'costComponents': [{'hourlyCost': '0.0010958904109589',
                                                                                        'hourlyQuantity': '0.010958904109589',
                                                                                        'monthlyCost': '0.8',
                                                                                        'monthlyQuantity': '8',
                                                                                        'name': 'Storage '
                                                                                                '(general '
                                                                                                'purpose '
                                                                                                'SSD, '
                                                                                                'gp2)',
                                                                                        'price': '0.1',
                                                                                        'unit': 'GB'}],
                                                                    'hourlyCost': '0.0010958904109589',
                                                                    'metadata': {},
                                                                    'monthlyCost': '0.8',
                                                                    'name': 'root_block_device'}]}],
                                    'totalHourlyCost': '0.0842958904109589',
                                    'totalMonthlyCost': '61.536'},
                            'metadata': {'path': '.',
                                        'type': 'terraform_dir',
                                        'vcsSubPath': 'example/cost_example'},
                            'name': '2222/pub/terrareg/example/cost_example',
                            'pastBreakdown': {'resources': [],
                                            'totalHourlyCost': '0',
                                            'totalMonthlyCost': '0'},
                            'summary': {'noPriceResourceCounts': {},
                                        'totalDetectedResources': 1,
                                        'totalNoPriceResources': 0,
                                        'totalSupportedResources': 1,
                                        'totalUnsupportedResources': 0,
                                        'totalUsageBasedResources': 1,
                                        'unsupportedResourceCounts': {}}}],
            'summary': {'noPriceResourceCounts': {},
                        'totalDetectedResources': 1,
                        'totalNoPriceResources': 0,
                        'totalSupportedResources': 1,
                        'totalUnsupportedResources': 0,
                        'totalUsageBasedResources': 1,
                        'unsupportedResourceCounts': {}},
            'timeGenerated': '2022-08-17T18:39:55.964808023Z',
            'totalHourlyCost': '0.0842958904109589',
            'totalMonthlyCost': '61.536',
            'version': '0.2',
        }

    def test_uploading_module_with_reference_to_inaccessible_remote_module(self):
        """Test uploading a module with reference to inaccessible external module."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='15.0.0')
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

    def test_uploading_module_with_custom_tab_files(self):
        """Test uploading a module with custom tab files."""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='16.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                with open(os.path.join(upload_directory, 'UNITTEST_NO_EXT'), 'w') as fh:
                    fh.write("""
# Unit test no extension
""")
                with open(os.path.join(upload_directory, 'unittest_backup_file'), 'w') as fh:
                    fh.write("""
# Unit test backup file name
""")

                with open(os.path.join(upload_directory, 'unittest_file.md'), 'w') as fh:
                    fh.write("""
# Unit test markdown file
""")

            with mock.patch('terrareg.config.Config.ADDITIONAL_MODULE_TABS',
                            """[["Does Not Exist", ["does_not_exist"]],
                                ["No Extension", ["UNITTEST_NO_EXT"]],
                                ["Multiple files", ["primary_file", "unittest_backup_file"]],
                                ["markdown file", ["unittest_file.md"]]]"""):
                UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

        # Ensure files were created correctly
        expected_files = {
            'UNITTEST_NO_EXT': b'\n# Unit test no extension\n',
            'unittest_backup_file': b'\n# Unit test backup file name\n',
            'unittest_file.md': b'\n# Unit test markdown file\n'
        }
        assert {
            file.path: file.content
            for file in module_version.module_version_files
        } == expected_files


    @skipif_unless_ci((not os.path.isfile('/usr/bin/zip')), reason="Zip must be installed on system")
    def test_upload_malicious_zip(self):
        """Test basic module upload with single depth."""
        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='17.0.0')
        module_version.prepare_module()

        # Root directory, outside of root path
        temp_directory = mkdtemp()
        zip_source_directory = os.path.join(temp_directory, 'upload')
        os.mkdir(zip_source_directory)

        # Create secure file
        secure_file = os.path.join(temp_directory, 'secure_file')
        with open(secure_file, 'w'):
            pass

        subprocess.check_call(['zip', '-r', '../upload.zip', '../secure_file'], cwd=zip_source_directory)

        class MockUpload:

            def __init__(self, source):
                self._source = source
                self.filename = 'upload.zip'
            
            def save(self, dest):
                shutil.copy(self._source, dest)

        mock_upload = MockUpload(os.path.join(temp_directory, 'upload.zip'))

        with ApiUploadModuleExtractor(upload_file=mock_upload, module_version=module_version) as me:

            # Ensure the upload raises an exception due to zip member being outside
            # the root directory
            with pytest.raises(terrareg.utils.PathIsNotWithinBaseDirectoryError):
                # Perform base module upload
                me.process_upload()

    def test_generate_archive(self):
        """Test _generate_archive method during module extraction"""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='21.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines(UploadTestModule.VALID_MAIN_TF_FILE.strip())
                # Create main.tf
                with open(os.path.join(upload_directory, 'second_file.tf'), 'w') as second_tf_fh:
                    second_tf_fh.writelines("""
                        # Test output file
                    """.strip())
                with open(os.path.join(upload_directory, '.hidden-file.tf'), 'w') as hidden_file_fh:
                    hidden_file_fh.writelines("""
                        # Hidden file
                    """.strip())
                os.mkdir(os.path.join(upload_directory, 'subdir'))
                with open(os.path.join(upload_directory, 'subdir', 'nested-file.tf'), 'w') as nested_file_fh:
                    nested_file_fh.writelines("""
                        # Nested file
                    """.strip())

                # Create .git directory with some files
                os.mkdir(os.path.join(upload_directory, '.git'))
                with open(os.path.join(upload_directory, '.git', 'example-git-file'), 'w') as git_file:
                    git_file.writelines("# Test git file")

            temp_dir = tempfile.mkdtemp()
            os.mkdir(os.path.join(temp_dir, 'modules'))

            # Mock tempfile to obtain temporary directory used for generating archives
            generate_archive_temp_dir = None
            def mock_temporary_directory(suffix=None, *args, **kwargs):
                nonlocal generate_archive_temp_dir
                temp_dir_ret = TemporaryDirectory(suffix=suffix, *args, **kwargs)
                if suffix == "generate-archive":
                    generate_archive_temp_dir = temp_dir_ret.name
                return temp_dir_ret

            try:
                local_storage = terrareg.file_storage.LocalFileStorage(base_directory=temp_dir)
                mock_local_file_storage = mock.MagicMock(wraps=local_storage)
                with mock.patch('terrareg.config.Config.DELETE_EXTERNALLY_HOSTED_ARTIFACTS', False), \
                        mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir), \
                        mock.patch('terrareg.file_storage.FileStorageFactory.get_file_storage', return_value=mock_local_file_storage), \
                        mock.patch('tempfile.TemporaryDirectory', side_effect=mock_temporary_directory):

                    UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

                    # Ensure the module version tar archive path looks correct
                    assert module_version.archive_path_tar_gz == "/modules/testprocessupload/test-module/aws/21.0.0/source.tar.gz"
                    full_tar_path = os.path.join(temp_dir, module_version.archive_path_tar_gz.lstrip(os.path.sep))
                    assert os.path.isfile(full_tar_path)

                    # Inspect generated archive
                    with tarfile.open(full_tar_path, "r:gz") as tar_fh:
                        # Create map of all files paths against the file contents for verification
                        tar_contents = {
                            member.name: tar_fh.extractfile(member).read().decode('utf-8')
                            for member in tar_fh.getmembers()
                            if tar_fh.extractfile(member)
                        }

                        assert tar_contents == {
                            '.hidden-file.tf': '# Hidden file',
                            'main.tf': UploadTestModule.VALID_MAIN_TF_FILE.strip(),
                            'second_file.tf': '# Test output file',
                            'subdir/nested-file.tf': '# Nested file',
                        }

                    # Ensure zip file was created correctly
                    assert module_version.archive_path_zip == "/modules/testprocessupload/test-module/aws/21.0.0/source.zip"

                    full_zip_path = os.path.join(temp_dir, module_version.archive_path_zip.lstrip(os.path.sep))
                    assert os.path.isfile(full_zip_path)

                    # Inspect generated archive
                    with zipfile.ZipFile(full_zip_path) as z:
                        zip_contents = {
                            fileobj.filename: z.open(fileobj.filename).read().decode('utf-8')
                            for fileobj in z.infolist()
                            if not fileobj.is_dir()
                        }

                        assert zip_contents == {
                            '.hidden-file.tf': '# Hidden file',
                            'main.tf': UploadTestModule.VALID_MAIN_TF_FILE.strip(),
                            'second_file.tf': '# Test output file',
                            'subdir/nested-file.tf': '# Nested file',
                        }

                    mock_local_file_storage.make_directory.assert_called_once_with("/modules/testprocessupload/test-module/aws/21.0.0")
                    mock_local_file_storage.upload_file.assert_has_calls(calls=[
                        mock.call(os.path.join(generate_archive_temp_dir, "source.tar.gz"), '/modules/testprocessupload/test-module/aws/21.0.0', 'source.tar.gz'),
                        mock.call(os.path.join(generate_archive_temp_dir, "source.zip"), '/modules/testprocessupload/test-module/aws/21.0.0', 'source.zip')
                    ])

            finally:
                shutil.rmtree(temp_dir)

    def test_terraform_version(self):
        """Test terraform_version data extraction"""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='21.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
terraform {
  required_version = ">= 1.0, < 1.1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 5.1.0"
    }
  }
}
""")

                # Create module with version data
                os.mkdir(os.path.join(upload_directory, 'modules'))
                os.mkdir(os.path.join(upload_directory, 'modules', 'test-module'))
                with open(os.path.join(upload_directory, 'modules', 'test-module', 'versions.tf'), 'w') as module_examples:
                    module_examples.writelines("""
terraform {
  required_version = ">= 1.0, < 1.2.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 5.2.0"
    }
  }
}
""")

                # Create example with version data
                os.mkdir(os.path.join(upload_directory, 'examples'))
                os.mkdir(os.path.join(upload_directory, 'examples', 'test-example'))
                with open(os.path.join(upload_directory, 'examples', 'test-example', 'versions.tf'), 'w') as example_versions:
                    example_versions.writelines("""
terraform {
  required_version = ">= 1.0, < 1.3.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 5.3.0"
    }
  }
}
""")

            temp_dir = tempfile.mkdtemp()
            try:

                with mock.patch('terrareg.config.Config.DELETE_EXTERNALLY_HOSTED_ARTIFACTS', False), \
                        mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):
                    UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

                    terraform_arch = "linux_amd64" if platform.uname().processor == "x86_64" else "linux_arm64"

                    # Ensure the module version terraform_version argument is correct
                    assert module_version.module_details.terraform_version == {
                        'platform': terraform_arch,
                        'provider_selections': {
                            'registry.terraform.io/hashicorp/aws': '5.0.1'
                        },
                        'terraform_outdated': True,
                        'terraform_version': '1.0.11'
                    }

                    # Ensure the sub-module terraform_version argument is correct
                    assert len(module_version.get_submodules()) == 1
                    assert module_version.get_submodules()[0].module_details.terraform_version == {
                        'platform': terraform_arch,
                        'provider_selections': {
                            'registry.terraform.io/hashicorp/aws': '5.1.0'
                        },
                        'terraform_outdated': True,
                        'terraform_version': '1.1.9'
                    }

                    # Ensure the example terraform_version argument is correct
                    assert len(module_version.get_examples()) == 1
                    assert module_version.get_examples()[0].module_details.terraform_version == {
                        'platform': terraform_arch,
                        'provider_selections': {
                            'registry.terraform.io/hashicorp/aws': '5.2.0'
                        },
                        'terraform_outdated': True,
                        'terraform_version': '1.2.9'
                    }

            finally:
                shutil.rmtree(temp_dir)

    def test_terraform_modules(self):
        """Test terraform_modules data extraction"""
        test_upload = UploadTestModule()

        namespace = Namespace.get(name='testprocessupload', create=True)
        module = Module(namespace=namespace, name='test-module')
        module_provider = ModuleProvider.get(module=module, name='aws', create=True)
        module_version = ModuleVersion(module_provider=module_provider, version='21.0.0')
        module_version.prepare_module()

        with test_upload as zip_file:
            with test_upload as upload_directory:
                # Create main.tf
                with open(os.path.join(upload_directory, 'main.tf'), 'w') as main_tf_fh:
                    main_tf_fh.writelines("""
# Call external module
module "external_module" {
  source  = "terraform-aws-modules/iam/aws"
  version = "5.30.0"
}

# Call submodule
module "internal_module" {
  source = "./internal_module"
}
""")

                os.mkdir(os.path.join(upload_directory, "internal_module"))
                with open(os.path.join(upload_directory, "internal_module", "main.tf"), "w") as internal_tf:
                    internal_tf.write("""
output "test" { value = "test" }
""")

                # Create module with version data
                os.mkdir(os.path.join(upload_directory, 'modules'))
                os.mkdir(os.path.join(upload_directory, 'modules', 'test-module'))
                with open(os.path.join(upload_directory, 'modules', 'test-module', 'main.tf'), 'w') as module_examples:
                    module_examples.writelines("""
# Call root module
module "root_module" {
  source = "../../"
}
# Call internal module
module "internal_module" {
  source = "../../internal_module"
}
""")

                # Create example with version data
                os.mkdir(os.path.join(upload_directory, 'examples'))
                os.mkdir(os.path.join(upload_directory, 'examples', 'test-example'))
                with open(os.path.join(upload_directory, 'examples', 'test-example', 'main.tf'), 'w') as example_versions:
                    example_versions.writelines("""
# Call root module
module "root" {
  source = "../../"
}
""")

            temp_dir = tempfile.mkdtemp()
            try:

                with mock.patch('terrareg.config.Config.DELETE_EXTERNALLY_HOSTED_ARTIFACTS', False), \
                        mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):
                    UploadTestModule.upload_module_version(module_version=module_version, zip_file=zip_file)

                    # Ensure the module version terraform_modules argument is correct
                    assert module_version.module_details.terraform_modules == {
                        'Modules': [
                            {
                                'Dir': '.',
                                'Key': '',
                                'Source': ''},
                            {
                                'Dir': '.terraform/modules/external_module',
                                'Key': 'external_module',
                                'Source': 'registry.terraform.io/terraform-aws-modules/iam/aws',
                                'Version': '5.30.0'},
                            {
                                'Dir': 'internal_module',
                                'Key': 'internal_module',
                                'Source': './internal_module'
                            }
                        ]
                    }

                    # Ensure the sub-module terraform_modules argument is correct
                    assert len(module_version.get_submodules()) == 1
                    assert module_version.get_submodules()[0].module_details.terraform_modules == {
                        'Modules': [
                            {
                                'Dir': '.',
                                'Key': '',
                                'Source': ''
                            },
                            {
                                'Dir': '../../internal_module',
                                'Key': 'internal_module',
                                'Source': '../../internal_module'
                            },
                            {
                                'Dir': '../..',
                                'Key': 'root_module',
                                'Source': '../..'
                            },
                            {
                                'Dir': '.terraform/modules/root_module.external_module',
                                'Key': 'root_module.external_module',
                                'Source': 'registry.terraform.io/terraform-aws-modules/iam/aws',
                                'Version': '5.30.0'
                            },
                            {
                                'Dir': '../../internal_module',
                                'Key': 'root_module.internal_module',
                                'Source': './internal_module'
                            }
                        ]
                    }

                    # Ensure the example terraform_modules argument is correct
                    assert len(module_version.get_examples()) == 1
                    assert module_version.get_examples()[0].module_details.terraform_modules == {
                        'Modules': [
                            {
                                'Dir': '.',
                                'Key': '',
                                'Source': ''
                            },
                            {
                                'Dir': '../..',
                                'Key': 'root',
                                'Source': '../..'
                            },
                            {
                                'Dir': '.terraform/modules/root.external_module',
                                'Key': 'root.external_module',
                                'Source': 'registry.terraform.io/terraform-aws-modules/iam/aws',
                                'Version': '5.30.0'
                            },
                            {
                                'Dir': '../../internal_module',
                                'Key': 'root.internal_module',
                                'Source': './internal_module'
                            }
                        ]
                    }

            finally:
                shutil.rmtree(temp_dir)

