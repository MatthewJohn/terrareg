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

from werkzeug.utils import secure_filename
import magic

from terrareg.models import ModuleVersion
from terrareg.database import Database
from terrareg.errors import (
    UnknownFiletypeError,
    InvalidTerraregMetadataFileError,
    MetadataDoesNotContainRequiredAttributeError
)
from terrareg.config import (
    DELETE_EXTERNALLY_HOSTED_ARTIFACTS,
    REQUIRED_MODULE_METADATA_ATTRIBUTES
)


class ModuleExtractor():
    """Provide extraction method of moduls."""

    TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']

    def __init__(self, upload_file, module_version: ModuleVersion):
        """Create temporary directories and store member variables."""
        self._upload_file = upload_file
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
            re.sub('\.zip$', '', self._module_version.archive_path_zip),
            'zip',
            self.extract_directory)

    def _insert_database(
        self,
        readme_content: str,
        terraform_docs_output: dict,
        terrareg_metadata: dict) -> int:
        """Insert module into DB, overwrite any pre-existing"""
        db = Database.get()

        conn = db.get_engine().connect()

        # Get module_provider row
        provider_select = db.module_provider.select().where(
            db.module_provider.c.namespace ==
            self._module_version._module_provider._module._namespace.name,
            db.module_provider.c.module ==
            self._module_version._module_provider._module.name,
            db.module_provider.c.provider == self._module_version._module_provider.name
        )
        module_provider_row = conn.execute(provider_select).fetchone()

        # Create module_provider if it does not exist
        if not module_provider_row:
            module_provider_insert = db.module_provider.insert().values(
                namespace=self._module_version._module_provider._module._namespace.name,
                module=self._module_version._module_provider._module.name,
                provider=self._module_version._module_provider.name
            )
            res = conn.execute(module_provider_insert)
            # Obtain newly inserted module_provider
            module_provider_row = conn.execute(
                db.module_provider.select().where(
                    db.module_provider.c.id == res.inserted_primary_key[0]
                )
            ).fetchone()

        # Delete module from module_version table
        delete_statement = db.module_version.delete().where(
            db.module_version.c.module_provider_id ==
            module_provider_row.id,
            db.module_version.c.version == self._module_version.version
        )
        conn.execute(delete_statement)

        # Insert new module into table
        insert_statement = db.module_version.insert().values(
            module_provider_id=module_provider_row.id,
            version=self._module_version.version,
            readme_content=readme_content,
            module_details=json.dumps(terraform_docs_output),

            published_at=datetime.datetime.now(),

            # Terrareg meta-data
            owner=terrareg_metadata.get('owner', None),
            description=terrareg_metadata.get('description', None),
            source=terrareg_metadata.get('source', None),
            variable_template=json.dumps(terrareg_metadata.get('variable_template', {})),
            verified=terrareg_metadata.get('verified', False),
            artifact_location=terrareg_metadata.get('artifact_location', None)
        )
        res = conn.execute(insert_statement)

        # Return primary key
        return res.inserted_primary_key[0]

    def _process_submodule(self, module_pk: int, submodule: str):
        """Process submodule."""
        submodule_dir = os.path.join(self.extract_directory, submodule['source'])

        if not os.path.isdir(submodule_dir):
            print('Submodule does not appear to be local: {0}'.format(submodule['source']))
            return

        tf_docs = self._run_terraform_docs(submodule_dir)
        readme_content = self._get_readme_content(submodule_dir)

        db = Database.get()
        conn = db.get_engine().connect()
        insert_statement = db.sub_module.insert().values(
            parent_module_version=module_pk,
            path=submodule['source'],
            readme_content=readme_content,
            module_details=json.dumps(tf_docs),
        )
        conn.execute(insert_statement)

    def process_upload(self):
        """Handle file upload of module source."""
        self._save_upload_file()
        self._check_file_type()
        self._extract_archive()

        # Run terraform-docs on module content and obtain README
        module_details = self._run_terraform_docs(self.extract_directory)
        readme_content = self._get_readme_content(self.extract_directory)

        # Check for any terrareg metadata files
        terrareg_metadata = self._get_terrareg_metadata(self.extract_directory)

        # Generate the archive, unless the module is externally hosted and
        # the config for deleting externally hosted artifacts is enabled.
        if not (terrareg_metadata.get('artifact_location', None) and DELETE_EXTERNALLY_HOSTED_ARTIFACTS):
            self._generate_archive()

        # Debug
        # print(module_details)
        print(json.dumps(module_details, sort_keys=False, indent=4))
        # print(readme_content)

        module_pk = self._insert_database(
            readme_content=readme_content,
            terraform_docs_output=module_details,
            terrareg_metadata=terrareg_metadata
        )

        for submodule in module_details['modules']:
            self._process_submodule(module_pk, submodule)
