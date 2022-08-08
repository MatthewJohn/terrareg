
import datetime
from importlib.util import module_for_loader
import os
from distutils.version import LooseVersion
import json
import re
import secrets
import sqlalchemy
import urllib.parse

import markdown

import terrareg.analytics
from terrareg.database import Database
import terrareg.config
from terrareg.errors import (
    InvalidModuleNameError, InvalidModuleProviderNameError,
    InvalidVersionError, NoModuleVersionAvailableError,
    InvalidGitTagFormatError, InvalidNamespaceNameError,
    RepositoryUrlDoesNotContainValidSchemeError,
    RepositoryUrlContainsInvalidSchemeError,
    RepositoryUrlDoesNotContainHostError,
    RepositoryDoesNotContainPathError,
    InvalidGitProviderConfigError,
    ModuleProviderCustomGitRepositoryUrlNotAllowedError,
    NoModuleDownloadMethodConfiguredError,
    ProviderNameNotPermittedError
)
from terrareg.utils import safe_join_paths
from terrareg.validators import GitUrlValidator



class Session:
    """Provide interface to get and set sessions"""

    SESSION_ID_LENGTH = 32

    @classmethod
    def create_session(cls):
        """Create new session object."""
        db = Database.get()
        with db.get_connection() as conn:
            session_id = secrets.token_urlsafe(cls.SESSION_ID_LENGTH)
            conn.execute(db.session.insert().values(
                id=session_id,
                expiry=(datetime.datetime.now() + datetime.timedelta(minutes=terrareg.config.Config().ADMIN_SESSION_EXPIRY_MINS))
            ))

            return cls(session_id=session_id)

    @classmethod
    def check_session(cls, session_id):
        """Get session object."""
        # Check session ID is not empty
        if not session_id:
            return None

        # Check if session exists in database and is still valid
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.session.select().where(
                db.session.c.id==session_id,
                db.session.c.expiry >= datetime.datetime.now()
            ))
            row = res.fetchone()
        # If no rows found in database, return None
        if not row:
            return None

        return cls(session_id=session_id)

    @classmethod
    def cleanup_old_sessions(cls):
        """Delete old sessions from database that have expired."""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.session.delete().where(
                db.session.c.expiry < datetime.datetime.now()
            ))

    @property
    def id(self):
        """Return ID of session"""
        return self._session_id

    def __init__(self, session_id):
        """Store current session ID."""
        self._session_id = session_id

    def delete(self):
        """Delete session from database"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.session.delete().where(
                db.session.c.id==self.id
            ))


class GitProvider:
    """Interface to specify how modules should interact with known git providers."""

    @staticmethod
    def initialise_from_config():
        """Load git providers from config into database."""
        git_provider_config = json.loads(terrareg.config.Config().GIT_PROVIDER_CONFIG)
        db = Database.get()
        for git_provider_config in git_provider_config:
            # Validate provider config
            for attr in ['name', 'base_url', 'clone_url', 'browse_url']:
                if attr not in git_provider_config:
                    raise InvalidGitProviderConfigError(
                        'Git provider config does not contain required attribute: {}'.format(attr))

            # Valid git URLs for git provider
            GitUrlValidator(git_provider_config['base_url']).validate(
                requires_namespace_placeholder=True,
                requires_module_placeholder=True,
                requires_tag_placeholder=False,
                requires_path_placeholder=False
            )
            GitUrlValidator(git_provider_config['clone_url']).validate(
                requires_namespace_placeholder=True,
                requires_module_placeholder=True,
                requires_tag_placeholder=False,
                requires_path_placeholder=False
            )
            GitUrlValidator(git_provider_config['browse_url']).validate(
                requires_namespace_placeholder=True,
                requires_module_placeholder=True,
                requires_tag_placeholder=True,
                requires_path_placeholder=True
            )

            # Check if git provider exists in DB
            existing_git_provider = GitProvider.get_by_name(name=git_provider_config['name'])
            if existing_git_provider:
                # Update existing row
                upsert = db.git_provider.update().where(
                    db.git_provider.c.id == existing_git_provider.pk
                ).values(
                    base_url_template=git_provider_config['base_url'],
                    clone_url_template=git_provider_config['clone_url'],
                    browse_url_template=git_provider_config['browse_url']
                )
            else:
                upsert = db.git_provider.insert().values(
                    name=git_provider_config['name'],
                    base_url_template=git_provider_config['base_url'],
                    clone_url_template=git_provider_config['clone_url'],
                    browse_url_template=git_provider_config['browse_url']
                )
            with db.get_connection() as conn:
                conn.execute(upsert)

    @classmethod
    def get_by_name(cls, name):
        """Return instance of git provider by name."""
        db = Database.get()
        # Obtain row from git providers table where name
        # matches git provider name
        select = db.git_provider.select().where(
            db.git_provider.c.name == name
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            row = res.fetchone()

        # If git provider found with name, return instance
        # of git provider object with ID
        if row:
            return cls(id=row['id'])

        # Otherwise return None
        return None

    @classmethod
    def get_all(cls):
        """Return all repository providers."""
        db = Database.get()
        # Obtain row from git providers table where name
        # matches git provider name
        select = db.git_provider.select()
        with db.get_connection() as conn:
            res = conn.execute(select)
            return [
                cls(id=row['id'])
                for row in res
            ]

    @classmethod
    def get(cls, id):
        """Create object and validate that it exists."""
        git_provider = cls(id=id)
        if git_provider._get_db_row() is None:
            return None
        return git_provider

    @property
    def pk(self):
        """Return DB ID for git provider."""
        return self._get_db_row()['id']

    @property
    def name(self):
        """Return name for git provider."""
        return self._get_db_row()['name']

    @property
    def clone_url_template(self):
        """Return clone_url_template for git provider."""
        return self._get_db_row()['clone_url_template']

    @property
    def base_url_template(self):
        """Return base_url for git provider."""
        return self._get_db_row()['base_url_template']

    @property
    def browse_url_template(self):
        """Return browse_url for git provider."""
        return self._get_db_row()['browse_url_template']

    def __init__(self, id):
        """Store member variable for ID."""
        self._id = id
        self._row_cache = None

    def _get_db_row(self):
        """Return DB row for git provider."""
        if self._row_cache is None:
            db = Database.get()
            # Obtain row from git providers table for git provider.
            select = db.git_provider.select().where(
                db.git_provider.c.id == self._id
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                return res.fetchone()
        return self._row_cache


class Namespace(object):

    @staticmethod
    def get_total_count():
        """Get total number of namespaces."""
        db = Database.get()
        counts = sqlalchemy.select(
            db.module_provider.c.namespace
        ).select_from(
            db.module_provider
        ).group_by(
            db.module_provider.c.namespace
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        with db.get_connection() as conn:
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
            # Check if analytics token is the example provided
            # in the config
            if namespace_split[0] == terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN:
                # Return None for analytics token, acting like one has
                # not been provided.
                return namespace_split[1], None

            return namespace_split[1], namespace_split[0]

        # If there were not two element (more or less),
        # return original value
        return namespace, None

    @staticmethod
    def get_all(only_published=False):
        """Return all namespaces."""
        db = Database.get()

        if only_published:
            # If only getting namespaces, with published/visible versions,
            # query module provider, joining to latest module version
            modules_query = db.select_module_provider_joined_latest_module_version(
                db.module_provider
            )

            modules_query = modules_query.where(
                db.module_version.c.published == True,
                db.module_version.c.beta == False
            )

            modules_query = modules_query.subquery()

            namespace_query = sqlalchemy.select(
                modules_query.c.namespace
            ).select_from(modules_query).group_by(
                modules_query.c.namespace
            ).order_by(
                modules_query.c.namespace
            )
        else:
            namespace_query = sqlalchemy.select(
                db.module_provider.c.namespace
            ).select_from(
                db.module_provider
            ).group_by(
                db.module_provider.c.namespace
            ).order_by(
                db.module_provider.c.namespace
            )

        with db.get_connection() as conn:
            res = conn.execute(namespace_query)

            namespaces = [r['namespace'] for r in res]
            return [
                Namespace(name=namespace)
                for namespace in namespaces
            ]

    @property
    def base_directory(self):
        """Return base directory."""
        return safe_join_paths(terrareg.config.Config().DATA_DIRECTORY, 'modules', self._name)

    @property
    def name(self):
        """Return name."""
        return self._name

    @property
    def is_auto_verified(self):
        """Whether the namespace is set to verfied in the config."""
        return self.name in terrareg.config.Config().VERIFIED_MODULE_NAMESPACES

    @property
    def trusted(self):
        """Whether namespace is trusted."""
        return self.name in terrareg.config.Config().TRUSTED_NAMESPACES

    @staticmethod
    def _validate_name(name):
        """Validate name of namespace"""
        if not re.match(r'^[0-9a-zA-Z][0-9a-zA-Z-_]+[0-9A-Za-z]$', name):
            raise InvalidNamespaceNameError('Namespace name is invalid')

    def __init__(self, name: str):
        """Validate name and store member variables"""
        self._validate_name(name)
        self._name = name

    def get_view_url(self):
        """Return view URL"""
        return '/modules/{namespace}'.format(namespace=self.name)

    def get_details(self):
        """Return custom terrareg details about namespace."""
        return {
            'is_auto_verified': self.is_auto_verified,
            'trusted': self.trusted
        }

    def get_all_modules(self):
        """Return all modules for namespace."""
        db = Database.get()
        select = sqlalchemy.select(
            db.module_provider.c.module
        ).select_from(db.module_provider).where(
            db.module_provider.c.namespace == self.name
        ).group_by(
            db.module_provider.c.module
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            modules = [r['module'] for r in res]

        return [
            Module(namespace=self, name=module)
            for module in modules
        ]

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)


class Module(object):

    @staticmethod
    def _validate_name(name):
        """Validate name of module"""
        if not re.match(r'^[0-9a-zA-Z][0-9a-zA-Z-_]+[0-9A-Za-z]$', name):
            raise InvalidModuleNameError('Module name is invalid')

    @property
    def name(self):
        """Return name."""
        return self._name

    @property
    def base_directory(self):
        """Return base directory."""
        return safe_join_paths(self._namespace.base_directory, self._name)

    def __init__(self, namespace: Namespace, name: str):
        """Validate name and store member variables."""
        self._validate_name(name)
        self._namespace = namespace
        self._name = name

    def get_view_url(self):
        """Return view URL"""
        return '{namespace_url}/{module}'.format(
            namespace_url=self._namespace.get_view_url(),
            module=self.name
        )

    def get_providers(self):
        """Return module providers for module."""
        db = Database.get()
        select = sqlalchemy.select(
            db.module_provider.c.provider
        ).select_from(
            db.module_provider
        ).where(
            db.module_provider.c.namespace == self._namespace.name,
            db.module_provider.c.module == self.name
        ).group_by(
            db.module_provider.c.provider
        )
        with db.get_connection() as conn:
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


class ModuleDetails:
    """Object to store common details between root module, submodules and examples."""

    @classmethod
    def create(cls):
        """Create instance of object in database."""
        # Create module details row
        db = Database.get()
        module_details_insert = db.module_details.insert().values()
        with db.get_connection() as conn:
            insert_res = conn.execute(module_details_insert)

        return cls(id=insert_res.inserted_primary_key[0])

    @property
    def pk(self):
        """Return ID of module details row."""
        return self._id

    @property
    def terraform_docs(self):
        """Return terraform_docs column"""
        if self._get_db_row():
            return self._get_db_row()['terraform_docs']
        return None

    @property
    def readme_content(self):
        """Return readme_content column"""
        if self._get_db_row():
            return self._get_db_row()['readme_content']
        return None

    @property
    def tfsec(self):
        """Return tfsec data."""
        # If module scanning is disabled, do not return the tfsec output
        if (terrareg.config.Config().ENABLE_SECURITY_SCANNING and
                self._get_db_row() is not None and
                self._get_db_row()['tfsec']):
            return json.loads(self._get_db_row()['tfsec'])
        return {'results': None}

    def __init__(self, id: int):
        """Store member variables."""
        self._id = id
        self._cache_db_row = None

    def _get_db_row(self):
        """Return database row for module details."""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_details.select(
            ).where(
                db.module_details.c.id == self.pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row

    def get_db_where(self, db: Database, statement):
        """Return DB where statement"""
        return statement.where(
            db.module_details.c.id == self.pk
        )

    def update_attributes(self, **kwargs):
        """Update DB row."""
        # Check for any blob and encode the values
        for kwarg in kwargs:
            if kwarg in ['readme_content', 'terraform_docs', 'tfsec']:
                kwargs[kwarg] = Database.encode_blob(kwargs[kwarg])

        db = Database.get()
        update = self.get_db_where(
            db=db, statement=db.module_details.update()
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None

    def delete(self):
        """Delete from database."""
        assert self.pk is not None
        db = Database.get()

        with db.get_connection() as conn:
            # Delete module details from module_details table
            delete_statement = db.module_details.delete().where(
                db.module_details.c.id == self.pk
            )
            conn.execute(delete_statement)


class ProviderLogo:

    INFO = {
        'aws': {
            'source': '/static/images/PB_AWS_logo_RGB_stacked.547f032d90171cdea4dd90c258f47373c5573db5.png',
            'tos': 'Amazon Web Services, AWS, the Powered by AWS logo are trademarks of Amazon.com, Inc. or its affiliates.',
            'alt': 'Powered by AWS Cloud Computing',
            'link': 'https://aws.amazon.com/'
        },
        'gcp': {
            'source': '/static/images/gcp.png',
            'tos': 'Google Cloud and the Google Cloud logo are trademarks of Google LLC.',
            'alt': 'Google Cloud',
            'link': 'https://cloud.google.com/'
        },
        'null': {
            'source': '/static/images/null.png',
            'tos': ' ',
            'alt': 'Null Provider',
            'link': '#'
        },
        'datadog': {
            'source': '/static/images/dd_logo_v_rgb.png',
            'tos': 'All \'Datadog\' modules are designed to work with Datadog. Modules are in no way affiliated with nor endorsed by Datadog Inc.',
            'alt': 'Works with Datadog',
            'link': 'https://www.datadoghq.com/'
        }
    }

    @staticmethod
    def get_all():
        """Return all provider logos"""
        return [
            ProviderLogo(provider)
            for provider in ProviderLogo.INFO
        ]

    def __init__(self, provider):
        """Store details and provider."""
        self._provider = provider
        self._details = ProviderLogo.INFO.get(provider, None)
        if self._details is not None:
            # Ensure required attributes exist for logo
            for attr in ['source', 'tos', 'alt', 'link']:
                assert attr in self._details and self._details[attr]

    @property
    def provider(self):
        """Return name of provider"""
        return self._provider

    @property
    def exists(self):
        """Determine whether logo exists."""
        return self._details is not None

    @property
    def source(self):
        """Return logo source URL."""
        return self._details['source'] if self._details is not None else None

    @property
    def tos(self):
        """Return logo terms of service."""
        return self._details['tos'] if self._details is not None else None

    @property
    def alt(self):
        """Return logo image alt text."""
        return self._details['alt'] if self._details is not None else None

    @property
    def link(self):
        """Return logo image link URL."""
        return self._details['link'] if self._details is not None else None


class ModuleProvider(object):

    @staticmethod
    def _validate_name(name):
        """Validate name of module"""
        if not re.match(r'^[0-9a-z]+$', name):
            raise InvalidModuleProviderNameError('Module provider name is invalid')

        # Check if providers allow-list is enabled
        # and check if name in list of allowed providers
        if terrareg.config.Config().ALLOWED_PROVIDERS and name not in terrareg.config.Config().ALLOWED_PROVIDERS:
            raise ProviderNameNotPermittedError(
                'Provider name is not in the list of alllowed providers.'
            )

    @staticmethod
    def get_total_count():
        """Get total number of module providers."""
        db = Database.get()
        counts = sqlalchemy.select(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).select_from(
            db.module_provider
        ).group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        with db.get_connection() as conn:
            res = conn.execute(select)

            return res.scalar()

    @classmethod
    def create(cls, module, name):
        """Create instance of object in database."""
        # Create module provider
        db = Database.get()
        module_provider_insert = db.module_provider.insert().values(
            namespace=module._namespace.name,
            module=module.name,
            provider=name,
            verified=module._namespace.is_auto_verified
        )
        with db.get_connection() as conn:
            conn.execute(module_provider_insert)

        return cls(module=module, name=name)

    @classmethod
    def get(cls, module, name, create=False):
        """Create object and ensure the object exists."""
        obj = cls(module=module, name=name)

        # If there is no row, the module provider does not exist
        if obj._get_db_row() is None:

            # If set to create and auto module-provider creation
            # is enabled in config, create the module provider
            if create and terrareg.config.Config().AUTO_CREATE_MODULE_PROVIDER:
                cls.create(module=module, name=name)

                return obj

            # If not creating, return None
            return None

        # Otherwise, return object
        return obj

    def get_logo(self):
        """Return logo for provider."""
        return ProviderLogo(provider=self.name)

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

    @property
    def pk(self):
        """Return database ID of module provider."""
        return self._get_db_row()['id']

    @property
    def verified(self):
        """Return whether module provider is verified."""
        return self._get_db_row()['verified']

    @property
    def git_tag_format(self):
        """Return git tag format"""
        if self._get_db_row()['git_tag_format']:
            return self._get_db_row()['git_tag_format']
        # Return default format template for just version
        return '{version}'

    @property
    def git_ref_format(self):
        return 'refs/tags/{}'.format(self.git_tag_format)

    @property
    def tag_ref_regex(self):
        """Return regex match for git ref to match version"""
        # Hacky method to replace placeholder with temporary string,
        # escape regex characters and then replace temporary string
        # with regex for version
        string_does_not_exist = 'th15w1lln3v3rc0m3up1Pr0m153'
        version_re = self.git_ref_format.format(version=string_does_not_exist)
        version_re = re.escape(version_re)
        # Add EOL and SOL characters
        version_re = '^{version_re}$'.format(version_re=version_re)
        # Replace temporary string with regex for symatec version
        version_re = version_re.replace(string_does_not_exist, r'(\d+\.\d+.\d+)')
        # Return copmiled regex
        return re.compile(version_re)

    def get_version_from_tag_ref(self, tag_ref):
        """Match tag ref against version number and return actual version number."""
        # Handle empty/None tag_ref
        if not tag_ref:
            return None

        res = self.tag_ref_regex.match(tag_ref)
        if res:
            return res.group(1)
        return None

    @property
    def tag_version_regex(self):
        """Return regex match for git ref to match version"""
        # Hacky method to replace placeholder with temporary string,
        # escape regex characters and then replace temporary string
        # with regex for version
        string_does_not_exist = 'th15w1lln3v3rc0m3up1Pr0m153'
        version_re = self.git_tag_format.format(version=string_does_not_exist)
        version_re = re.escape(version_re)
        # Add EOL and SOL characters
        version_re = '^{version_re}$'.format(version_re=version_re)
        # Replace temporary string with regex for symatec version
        version_re = version_re.replace(string_does_not_exist, r'(\d+\.\d+.\d+)')
        # Return copmiled regex
        return re.compile(version_re)

    def get_version_from_tag(self, tag):
        """Match tag against version number and return actual version number."""
        # Handle empty/None tag_ref
        if not tag:
            return None

        res = self.tag_version_regex.match(tag)
        if res:
            return res.group(1)
        return None

    @property
    def base_directory(self):
        """Return base directory."""
        return safe_join_paths(self._module.base_directory, self._name)

    def __init__(self, module: Module, name: str):
        """Validate name and store member variables."""
        self._validate_name(name)
        self._module = module
        self._name = name
        self._cache_db_row = None

    def get_db_where(self, db, statement):
        """Filter DB query by where for current object."""
        return statement.where(
            db.module_provider.c.namespace == self._module._namespace.name,
            db.module_provider.c.module == self._module.name,
            db.module_provider.c.provider == self.name
        )

    def _get_db_row(self):
        """Return database row for module provider."""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_provider.select(
            ).where(
                db.module_provider.c.namespace == self._module._namespace.name,
                db.module_provider.c.module == self._module.name,
                db.module_provider.c.provider == self.name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row

    def delete(self):
        """DELETE module provider, all module version and all associated subversions."""
        # Delete all versions
        for module_version in self.get_versions(include_beta=True, include_unpublished=True):
            module_version.delete()

        db = Database.get()

        with db.get_connection() as conn:
            # Delete module from module_version table
            delete_statement = db.module_provider.delete().where(
                db.module_provider.c.id == self.pk
            )
            conn.execute(delete_statement)

    def get_git_provider(self):
        """Return the git provider associated with this module provider."""
        if self._get_db_row()['git_provider_id']:
            return GitProvider.get(id=self._get_db_row()['git_provider_id'])
        return None

    def get_git_clone_url(self):
        """Return URL to perform git clone"""
        template = None

        # Check if allowed and module provider has custom git URL
        if (terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER and
                self._get_db_row()['repo_clone_url_template']):
            template = self._get_db_row()['repo_clone_url_template']

        # Otherwise, check if module provider is configured with git provider
        elif self.get_git_provider():
            template = self.get_git_provider().clone_url_template

        # Return rendered version of template
        if template:
            rendered_url = template.format(
                namespace=self._module._namespace.name,
                module=self._module.name,
                provider=self.name
            )

            return rendered_url

        return None

    def update_attributes(self, **kwargs):
        """Update DB row."""
        db = Database.get()
        update = self.get_db_where(
            db=db, statement=db.module_provider.update()
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None

    def update_git_provider(self, git_provider: GitProvider):
        """Update git provider associated with module provider."""
        self.update_attributes(
            git_provider_id=(git_provider.pk if git_provider is not None else None)
        )

    def update_git_tag_format(self, git_tag_format):
        """Update git_tag_format."""
        sanitised_git_tag_format = urllib.parse.quote(git_tag_format, safe='/{}')

        if git_tag_format:
            # If tag format was provided, ensured it can be passed with 'format'
            try:
                sanitised_git_tag_format.format(version='1.1.1')
                assert '{version}' in sanitised_git_tag_format
            except (ValueError, AssertionError):
                raise InvalidGitTagFormatError('Invalid git tag format. Must contain one placeholder: {version}.')
        else:
            # If not value was provided, default to None
            sanitised_git_tag_format = None
        self.update_attributes(git_tag_format=sanitised_git_tag_format)

    def update_repo_clone_url_template(self, repo_clone_url_template):
        """Update repository URL for module provider."""
        if repo_clone_url_template:
            # Check whether custom URLs is disabled, if attempting to set to a URL
            if not terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER:
                raise ModuleProviderCustomGitRepositoryUrlNotAllowedError(
                    'Custom module provider git repository URL cannot be set.'
                )

            converted_template = repo_clone_url_template.format(
                namespace=self._module._namespace.name,
                module=self._module.name,
                provider=self.name)

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'Repository URL does not contain a scheme (e.g. ssh://)'
                )
            if url.scheme not in ['http', 'https', 'ssh']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'Repository URL contains an unknown scheme (e.g. https/ssh/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'Repository URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryDoesNotContainPathError(
                    'Repository URL does not contain a path'
                )

            repo_clone_url_template = urllib.parse.quote(repo_clone_url_template, safe='\{\}/:@%?=')

        self.update_attributes(repo_clone_url_template=repo_clone_url_template)

    def update_repo_browse_url_template(self, repo_browse_url_template):
        """Update browse URL template for module provider."""
        if repo_browse_url_template:
            # Check whether custom URLs is disabled, if attempting to set to a URL
            if not terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER:
                raise ModuleProviderCustomGitRepositoryUrlNotAllowedError(
                    'Custom module provider git repository URL cannot be set.'
                )
            GitUrlValidator(repo_browse_url_template).validate(
                requires_path_placeholder=True,
                requires_tag_placeholder=True
            )

            converted_template = repo_browse_url_template.format(
                namespace=self._module._namespace.name,
                module=self._module.name,
                provider=self.name,
                tag='',
                path='')

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'Repository URL does not contain a scheme (e.g. https://)'
                )
            if url.scheme not in ['http', 'https']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'Repository URL contains an unknown scheme (e.g. https/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'Repository URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryDoesNotContainPathError(
                    'Repository URL does not contain a path'
                )

            repo_browse_url_template = urllib.parse.quote(repo_browse_url_template, safe='\{\}/:@%?=')

        self.update_attributes(repo_browse_url_template=repo_browse_url_template)

    def update_repo_base_url_template(self, repo_base_url_template):
        """Update browse URL template for module provider."""
        if repo_base_url_template:
            # Check whether custom URLs is disabled, if attempting to set to a URL
            if not terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER:
                raise ModuleProviderCustomGitRepositoryUrlNotAllowedError(
                    'Custom module provider git repository URL cannot be set.'
                )

            converted_template = repo_base_url_template.format(
                namespace=self._module._namespace.name,
                module=self._module.name,
                provider=self.name)

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'Repository URL does not contain a scheme (e.g. https://)'
                )
            if url.scheme not in ['http', 'https']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'Repository URL contains an unknown scheme (e.g. https/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'Repository URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryDoesNotContainPathError(
                    'Repository URL does not contain a path'
                )

            repo_base_url_template = urllib.parse.quote(repo_base_url_template, safe='\{\}/:@%?=')

        self.update_attributes(repo_base_url_template=repo_base_url_template)

    def get_view_url(self):
        """Return view URL"""
        return '{module_url}/{module}'.format(
            module_url=self._module.get_view_url(),
            module=self.name
        )

    def get_latest_version(self):
        """Return latest published version of module."""
        db = Database.get()
        select = sqlalchemy.select(db.module_version.c.version).select_from(db.module_provider).join(
            db.module_version,
            db.module_provider.c.latest_version_id==db.module_version.c.id
        ).where(
            db.module_provider.c.id==self.pk
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            version = res.fetchone()

        if version is None:
            return None

        return ModuleVersion(module_provider=self, version=version['version'])

    def calculate_latest_version(self):
        """Obtain all versions of module and sort by semantec version numbers to obtain latest version."""
        db = Database.get()
        select = db.select_module_version_joined_module_provider(
            db.module_version.c.version
        ).where(
            db.module_provider.c.namespace == self._module._namespace.name,
            db.module_provider.c.module == self._module.name,
            db.module_provider.c.provider == self.name,
            db.module_version.c.published == True,
            db.module_version.c.beta == False
        )
        with db.get_connection() as conn:
            res = conn.execute(select)

            # Convert to list
            rows = [r for r in res]

        # Sort rows by semantec versioning
        rows.sort(key=lambda x: LooseVersion(x['version']), reverse=True)

        # Ensure at least one row
        if not rows:
            return None

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

    def get_versions(self, include_beta=True, include_unpublished=False):
        """Return all module provider versions."""
        db = Database.get()

        select = db.select_module_version_joined_module_provider(
            db.module_version.c.version
        ).where(
            db.module_provider.c.namespace == self._module._namespace.name,
            db.module_provider.c.module == self._module.name,
            db.module_provider.c.provider == self.name
        )
        # Remove unpublished versions, it not including them
        if not include_unpublished:
            select = select.where(
                db.module_version.c.published == True
            )
        # Remove beta versions if not including them
        if not include_beta:
            select = select.where(
                db.module_version.c.beta == False
            )

        with db.get_connection() as conn:
            res = conn.execute(select)
            module_versions = [
                ModuleVersion(module_provider=self, version=r['version'])
                for r in res
            ]
        module_versions.sort(
            key=lambda x: LooseVersion(x.version),
            reverse=True
        )
        return module_versions

    def get_api_outline(self):
        """Return dict of basic provider details for API response."""
        return {
            "id": self.id,
            "namespace": self._module._namespace.name,
            "name": self._module.name,
            "provider": self.name,
            "verified": self.verified,
            "trusted": self._module._namespace.trusted
        }

    def get_api_details(self, include_beta=True):
        """Return dict of provider details for API response."""
        api_details = self.get_api_outline()
        api_details.update({
            "versions": [v.version for v in self.get_versions(include_beta=include_beta)]
        })
        return api_details

    def get_terrareg_api_details(self):
        """Return dict of module details with additional attributes used by terrareg UI."""
        git_provider = self.get_git_provider()
        # Obtain base API details, but do not include beta versions,
        # as these should not be displayed in the UI
        api_details = self.get_api_details(include_beta=False)
        api_details.update({
            "module_provider_id": self.id,
            "git_provider_id": git_provider.pk if git_provider else None,
            "git_tag_format": self.git_tag_format,
            "repo_base_url_template": self._get_db_row()['repo_base_url_template'],
            "repo_clone_url_template": self._get_db_row()['repo_clone_url_template'],
            "repo_browse_url_template": self._get_db_row()['repo_browse_url_template']
        })
        return api_details

    def get_integrations(self):
        """Return integration URL and details"""
        integrations = {
            'import': {
                'method': 'POST',
                'url': f'/v1/terrareg/modules/{self.id}/${{version}}/import',
                'description': 'Trigger module version import',
                'notes': ''
            },
            'hooks_bitbucket': {
                'method': None,
                'url': f'/v1/terrareg/modules/{self.id}/hooks/bitbucket',
                'description': 'Bitbucket hook trigger',
                'notes': ''
            },
            'hooks_github': {
                'method': None,
                'url': f'/v1/terrareg/modules/{self.id}/hooks/github',
                'description': 'Github hook trigger',
                'notes': 'Only accepts `Releases` events, all other events will return an error.',
            },
            'hooks_gitlab': {
                'method': None,
                'url': f'/v1/terrareg/modules/{self.id}/hooks/gitlab',
                'description': 'Gitlab hook trigger',
                'notes': '',
                'coming_soon': True
            },
            'publish': {
                'method': 'POST',
                'url': f'/v1/terrareg/modules/{self.id}/${{version}}/publish',
                'description': 'Mark module version as published',
                'notes': ''
            }
        }
        if terrareg.config.Config().ALLOW_MODULE_HOSTING:
            integrations['upload'] = {
                'method': 'POST',
                'url': f'/v1/terrareg/modules/{self.id}/${{version}}/upload',
                'description': 'Create module version using source archive',
                'notes': 'Source ZIP file must be provided as data.'
            }
        return integrations


class TerraformSpecsObject(object):
    """Base terraform object, that has terraform-docs available."""

    @classmethod
    def get(cls, *args, **kwargs):
        """Create object and ensure the object exists."""
        obj = cls(*args, **kwargs)

        # If there is no row, return None
        if obj._get_db_row() is None:
            return None
        # Otherwise, return object
        return obj

    def __init__(self):
        """Setup member variables."""
        self._module_specs = None

    @property
    def module_version(self):
        """Return module version"""
        raise NotImplementedError

    @property
    def path(self):
        """Return module path"""
        raise NotImplementedError
    @property
    def is_submodule(self):
        """Whether object is submodule."""
        raise NotImplementedError

    @property
    def is_example(self):
        """Whether object is an example."""
        return False

    @property
    def pk(self):
        """Return primary key of database row"""
        return self._get_db_row()['id']

    @property
    def registry_id(self):
        """Return registry path ID (with excludes version)."""
        raise NotImplementedError

    def _get_db_row(self):
        """Must be implemented by object. Return row from DB."""
        raise NotImplementedError

    def get_tfsec_failure_count(self):
        """Return number of tfsec failures."""
        # Handle when results in None
        module_details = self.module_details
        if module_details is None or module_details.tfsec['results'] is None:
            return 0

        count = 0
        # Count each of the test failures
        for result in self.module_details.tfsec['results']:
            # TFsec status of 0 is a fail
            if result['status'] == 0:
                count += 1
        return count

    def get_module_specs(self):
        """Return module specs"""
        if self._module_specs is None:
            module_specs = {}

            module_details = self.module_details
            if module_details:
                raw_json = Database.decode_blob(module_details.terraform_docs)
                if raw_json:
                    module_specs = json.loads(raw_json)
            self._module_specs = module_specs
        return self._module_specs

    def get_readme_html(self, server_hostname):
        """Replace examples in README and convert readme markdown to HTML"""
        readme_md = self.get_readme_content()
        if readme_md:
            readme_md = self.replace_source_in_file(
                readme_md, server_hostname)
            return markdown.markdown(
                readme_md,
                extensions=['fenced_code', 'tables']
            )
        return None

    @property
    def module_details(self):
        """Return instance of ModuleDetails for object."""
        if self._get_db_row() and self._get_db_row()['module_details_id']:
            return ModuleDetails(id=self._get_db_row()['module_details_id'])
        else:
            return None

    def get_readme_content(self):
        """Get readme contents"""
        module_details = self.module_details
        if module_details:
            return Database.decode_blob(module_details.readme_content)
        return None

    def get_terraform_inputs(self):
        """Obtain module inputs"""
        return self.get_module_specs().get('inputs', [])

    def get_terraform_outputs(self):
        """Obtain module inputs"""
        return self.get_module_specs().get('outputs', [])

    def get_terraform_resources(self):
        """Obtain module resources."""
        return self.get_module_specs().get('resources', [])

    def get_terraform_dependencies(self):
        """Obtain module dependencies."""
        #return self.get_module_specs()['requirements']
        # @TODO Verify what this should be - terraform example is empty and real-world examples appears to
        # be empty, but do have an undocumented 'provider_dependencies'
        return []

    def get_terraform_provider_dependencies(self):
        """Obtain module dependencies."""
        providers = []
        for provider in self.get_module_specs().get('providers', []):

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

    def replace_source_in_file(self, content: str, server_hostname: str):
        """Replace single 'source' line in example/readme files."""
        def callback(match):
            # Convert relative path to absolute
            module_path = os.path.abspath(
                # Since example path does not contain a leading slash,
                # prepend one to perform the abspath relative to root.
                os.path.join('/', self.path, match.group(3))
            )
            # If the path is empty (at root),
            # leave the path blank
            if module_path == '/':
                module_path = ''
            else:
                # Otherwise, prepend with additional leading slash,
                # for the terraform annotation for a sub-directory within
                # the module
                module_path = '/{module_path}'.format(module_path=module_path)

            trailing_space_count = len(match.group(2))
            # If only 1 leading space before source '=' character,
            # increment by 1 to align to 'version'
            if trailing_space_count < 2:
                trailing_space_count = 2

            return ('\n{leading_space}source{trailing_space}= "{server_hostname}/{module_provider_id}{sub_dir}"\n'
                    '{leading_space}version{version_trailing_space}= "{version_string}"\n').format(
                leading_space=match.group(1),
                trailing_space=(' ' * trailing_space_count),
                version_trailing_space=(' ' * (trailing_space_count - 1)),
                server_hostname=server_hostname,
                module_provider_id=self.module_version.module_provider.id,
                sub_dir=module_path,
                version_string=self.module_version.get_terraform_example_version_string()
            )

        return re.sub(
            r'\n([ \t]*)source(\s+)=\s+"(\..*)"[ \t]*\n',
            callback,
            content,
            re.MULTILINE
        )


class ModuleVersion(TerraformSpecsObject):

    @staticmethod
    def get_total_count():
        """Get total number of module versions."""
        db = Database.get()
        counts = db.select_module_version_joined_module_provider(
            db.module_version.c.version
        ).group_by(
            db.module_provider.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider,
            db.module_version.c.version
        ).subquery()

        select = sqlalchemy.select([sqlalchemy.func.count()]).select_from(counts)

        with db.get_connection() as conn:
            res = conn.execute(select)

            return res.scalar()

    @staticmethod
    def _validate_version(version):
        """Validate version, checking if version is a beta version."""
        match = re.match(r'^[0-9]+\.[0-9]+\.[0-9]+((:?-[a-z0-9]+)?)$', version)
        if not match:
            raise InvalidVersionError('Version is invalid')
        return bool(match.group(1))

    @property
    def is_submodule(self):
        """Whether object is submodule."""
        return False

    @property
    def publish_date_display(self):
        """Return display view of date of module published."""
        published_at = self._get_db_row()['published_at']
        if published_at:
            return published_at.strftime('%B %d, %Y')
        return None

    @property
    def owner(self):
        """Return owner of module."""
        return self._get_db_row()['owner']

    @property
    def published(self):
        """Return whether module is published"""
        return bool(self._get_db_row()['published'])

    @property
    def description(self):
        """Return description."""
        return self._get_db_row()['description']

    @property
    def version(self):
        """Return version."""
        return self._version

    @property
    def source_git_tag(self):
        """Return git tag used for extraction clone"""
        return self._module_provider.git_tag_format.format(version=self._version)

    @property
    def git_tag_ref(self):
        """Return git tag ref for extraction."""
        return self._module_provider.git_ref_format.format(version=self._version)

    @property
    def base_directory(self):
        """Return base directory."""
        return safe_join_paths(self._module_provider.base_directory, self._version)

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
        return safe_join_paths(self.base_directory, self.archive_name_tar_gz)

    @property
    def archive_path_zip(self):
        """Return full path of the archive file."""
        return safe_join_paths(self.base_directory, self.archive_name_zip)

    @property
    def beta(self):
        """Return whether module version is a beta version."""
        return self._get_db_row()['beta']

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
    def registry_id(self):
        """Return registry path ID (with excludes version)."""
        return self._module_provider.id

    @property
    def variable_template(self):
        """Return variable template for module version."""
        raw_json = Database.decode_blob(self._get_db_row()['variable_template'])
        variables = json.loads(raw_json) if raw_json else []

        # Detect bad type for variable template and replace
        # with empty array
        if type(variables) is not list:
            variables = []

        if terrareg.config.Config().AUTOGENERATE_USAGE_BUILDER_VARIABLES:
            for input_variable in self.get_terraform_inputs():
                if not input_variable['required']:
                    continue
                if input_variable['name'] not in [v['name'] for v in variables]:

                    converted_type = 'text'
                    if input_variable['type'] == 'bool':
                        converted_type = 'boolean'
                    elif input_variable['type'].startswith('list('):
                        converted_type = 'list'

                    variables.append({
                        'name': input_variable['name'],
                        'type': converted_type,
                        'additional_help': input_variable['description'],
                        'quote_value': True
                    })
        return variables

    @property
    def module_provider(self):
        """Return module provider"""
        return self._module_provider

    @property
    def module_version(self):
        """Return module version"""
        return self

    def __init__(self, module_provider: ModuleProvider, version: str):
        """Setup member variables."""
        self._extracted_beta_flag = self._validate_version(version)
        self._module_provider = module_provider
        self._version = version
        self._cache_db_row = None
        super(ModuleVersion, self).__init__()

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_version.select().join(
                db.module_provider, db.module_version.c.module_provider_id == db.module_provider.c.id
            ).where(
                db.module_provider.c.namespace == self._module_provider._module._namespace.name,
                db.module_provider.c.module == self._module_provider._module.name,
                db.module_provider.c.provider == self._module_provider.name,
                db.module_version.c.version == self.version
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def get_terraform_example_version_string(self):
        """Return formatted string of version parameter for example terraform."""
        # For beta versions, pass an exact version constraint.
        if self.beta:
            return self.version

        # Generate list of template values for formatting
        major, minor, patch = self.version.split('.')
        kwargs = {'major': major, 'minor': minor, 'patch': patch}
        for i in ['major', 'minor', 'patch']:
            val = int(kwargs[i])
            kwargs['{}_plus_one'.format(i)] = val + 1
            # Default minus_one values to 0, if they are already 0
            kwargs['{}_minus_one'.format(i)] = val - 1 if val > 0 else 0

        # Return formatted example template
        return terrareg.config.Config().TERRAFORM_EXAMPLE_VERSION_TEMPLATE.format(
            **kwargs
        )

    def get_view_url(self):
        """Return view URL"""
        return '{module_provider_url}/{version}'.format(
            module_provider_url=self._module_provider.get_view_url(),
            version=self.version
        )

    def get_git_clone_url(self):
        """Return URL to perform git clone"""
        template = None

        # Check if allowed, and module version has custom git URL
        if (terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION and
                self._get_db_row()['repo_clone_url_template']):
            template = self._get_db_row()['repo_clone_url_template']

        # Otherwise, get git clone URL from module provider
        elif self._module_provider.get_git_clone_url():
            template = self._module_provider.get_git_clone_url()

        # Return rendered version of template
        if template:
            rendered_url = template.format(
                namespace=self._module_provider._module._namespace.name,
                module=self._module_provider._module.name,
                provider=self._module_provider.name
            )

            return rendered_url

        return None

    def get_source_download_url(self, path=None):
        """Return URL to download source file."""
        rendered_url = None

        rendered_url = self.get_git_clone_url()

        # Return rendered version of template
        if rendered_url:
            # Check if scheme starts with git::, which is required
            # by terraform to acknowledge a git repository
            # and add if it not
            parsed_url = urllib.parse.urlparse(rendered_url)
            if not parsed_url.scheme.startswith('git::'):
                rendered_url = 'git::{rendered_url}'.format(rendered_url=rendered_url)

            # Check if path is present for module (only used for submodules)
            if path:
                rendered_url = '{rendered_url}//{path}'.format(
                    rendered_url=rendered_url,
                    path=path)

            # Add tag to URL
            rendered_url = '{rendered_url}?ref={tag}'.format(
                rendered_url=rendered_url,
                tag=self.source_git_tag
            )

            return rendered_url

        if terrareg.config.Config().ALLOW_MODULE_HOSTING:
            return '/v1/terrareg/modules/{0}/{1}'.format(self.id, self.archive_name_zip)

        raise NoModuleDownloadMethodConfiguredError(
            'Module is not configured with a git URL and direct downloads are disabled'
        )

    def get_source_browse_url(self, path=None):
        """Return URL to browse the source doe."""
        template = None

        # Check if allowed, and module version has custom git URL
        if terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION and self._get_db_row()['repo_browse_url_template']:
            template = self._get_db_row()['repo_browse_url_template']

        # Otherwise, check if allowed and module provider has custom git URL
        elif (terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER and
                self._module_provider._get_db_row()['repo_browse_url_template']):
            template = self._module_provider._get_db_row()['repo_browse_url_template']

        # Otherwise, check if module provider is configured with git provider
        elif self._module_provider.get_git_provider():
            template = self._module_provider.get_git_provider().browse_url_template

        # Return rendered version of template
        if template:
            validator = GitUrlValidator(template)
            return validator.get_value(
                namespace=self._module_provider._module._namespace.name,
                module=self._module_provider._module.name,
                provider=self._module_provider.name,
                tag=self.source_git_tag,
                # Default path to empty string to avoid
                # adding 'None' to string
                path=(path if path else '')
            )

        return None

    def get_source_base_url(self, path=None):
        """Return URL to view the source repository."""
        template = None

        # Check if allowed, and module version has custom git URL
        if (terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_VERSION and
                self._get_db_row()['repo_base_url_template']):
            template = self._get_db_row()['repo_base_url_template']

        # Otherwise, check if allowed and module provider has custom git URL
        elif (terrareg.config.Config().ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER and
                self._module_provider._get_db_row()['repo_base_url_template']):
            template = self._module_provider._get_db_row()['repo_base_url_template']

        # Otherwise, check if module provider is configured with git provider
        elif self._module_provider.get_git_provider():
            template = self._module_provider.get_git_provider().base_url_template

        # Return rendered version of template
        if template:
            return template.format(
                namespace=self._module_provider._module._namespace.name,
                module=self._module_provider._module.name,
                provider=self._module_provider.name
            )

        return None

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._module_provider.base_directory):
            self._module_provider.create_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    def publish(self):
        """Publish module version."""
        # Mark module version as published
        self.update_attributes(published=True)

        # Calculate latest version will take beta flag into account and will only match
        # the current version if the current version is latest and is capable of being the
        # latest version.
        if (self._module_provider.calculate_latest_version() is not None and
                self._module_provider.calculate_latest_version().version == self.version):
            self._module_provider.update_attributes(latest_version_id=self.pk)

    def get_api_outline(self):
        """Return dict of basic version details for API response."""
        row = self._get_db_row()
        api_outline = self._module_provider.get_api_outline()
        api_outline.update({
            "id": self.id,
            "owner": row['owner'],
            "version": self.version,
            "description": row['description'],
            "source": self.get_source_base_url(),
            "published_at": row['published_at'].isoformat() if row['published_at'] else None,
            "downloads": self.get_total_downloads(),
            "internal": self._get_db_row()['internal']
        })
        return api_outline

    def get_total_downloads(self):
        """Obtain total number of downloads for module version."""
        return terrareg.analytics.AnalyticsEngine.get_module_version_total_downloads(
            module_version=self
        )

    def get_api_details(self):
        """Return dict of version details for API response."""#
        api_details = self._module_provider.get_api_details()
        api_details.update(self.get_api_outline())
        api_details.update({
            "root": self.get_api_module_specs(),
            "submodules": [sm.get_api_module_specs() for sm in self.get_submodules()],
            "providers": [p.name for p in self._module_provider._module.get_providers()]
        })
        return api_details

    def get_terrareg_api_details(self):
        """Return dict of version details with additional attributes used by terrareg UI."""
        api_details = self._module_provider.get_terrareg_api_details()

        # Capture versions from module provider API output, as this limits
        # some versions, which are normally displayed in the Terraform APIs
        versions = api_details['versions']

        api_details.update(self.get_api_details())

        source_browse_url = self.get_source_browse_url()
        api_details.update({
            "published_at_display": self.publish_date_display,
            "display_source_url": source_browse_url if source_browse_url else self.get_source_base_url(),
            "terraform_example_version_string": self.get_terraform_example_version_string(),
            "versions": versions,
            "beta": self.beta,
            "published": self.published,
            "security_failures": self.get_tfsec_failure_count()
        })
        return api_details

    def prepare_module(self):
        """Handle file upload of module version."""
        self.create_data_directory()
        self._create_db_row()

    def get_db_where(self, db, statement):
        """Filter DB query by where for current object."""
        return statement.where(
            db.module_version.c.module_provider_id == self._module_provider.pk,
            db.module_version.c.version == self.version
        )

    def update_attributes(self, **kwargs):
        """Update attributes of module version in database row."""
        # Check for any blob and encode the values
        for kwarg in kwargs:
            if kwarg in ['variable_template']:
                kwargs[kwarg] = Database.encode_blob(kwargs[kwarg])

        db = Database.get()
        update = self.get_db_where(
            db=db, statement=db.module_version.update()
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Clear cached DB row
        self._cache_db_row = None

    def delete(self, delete_related_analytics=True):
        """Delete module version and all associated submodules."""
        for example in self.get_examples():
            example.delete()

        for submodule in self.get_submodules():
            submodule.delete()

        if delete_related_analytics:
            terrareg.analytics.AnalyticsEngine.delete_analytics_for_module_version(self)

        # Delete associated module details
        module_details = self.module_details
        if module_details:
            module_details.delete()

        db = Database.get()

        with db.get_connection() as conn:
            # Delete module from module_version table
            delete_statement = db.module_version.delete().where(
                db.module_version.c.id == self.pk
            )
            conn.execute(delete_statement)

            # Invalidate cache for previous DB row
            self._cache_db_row = None

        # Update latest version of parent module
        new_latest_version = self._module_provider.calculate_latest_version()
        self._module_provider.update_attributes(
            latest_version_id=(new_latest_version.pk if new_latest_version is not None else None)
        )

    def _create_db_row(self):
        """Insert into datadabase, removing any existing duplicate versions."""
        db = Database.get()

        # Delete pre-existing version, if it exists
        old_module_version_pk = None
        if self._get_db_row():
            old_module_version_pk = self.pk
            self.delete(delete_related_analytics=False)

        with db.get_connection() as conn:
            # Insert new module into table
            insert_statement = db.module_version.insert().values(
                module_provider_id=self._module_provider.pk,
                version=self.version,
                published=False,
                beta=self._extracted_beta_flag,
                internal=False
            )
            conn.execute(insert_statement)

        # Migrate analytics from old module version ID to new module version
        if old_module_version_pk is not None:
            terrareg.analytics.AnalyticsEngine.migrate_analytics_to_new_module_version(
                old_version_version_pk=old_module_version_pk,
                new_module_version=self)

    def get_submodules(self):
        """Return list of submodules."""
        db = Database.get()
        select = db.sub_module.select(
        ).join(db.module_version, db.module_version.c.id == db.sub_module.c.parent_module_version).where(
            db.module_version.c.id == self.pk,
            db.sub_module.c.type == Submodule.TYPE
        )
        with db.get_connection() as conn:
            res = conn.execute(select)

            return [
                Submodule(module_version=self, module_path=r['path'])
                for r in res
            ]

    def get_examples(self):
        """Return list of submodules."""
        db = Database.get()
        select = db.sub_module.select(
        ).join(db.module_version, db.module_version.c.id == db.sub_module.c.parent_module_version).where(
            db.module_version.c.id == self.pk,
            db.sub_module.c.type == Example.TYPE
        )
        with db.get_connection() as conn:
            res = conn.execute(select)

            return [
                Example(module_version=self, module_path=r['path'])
                for r in res
            ]


class BaseSubmodule(TerraformSpecsObject):
    """Base submodule, for submodule and examples from a module version."""

    TYPE = None

    @classmethod
    def get_by_id(cls, module_version: ModuleVersion, pk: int):
        """Return instance of submodule based on ID of submodule"""
        db = Database.get()
        select = db.sub_module.select().where(
            db.sub_module.c.id == pk,
            db.sub_module.c.type == cls.TYPE
        )
        with db.get_connection() as conn:
            row = conn.execute(select).fetchone()
        if row is None:
            return None

        return cls(module_version=module_version, module_path=row['path'])

    @classmethod
    def create(cls, module_version: ModuleVersion, module_path: str):
        """Create instance of object in database."""
        # Create submodule
        db = Database.get()
        insert_statement = db.sub_module.insert().values(
            parent_module_version=module_version.pk,
            type=cls.TYPE,
            path=module_path
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)

        # Return instance of object
        return cls(module_version=module_version, module_path=module_path)

    @property
    def pk(self):
        """Return DB primary key."""
        return self._get_db_row()['id']

    @property
    def path(self):
        """Return module path"""
        return self._module_path

    @property
    def id(self):
        """Return ID for module"""
        return '{0}//{1}'.format(self._module_version.id, self.path)

    @property
    def registry_id(self):
        """Return registry path ID (with excludes version)."""
        return '{0}//{1}'.format(self._module_version.registry_id, self.path)

    @property
    def is_submodule(self):
        """Whether object is submodule."""
        return True

    @property
    def module_version(self):
        """Return module version"""
        return self._module_version

    def __init__(self, module_version: ModuleVersion, module_path: str):
        self._module_version = module_version
        self._module_path = module_path
        self._cache_db_row = None
        super(BaseSubmodule, self).__init__()

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.sub_module.select().where(
                db.sub_module.c.parent_module_version == self._module_version.pk,
                db.sub_module.c.path == self._module_path,
                db.sub_module.c.type == self.TYPE
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def update_attributes(self, **kwargs):
        """Update DB row."""
        db = Database.get()
        update = db.sub_module.update().where(
            db.sub_module.c.id == self.pk
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None

    def delete(self):
        """Delete submodule from DB."""
        # Delete associated module details
        module_details = self.module_details
        if module_details:
            module_details.delete()

        db = Database.get()

        with db.get_connection() as conn:
            delete_statement = db.sub_module.delete().where(
                db.sub_module.c.id == self.pk
            )
            conn.execute(delete_statement)

        # Invalidate DB row cache
        self._cache_db_row = None

    def get_source_browse_url(self):
        """Get formatted source browse URL"""
        return self._module_version.get_source_browse_url(path=self.path)

    def get_source_download_url(self):
        """Get formatted source download URL"""
        return self._module_version.get_source_download_url(path=self.path)

    def get_view_url(self):
        """Return view URL"""
        return '{module_version_url}/{submodules_type}/{submodule_path}'.format(
            module_version_url=self._module_version.get_view_url(),
            submodules_type=self.TYPE,
            submodule_path=self.path
        )

    def get_terrareg_api_details(self):
        """Return dict of submodule details with additional attributes used by terrareg UI."""
        api_details = self.get_api_module_specs()
        source_browse_url = self.get_source_browse_url()
        api_details.update({
            "display_source_url": source_browse_url if source_browse_url else None,
            "security_failures": self.get_tfsec_failure_count()
        })
        return api_details


class Submodule(BaseSubmodule):

    TYPE = 'submodule'


class Example(BaseSubmodule):

    TYPE = 'example'

    @property
    def is_example(self):
        """Whether object is an example."""
        return True

    def get_files(self):
        """Return example files associated with example."""
        db = Database.get()
        select = db._example_file.select().where(
            db._example_file.c.submodule_id == self.pk
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            return [ExampleFile(example=self, path=row['path']) for row in res]

    def delete(self):
        """Delete any example files and self."""
        # Delete any example files
        for example_file in self.get_files():
            example_file.delete()

        # Call super method to delete self
        super(Example, self).delete()


class ExampleFile:

    @classmethod
    def create(cls, example: Example, path: str):
        """Create instance of object in database."""
        # Insert example file into database
        db = Database.get()
        insert_statement = db.example_file.insert().values(
            submodule_id=example.pk,
            path=path
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)

        # Return instance of object
        return cls(example=example, path=path)

    @staticmethod
    def get_by_path(module_version: ModuleVersion, file_path: str):
        """Return example file object by file path and module version"""
        db = Database.get()
        select = sqlalchemy.select(
            db.example_file.c.submodule_id
        ).select_from(
            db.example_file
        ).join(
            db.sub_module,
            db.example_file.c.submodule_id == db.sub_module.c.id
        ).join(
            db.module_version,
            db.sub_module.c.parent_module_version == db.module_version.c.id
        ).where(
            db.module_version.c.id == module_version.pk,
            db.example_file.c.path == file_path
        )
        with db.get_connection() as conn:
            row = conn.execute(select).fetchone()
        if not row:
            return None

        example = Example.get_by_id(module_version=module_version, pk=row['submodule_id'])

        if example is None:
            return None

        return ExampleFile(example=example, path=file_path)

    @property
    def file_name(self):
        """Return name of file"""
        return self._path.split('/')[-1]

    @property
    def path(self):
        """Return path of example file."""
        return self._path

    @property
    def pk(self):
        """Get ID from DB row"""
        return self._get_db_row()['id']

    @property
    def content(self):
        """Return content of example file."""
        return Database.decode_blob(self._get_db_row()['content'])

    def __init__(self, example: Example, path: str):
        """Store identifying data."""
        self._example = example
        self._path = path
        self._cache_db_row = None

    def __lt__(self, other):
        """Implement less than for sorting example files."""
        # If current file is main, it is always at the top, i.e. least value
        if self.file_name == 'main.tf':
            return True
        # If other file is main.tf, this file is more
        elif other.file_name == 'main.tf':
            return False

        return (self.file_name < other.file_name)

    def _get_db_row(self):
        """Return DB row for git provider."""
        if self._cache_db_row is None:
            db = Database.get()
            # Obtain row from git providers table for git provider.
            select = db.example_file.select().where(
                db.example_file.c.submodule_id == self._example.pk,
                db.example_file.c.path == self._path
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                return res.fetchone()
        return self._cache_db_row

    def update_attributes(self, **kwargs):
        """Update DB row."""
        # Encode columns that are binary blobs in the database
        for kwarg in kwargs:
            if kwarg in ['content']:
                kwargs[kwarg] = Database.encode_blob(kwargs[kwarg])

        db = Database.get()
        update = db.example_file.update().where(
            db.example_file.c.id == self.pk
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None

    def delete(self):
        """Delete example file from DB."""
        db = Database.get()

        with db.get_connection() as conn:
            delete_statement = db.example_file.delete().where(
                db.example_file.c.id == self.pk
            )
            conn.execute(delete_statement)

        # Invalidate DB row cache
        self._cache_db_row = None

    def get_content(self, server_hostname):
        """Return content with source replaced"""
        # Replace source lines that use relative paths
        return self._example.replace_source_in_file(
            content=self.content,
            server_hostname=server_hostname)
