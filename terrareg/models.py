
import os
from distutils.version import StrictVersion
import json
import re
import jinja2
import sqlalchemy

import markdown

import terrareg.analytics
from terrareg.database import Database
from terrareg.config import DATA_DIRECTORY
from terrareg.errors import (NoModuleVersionAvailableError)


class Namespace(object):

    @staticmethod
    def get_total_count():
        """Get total number of namespaces."""
        db = Database.get()
        counts = db.module_provider.select(
        ).group_by(
            db.module_provider.c.namespace
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        return res.scalar()

    @staticmethod
    def extract_analytics_token(namespace: str):
        """Extract analytics token from start of namespace."""
        namespace_split = re.split(r'__', namespace)

        # If there are two values in the split,
        # return first as analytics token and
        # second as namespace
        if len(namespace_split) == 2:
            return namespace_split[1], namespace_split[0]

        # If there were not two element (more or less),
        # return original value
        return namespace, None

    @staticmethod
    def get_all():
        """Return all namespaces."""
        db = Database.get()
        select = db.module_provider.select().group_by(
            db.module_provider.c.namespace
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
        select = db.module_provider.select(
        ).where(
            db.module_provider.c.namespace == self.name
        ).group_by(
            db.module_provider.c.module
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
        select = db.module_provider.select(
        ).where(
            db.module_provider.c.namespace == self._namespace.name
        ).where(
            db.module_provider.c.module == self.name
        ).group_by(
            db.module_provider.c.provider
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

    @staticmethod
    def get_total_count():
        """Get total number of module providers."""
        db = Database.get()
        counts = db.module_provider.select(
        ).group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        return res.scalar()

    def __init__(self, module: Module, name: str):
        self._module = module
        self._name = name

    @property
    def name(self):
        """Return name."""
        return self._name

    @property
    def id(self):
        """Return ID in form of namespace/name/provider/version"""
        return '{namespace}/{name}/{provider}'.format(
            namespace=self._module._namespace.name,
            name=self._module.name,
            provider=self.name
        )

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
        select = db.get_module_version_joined_to_module_provider(
            db.module_version.select()
        ).where(
            db.module_provider.c.namespace == self._module._namespace.name
        ).where(
            db.module_provider.c.module == self._module.name
        ).where(
            db.module_provider.c.provider == self.name
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

        select = db.get_module_version_joined_to_module_provider(
            db.module_version.select()
        ).where(
            db.module_provider.c.namespace == self._module._namespace.name
        ).where(
            db.module_provider.c.module == self._module.name
        ).where(
            db.module_provider.c.provider == self.name
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
        providers = []
        for provider in self.get_module_specs()['providers']:

            name_split = provider['name'].split('/')
            # Default to name being the name and hashicorp
            # as the namespace
            name = provider['name']
            namespace = 'hashicorp'
            if len(name_split) > 1:
                # If name contains slash, assume
                # namespace is the first element
                namespace = name_split[0]
                name = '/'.join(name_split[1:])

            providers.append(            {
                'name': name,
                'namespace': namespace,
                'source': '',  # This data is not available
                'version': provider['version'] if provider['version'] else ''
            })
        return providers

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

    @staticmethod
    def get_total_count():
        """Get total number of module versions."""
        db = Database.get()
        counts = db.get_module_version_joined_to_module_provider(
            db.module_version.select()
        ).group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider,
            db.module_version.c.version
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        conn = db.get_engine().connect()
        res = conn.execute(select)

        return res.scalar()

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
        if self._get_db_row()['artifact_location']:
            return self._get_db_row()['artifact_location'].format(module_version=self.version)

        return '/static/modules/{0}/{1}'.format(self.id, self.archive_name_zip)

    @property
    def publish_date_display(self):
        """Return display view of date of module published."""
        return self._get_db_row()['published_at'].strftime('%B %d, %Y')

    @property
    def owner(self):
        """Return owner of module."""
        return self._get_db_row()['owner']

    @property
    def source_code_url(self):
        """Return source code URL."""
        return self._get_db_row()['source']

    @property
    def description(self):
        """Return description."""
        return self._get_db_row()['description']

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
        return '{provider_id}/{version}'.format(
            provider_id=self._module_provider.id,
            version=self.version
        )

    @property
    def variable_template(self):
        """Return variable template for module version."""
        return json.loads(self._get_db_row()['variable_template'])

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
            "downloads": self.get_total_downloads(),
            "verified": True,
        }

    def get_total_downloads(self):
        """Obtain total number of downloads for module version."""
        return terrareg.analytics.AnalyticsEngine.get_module_version_total_downloads(
            module_version=self
        )

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
        select = db.module_version.select().join(
            db.module_provider, db.module_version.c.module_provider_id == db.module_provider.c.id
        ).where(
            db.module_provider.c.namespace == self._module_provider._module._namespace.name,
            db.module_provider.c.module == self._module_provider._module.name,
            db.module_provider.c.provider == self._module_provider.name,
            db.module_version.c.version == self.version
        )
        conn = db.get_engine().connect()
        res = conn.execute(select)
        return res.fetchone()

    def get_readme_html(self):
        """Convert readme markdown to HTML"""
        if self.get_readme_content():
            return markdown.markdown(
                self.get_readme_content(),
                extensions=['fenced_code', 'tables']
            )
        
        # Return string when no readme is present
        return '<h5 class="title is-5">No README present in the module</h3>'

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