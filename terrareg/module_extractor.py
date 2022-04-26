"""Provide extraction method of modules."""

import os
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

from terrareg.models import ModuleVersion
from terrareg.database import Database
from terrareg.errors import (
    UnknownFiletypeError,
    InvalidTerraregMetadataFileError,
    MetadataDoesNotContainRequiredAttributeError
)
from terrareg.utils import PathDoesNotExistError, safe_join_paths
from terrareg.config import (
    DELETE_EXTERNALLY_HOSTED_ARTIFACTS,
    REQUIRED_MODULE_METADATA_ATTRIBUTES,
    AUTO_PUBLISH_MODULE_VERSIONS,
    MODULES_DIRECTORY
)


class ModuleExtractor:
    """Provide extraction method of moduls."""

    TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']

    def __init__(self, module_version: ModuleVersion):
        """Create temporary directories and store member variables."""
        self._module_version = module_version
        self._extract_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._upload_directory = tempfile.TemporaryDirectory()  # noqa: R1732
        self._source_file = None

    @property
    def source_file(self):
        """Generate/return source filename."""
        if self._source_file is None:
            filename = secure_filename(self._upload_file.filename)
            self._source_file = os.path.join(self.upload_directory, filename)
        return self._source_file

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
        terradocs_output = subprocess.check_output(['terraform-docs', 'json', module_path])
        return json.loads(terradocs_output)

    @staticmethod
    def _get_readme_content(module_path):
        """Obtain README contents for given module."""
        readme_path = os.path.join(module_path, 'README.md')
        if os.path.isfile(readme_path):
            with open(readme_path, 'r') as readme_fd:
                return ''.join(readme_fd.readlines())

        # If no README found, return None
        return None

    def _get_terrareg_metadata(self, module_path):
        """Obtain terrareg metadata for module, if it exists."""
        terrareg_metadata = {}
        for terrareg_file in self.TERRAREG_METADATA_FILES:
            path = os.path.join(module_path, terrareg_file)
            if os.path.isfile(path):
                with open(path, 'r') as terrareg_fh:
                    try:
                        terrareg_metadata = json.loads(''.join(terrareg_fh.readlines()))
                    except:
                        raise InvalidTerraregMetadataFileError(
                            'An error occured whilst processing the terrareg metadata file.'
                        )

                # Remove the meta-data file, so it is not added to the archive
                os.unlink(path)

        for required_attr in REQUIRED_MODULE_METADATA_ATTRIBUTES:
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
            source=terrareg_metadata.get('source', None),
            variable_template=json.dumps(terrareg_metadata.get('variable_template', {})),
            artifact_location=terrareg_metadata.get('artifact_location', None),
            published=AUTO_PUBLISH_MODULE_VERSIONS
        )
        print(AUTO_PUBLISH_MODULE_VERSIONS)

    def _process_submodule(self, submodule: str):
        """Process submodule."""
        print('Processing submodule: {0}'.format(submodule))
        submodule_dir = safe_join_paths(self.extract_directory, submodule)

        tf_docs = self._run_terraform_docs(submodule_dir)
        readme_content = self._get_readme_content(submodule_dir)

        db = Database.get()
        conn = db.get_engine().connect()
        insert_statement = db.sub_module.insert().values(
            parent_module_version=self._module_version.pk,
            path=submodule,
            readme_content=readme_content,
            module_details=json.dumps(tf_docs),
        )
        conn.execute(insert_statement)

    def _scan_submodules(self):
        """Scan for submodules and extract details."""
        try:
            submodule_base_directory = safe_join_paths(self.extract_directory, MODULES_DIRECTORY, is_dir=True)
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
        for submodule in submodules:
            self._process_submodule(submodule)

    def process_upload(self):
        """Handle data extraction from module source."""
        # Run terraform-docs on module content and obtain README
        module_details = self._run_terraform_docs(self.extract_directory)
        readme_content = self._get_readme_content(self.extract_directory)

        # Check for any terrareg metadata files
        terrareg_metadata = self._get_terrareg_metadata(self.extract_directory)

        # Generate the archive, unless the module is externally hosted and
        # the config for deleting externally hosted artifacts is enabled.
        if not (terrareg_metadata.get('artifact_location', None) and DELETE_EXTERNALLY_HOSTED_ARTIFACTS):
            self._generate_archive()

        self._insert_database(
            readme_content=readme_content,
            terraform_docs_output=module_details,
            terrareg_metadata=terrareg_metadata
        )

        self._scan_submodules()


class ApiUploadModuleExtractor(ModuleExtractor):
    """Extraction of module uploaded via API."""

    def __init__(self, upload_file, *args, **kwargs):
        """Store member variables."""
        super(ApiUploadModuleExtractor, self).__init__(*args, **kwargs)
        self._upload_file = upload_file

    def _save_upload_file(self):
        """Save uploaded file to uploads directory."""
        filename = secure_filename(self._upload_file.filename)
        source_file = os.path.join(self.upload_directory, filename)
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

        subprocess.check_call([
            'git', 'clone', '--single-branch',
            '--branch', self._module_version.source_git_tag,
            self._module_version._module_provider.repository_url,
            self.extract_directory
        ], env=env)

    def process_upload(self):
        """Extract archive and perform data extraction from module source."""
        self._clone_repository()

        super(GitModuleExtractor, self).process_upload()
