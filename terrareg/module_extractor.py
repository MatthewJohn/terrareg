"""Provide extraction method of modules."""

import os
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

from terrareg.models import BaseSubmodule, Example, ModuleVersion, Submodule
from terrareg.database import Database
from terrareg.errors import (
    UnableToProcessTerraformError,
    UnknownFiletypeError,
    InvalidTerraregMetadataFileError,
    MetadataDoesNotContainRequiredAttributeError,
    GitCloneError
)
from terrareg.utils import PathDoesNotExistError, safe_iglob, safe_join_paths
from terrareg.config import Config


class ModuleExtractor:
    """Provide extraction method of moduls."""

    TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']

    def __init__(self, module_version: ModuleVersion):
        """Create temporary directories and store member variables."""
        self._module_version = module_version
        self._extract_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._upload_directory = tempfile.TemporaryDirectory()  # noqa: R1732

    @property
    def extract_directory(self):
        """Return path of extract directory."""
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

    def _insert_database(
        self,
        readme_content: str,
        terraform_docs_output: dict,
        terrareg_metadata: dict) -> int:
        """Insert module into DB, overwrite any pre-existing"""
        # Update attributes of module_version in database
        self._module_version.update_attributes(
            readme_content=readme_content,
            module_details=json.dumps(terraform_docs_output),

            published_at=datetime.datetime.now(),

            # Terrareg meta-data
            owner=terrareg_metadata.get('owner', None),
            description=terrareg_metadata.get('description', None),
            repo_clone_url_template=terrareg_metadata.get('repo_clone_url', None),
            repo_browse_url_template=terrareg_metadata.get('repo_browse_url', None),
            repo_base_url_template=terrareg_metadata.get('repo_base_url', None),
            variable_template=json.dumps(terrareg_metadata.get('variable_template', {})),
            published=Config().AUTO_PUBLISH_MODULE_VERSIONS
        )

    def _process_submodule(self, submodule: BaseSubmodule):
        """Process submodule."""
        print('Processing submodule: {0}'.format(submodule.path))
        submodule_dir = safe_join_paths(self.extract_directory, submodule.path)

        tf_docs = self._run_terraform_docs(submodule_dir)
        readme_content = self._get_readme_content(submodule_dir)

        db = Database.get()
        insert_statement = db.sub_module.insert().values(
            parent_module_version=self._module_version.pk,
            type=submodule.TYPE,
            path=submodule.path,
            readme_content=Database.encode_blob(readme_content),
            module_details=Database.encode_blob(json.dumps(tf_docs))
        )
        with db.get_engine().connect() as conn:
            conn.execute(insert_statement)

        if isinstance(submodule, Example):
            self._extract_example_files(example=submodule)

    def _extract_example_files(self, example: Example):
        """Extract all terraform files in example and insert into DB"""
        example_base_dir = safe_join_paths(self.extract_directory, example.path)
        for tf_file_path in safe_iglob(base_dir=example_base_dir,
                                       pattern='*.tf',
                                       recursive=False,
                                       is_file=True):
            # Remove extraction directory from file path
            tf_file = re.sub('^{}/'.format(self.extract_directory), '', tf_file_path)

            # Obtain contents of file and insert into DB
            with open(tf_file_path, 'r') as file_fd:
                content = ''.join(file_fd.readlines())

            db = Database.get()
            insert_statement = db._example_file.insert().values(
                submodule_id=example.pk,
                path=tf_file,
                content=Database.encode_blob(content)
            )
            with db.get_engine().connect() as conn:
                conn.execute(insert_statement)

    def _scan_submodules(self, subdirectory: str, submodule_class: Type[BaseSubmodule]):
        """Scan for submodules and extract details."""
        try:
            submodule_base_directory = safe_join_paths(self.extract_directory, subdirectory, is_dir=True)
        except PathDoesNotExistError:
            # If the modules directory does not exist,
            # ignore and return
            print('No modules directory found')
            return

        extract_directory_re = re.compile('^{}'.format(
            re.escape(
                '{0}/'.format(self.extract_directory)
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
            submodule_name = extract_directory_re.sub('', submodule_dir)

            # Check submodule is not in the root of the submodules
            if not submodule_name:
                print('WARNING: submodule is in root of submodules directory.')
                continue

            # Add submodule to list if not already there
            if submodule_name not in submodules:
                submodules.append(submodule_name)

        # Extract all submodules
        for submodule_path in submodules:
            obj = submodule_class(
                module_version=self._module_version,
                module_path=submodule_path)
            self._process_submodule(submodule=obj)

    def process_upload(self):
        """Handle data extraction from module source."""
        # Run terraform-docs on module content and obtain README
        module_details = self._run_terraform_docs(self.extract_directory)
        readme_content = self._get_readme_content(self.extract_directory)

        # Check for any terrareg metadata files
        terrareg_metadata = self._get_terrareg_metadata(self.extract_directory)

        self._insert_database(
            readme_content=readme_content,
            terraform_docs_output=module_details,
            terrareg_metadata=terrareg_metadata
        )

        # Generate the archive, unless the module has a git clone URL and
        # the config for deleting externally hosted artifacts is enabled.
        if not (self._module_version.get_git_clone_url() and
                Config().DELETE_EXTERNALLY_HOSTED_ARTIFACTS):
            self._generate_archive()

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
                env=env
            )
        except subprocess.CalledProcessError as exc:
            error = 'Unknown error occurred during git clone'
            for line in exc.output.decode('ascii').split('\n'):
                print(line)
                if line.startswith('fatal:'):
                    error = 'Error occurred during git clone: {}'.format(line)
            raise GitCloneError(error)

    def process_upload(self):
        """Extract archive and perform data extraction from module source."""
        self._clone_repository()

        super(GitModuleExtractor, self).process_upload()
