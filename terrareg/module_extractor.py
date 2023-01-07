"""Provide extraction method of modules."""

from contextlib import contextmanager
import os
import threading
from typing import Type
import tempfile
import zipfile
import tarfile
import subprocess
import json
import datetime
import shutil
import re
import glob
import pathlib

from werkzeug.utils import secure_filename
import magic
from bs4 import BeautifulSoup
import markdown

from terrareg.models import BaseSubmodule, Example, ExampleFile, ModuleVersion, Submodule, ModuleDetails, ModuleVersionFile
from terrareg.database import Database
from terrareg.errors import (
    UnableToProcessTerraformError,
    UnknownFiletypeError,
    InvalidTerraregMetadataFileError,
    MetadataDoesNotContainRequiredAttributeError,
    GitCloneError,
    UnableToGetGlobalTerraformLockError,
    TerraformVersionSwitchError
)
from terrareg.utils import PathDoesNotExistError, safe_iglob, safe_join_paths
from terrareg.config import Config


class ModuleExtractor:
    """Provide extraction method of moduls."""

    TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']
    TERRAFORM_LOCK = threading.Lock()

    def __init__(self, module_version: ModuleVersion):
        """Create temporary directories and store member variables."""
        self._module_version = module_version
        self._extract_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._upload_directory = tempfile.TemporaryDirectory()  # noqa: R1732

    @property
    def terraform_binary(self):
        """Return path of terraform binary"""
        return os.path.join(os.getcwd(), "bin", "terraform")

    @property
    def extract_directory(self):
        """Return path of extract directory."""
        return self._extract_directory.name

    @property
    def module_directory(self):
        """Return path of module directory, based on configured git path."""
        if self._module_version.git_path:
            return safe_join_paths(self._extract_directory.name, self._module_version.git_path)
        else:
            return self._extract_directory.name

    @property
    def upload_directory(self):
        """Return path of extract directory."""
        return self._upload_directory.name

    def __enter__(self):
        """Run enter of upstream context managers."""
        self._extract_directory.__enter__()
        self._upload_directory.__enter__()
        return self

    def __exit__(self, *args, **kwargs):
        """Run exit of upstream context managers."""
        self._extract_directory.__exit__(*args, **kwargs)
        self._upload_directory.__exit__(*args, **kwargs)

    @staticmethod
    def _run_terraform_docs(module_path):
        """Run terraform docs and return output."""
        # Check if a terraform docs configuration file exists and remove it
        for terraform_docs_config_file in ['.terraform-docs.yml', '.terraform-docs.yaml']:
            terraform_docs_config_path = os.path.join(module_path, terraform_docs_config_file)
            if os.path.isfile(terraform_docs_config_path):
                os.unlink(terraform_docs_config_path)

        try:
            terradocs_output = subprocess.check_output(['terraform-docs', 'json', module_path])
        except subprocess.CalledProcessError as exc:
            raise UnableToProcessTerraformError('An error occurred whilst processing the terraform code.')

        return json.loads(terradocs_output)

    @contextmanager
    def _switch_terraform_versions(self, module_path):
        """Switch terraform to required version for module"""
        # Wait for global lock on terraform, so that only
        # instance can run terraform at a time
        if not ModuleExtractor.TERRAFORM_LOCK.acquire(blocking=True, timeout=60):
            raise UnableToGetGlobalTerraformLockError(
                "Unable to obtain global Terraform lock in 60 seconds"
            )
        try:
            default_terraform_version = Config().DEFAULT_TERRAFORM_VERSION
            tfswitch_env = os.environ.copy()

            if default_terraform_version:
                tfswitch_env["TF_VERSION"] = default_terraform_version

            # Run tfswitch
            try:
                subprocess.check_output(
                    ["tfswitch", "--mirror", Config().TERRAFORM_ARCHIVE_MIRROR, "--bin", self.terraform_binary],
                    env=tfswitch_env,
                    cwd=module_path
                )
            except subprocess.CalledProcessError as exc:
                print("An error occured whilst running tfswitch:", str(exc))
                raise TerraformVersionSwitchError("An error occurred whilst initialising Terraform version")

            yield
        finally:
            ModuleExtractor.TERRAFORM_LOCK.release()

    def _run_tfsec(self, module_path):
        """Run tfsec and return output."""
        try:
            raw_output = subprocess.check_output([
                'tfsec',
                '--ignore-hcl-errors', '--format', 'json', '--no-module-downloads', '--soft-fail',
                '--no-colour', '--include-ignored', '--include-passed', '--disable-grouping',
                module_path
            ])
        except subprocess.CalledProcessError as exc:
            raise UnableToProcessTerraformError('An error occurred whilst performing security scan of code.')

        tfsec_results = json.loads(raw_output)

        # Strip the extraction directory from all paths in results
        if tfsec_results['results']:
            for result in tfsec_results['results']:
                result['location']['filename'] = result['location']['filename'].replace(self._extract_directory.name + '/', '')

        return tfsec_results

    def _run_tf_init(self, module_path):
        """Perform terraform init"""
        try:
            subprocess.check_call([self.terraform_binary, "init"], cwd=module_path)
        except subprocess.CalledProcessError:
            return False
        return True

    def _get_graph_data(self, module_path):
        """Run inframap and generate graphiz"""
        try:
            terraform_graph_data = subprocess.check_output(
                [self.terraform_binary, "graph"],
                cwd=module_path
            )
        except subprocess.CalledProcessError as exc:
            print("Failed to generate Terraform graph data:", str(exc))
            raise UnableToProcessTerraformError("Failed to generate Terraform graph data")

        terraform_graph_data = terraform_graph_data.decode("utf-8")

        return terraform_graph_data

    @staticmethod
    def _get_readme_content(module_path):
        """Obtain README contents for given module."""
        try:
            readme_path = safe_join_paths(module_path, 'README.md', is_file=True)
        except PathDoesNotExistError:
            # If no README found, return None
            return None

        with open(readme_path, 'r') as readme_fd:
            return ''.join(readme_fd.readlines())

    def _get_terrareg_metadata(self, module_path):
        """Obtain terrareg metadata for module, if it exists."""
        terrareg_metadata = {}
        for terrareg_file in self.TERRAREG_METADATA_FILES:
            try:
                path = safe_join_paths(module_path, terrareg_file, is_file=True)
            except PathDoesNotExistError:
                continue

            with open(path, 'r') as terrareg_fh:
                try:
                    terrareg_metadata = json.loads(''.join(terrareg_fh.readlines()))
                except:
                    raise InvalidTerraregMetadataFileError(
                        'An error occured whilst processing the terrareg metadata file.'
                    )

            # Remove the meta-data file, so it is not added to the archive
            os.unlink(path)

            break

        for required_attr in Config().REQUIRED_MODULE_METADATA_ATTRIBUTES:
            if not terrareg_metadata.get(required_attr, None):
                raise MetadataDoesNotContainRequiredAttributeError(
                    'terrareg metadata file does not contain required attribute: {}'.format(required_attr)
                )

        return terrareg_metadata

    def _extract_additional_tab_files(self):
        """Extract addition files for populating tabs in UI"""
        config = Config()

        files_extracted = []
        # Iterate through all files of all additionally defined tabs
        for tab_config in json.loads(config.ADDITIONAL_MODULE_TABS):
            for file_name in tab_config[1]:
                path = safe_join_paths(self.extract_directory, file_name)
                # Check if file exists
                if file_name in files_extracted or not os.path.exists(path):
                    continue

                # Read file contents and create DB record
                with open(path, 'r') as fh:
                    file_content = ''.join(fh.readlines())

                module_version_file = ModuleVersionFile.create(module_version=self._module_version, path=file_name)
                module_version_file.update_attributes(content=file_content)

    def _generate_archive(self):
        """Generate archive of extracted module"""
        # Create tar.gz
        with tarfile.open(self._module_version.archive_path_tar_gz, "w:gz") as tar:
            tar.add(self.extract_directory, arcname='', recursive=True)
        # Create zip
        shutil.make_archive(
            re.sub(r'\.zip$', '', self._module_version.archive_path_zip),
            'zip',
            self.extract_directory)

    def _create_module_details(self, readme_content, terraform_docs, tfsec, terraform_graph, infracost=None):
        """Create module details row."""
        module_details = ModuleDetails.create()
        module_details.update_attributes(
            readme_content=readme_content,
            terraform_docs=json.dumps(terraform_docs),
            tfsec=json.dumps(tfsec),
            infracost=json.dumps(infracost) if infracost else None,
            terraform_graph=terraform_graph
        )
        return module_details

    def _insert_database(
        self,
        description: str,
        readme_content: str,
        terraform_docs: dict,
        tfsec: dict,
        terrareg_metadata: dict,
        terraform_graph: str) -> int:
        """Insert module into DB, overwrite any pre-existing"""
        # Create module details row
        module_details = self._create_module_details(
            terraform_docs=terraform_docs,
            readme_content=readme_content,
            tfsec=tfsec,
            terraform_graph=terraform_graph
        )
        
        # Update attributes of module_version in database
        self._module_version.update_attributes(
            module_details_id=module_details.pk,

            published_at=datetime.datetime.now(),

            # Terrareg meta-data
            owner=terrareg_metadata.get('owner', None),
            description=description,
            repo_clone_url_template=terrareg_metadata.get('repo_clone_url', None),
            repo_browse_url_template=terrareg_metadata.get('repo_browse_url', None),
            repo_base_url_template=terrareg_metadata.get('repo_base_url', None),
            variable_template=json.dumps(terrareg_metadata.get('variable_template', {})),
            published=Config().AUTO_PUBLISH_MODULE_VERSIONS,
            internal=terrareg_metadata.get('internal', False)
        )

    def _process_submodule(self, submodule: BaseSubmodule):
        """Process submodule."""
        submodule_dir = safe_join_paths(self.module_directory, submodule.path)

        tf_docs = self._run_terraform_docs(submodule_dir)
        tfsec = self._run_tfsec(submodule_dir)
        readme_content = self._get_readme_content(submodule_dir)

        terraform_graph = None
        with self._switch_terraform_versions(submodule_dir):
            if self._run_tf_init(submodule_dir):
                terraform_graph = self._get_graph_data(submodule_dir)

        infracost = None
        # Run infracost on examples, if API key is set
        if isinstance(submodule, Example) and Config().INFRACOST_API_KEY:
            try:
                infracost = self._run_infracost(example=submodule)
            except UnableToProcessTerraformError as exc:
                print('An error occured whilst running infracost against example')

        # Create module details row
        module_details = self._create_module_details(
            terraform_docs=tf_docs,
            readme_content=readme_content,
            tfsec=tfsec,
            infracost=infracost,
            terraform_graph=terraform_graph
        )

        submodule.update_attributes(
            module_details_id=module_details.pk
        )

        if isinstance(submodule, Example):
            self._extract_example_files(example=submodule)

    def _run_infracost(self, example: Example):
        """Run infracost to obtain cost of examples."""
        # Ensure example path is within root module
        safe_join_paths(self.module_directory, example.path)

        infracost_env = dict(os.environ)
        if Config().DOMAIN_NAME:
            infracost_env['INFRACOST_TERRAFORM_CLOUD_TOKEN'] = Config().INTERNAL_EXTRACTION_ANALYITCS_TOKEN
            infracost_env['INFRACOST_TERRAFORM_CLOUD_HOST'] = Config().DOMAIN_NAME

        # Create temporary file safely and immediately close to
        # pass path to infracost
        with tempfile.NamedTemporaryFile(delete=False) as output_file:
            output_file.close()
            try:
                subprocess.check_output(
                    ['infracost', 'breakdown', '--path', example.path,
                     '--format', 'json', '--out-file', output_file.name],
                    cwd=self.module_directory,
                    env=infracost_env
                )
            except subprocess.CalledProcessError as exc:
                raise UnableToProcessTerraformError('An error occurred whilst performing cost analysis of code.')

            with open(output_file.name, 'r') as output_file_fh:
                infracost_result = json.load(output_file_fh)

            os.unlink(output_file.name)

        return infracost_result

    def _extract_example_files(self, example: Example):
        """Extract all terraform files in example and insert into DB"""
        example_base_dir = safe_join_paths(self.module_directory, example.path)
        for tf_file_path in safe_iglob(base_dir=example_base_dir,
                                       pattern='*.tf',
                                       recursive=False,
                                       is_file=True):
            # Remove extraction directory from file path
            tf_file = re.sub('^{}/'.format(self.module_directory), '', tf_file_path)

            # Obtain contents of file
            with open(tf_file_path, 'r') as file_fd:
                content = ''.join(file_fd.readlines())

            # Create example file and update content attribute
            example_file = ExampleFile.create(example=example, path=tf_file)
            example_file.update_attributes(
                content=content
            )

    def _scan_submodules(self, subdirectory: str, submodule_class: Type[BaseSubmodule]):
        """Scan for submodules and extract details."""
        try:
            submodule_base_directory = safe_join_paths(self.module_directory, subdirectory, is_dir=True)
        except PathDoesNotExistError:
            # If the modules directory does not exist,
            # ignore and return
            print('No modules directory found')
            return

        module_directory_re = re.compile('^{}'.format(
            re.escape(
                '{0}/'.format(self.module_directory)
            )
        ))

        submodules = []
        # Search for all subdirectories containing terraform
        for terraform_file_path in glob.iglob('{modules_path}/**/*.tf'.format(modules_path=submodule_base_directory), recursive=True):
            # Get parent directory of terraform file
            tf_file_path_obj = pathlib.Path(terraform_file_path)
            submodule_dir = str(tf_file_path_obj.parent)

            # Strip extraction directory base path from submodule directory
            # to return relative path from base of extracted module
            submodule_name = module_directory_re.sub('', submodule_dir)

            # Check submodule is not in the root of the submodules
            if not submodule_name:
                print('WARNING: submodule is in root of submodules directory.')
                continue

            # Add submodule to list if not already there
            if submodule_name not in submodules:
                submodules.append(submodule_name)

        # Extract all submodules
        for submodule_path in submodules:
            obj = submodule_class.create(
                module_version=self._module_version,
                module_path=submodule_path)
            self._process_submodule(submodule=obj)

    def _extract_description(self, readme_content):
        """Extract description from README"""
        # If module description extraction is disabled, skip
        if not Config().AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION:
            return None

        # If README is empty, return early
        if not readme_content:
            return None

        # Convert README to HTML
        html_readme = markdown.markdown(
            readme_content,
            extensions=['fenced_code', 'tables']
        )

        # Convert HTML to plain text
        plain_text = BeautifulSoup(html_readme, features='html.parser').get_text()
        for line in plain_text.split('\n'):
            # Skip if line is empty
            if not line.strip():
                continue

            # Check number of characters in string
            if len(re.sub(r'[^a-zA-Z]', '', line)) < 20:
                continue

            # Check number of words
            word_match = re.findall(r'(?:([a-zA-Z]+)(?:\s|$|\.))', line)
            if word_match is None or len(word_match) < 6:
                continue

            # Check if description line contains unwanted text
            found_unwanted_text = False
            for unwanted_text in ['http://', 'https://', '@']:
                if unwanted_text in line:
                    found_unwanted_text = True
                    break
            if found_unwanted_text:
                continue

            # Get sentences
            extracted_description = ''
            for scentence in line.split('. '):
                new_description = extracted_description
                if extracted_description:
                    new_description += '. '
                new_description += scentence.strip()

                # Check length of combined sentences.
                # For combining a new sentence, check overall description
                # length of 100 chracters.
                # If this is the first sentence, give a higher allowance, as it's
                # preferable to extract a description.
                if ((new_description and len(new_description) >= 80) or
                        (not extracted_description and len(new_description) >= 130)):
                    # Otherwise, break from iterations
                    break
                extracted_description = new_description
         
            return extracted_description if extracted_description else None

        return None

    def process_upload(self):
        """Handle data extraction from module source."""
        # Run terraform-docs on module content and obtain README
        terraform_docs = self._run_terraform_docs(self.module_directory)
        tfsec = self._run_tfsec(self.module_directory)
        readme_content = self._get_readme_content(self.module_directory)

        terraform_graph = None
        with self._switch_terraform_versions(self.module_directory):
            if self._run_tf_init(self.module_directory):
                terraform_graph = self._get_graph_data(self.module_directory)

        # Check for any terrareg metadata files
        terrareg_metadata = self._get_terrareg_metadata(self.module_directory)

        # Check if description is available in metadata
        description = terrareg_metadata.get('description', None)
        if not description:
            # Otherwise, attempt to extract description from README
            description = self._extract_description(readme_content)

        self._insert_database(
            description=description,
            readme_content=readme_content,
            tfsec=tfsec,
            terraform_docs=terraform_docs,
            terrareg_metadata=terrareg_metadata,
            terraform_graph=terraform_graph
        )

        # Generate the archive, unless the module has a git clone URL and
        # the config for deleting externally hosted artifacts is enabled.
        if not (self._module_version.get_git_clone_url() and
                Config().DELETE_EXTERNALLY_HOSTED_ARTIFACTS):
            self._generate_archive()

        self._extract_additional_tab_files()

        self._scan_submodules(
            submodule_class=Submodule,
            subdirectory=Config().MODULES_DIRECTORY)
        self._scan_submodules(
            submodule_class=Example,
            subdirectory=Config().EXAMPLES_DIRECTORY)


class ApiUploadModuleExtractor(ModuleExtractor):
    """Extraction of module uploaded via API."""

    def __init__(self, upload_file, *args, **kwargs):
        """Store member variables."""
        super(ApiUploadModuleExtractor, self).__init__(*args, **kwargs)
        self._upload_file = upload_file
        self._source_file = None

    @property
    def source_file(self):
        """Generate/return source filename."""
        if self._source_file is None:
            filename = secure_filename(self._upload_file.filename)
            self._source_file = safe_join_paths(self.upload_directory, filename)
        return self._source_file

    def _save_upload_file(self):
        """Save uploaded file to uploads directory."""
        filename = secure_filename(self._upload_file.filename)
        source_file = safe_join_paths(self.upload_directory, filename)
        self._upload_file.save(source_file)

    def _check_file_type(self):
        """Check filetype"""
        file_type = magic.from_file(self.source_file, mime=True)
        if file_type == 'application/zip':
            pass
        else:
            raise UnknownFiletypeError('Upload file is of unknown filetype. Must by zip, tar.gz')

    def _extract_archive(self):
        """Extract uploaded archive into extract directory."""
        with zipfile.ZipFile(self.source_file, 'r') as zip_ref:
            for name in zip_ref.namelist():
                safe_join_paths(self.extract_directory, name)
            zip_ref.extractall(self.extract_directory)

    def process_upload(self):
        """Extract archive and perform data extraction from module source."""
        self._save_upload_file()
        self._check_file_type()
        self._extract_archive()

        super(ApiUploadModuleExtractor, self).process_upload()


class GitModuleExtractor(ModuleExtractor):
    """Extraction of module via git."""

    def __init__(self, *args, **kwargs):
        """Store member variables."""
        super(GitModuleExtractor, self).__init__(*args, **kwargs)
        # # Sanitise URL and tag name
        # self._git_url = urllib.parse.quote(git_url, safe='/:@%?=')
        # self._tag_name = urllib.parse.quote(tag_name, safe='/')

    def _clone_repository(self):
        """Extract uploaded archive into extract directory."""
        # Copy current environment variables to add GIT SSH option
        env = os.environ.copy()
        # Set SSH to autoaccept new host keys
        env['GIT_SSH_COMMAND'] = 'ssh -o StrictHostKeyChecking=accept-new'

        git_url = self._module_version._module_provider.get_git_clone_url()

        try:
            subprocess.check_output([
                    'git', 'clone', '--single-branch',
                    '--branch', self._module_version.source_git_tag,
                    git_url,
                    self.extract_directory
                ],
                stderr=subprocess.STDOUT,
                env=env,
                timeout=Config().GIT_CLONE_TIMEOUT
            )
        except subprocess.CalledProcessError as exc:
            error = 'Unknown error occurred during git clone'
            for line in exc.output.decode('ascii').split('\n'):
                if line.startswith('fatal:'):
                    error = 'Error occurred during git clone: {}'.format(line)
            raise GitCloneError(error)

    def process_upload(self):
        """Extract archive and perform data extraction from module source."""
        self._clone_repository()

        super(GitModuleExtractor, self).process_upload()
