
import os
from distutils.version import StrictVersion
import json

import markdown

from terrareg.database import Database
from terrareg.config import DATA_DIRECTORY
from terrareg.errors import (NoModuleVersionAvailableError)


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
        rows.sort(key=lambda x: StrictVersion(x['version']), reverse=True)

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
        return [
            {
                'name': provider['name'],
                'namespace': '',  # This data is not available
                'source': '',  # This data is not available
                'version': provider['version']
            }
            for provider in self.get_module_specs()['providers']
        ]

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

    def get_source_download_url(self):
        """Return URL to download source file."""
        return '/static/modules/{0}/{1}'.format(self.id, self.archive_name_zip)

    @property
    def version(self):
        """Return version."""
        return self._version

    @property
    def base_directory(self):
        """Return base directory."""
        return os.path.join(self._module_provider.base_directory, self._version)

    @property
    def source_file_prefix(self):
        """Prefix of source file"""
        return 'source'

    @property
    def archive_name_tar_gz(self):
        """Return name of the archive file"""
        return '{0}.tar.gz'.format(self.source_file_prefix)

    @property
    def archive_name_zip(self):
        """Return name of the archive file"""
        return '{0}.zip'.format(self.source_file_prefix)

    @property
    def archive_path_tar_gz(self):
        """Return full path of the archive file."""
        return os.path.join(self.base_directory, self.archive_name_tar_gz)

    @property
    def archive_path_zip(self):
        """Return full path of the archive file."""
        return os.path.join(self.base_directory, self.archive_name_zip)

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

    def prepare_module(self):
        """Handle file upload of module version."""
        self.create_data_directory()

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