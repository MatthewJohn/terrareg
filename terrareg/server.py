
import os
import tempfile
import mimetypes
import zipfile
import subprocess
import json
import tarfile

import magic

from werkzeug.utils import secure_filename
from flask import Flask, request


DATA_DIRECTORY = os.path.join(os.environ.get('DATA_DIRECTORY', '.'), 'data')

class ModuleFactory(object):
    pass

class UnknownFiletypeError(Exception):
    """Uploaded filetype is unknown."""

    pass


class Namespace(object):

    def __init__(self, name: str):
        self._name = name

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(DATA_DIRECTORY, 'modules', self._name)

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

class Module(object):
    
    def __init__(self, namespace: Namespace, name: str):
        self._namespace = namespace
        self._name = name

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._namespace.base_directory):
            self._namespace.create_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(self._namespace.base_directory, self._name)

class ModuleProvider(object):

    def __init__(self, module: Module, name: str):
        self._module = module
        self._name = name

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(self._module.base_directory, self._name)

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._module.base_directory):
            self._module.create_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    def get_versions(self):
        """Return all module provider versions."""
        return []

class ModuleVersion(object):

    def __init__(self, module_provider: ModuleProvider, version: str):
        self._module_provider = module_provider
        self._version = version

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(self._module_provider.base_directory, self._version)

    @property
    def archive_name(self):
        """Return name of the archive file"""
        return "source.tar.gz"

    @property
    def archive_path(self):
        """Return full path of the archive file."""
        return os.path.join(self.base_directory, self.archive_name)

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._module_provider.base_directory):
            self._module_provider.create_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    def handle_file_upload(self, file):
        """Handle file upload of module source."""
        with tempfile.TemporaryDirectory() as upload_d:
            # Save uploaded file to uploads directory
            filename = secure_filename(file.filename)
            source_file = os.path.join(upload_d, filename)
            file.save(source_file)

            # Check filetype and extract archive
            file_type = magic.from_file(source_file, mime=True)
            if file_type == 'application/zip':
                pass
            else:
                raise UnknownFiletypeError('Upload file is of unknown filetype. Must by zip, tar.gz')

            with tempfile.TemporaryDirectory() as extract_d:
                # Extract archive into temporary directory
                with zipfile.ZipFile(source_file, 'r') as zip_ref:
                    zip_ref.extractall(extract_d)

                # Run terraform-docs on module content
                terradocs_output = subprocess.check_output(['terraform-docs', 'json', extract_d])
                module_details = json.loads(terradocs_output)

                # Read readme file
                readme_content = None
                if os.path.isfile(os.path.join(extract_d, 'README.md')):
                    with open(os.path.join(extract_d, 'README.md'), 'r') as readme_fd:
                        readme_content = readme_fd.readlines()

                # Generate various archive formats for downloads
                ## Generate zip file
                self.create_data_directory()
                with tarfile.open(self.archive_path, "w:gz") as tar:
                    tar.add(extract_d, arcname='', recursive=True)

            print(module_details)
            print(readme_content)

class Server(object):
    """Manage web server and route requests"""

    ALLOWED_EXTENSIONS = {'zip'}

    def __init__(self):
        """Create flask app and store member variables"""
        self._app = Flask(__name__)
        self.host = '127.0.0.1'
        self.port = 5000
        self.debug = True

        if not os.path.isdir(DATA_DIRECTORY):
            os.mkdir(DATA_DIRECTORY)
        if not os.path.isdir(self._get_upload_directory()):
            os.mkdir(self._get_upload_directory())
        if not os.path.isdir(os.path.join(DATA_DIRECTORY, 'modules')):
            os.mkdir(os.path.join(DATA_DIRECTORY, 'modules'))

        self._app.config['UPLOAD_FOLDER'] = self._get_upload_directory()

        self._register_routes()

    def _get_upload_directory(self):
        return os.path.join(DATA_DIRECTORY, 'upload')

    def _register_routes(self):
        """Register routes with flask."""

        # Upload module
        self._app.route('/v1/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload', methods=['POST'])(self._upload_module_version)

        # Terraform registry routes
        self._app.route('/v1/<string:namespace>/<string:name>/<string:provider>/versions')(self._module_versions)
        self._app.route('/v1/<string:namespace>/<string:name>/<string:provider>/<string:version>/download')(self._module_version_download)

    def run(self):
        """Run flask server."""
        self._app.run(host=self.host, port=self.port, debug=self.debug)

    def allowed_file(self, filename):
        return '.' in filename and \
               filename.rsplit('.', 1)[1].lower() in self.ALLOWED_EXTENSIONS


    def _upload_module_version(self, namespace, name, provider, version):
        """Handle module version upload."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = ModuleVersion(module_provider=module_provider, version=version)

        if len(request.files) != 1:
            return 'One file can be uploaded'

        file = request.files[[f for f in request.files.keys()][0]]

        # If the user does not select a file, the browser submits an
        # empty file without a filename.
        if file.filename == '':
            return 'No selected file'
        if file and self.allowed_file(file.filename):
            module_version.handle_file_upload(file)
            return 'Upload sucessful'

        return 'Error occurred - unknown file extension'

    def _module_versions(self, namespace, name, provider):
        """Return list of version."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        return [v for v in module_provider.get_versions()]

    def _module_version_download(self, namespace, name, provider, version):
        return ''
