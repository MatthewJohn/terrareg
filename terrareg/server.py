
from ast import Sub
import os
import tempfile
import zipfile
import subprocess
import json
import tarfile
from distutils.version import StrictVersion
import datetime

import markdown
import sqlalchemy
import magic
from werkzeug.utils import secure_filename
from flask import Flask, request, jsonify, render_template, redirect
from flask_restful import Resource, Api, reqparse


DATA_DIRECTORY = os.path.join(os.environ.get('DATA_DIRECTORY', '.'), 'data')


class UnknownFiletypeError(Exception):
    """Uploaded filetype is unknown."""

    pass


class NoModuleVersionAvailableError(Exception):
    """No version of this module available."""

    pass


class InvalidTerraregMetadataFileError(Exception):
    """Error whilst reading terrareg metadata file."""

    pass


class Database(object):

    _META = None
    _ENGINE = None
    _INSTANCE = None

    def __init__(self):
        pass

    @classmethod
    def get(cls):
        if cls._INSTANCE is None:
            cls._INSTANCE = Database()
        return cls._INSTANCE

    @classmethod
    def get_meta(cls):
        """Return meta object"""
        if cls._META is None:
            cls._META = sqlalchemy.MetaData()
        return cls._META

    @classmethod
    def get_engine(cls):
        if cls._ENGINE is None:
            cls._ENGINE = sqlalchemy.create_engine('sqlite:///modules.db', echo = True)
        return cls._ENGINE

    def initialise(self):
        """Initialise database schema"""
        meta = self.get_meta()
        engine = self.get_engine()

        self.module_version = sqlalchemy.Table(
            'module_version', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('namespace', sqlalchemy.String),
            sqlalchemy.Column('module', sqlalchemy.String),
            sqlalchemy.Column('provider', sqlalchemy.String),
            sqlalchemy.Column('version', sqlalchemy.String),
            sqlalchemy.Column('owner', sqlalchemy.String),
            sqlalchemy.Column('description', sqlalchemy.String),
            sqlalchemy.Column('source', sqlalchemy.String),
            sqlalchemy.Column('published_at', sqlalchemy.DateTime),
            sqlalchemy.Column('readme_content', sqlalchemy.String),
            sqlalchemy.Column('module_details', sqlalchemy.String)
        )

        self.sub_module = sqlalchemy.Table(
            'submodule', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'parent_module_version',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('path', sqlalchemy.String),
            sqlalchemy.Column('name', sqlalchemy.String),
            sqlalchemy.Column('readme_content', sqlalchemy.String),
            sqlalchemy.Column('module_details', sqlalchemy.String)
        )

        meta.create_all(engine)


class ModuleSearch(object):

    @staticmethod
    def search_module_providers(
        offset: int,
        limit: int,
        query: str=None,
        namespace: str=None,
        provider: str=None,
        verified: bool=False):

        db = Database.get()
        select = db.module_version.select()

        if query:
            for query_part in query.split():
                select = select.where(
                    sqlalchemy.or_(
                        db.module_version.c.namespace.like(query_part),
                        db.module_version.c.module.like(query_part),
                        db.module_version.c.provider.like(query_part),
                        db.module_version.c.version.like(query_part),
                        db.module_version.c.description.like(query_part),
                        db.module_version.c.owner.like(query_part)
                    )
                )

        # If provider has been supplied, select by that
        if provider:
            select = select.where(
                db.module_version.c.provider == provider
            )

        # If namespace has been supplied, select by that
        if namespace:
            select = select.where(
                db.module_version.c.namespace == namespace
            )

        # Filter by verified modules, if requested
        if verified:
            select = select.where(
                db.module_version.c.verified == True
            )

        # Group by and order by namespace, module and provider
        select = select.group_by(
            db.module_version.c.namespace,
            db.module_version.c.module,
            db.module_version.c.provider
        ).order_by(
            db.module_version.c.namespace.asc(),
            db.module_version.c.module.asc(),
            db.module_version.c.provider.asc()
        ).limit(limit).offset(offset)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        module_providers = []
        for r in res:
            namespace = Namespace(name=r['namespace'])
            module = Module(namespace=namespace, name=r['module'])
            module_providers.append(ModuleProvider(module=module, name=r['provider']))

        return module_providers


class Namespace(object):

    @staticmethod
    def get_all():
        """Return all namespaces."""
        db = Database.get()
        select = db.module_version.select().group_by(
            db.module_version.c.namespace
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        namespaces = [r['namespace'] for r in res]
        return [
            Namespace(name=namespace)
            for namespace in namespaces
        ]

    def __init__(self, name: str):
        self._name = name

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(DATA_DIRECTORY, 'modules', self._name)

    def get_view_url(self):
        """Return view URL"""
        return '/modules/{namespace}'.format(namespace=self.name)

    def get_all_modules(self):
        """Return all modules for namespace."""
        db = Database.get()
        select = db.module_version.select(
        ).where(
            db.module_version.c.namespace == self.name
        ).group_by(
            db.module_version.c.module
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        modules = [r['module'] for r in res]
        return [
            Module(namespace=self, name=module)
            for module in modules
        ]

    @property
    def name(self):
        """Return name."""
        return self._name

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

class Module(object):
    
    def __init__(self, namespace: Namespace, name: str):
        self._namespace = namespace
        self._name = name

    @property
    def name(self):
        """Return name."""
        return self._name

    def get_view_url(self):
        """Return view URL"""
        return '{namespace_url}/{module}'.format(
            namespace_url=self._namespace.get_view_url(),
            module=self.name
        )

    def get_providers(self):
        """Return module providers for module."""
        db = Database.get()
        select = db.module_version.select(
        ).where(
            db.module_version.c.namespace == self._namespace.name
        ).where(
            db.module_version.c.module == self.name
        ).group_by(
            db.module_version.c.provider
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        providers = [r['provider'] for r in res]
        return [
            ModuleProvider(module=self, name=provider)
            for provider in providers
        ]

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
    def name(self):
        """Return name."""
        return self._name

    def get_view_url(self):
        """Return view URL"""
        return '{module_url}/{module}'.format(
            module_url=self._module.get_view_url(),
            module=self.name
        )

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(self._module.base_directory, self._name)

    def get_latest_version(self):
        """Get latest version of module."""
        db = Database.get()
        select = db.module_version.select().where(
            db.module_version.c.namespace == self._module._namespace.name
        ).where(
            db.module_version.c.module == self._module.name
        ).where(
            db.module_version.c.provider == self.name
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        # Convert to list
        rows = [r for r in res]

        # Sort rows by semantec versioning
        rows.sort(key=lambda x: StrictVersion(x['version']))

        # Ensure at least one row
        if not rows:
            raise NoModuleVersionAvailableError('No module version available.')

        # Obtain latest row
        return ModuleVersion(module_provider=self, version=rows[0]['version'])

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
        db = Database.get()

        select = db.module_version.select().where(
            db.module_version.c.namespace == self._module._namespace.name
        ).where(
            db.module_version.c.module == self._module.name
        ).where(
            db.module_version.c.provider == self.name
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        return [
            ModuleVersion(module_provider=self, version=r['version'])
            for r in res
        ]


class TerraformSpecsObject(object):
    """Base terraform object, that has terraform-docs available."""

    def __init__(self):
        """Setup member variables."""
        self._module_specs = None

    @property
    def path(self):
        """Return module path"""
        raise NotImplementedError

    def _get_db_row(self):
        """Must be implemented by object. Return row from DB."""
        raise NotImplementedError

    def get_module_specs(self):
        """Return module specs"""
        if self._module_specs is None:
            self._module_specs = json.loads(self._get_db_row()['module_details'])
        return self._module_specs

    def get_readme_content(self):
        """Get readme contents"""
        return self._get_db_row()['readme_content']

    def get_terraform_inputs(self):
        """Obtain module inputs"""
        return self.get_module_specs()['inputs']

    def get_terraform_outputs(self):
        """Obtain module inputs"""
        return self.get_module_specs()['outputs']

    def get_terraform_resources(self):
        """Obtain module resources."""
        return self.get_module_specs()['resources']

    def get_terraform_dependencies(self):
        """Obtain module dependencies."""
        #return self.get_module_specs()['requirements']
        # @TODO Verify what this should be - terraform example is empty and real-world examples appears to
        # be empty, but do have an undocumented 'provider_dependencies'
        return []

    def get_terraform_provider_dependencies(self):
        """Obtain module dependencies."""
        # @TODO See Above
        #return self.get_module_specs()['providers']
        return []

    def get_api_module_specs(self):
        """Return module specs for API."""
        return {
            "path": self.path,
            "readme": self.get_readme_content(),
            "empty": False,
            "inputs": self.get_terraform_inputs(),
            "outputs": self.get_terraform_outputs(),
            "dependencies": self.get_terraform_dependencies(),
            "provider_dependencies": self.get_terraform_provider_dependencies(),
            "resources": self.get_terraform_resources(),
        }


class ModuleVersion(TerraformSpecsObject):

    def __init__(self, module_provider: ModuleProvider, version: str):
        """Setup member variables."""
        self._module_provider = module_provider
        self._version = version
        super(ModuleVersion, self).__init__()

    def get_view_url(self):
        """Return view URL"""
        return '{module_provider_url}/{version}'.format(
            module_provider_url=self._module_provider.get_view_url(),
            version=self.version
        )

    @property
    def version(self):
        """Return version."""
        return self._version

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

    @property
    def pk(self):
        """Return database ID of module version."""
        return self._get_db_row()['id']

    @property
    def path(self):
        """Return module path"""
        # Root module is always empty
        return ''

    @property
    def id(self):
        """Return ID in form of namespace/name/provider/version"""
        return "{namespace}/{name}/{provider}/{version}".format(
            namespace=self._module_provider._module._namespace.name,
            name=self._module_provider._module.name,
            provider=self._module_provider.name,
            version=self.version
        )

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._module_provider.base_directory):
            self._module_provider.create_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    def get_api_outline(self):
        """Return dict of basic version details for API response."""
        row = self._get_db_row()
        return {
            "id": self.id,
            "owner": row['owner'],
            "namespace": self._module_provider._module._namespace.name,
            "name": self._module_provider._module.name,
            "version": self.version,
            "provider": self._module_provider.name,
            "description": row['description'],
            "source": row['source'],
            "published_at": row['published_at'].isoformat(),
            "downloads": 0,
            "verified": True,
        }

    def get_api_details(self):
        """Return dict of version details for API response."""
        api_details = self.get_api_outline()
        api_details.update({
            "root": self.get_api_module_specs(),
            "submodules": [sm.get_api_module_specs() for sm in self.get_submodules()],
            "providers": [p.name for p in self._module_provider._module.get_providers()],
            "versions": [v.version for v in self._module_provider.get_versions()]
        })
        return api_details

    def _get_db_row(self):
        """Get object from database"""
        db = Database.get()
        select = db.module_version.select().where(
            db.module_version.c.namespace == self._module_provider._module._namespace.name
        ).where(
            db.module_version.c.module == self._module_provider._module.name
        ).where(
            db.module_version.c.provider == self._module_provider.name
        ).where(
            db.module_version.c.version == self.version
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)
        return res.fetchone()

    def get_readme_html(self):
        """Convert readme markdown to HTML"""
        return markdown.markdown(self.get_readme_content(), extensions=['fenced_code'])

    def handle_file_upload(self, file):
        """Handle file upload of module version."""
        self.create_data_directory()
        with ModuleExtractor(upload_file=file, module_version=self) as me:
            me.process_upload()

    def get_submodules(self):
        """Return list of submodules."""
        db = Database.get()
        select = db.sub_module.select(
        ).join(db.module_version, db.module_version.c.id == db.sub_module.c.parent_module_version).where(
            db.module_version.c.id == self.pk,
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)

        return [
            Submodule(module_version=self, module_path=r['path'])
            for r in res
        ]


class Submodule(TerraformSpecsObject):
    """Sub module from a module version."""

    def __init__(self, module_version: ModuleVersion, module_path: str):
        self._module_version = module_version
        self._module_path = module_path
        super(Submodule, self).__init__()

    @property
    def path(self):
        """Return module path"""
        return self._module_path

    def _get_db_row(self):
        """Get object from database"""
        db = Database.get()
        select = db.sub_module.select().where(
            db.sub_module.c.parent_module_version == self._module_version.pk,
            db.sub_module.c.path == self._module_path
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)
        return res.fetchone()


class ModuleExtractor(object):

    TERRAREG_METADATA_FILES = ['terrareg.json', '.terrareg.json']

    def __init__(self, upload_file, module_version: ModuleVersion):
        """Create temporary directories and store member variables."""
        self._upload_file = upload_file
        self._module_version = module_version
        self._extract_directory = tempfile.TemporaryDirectory()
        self._upload_directory = tempfile.TemporaryDirectory()
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

    def _run_terraform_docs(self, module_path):
        """Run terraform docs and return output."""
        terradocs_output = subprocess.check_output(['terraform-docs', 'json', module_path])
        return json.loads(terradocs_output)

    def _get_readme_content(self, module_path):
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
        
        return terrareg_metadata

    def _generate_archive(self):
        """Generate archive of extracted module"""
        with tarfile.open(self._module_version.archive_path, "w:gz") as tar:
            tar.add(self.extract_directory, arcname='', recursive=True)

    def _insert_database(self, readme_content: str, terraform_docs_output: dict, terrareg_metadata: dict) -> int:
        """Insert module into DB, overwrite any pre-existing"""
        db = Database.get()

        # Delete module from module_version table
        delete_statement = db.module_version.delete().where(
            db.module_version.c.namespace == self._module_version._module_provider._module._namespace.name
        ).where(
            db.module_version.c.module == self._module_version._module_provider._module.name
        ).where(
            db.module_version.c.provider == self._module_version._module_provider.name
        ).where(
            db.module_version.c.version == self._module_version.version
        )
        conn = db.get_engine().connect()
        conn.execute(delete_statement)

        # Insert new module into table
        insert_statement = db.module_version.insert().values(
            namespace=self._module_version._module_provider._module._namespace.name,
            module=self._module_version._module_provider._module.name,
            provider=self._module_version._module_provider.name,
            version=self._module_version.version,
            readme_content=readme_content,
            module_details=json.dumps(terraform_docs_output),

            published_at=datetime.datetime.now(),

            # Terrareg meta-data
            owner=terrareg_metadata.get('owner', None),
            description=terrareg_metadata.get('description', None),
            source=terrareg_metadata.get('source', None)
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


class Server(object):
    """Manage web server and route requests"""

    ALLOWED_EXTENSIONS = {'zip'}

    def __init__(self):
        """Create flask app and store member variables"""
        self._app = Flask(
            __name__,
            static_folder='static',
            template_folder='templates'
        )
        self._api = Api(
            self._app,
            #prefix='v1'
        )

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

        # Initialise database
        Database.get().initialise()

        self._register_routes()

    def _get_upload_directory(self):
        return os.path.join(DATA_DIRECTORY, 'upload')

    def _register_routes(self):
        """Register routes with flask."""

        # Upload module
        self._app.route(
            '/v1/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload',
            methods=['POST']
        )(self._upload_module_version)

        # Terraform registry routes
        self._api.add_resource(
            ApiModuleList,
            '/v1/modules',
            '/v1/modules/'
        )
        self._api.add_resource(
            ApiModuleSearch,
            '/v1/modules/search',
            '/v1/modules/search/'
        )
        self._api.add_resource(
            ApiModuleDetails,
            '/v1/modules/<string:namespace>/<string:name>',
            '/v1/modules/<string:namespace>/<string:name>/'
        )
        self._api.add_resource(
            ApiModuleProviderDetails,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>',
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/')
        self._api.add_resource(
            ApiModuleVersions,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions',
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions/'
        )
        self._api.add_resource(
            ApiModuleVersionDownload,
            '/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/download'
        )

        # Views
        self._app.route('/')(self._view_serve_static_index)
        self._app.route(
            '/modules'
        )(self._view_serve_namespace_list)
        self._app.route(
            '/modules/'
        )(self._view_serve_namespace_list)
        self._app.route(
            '/modules/<string:namespace>'
        )(self._view_serve_namespace)
        self._app.route(
            '/modules/<string:namespace>/'
        )(self._view_serve_namespace)
        self._app.route(
            '/modules/<string:namespace>/<string:name>'
        )(self._view_serve_module)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/'
        )(self._view_serve_module)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>'
        )(self._view_serve_module_provider)
        self._app.route(
            '/modules/<string:namespace>/<string:name>/<string:provider>/'
        )(self._view_serve_module_provider)

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

    def _view_serve_static_index(self):
        """Serve static index"""
        return render_template('index.html')

    def _view_serve_namespace_list(self):
        """Render view for display module."""
        namespaces = Namespace.get_all()

        # If only one provider for module, redirect to it.
        if len(namespaces) == 1:
            return redirect(namespaces[0].get_view_url())
        else:
            return render_template(
                'namespace_list.html',
                namespaces=namespaces
            )

    def _view_serve_namespace(self, namespace):
        """Render view for namespace."""
        namespace = Namespace(namespace)
        modules = namespace.get_all_modules()

        return render_template(
            'namespace.html',
            namespace=namespace,
            modules=modules
        )

    def _view_serve_module(self, namespace, name):
        """Render view for display module."""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_providers = module.get_providers()

        # If only one provider for module, redirect to it.
        if len(module_providers) == 1:
            return redirect(module_providers[0].get_view_url())
        else:
            return render_template(
                'module.html',
                namespace=namespace,
                module=module,
                module_providers=module_providers
            )

    def _view_serve_module_provider(self, namespace, name, provider, version=None):
        """Render view for displaying module provider information"""
        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        if version is None:
            module_version = module_provider.get_latest_version()
        else:
            module_version = ModuleVersion(module_provider=module_provider, version=version)

        return render_template(
            'module_provider.html',
            namespace=namespace,
            module=module,
            module_provider=module_provider,
            module_version=module_version
        )

class ApiModuleList(Resource):
    def get(self):
        """Return list of modules."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'provider', type=str,
            default=None, help='Limits modules to a specific provider.'
        )
        parser.add_argument(
            'verified', type=bool,
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        # Limit the limits
        limit = 50 if args.limit > 50 else args.limit
        limit = 1 if limit < 1 else limit
        current_offset = 0 if args.offset < 0 else args.offset

        module_providers = ModuleSearch.search_module_providers(
            provider=args.provider,
            verified=args.verified,
            offset=current_offset,
            limit=limit
        )

        return {
            "meta": {
                "limit": limit,
                "current_offset": current_offset,
                "next_offset": (current_offset + limit),
                "prev_offset": (current_offset - limit) if (current_offset >= limit) else 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers
            ]
        }


class ApiModuleSearch(Resource):

    def get(self):
        """Search for modules, given query string, namespace or provider."""
        parser = reqparse.RequestParser()
        parser.add_argument(
            'q', type=str,
            required=True,
            help='The search string.'
        )
        parser.add_argument(
            'offset', type=int,
            default=0, help='Pagination offset')
        parser.add_argument(
            'limit', type=int,
            default=10, help='Pagination limit'
        )
        parser.add_argument(
            'provider', type=str,
            default=None, help='Limits modules to a specific provider.'
        )
        parser.add_argument(
            'namespace', type=str,
            default=None, help='Limits modules to a specific namespace.'
        )
        parser.add_argument(
            'verified', type=bool,
            default=False, help='Limits modules to only verified modules.'
        )

        args = parser.parse_args()

        # Limit the limits
        limit = 50 if args.limit > 50 else args.limit
        limit = 1 if limit < 1 else limit
        current_offset = 0 if args.offset < 0 else args.offset

        module_providers = ModuleSearch.search_module_providers(
            query=args.q,
            namespace=args.namespace,
            provider=args.provider,
            verified=args.verified,
            offset=current_offset,
            limit=limit
        )

        return {
            "meta": {
                "limit": limit,
                "current_offset": current_offset,
                "next_offset": (current_offset + limit),
                "prev_offset": (current_offset - limit) if (current_offset >= limit) else 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module_providers
            ]
        }

class ApiModuleDetails(Resource):
    def get(self, namespace, name):
        """Return latest version for each module provider."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        return {
            "meta": {
                "limit": 5,
                "offset": 0
            },
            "modules": [
                module_provider.get_latest_version().get_api_outline()
                for module_provider in module.get_providers()
            ]
        }

class ApiModuleProviderDetails(Resource):

    def get(self, namespace, name, provider):
        """Return list of version."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        module_version = module_provider.get_latest_version()
        return module_version.get_api_details()


class ApiModuleVersions(Resource):

    def get(self, namespace, name, provider):
        """Return list of version."""

        namespace = Namespace(namespace)
        module = Module(namespace=namespace, name=name)
        module_provider = ModuleProvider(module=module, name=provider)
        return {
            "modules": [
                {
                    "source": "{namespace}/{module}/{provider}".format(
                        namespace=namespace.name,
                        module=module.name,
                        provider=module_provider.name
                    ),
                    "versions": [
                        {
                            "version": v.version,
                            "root": {
                                # @TODO: Add providers/depdencies
                                "providers": [],
                                "dependencies": []
                            },
                            # @TODO: Add submodule information
                            "submodules": []
                        }
                        for v in module_provider.get_versions()
                    ]
                }
            ]
        }

class ApiModuleVersionDownload(Resource):
    def get(self, namespace, name, provider, version):
        return ''
