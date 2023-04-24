
import datetime
from enum import Enum
import os
from distutils.version import LooseVersion
import json
import re
import secrets
import sqlalchemy
import urllib.parse

import markdown
import pygraphviz
import networkx as nx
import terrareg.analytics
from terrareg.database import Database
import terrareg.config
import terrareg.audit
import terrareg.audit_action
from terrareg.errors import (
    DuplicateNamespaceDisplayNameError, InvalidModuleNameError, InvalidModuleProviderNameError, InvalidNamespaceDisplayNameError, InvalidUserGroupNameError,
    InvalidVersionError, NamespaceAlreadyExistsError, NoModuleVersionAvailableError,
    InvalidGitTagFormatError, InvalidNamespaceNameError, ReindexingExistingModuleVersionsIsProhibitedError, RepositoryUrlContainsInvalidPortError, RepositoryUrlContainsInvalidTemplateError,
    RepositoryUrlDoesNotContainValidSchemeError,
    RepositoryUrlContainsInvalidSchemeError,
    RepositoryUrlDoesNotContainHostError,
    RepositoryUrlDoesNotContainPathError,
    InvalidGitProviderConfigError,
    ModuleProviderCustomGitRepositoryUrlNotAllowedError,
    NoModuleDownloadMethodConfiguredError,
    ProviderNameNotPermittedError, RepositoryUrlParseError
)
from terrareg.utils import convert_markdown_to_html, get_public_url_details, safe_join_paths, sanitise_html_content
from terrareg.validators import GitUrlValidator
from terrareg.constants import EXTRACTION_VERSION


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


class UserGroup:

    @staticmethod
    def _validate_name(name):
        """Validate name of user group"""
        if not re.match(r'^[\s0-9a-zA-Z-_]+$', name):
            raise InvalidUserGroupNameError('User group name is invalid')
        return True

    @classmethod
    def get_by_group_name(cls, name):
        """Obtain group by name."""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.user_group.select().where(
                db.user_group.c.name==name
            ))
            if row := res.fetchone():
                return cls(name=row['name'])

            return None

    @classmethod
    def create(cls, name, site_admin):
        """Create usergroup"""
        # Check if group exists with name
        if cls.get_by_group_name(name=name):
            return None

        # Check group name
        if not cls._validate_name(name):
            return None

        cls._insert_into_database(name=name, site_admin=site_admin)

        obj = cls(name=name)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_GROUP_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.name,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def _insert_into_database(cls, name, site_admin):
        """Insert new user group into database."""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group.insert().values(
                name=name, site_admin=site_admin
            ))

    @classmethod
    def get_all_user_groups(cls):
        """Obtain all user groups."""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.user_group.select())
            return [
                cls(row['name'])
                for row in res.fetchall()
            ]

    @property
    def pk(self):
        """Return DB ID of user group"""
        return self._get_db_row()['id']

    @property
    def name(self):
        """Return name of user group"""
        return self._name

    @property
    def site_admin(self):
        """Return site_admin property of user group"""
        return self._get_db_row()['site_admin']

    def __init__(self, name):
        """Store member variables"""
        self._name = name
        self._row_cache = None

    def __eq__(self, __o):
        """Check if two user groups are the same"""
        if isinstance(__o, self.__class__):
            return self.pk == __o.pk and self.name == __o.name and self.site_admin == __o.site_admin
        return super(UserGroup, self).__eq__(__o)

    def _get_db_row(self):
        """Return DB row for user group."""
        if self._row_cache is None:
            db = Database.get()
            # Obtain row from user group table.
            select = db.user_group.select().where(
                db.user_group.c.name == self._name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._row_cache = res.fetchone()
        return self._row_cache

    def delete(self):
        """Delete user group"""
        for group_permission in UserGroupNamespacePermission.get_permissions_by_user_group(self):
            group_permission.delete()

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_GROUP_DELETE,
            object_type=self.__class__.__name__,
            object_id=self.name,
            old_value=None, new_value=None
        )

        self._delete_from_database()

    def _delete_from_database(self):
        """Delete row from database"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group.delete().where(
                db.user_group.c.id==self.pk
            ))


class UserGroupNamespacePermission:

    @classmethod
    def get_permissions_by_user_group(cls, user_group):
        """Return permissions by user group"""
        db = Database.get()
        with db.get_connection() as conn:
            query = sqlalchemy.select(
                db.user_group.c.name.label('user_group_name'),
                db.namespace.c.namespace.label('namespace_name')
            ).select_from(
                db.user_group_namespace_permission
            ).join(
                db.user_group,
                db.user_group_namespace_permission.c.user_group_id==db.user_group.c.id
            ).join(
                db.namespace,
                db.user_group_namespace_permission.c.namespace_id==db.namespace.c.id
            ).where(
                db.user_group.c.id==user_group.pk
            )
            res = conn.execute(query)

            return [
                cls(
                    user_group=UserGroup(name=r['user_group_name']),
                    namespace=Namespace(name=r['namespace_name'])
                )
                for r in res.fetchall()
            ]

    @classmethod
    def get_permissions_by_user_group_and_namespace(cls, user_group, namespace):
        """Return permission by user group and namespace"""
        permissions = cls.get_permissions_by_user_groups_and_namespace([user_group], namespace)
        if len(permissions) > 1:
            raise Exception('Found more than 1 permission for user group/namespace')
        return permissions[0] if permissions else None

    @classmethod
    def get_permissions_by_user_groups_and_namespace(cls, user_groups, namespace):
        """Obtain user permission by multiple user groups for a single namespace"""
        db = Database.get()
        user_group_ids = [user_group.pk for user_group in user_groups]
        user_group_mapping = {
            user_group.name: user_group
            for user_group in user_groups
        }
        with db.get_connection() as conn:
            query = sqlalchemy.select(
                db.user_group.c.name.label('user_group_name')
            ).join(
                db.user_group,
                db.user_group_namespace_permission.c.user_group_id==db.user_group.c.id
            ).where(
                db.user_group.c.id.in_(user_group_ids),
                db.user_group_namespace_permission.c.namespace_id==namespace.pk
            )
            res = conn.execute(query)
            permissions = [
                cls(user_group=user_group_mapping[row['user_group_name']], namespace=namespace)
                for row in res
            ]

            return permissions

    @classmethod
    def create(cls, user_group, namespace, permission_type):
        """Create user group namespace permission"""
        # Check if permission already exists
        if cls.get_permissions_by_user_group_and_namespace(
                user_group=user_group,
                namespace=namespace):
            return None

        cls._insert_into_database(
            user_group=user_group,
            namespace=namespace,
            permission_type=permission_type)

        obj = cls(user_group=user_group, namespace=namespace)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_GROUP_NAMESPACE_PERMISSION_ADD,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def _insert_into_database(cls, user_group, namespace, permission_type):
        """Insert permission into database"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group_namespace_permission.insert().values(
                user_group_id=user_group.pk,
                namespace_id=namespace.pk,
                permission_type=permission_type
            ))

    @property
    def id(self):
        """Return identifiable name of object"""
        return '{user_group}/{namespace}'.format(
            user_group=self.user_group.name,
            namespace=self.namespace.name
        )

    @property
    def user_group(self):
        """Return user group."""
        return self._user_group

    @property
    def namespace(self):
        """Return namespace."""
        return self._namespace

    @property
    def permission_type(self):
        """Return permission."""
        return self._get_db_row()['permission_type']

    def __init__(self, user_group, namespace):
        """Store member variables."""
        self._user_group = user_group
        self._namespace = namespace
        self._row_cache = None

    def __eq__(self, __o):
        """Check if two user group namespace permissions are the same"""
        if isinstance(__o, self.__class__):
            return (
                self.namespace.pk == __o.namespace.pk and
                self.user_group.pk == __o.user_group.pk and
                self.permission_type == __o.permission_type
            )
        return super(UserGroupNamespacePermission, self).__eq__(__o)

    def _get_db_row(self):
        """Return DB row for user group."""
        if self._row_cache is None:
            db = Database.get()
            # Obtain row from user group table.
            select = sqlalchemy.select(
                db.user_group_namespace_permission
            ).where(
                db.user_group_namespace_permission.c.user_group_id==self._user_group.pk,
                db.user_group_namespace_permission.c.namespace_id==self._namespace.pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._row_cache = res.fetchone()
        return self._row_cache

    def delete(self):
        """Delete user group namespace permission."""
        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.USER_GROUP_NAMESPACE_PERMISSION_DELETE,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None, new_value=None
        )

        self._delete_from_database()

    def _delete_from_database(self):
        """Delete row from database"""
        db = Database.get()
        with db.get_connection() as conn:
            conn.execute(db.user_group_namespace_permission.delete().where(
                db.user_group_namespace_permission.c.user_group_id==self.user_group.pk,
                db.user_group_namespace_permission.c.namespace_id==self.namespace.pk
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

    def __eq__(self, __o):
        """Check if two git providers are the same"""
        if isinstance(__o, self.__class__):
            return self.pk == __o.pk
        return super(GitProvider, self).__eq__(__o)

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

    @classmethod
    def get(cls, name, create=False):
        """Create object and ensure the object exists."""
        obj = cls(name=name)

        # If there is no row, the module provider does not exist
        if obj._get_db_row() is None:

            # If set to create and auto module-provider creation
            # is enabled in config, create the module provider
            if create and terrareg.config.Config().AUTO_CREATE_NAMESPACE:
                return cls.create(name=name, display_name=None)

            # If not creating, return None
            return None

        # Otherwise, return object
        return obj

    @classmethod
    def get_by_display_name(cls, display_name):
        """Create object and ensure the object exists."""
        if not display_name:
            return None

        db = Database.get()
        display_name_query = sqlalchemy.select(
            db.namespace.c.namespace
        ).select_from(
            db.namespace
        ).where(
            # Use a like to use case-insentive
            # match for pre-existing namespaces
            db.namespace.c.display_name.like(display_name)
        )
        with db.get_connection() as conn:
            res = conn.execute(display_name_query).fetchone()
            if res:
                return cls.get(res.namespace)

        return None

    @classmethod
    def get_by_case_insensitive_name(cls, name):
        """Get namespace by case-insensitive name match."""
        db = Database.get()

        select = sqlalchemy.select(
            db.namespace
        ).select_from(
            db.namespace
        ).where(
            # Use a like to use case-insentive
            # match for pre-existing namespaces
            db.namespace.c.namespace.like(name)
        )
        with db.get_connection() as conn:
            res = conn.execute(select).fetchone()

            if res:
                return cls.get(res.namespace)
        return None

    @classmethod
    def insert_into_database(cls, name, display_name):
        """Insert new namespace into database"""
        db = Database.get()
        module_provider_insert = db.namespace.insert().values(
            namespace=name,
            display_name=display_name if display_name else None
        )
        with db.get_connection() as conn:
            conn.execute(module_provider_insert)

    @classmethod
    def create(cls, name, display_name=None):
        """Create instance of object in database."""
        # Validate name
        cls._validate_name(name)
        cls._validate_display_name(display_name)

        if cls.get_by_case_insensitive_name(name):
            raise NamespaceAlreadyExistsError("A namespace already exists with this name")

        if cls.get_by_display_name(display_name):
            raise DuplicateNamespaceDisplayNameError("A namespace already has this display name")

        # Create namespace
        cls.insert_into_database(name=name, display_name=display_name)

        obj = cls(name=name)

        # Indicate that object was created
        obj.created = True

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.NAMESPACE_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.name,
            old_value=None, new_value=None
        )
        return obj

    @staticmethod
    def get_total_count():
        """Get total number of namespaces."""
        db = Database.get()
        counts = sqlalchemy.select(
            db.namespace.c.namespace
        ).select_from(
            db.namespace
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
                db.module_version.c.beta == False,
                db.module_version.c.extraction_complete == True
            )

            modules_query = modules_query.subquery()

            namespace_query = sqlalchemy.select(
                db.namespace.c.namespace
            ).select_from(
                modules_query
            ).join(
                db.namespace,
                db.namespace.c.id==modules_query.c.namespace_id
            ).group_by(
                db.namespace.c.namespace
            ).order_by(
                db.namespace.c.namespace
            )
        else:
            namespace_query = sqlalchemy.select(
                db.namespace.c.namespace
            ).select_from(
                db.namespace
            ).group_by(
                db.namespace.c.namespace
            ).order_by(
                db.namespace.c.namespace
            )

        with db.get_connection() as conn:
            res = conn.execute(namespace_query)

            return [
                Namespace(name=r['namespace'])
                for r in res
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
        if (name is None or
                not re.match(r'^[0-9a-zA-Z][0-9a-zA-Z-_]+[0-9A-Za-z]$', name) or
                '__' in name):
            raise InvalidNamespaceNameError(
                'Namespace name is invalid - '
                'it can only contain alpha-numeric characters, '
                'hyphens and underscores, and must start/end with '
                'an alphanumeric character. '
                'Sequential underscores are not allowed.'
            )

    @staticmethod
    def _validate_display_name(display_name):
        """Determine if display name is valid"""
        if not display_name:
            return

        if not re.match(r'^[A-Za-z0-9][0-9A-Za-z\s\-_]+[A-Za-z0-9]$', display_name):
            raise InvalidNamespaceDisplayNameError('Namespace display name is invalid')

    @property
    def pk(self):
        """Return database ID of namespace."""
        db_row = self._get_db_row()
        if not db_row:
            return None
        return db_row['id']

    @property
    def display_name(self):
        """Return display name for namespace"""
        return self._get_db_row()["display_name"]

    def __init__(self, name: str):
        """Validate name and store member variables"""
        self._name = name
        self.created = False
        self._cache_db_row = None

    def _get_db_row(self):
        """Return database row for namespace."""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.namespace.select(
            ).where(
                db.namespace.c.namespace == self._name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row

    def delete(self):
        """Delete namespace"""
        assert self.pk is not None

        # Delete all module providers
        for module in self.get_all_modules():
            for provider in module.get_providers():
                provider.delete()

        # Delete namespace from database
        db = Database.get()

        with db.get_connection() as conn:
            # Delete module details from module_details table
            delete_statement = db.namespace.delete().where(
                db.namespace.c.id == self.pk
            )
            conn.execute(delete_statement)

        # Clear cached DB row
        self._cache_db_row = None

    def get_view_url(self):
        """Return view URL"""
        return '/modules/{namespace}'.format(namespace=self.name)

    def get_details(self):
        """Return custom terrareg details about namespace."""
        return {
            'is_auto_verified': self.is_auto_verified,
            'trusted': self.trusted,
            'display_name': self.display_name
        }

    def get_all_modules(self):
        """Return all modules for namespace."""
        db = Database.get()
        select = sqlalchemy.select(
            db.module_provider.c.module
        ).select_from(db.module_provider).join(
            db.namespace, db.module_provider.c.namespace_id==db.namespace.c.id
        ).where(
            db.namespace.c.namespace == self.name
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

    def get_module_custom_links(self):
        """Obtain module links that are applicable to namespace"""
        links = filter(
            lambda x: x.get('namespaces', None) is None or self.name in x.get('namespaces', []),
            json.loads(terrareg.config.Config().MODULE_LINKS))
        return links


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
        ).join(
            db.namespace,
            db.module_provider.c.namespace_id==db.namespace.c.id
        ).where(
            db.namespace.c.namespace == self._namespace.name,
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

    @property
    def infracost(self):
        """Return infracost data."""
        db_row = self._get_db_row()
        if (db_row is not None and
                db_row['infracost']):
            return json.loads(db_row['infracost'])
        return {}

    @property
    def terraform_graph(self):
        """Return decoded terraform graph data."""
        db_row = self._get_db_row()
        if db_row and db_row["terraform_graph"]:
            return Database.decode_blob(db_row["terraform_graph"])
        return None

    def get_graph_json(self, full_resource_names=False, full_module_names=False):
        """Return graph JSON for resources."""
        terraform_graph = self.terraform_graph
        if not terraform_graph:
            return None

        # Generate NX graph from terraform graphviz output
        graph = pygraphviz.AGraph(terraform_graph)
        nx_graph = nx.nx_agraph.from_agraph(graph)

        infracost = self.infracost
        resource_costs = {}
        remove_item_iteration_re = re.compile(r'\[[^\]]+\]')
        if infracost:
            for resource in self.infracost["projects"][0]["breakdown"]["resources"]:
                if not resource["monthlyCost"]:
                    continue

                name = remove_item_iteration_re.sub("", resource["name"])
                if name not in resource_costs:
                    resource_costs[name] = 0
                resource_costs[name] += round((float(resource["monthlyCost"]) * 12), 2)

        module_var_output_local_re = re.compile(r'^(module\.[^\.]+\.)+(var|local|output)\.[^\.]+$')
        # Capture modules resources, such as:
        # module.module1
        # module.module1.module.module2
        module_re = re.compile(r'^(?:module\.[^\.]+\.)*(?:module\.([^\.]+))$')
        # Capture data resources, such as:
        # data.aws_s3_bucket.test
        # module.module1.data.aws_s3_bucket.test
        # module.module1.module.module2.data.aws_s3_bucket.test
        data_re = re.compile(r'^((?:module\.[^\.]+\.)+)data\.([^\.]+)\.([^\.])+$')
        # Capture resources, such as:
        # aws_s3_bucket.test
        # module.module1.aws_s3_bucket.test
        # module.module1.module.module2.aws_s3_bucket.test
        resource_re = re.compile(r'^((?:module\.[^\.]+\.)*)([^\.]+)\.([^\.]+)$')

        # Store node renames, to be renamed after initial iteration
        renames = {}
        # Store nodes to be removed
        to_remove = []
        # Store labels to be pushed to graph JSON
        labels = {}
        # Stoe type mappings for determing node attributes
        type_mapping = {}
        # Store parents of attirbutes to modules, used for
        # parent mapping in JSON
        parents = {}

        def remove_node(node):
            """Add a node to the remove_nodes list, if they are not already present"""
            if node not in to_remove:
                to_remove.append(node)

        for node_label in nx_graph.nodes:
            # Remove leading '[root] ' name and expand/close suffices from node names
            name = node_label.replace('[root] ', '').replace(' (expand)', '').replace(' (close)', '')

            # Check for root vars, outputs and locals
            if name.startswith('output.') or name.startswith('var.') or name.startswith('local.'):
                remove_node(node_label)

            # Remove any module vars/outputs/locals
            elif module_var_output_local_re.match(name):
                remove_node(node_label)

            # handle all other nodes
            else:
                # Rename to shortened name
                renames[node_label] = name

                # Match node name to type regexes
                module_match = module_re.match(name)
                resource_match = resource_re.match(name)
                data_match = data_re.match(name)

                # Create labels and type mapping
                if name == "root":
                    # Match root module
                    labels[name] = "Root Module"
                    type_mapping[name] = "module"

                # Match submodules
                elif module_match:
                    if full_module_names:
                        labels[name] = name
                    else:
                        labels[name] = module_match.group(1)

                    type_mapping[name] = "module"

                elif data_match:
                    type_mapping[name] = "data"
                    parents[name] = data_match.group(1).strip(".") or "root"

                    if full_resource_names:
                        labels[name] = name
                    else:
                        labels[name] = f"(data) {data_match.group(2)}.{data_match.group(3)}"

                # Ensure resource RE is performed last,
                # as this could also match module_re
                elif resource_match:
                    type_mapping[name] = "resource"
                    if full_resource_names:
                        labels[name] = name
                    else:
                        labels[name] = f"{resource_match.group(2)}.{resource_match.group(3)}"

                    # Add cost to label, if available
                    if name in resource_costs:
                        labels[name] += f" (${resource_costs[name]}/year)"
                    parents[name] = resource_match.group(1).strip(".") or "root"

                # Discard any unrecognised types
                else:
                    remove_node(name)
                    print("Unable to match node to type", name)

        # Perform rename of nodes
        nx_graph = nx.relabel_nodes(nx_graph, renames)

        # Remove any nodes marked for removal
        for node in to_remove:
            nx_graph.remove_node(node)

        # Convert to JSON for cytoscape
        cytoscape_json = {
            "nodes": [],
            "edges": []
        }

        for node in nx_graph.nodes:
            data = {
                "id": node,
                "label": labels.get(node),
                "child_count": list(parents.values()).count(node)
            }

            style = {}
            if type_mapping[node] == "module":
                style = {
                    'color': '#000000',
                    'background-color': '#F8F7F9',
                    'font-weight': 'bold',
                    'text-valign': 'top',
                }
            # Add red outline to resources that have an associated cost
            if node in resource_costs:
                style['border-style'] = 'solid'
                style['border-width'] = '2px'
                style['border-color'] = 'red'

            # Add parent if available
            parent = parents.get(node, None)
            if parent:
                data["parent"] = parent

            cytoscape_json["nodes"].append({
                "data": data,
                "style": style
            })

        # Add edges to graph
        seen_module_links = []
        for edge in nx_graph.edges:
            # Only add edges for module-module links
            if (type_mapping[edge[0]] == "module" and type_mapping[edge[1]] == "module" and
                    # Only link modules in one direction, where module is a sub-module of another,
                    # to avoid links in both directions
                    edge[0] in edge[1]):
                # Mark module as having been seen in edges
                seen_module_links.append(edge[1])

                cytoscape_json["edges"].append({
                    "data": {
                        "id": f"{edge[0]}.{edge[1]}",
                        "source": edge[0],
                        "target": edge[1]
                    },
                    "classes": [
                        f"{type_mapping[edge[0]]}-{type_mapping[edge[1]]}"
                    ]
                })

        # Iterate through all modukes...
        for module, type_mapping in type_mapping.items():
            if type_mapping == "module":
                # If a module link has not already been seen,
                # add a link to root module
                if module not in seen_module_links and module != "root":
                    cytoscape_json["edges"].append({
                        "data": {
                            "id": f"root.{module}",
                            "source": module,
                            "target": "root"
                        },
                        "classes": [
                            f"{module}-root"
                        ]
                    })

        return cytoscape_json

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
            if kwarg in ['readme_content', 'terraform_docs', 'tfsec', 'infracost', 'terraform_graph']:
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
        },
        'consul': {
            'source': '/static/images/consul.png',
            'tos': 'All \'Consul\' modules are designed to work with HashiCorp Consul. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Consul and the HashiCorp Consul logo are trademarks of HashiCorp.',
            'alt': 'Hashicorp Consul',
            'link': '#'
        },
        'nomad': {
            'source': '/static/images/nomad.png',
            'tos': 'All \'Nomad\' modules are designed to work with HashiCorp Nomad. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Nomad and the HashiCorp Nomad logo are trademarks of HashiCorp.',
            'alt': 'Hashicorp Nomad',
            'link': '#'
        },
        'vagrant': {
            'source': '/static/images/vagrant.png',
            'tos': 'All \'Vagrant\' modules are designed to work with HashiCorp Vagrant. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vagrant and the HashiCorp Vagrant logo are trademarks of HashiCorp.',
            'alt': 'Hashicorp Vagrant',
            'link': '#'
        },
        'vault': {
            'source': '/static/images/vault.png',
            'tos': 'All \'Vault\' modules are designed to work with HashiCorp Vault. Terrareg and modules hosted within it are in no way affiliated with, nor endorsed by, HashiCorp. HashiCorp, HashiCorp Vault and the HashiCorp Vault logo are trademarks of HashiCorp.',
            'alt': 'Hashicorp Vault',
            'link': '#'
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
    def get_total_count(only_published=False):
        """Get total number of module providers."""
        db = Database.get()
        counts = sqlalchemy.select(
            db.namespace.c.namespace,
            db.module_provider.c.module,
            db.module_provider.c.provider
        ).select_from(
            db.module_version
        ).join(
            db.module_provider,
            db.module_version.c.module_provider_id==db.module_provider.c.id
        ).join(
            db.namespace,
            db.module_provider.c.namespace_id==db.namespace.c.id
        ).where(
            db.module_version.c.extraction_complete==True
        )
        if only_published:
            counts = counts.where(
                db.module_version.c.published == True,
                db.module_version.c.beta == False
            )

        counts = counts.group_by(
            db.namespace.c.namespace,
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
            namespace_id=module._namespace.pk,
            module=module.name,
            provider=name,
            verified=module._namespace.is_auto_verified
        )
        with db.get_connection() as conn:
            conn.execute(module_provider_insert)

        obj = cls(module=module, name=name)
        obj.created = True

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def get(cls, module, name, create=False):
        """Create object and ensure the object exists."""
        obj = cls(module=module, name=name)

        # If there is no row, the module provider does not exist
        if obj._get_db_row() is None:

            # If set to create and auto module-provider creation
            # is enabled in config, create the module provider
            if create and terrareg.config.Config().AUTO_CREATE_MODULE_PROVIDER:
                return cls.create(module=module, name=name)

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
    def module(self):
        """Return module object of provider"""
        return self._module

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

    @property
    def git_path(self):
        """Return path of module within git"""
        row_value = self._get_db_row()['git_path']
        # Strip leading slash or dot-slash
        if row_value:
            # Use safe_join_path twice:
            # - check it doesn't traverse back any paths
            # - remove any relative paths and return absolute path against root
            #   and then trim the leading slash
            safe_join_paths('/test_dir', row_value, allow_same_directory=True)
            row_value = safe_join_paths('/', row_value, allow_same_directory=True)[1:]

        # If git path is empty or is a path for the root directory, return None
        if not row_value:
            return None

        # Otherwise, return the path
        return row_value

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
        self.created = False

    def get_db_where(self, db, statement):
        """Filter DB query by where for current object."""
        return statement.where(
            db.module_provider.c.id==self.pk
        )

    def _get_db_row(self):
        """Return database row for module provider."""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_provider.select(
            ).join(
                db.namespace,
                db.module_provider.c.namespace_id==db.namespace.c.id
            ).where(
                db.namespace.c.id == self._module._namespace.pk,
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

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_DELETE,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None, new_value=None
        )

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

    def update_verified(self, verified):
        """Update verified flag of module provider."""
        if verified in [True, False] and verified != self.verified:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_VERIFIED,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=self.verified,
                new_value=verified
            )
            self.update_attributes(
                verified=verified
            )

    def update_git_provider(self, git_provider: GitProvider):
        """Update git provider associated with module provider."""
        original_git_provider = self.get_git_provider()
        if original_git_provider != git_provider:

            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_PROVIDER,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=original_git_provider.name if original_git_provider else None,
                new_value=git_provider.name if git_provider else None
            )
        self.update_attributes(
            git_provider_id=(git_provider.pk if git_provider is not None else None)
        )

    def update_git_tag_format(self, git_tag_format):
        """Update git_tag_format."""
        sanitised_git_tag_format = urllib.parse.quote(git_tag_format, safe=r'/{}')

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

        if sanitised_git_tag_format != self.git_tag_format:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_TAG_FORMAT,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=self.git_tag_format,
                new_value=sanitised_git_tag_format
            )

        self.update_attributes(git_tag_format=sanitised_git_tag_format)

    def update_git_path(self, git_path):
        """Update git_path attribute"""
        # If git path is not empty or specifies the root directory,
        # check that it doesn't contain relative paths to escape the root directory
        if git_path and git_path != '/':
            # Sanity check path
            safe_join_paths('/somepath/somesubpath', git_path, allow_same_directory=True)

        original_value = self._get_db_row()['git_path']
        if original_value != git_path:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_PATH,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=original_value,
                new_value=git_path
            )
        self.update_attributes(git_path=git_path)

    def update_repo_clone_url_template(self, repo_clone_url_template):
        """Update repository URL for module provider."""
        if repo_clone_url_template:
            try:
                converted_template = repo_clone_url_template.format(
                    namespace=self._module._namespace.name,
                    module=self._module.name,
                    provider=self.name)
            except KeyError:
                # KeyError thrown when template value
                # contains a unknown template
                raise RepositoryUrlContainsInvalidTemplateError(
                    'URL contains invalid template value. '
                    'Only the following template values are allowed: {namespace}, {module}, {provider}'
                )

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'URL does not contain a scheme (e.g. ssh://)'
                )
            if url.scheme not in ['http', 'https', 'ssh']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'URL contains an unknown scheme (e.g. https/ssh/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryUrlDoesNotContainPathError(
                    'URL does not contain a path'
                )
            try:
                int(url.port)
            except ValueError:
                # Value error is thrown when port contains a value, but is
                # not convertable to an int
                raise RepositoryUrlContainsInvalidPortError(
                    'URL contains a invalid port. '
                    'Only use a colon to for specifying a port, otherwise a forward slash should be used.'
                )
            except TypeError:
                # TypeError is thrown when port is None when trying to convert to an int
                pass

            repo_clone_url_template = urllib.parse.quote(repo_clone_url_template, safe=r'\{\}/:@%?=')

        original_value = self._get_db_row()['repo_clone_url_template']
        if original_value != repo_clone_url_template:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_CLONE_URL,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=original_value,
                new_value=repo_clone_url_template
            )

        self.update_attributes(repo_clone_url_template=repo_clone_url_template)

    def update_repo_browse_url_template(self, repo_browse_url_template):
        """Update browse URL template for module provider."""
        if repo_browse_url_template:

            try:
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
            except KeyError:
                # KeyError thrown when template value
                # contains a unknown template
                raise RepositoryUrlContainsInvalidTemplateError(
                    'URL contains invalid template value. '
                    'Only the following template values are allowed: {namespace}, {module}, {provider}, {tag}, {path}'
                )

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'URL does not contain a scheme (e.g. https://)'
                )
            if url.scheme not in ['http', 'https']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'URL contains an unknown scheme (e.g. https/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryUrlDoesNotContainPathError(
                    'URL does not contain a path'
                )
            try:
                int(url.port)
            except ValueError:
                # Value error is thrown when port contains a value, but is
                # not convertable to an int
                raise RepositoryUrlContainsInvalidPortError(
                    'URL contains a invalid port. '
                    'Only use a colon to for specifying a port, otherwise a forward slash should be used.'
                )
            except TypeError:
                # TypeError is thrown when port is None when trying to convert to an int
                pass

            repo_browse_url_template = urllib.parse.quote(repo_browse_url_template, safe=r'\{\}/:@%?=')

        original_value = self._get_db_row()['repo_browse_url_template']
        if original_value != repo_browse_url_template:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BROWSE_URL,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=original_value,
                new_value=repo_browse_url_template
            )

        self.update_attributes(repo_browse_url_template=repo_browse_url_template)

    def update_repo_base_url_template(self, repo_base_url_template):
        """Update browse URL template for module provider."""
        if repo_base_url_template:

            try:
                converted_template = repo_base_url_template.format(
                    namespace=self._module._namespace.name,
                    module=self._module.name,
                    provider=self.name)
            except KeyError:
                # KeyError thrown when template value
                # contains a unknown template
                raise RepositoryUrlContainsInvalidTemplateError(
                    'URL contains invalid template value. '
                    'Only the following template values are allowed: {namespace}, {module}, {provider}'
                )

            url = urllib.parse.urlparse(converted_template)
            if not url.scheme:
                raise RepositoryUrlDoesNotContainValidSchemeError(
                    'URL does not contain a scheme (e.g. https://)'
                )
            if url.scheme not in ['http', 'https']:
                raise RepositoryUrlContainsInvalidSchemeError(
                    'URL contains an unknown scheme (e.g. https/http)'
                )
            if not url.hostname:
                raise RepositoryUrlDoesNotContainHostError(
                    'URL does not contain a host/domain'
                )
            if not url.path:
                raise RepositoryUrlDoesNotContainPathError(
                    'URL does not contain a path'
                )
            try:
                int(url.port)
            except ValueError:
                # Value error is thrown when port contains a value, but is
                # not convertable to an int
                raise RepositoryUrlContainsInvalidPortError(
                    'URL contains a invalid port. '
                    'Only use a colon to for specifying a port, otherwise a forward slash should be used.'
                )
            except TypeError:
                # TypeError is thrown when port is None when trying to convert to an int
                pass

            repo_base_url_template = urllib.parse.quote(repo_base_url_template, safe=r'\{\}/:@%?=')

        original_value = self._get_db_row()['repo_base_url_template']
        if original_value != repo_base_url_template:
            terrareg.audit.AuditEvent.create_audit_event(
                action=terrareg.audit_action.AuditAction.MODULE_PROVIDER_UPDATE_GIT_CUSTOM_BASE_URL,
                object_type=self.__class__.__name__,
                object_id=self.id,
                old_value=original_value,
                new_value=repo_base_url_template
            )

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
            db.module_provider.c.id == self.pk,
            db.module_version.c.published == True,
            db.module_version.c.beta == False,
            db.module_version.c.extraction_complete == True
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
            db.module_provider.c.id == self.pk,
            db.module_version.c.extraction_complete == True
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
            "git_path": self.git_path,
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


class TfsecResultStatus(Enum):
    """tfsec result status"""

    FAIL = 0
    PASS = 1
    SKIP = 2
    UNSPECIFIED = 'unspecified'


class TerraformSpecsObject(object):
    """Base Terraform object, that has terraform-docs available."""

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
        self._tfsec_results = None

    @property
    def module_version(self):
        """Return module version"""
        raise NotImplementedError

    @property
    def path(self):
        """Return module path"""
        raise NotImplementedError

    @property
    def git_path(self):
        """Return path of module within git"""
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

    def get_tfsec_failures(self):
        """Get tfsec failure results"""
        failures = []
        if self._tfsec_results is None:
            module_details = self.module_details
            if module_details is None:
                self._tfsec_results = None
            else:
                self._tfsec_results = module_details.tfsec.get("results", None)

        if self._tfsec_results is None:
            return None

        for result in self._tfsec_results:
            
            # Only return failed results
            if (TfsecResultStatus(
                        result.get('status', TfsecResultStatus.UNSPECIFIED.value)
                    ) == TfsecResultStatus.FAIL):
                failures.append(result)

        return failures

    def get_usage_example(self, request_domain):
        """Base method to create usage example terraform"""
        source_url, version = self.get_terraform_url_and_version_strings(request_domain=request_domain, module_path=self.path)
        terraform = f"""
module "{self.module_version.module_provider.module.name}" {{
{self.get_source_version_terraform(source_url, version)}

  # Provide variables here\n}}
""".strip()
        return terraform

    def get_source_version_terraform(self, source, version, leading_indentation="  ", trailing_indentation=None):
        """Return terraform"""
        calculated_extra_source_indentation = "  " if version else " "
        if trailing_indentation and len(trailing_indentation) >= len(calculated_extra_source_indentation):
            actual_trailing_indentation = trailing_indentation
        else:
            actual_trailing_indentation = calculated_extra_source_indentation

        terraform = f'{leading_indentation}source{actual_trailing_indentation}= "{source}"'
        if version:
            version_comments = self.module_version.get_terraform_example_version_comment()
            for version_comment in version_comments:
                terraform += f'\n{leading_indentation}# {version_comment}'
            terraform += f'\n{leading_indentation}version{actual_trailing_indentation[1:]}= "{version}"'
        return terraform

    def get_terraform_url_and_version_strings(self, request_domain, module_path):
        """Return terraform source URL and version values for given requested protoco, domain, port and module path"""
        protocol, domain, port = get_public_url_details(fallback_domain=request_domain)

        isHttps = protocol.lower() == "https"

        # Set default port if port is None or empty string, or port matches the default port for the protocol
        isDefaultPort = not port or (str(port) == "443" and isHttps) or (str(port) == "80" and not isHttps)

        # Add protocl for http
        source_url = '' if isHttps else 'http://'
        # Add domain name
        source_url += domain
        # Add port is non-default
        source_url += '' if isDefaultPort else f':{port}'
        # Add /modules URL path if over http
        source_url += '' if isHttps else '/modules'
        source_url += '/'
        # Add example analytics token
        source_url += (
            (terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN + '__')
            if terrareg.config.Config().EXAMPLE_ANALYTICS_TOKEN and (not terrareg.config.Config().DISABLE_ANALYTICS)
            else ''
        )
        # Add module provider ID
        source_url += self.module_version.module_provider.id
        # Add exact module version, if using http
        source_url += '' if isHttps else f'/{self.module_version.version}'
        # Remove any leading slashes from module_path
        module_path = re.sub(r'^\/+', '', module_path)
        # Add sub-module path, if it exists
        source_url += f'//{module_path}' if module_path else ''

        # Use module version example terraform version string, if HTTPS otherwise provide a None version string, as
        # the version is incorporated into the URL and http downloads don't support 'version' attribute
        version_string = self.module_version.get_terraform_example_version_string() if isHttps else None

        # Return source URL and version
        return source_url, version_string


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
        readme_md = self.get_readme_content(sanitise=False)
        if readme_md:
            readme_md = self.replace_source_in_file(
                readme_md, server_hostname)
            readme_html = convert_markdown_to_html(file_name='README.md', markdown_html=readme_md)
            return sanitise_html_content(readme_html, allow_markdown_html=True)
        return None

    @property
    def module_details(self):
        """Return instance of ModuleDetails for object."""
        if self._get_db_row() and self._get_db_row()['module_details_id']:
            return ModuleDetails(id=self._get_db_row()['module_details_id'])
        else:
            return None

    def get_readme_content(self, sanitise=True):
        """Get readme contents"""
        module_details = self.module_details
        if module_details:
            content = Database.decode_blob(module_details.readme_content)
            if sanitise:
                content = sanitise_html_content(content)
            return content
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
        # @TODO Verify what this should be - Terraform example is empty and real-world examples appears to
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

            providers.append({
                'name': sanitise_html_content(name),
                'namespace': sanitise_html_content(namespace),
                'source': '',  # This data is not available
                'version': sanitise_html_content(provider['version']) if provider['version'] else ''
            })
        return providers

    def get_terraform_version_constraints(self):
        """Obtain terraform version requirement"""
        for requirement in self.get_module_specs().get("requirements", []):
            if requirement["name"] == "terraform":
                return requirement["version"]
        return None

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
            "resources": self.get_terraform_resources()
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

            leading_space = match.group(1)
            trailing_space = match.group(2)

            source_url, version_string = self.get_terraform_url_and_version_strings(request_domain=server_hostname, module_path=module_path)

            return (
                '\n' +
                self.get_source_version_terraform(
                    source_url,
                    version_string,
                    leading_indentation=leading_space,
                    trailing_indentation=trailing_space
                ) +
                '\n'
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
        ).where(
            db.module_version.c.extraction_complete == True
        ).group_by(
            db.namespace.c.namespace,
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
        return sanitise_html_content(self._get_db_row()['owner'])

    @property
    def published(self):
        """Return whether module is published"""
        return bool(self._get_db_row()['published'])

    @property
    def description(self):
        """Return description."""
        return sanitise_html_content(self._get_db_row()['description'])

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
    def git_path(self):
        """Return path of module within git"""
        return self._module_provider.git_path

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

        # Set default values for each user-defined variable
        for variable in variables:
            variable['required'] = variable.get('required', True)
            variable['type'] = variable.get('type', 'text')
            variable['additional_help'] = variable.get('additional_help', '')
            variable['quote_value'] = variable.get('quote_value', True)
            variable['default_value'] = variable.get('default_value', None)

        # Detect bad type for variable template and replace
        # with empty array
        if type(variables) is not list:
            variables = []

        if terrareg.config.Config().AUTOGENERATE_USAGE_BUILDER_VARIABLES:
            for input_variable in self.get_terraform_inputs():
                if input_variable['name'] not in [v['name'] for v in variables]:
                    converted_type = 'text'
                    quote_value = True
                    default_value = input_variable['default']
                    if input_variable['type'] == 'bool':
                        converted_type = 'boolean'
                        quote_value = False
                    elif input_variable['type'].startswith('list('):
                        list_type = re.match(r'list\((.*)\)', input_variable['type'])
                        if list_type and list_type.group(1) == 'number':
                            quote_value = False
                        else:
                            quote_value = True
                        converted_type = 'list'
                    elif input_variable['type'] == 'number':
                        converted_type = 'number'
                        quote_value = False
                    elif input_variable['type'].startswith('map('):
                        converted_type = 'text'
                        quote_value = False
                    
                    variables.append({
                        'name': input_variable['name'],
                        'type': converted_type,
                        'additional_help': input_variable['description'],
                        'quote_value': quote_value,
                        'required': input_variable['required'],
                        'default_value': default_value
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

    @property
    def module_version_files(self):
        """Return list of module version files for module version"""
        db = Database.get()
        select = db.module_version_file.select().join(
            db.module_version, db.module_version_file.c.module_version_id==db.module_version.c.id
        ).where(
            db.module_version.c.id==self.pk
        )
        with db.get_connection() as conn:
            res = conn.execute(select)
            return res.fetchall()

    @property
    def graph_data_url(self):
        """Return URl for graph data"""
        return f"/v1/terrareg/modules/{self.id}/graph/data"

    @property
    def custom_links(self):
        """Return list of links to be displayed in UI"""
        links = []
        placeholders = {
            'namespace': self._module_provider._module._namespace.name,
            'module': self.module_provider._module.name,
            'provider': self.module_provider.name,
            'version': self._version
        }
        for link in self._module_provider._module._namespace.get_module_custom_links():
            links.append({
                'text': link.get('text', '').format(**placeholders),
                'url': link.get('url', '#').format(**placeholders)
            })
        return links

    @property
    def module_extraction_up_to_date(self):
        """Whether the extracted module version data is up-to-date"""
        return self._get_db_row()["extraction_version"] == EXTRACTION_VERSION

    @property
    def is_latest_version(self):
        """Return whether the version is the latest version for the module provider"""
        return self._module_provider.get_latest_version() == self

    def __init__(self, module_provider: ModuleProvider, version: str, pk: int=None):
        """Setup member variables."""
        self._extracted_beta_flag = self._validate_version(version)
        self._module_provider = module_provider
        self._version = version
        self._cache_db_row = None
        self._pk = pk
        super(ModuleVersion, self).__init__()

    def __eq__(self, __o):
        """Check if two module versions are the same"""
        if isinstance(__o, self.__class__):
            return self.pk == __o.pk
        return super(ModuleVersion, self).__eq__(__o)

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_version.select().join(
                db.module_provider, db.module_version.c.module_provider_id == db.module_provider.c.id
            ).where(
                db.module_provider.c.id == self._module_provider.pk,
                db.module_version.c.version == self.version
            )
            # If a specific primary key has been specified,
            # select based on this
            if self._pk:
                select = select.where(
                    db.module_version.c.id == self._pk
                )
            # Otherwise, obtain based on whether extraction is complete
            else:
                select = select.where(
                    db.module_version.c.extraction_complete == True
                )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def get_terraform_example_version_string(self):
        """Return formatted string of version parameter for example Terraform."""
        # For beta versions, pass an exact version constraint.
        if self.beta or not self.is_latest_version:
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

    def get_terraform_example_version_comment(self):
        """Get comment displayed above version string in Terraform, used for warning about specific versions."""
        if not self.published:
            return ["This version of this module has not yet been published,", "meaning that it cannot yet be used by Terraform"]
        elif self.beta:
            return ["This version of the module is a beta version.", "To use this version, it must be pinned in Terraform"]
        elif not self.is_latest_version:
            return ["This version of the module is not the latest version.", "To use this specific version, it must be pinned in Terraform"]
        return []

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
            # by Terraform to acknowledge a git repository
            # and add if it not
            parsed_url = urllib.parse.urlparse(rendered_url)
            if not parsed_url.scheme.startswith('git::'):
                rendered_url = 'git::{rendered_url}'.format(rendered_url=rendered_url)

            # Check if git_path has been set and prepend to path, if set.
            path = os.path.join(self.git_path or '', path or '')

            # Remove any trailing slashses from path
            if path and path.endswith('/'):
                path = path[:-1]

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

        # Check if git_path has been set and prepend to path, if set.
        path = os.path.join(self.git_path or '', path or '')

        # Remove any trailing slashses from path
        if path and path.endswith('/'):
            path = path[:-1]

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
                path=path
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
        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.MODULE_VERSION_PUBLISH,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None,
            new_value=None
        )

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

    def get_terrareg_api_details(self, request_domain):
        """Return dict of version details with additional attributes used by terrareg UI."""
        api_details = self._module_provider.get_terrareg_api_details()

        # Capture versions from module provider API output, as this limits
        # some versions, which are normally displayed in the Terraform APIs
        versions = api_details['versions']

        api_details.update(self.get_api_details())

        tab_files = [module_version_file.path for module_version_file in self.module_version_files]
        additional_module_tabs = json.loads(terrareg.config.Config().ADDITIONAL_MODULE_TABS)
        tab_file_mapping = {}
        for tab_config in additional_module_tabs:
            for file in tab_config[1]:
                if file in tab_files:
                    tab_file_mapping[tab_config[0]] = file

        source_browse_url = self.get_source_browse_url()
        tfsec_failures = self.get_tfsec_failures()
        api_details.update({
            "published_at_display": self.publish_date_display,
            "display_source_url": source_browse_url if source_browse_url else self.get_source_base_url(),
            "terraform_example_version_string": self.get_terraform_example_version_string(),
            "terraform_example_version_comment": self.get_terraform_example_version_comment(),
            "versions": versions,
            "beta": self.beta,
            "published": self.published,
            "security_failures": len(tfsec_failures) if tfsec_failures is not None else 0,
            "security_results": tfsec_failures,
            "additional_tab_files": tab_file_mapping,
            "custom_links": self.custom_links,
            "graph_url": f"/modules/{self.id}/graph",
            "terraform_version_constraint": self.get_terraform_version_constraints(),
            "module_extraction_up_to_date": self.module_extraction_up_to_date,
            "usage_example": self.get_usage_example(request_domain)
        })
        return api_details

    def prepare_module(self):
        """
        Handle file upload of module version.

        Returns boolean whethe previous DB row (if exists) was published.
        """
        self.create_data_directory()
        self._create_db_row()

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.MODULE_VERSION_INDEX,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None,
            new_value=None
        )

    def finalise_module(self):
        """Finalise module, after import. Delete old versions and activate the new version."""
        previous_version_published = False
        # Get previous instances of module
        db = Database.get()
        with db.start_transaction():
            with db.get_connection() as conn:
                res = conn.execute(sqlalchemy.select(
                    db.module_version
                ).where(
                    db.module_version.c.module_provider_id==self.module_provider.pk,
                    db.module_version.c.version==self.version,
                    db.module_version.c.id!=self.pk
                )).fetchall()
                res = [r for r in res]

            if res:
                for row in res:
                    old_module_version = ModuleVersion.get(
                        module_provider=self.module_provider,
                        version=self.version,
                        pk=row['id']
                    )

                    # If configured to auto re-publish module versions, use
                    # the current published state of previous module version
                    # to determine if publishing should happen
                    if terrareg.config.Config().MODULE_VERSION_REINDEX_MODE is terrareg.config.ModuleVersionReindexMode.AUTO_PUBLISH:
                        previous_version_published = previous_version_published or (row['extraction_complete'] and old_module_version.published)

                    # Delete previous module version
                    old_module_version.delete(delete_related_analytics=False)

                    # Migrate analytics from old module version ID to new module version
                    terrareg.analytics.AnalyticsEngine.migrate_analytics_to_new_module_version(
                        old_version_version_pk=row['id'],
                        new_module_version=self)

            self.update_attributes(extraction_complete=True)

            # Publish new module version, if previous module version
            # was found to be.
            if previous_version_published:
                self.publish()

    def get_db_where(self, db, statement):
        """Filter DB query by where for current object."""
        return statement.where(
            db.module_version.c.id==self.pk
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

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.MODULE_VERSION_DELETE,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None,
            new_value=None
        )

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
        """
        Insert into datadabase, removing any existing duplicate versions.

        Returns boolean whethe previous DB row (if exists) was published.
        """
        db = Database.get()

        # Delete pre-existing version, if it exists
        previous_version_published = False
        if self._get_db_row():
            # Determine if re-indexing of modules is allowed
            if terrareg.config.Config().MODULE_VERSION_REINDEX_MODE is terrareg.config.ModuleVersionReindexMode.PROHIBIT:
                raise ReindexingExistingModuleVersionsIsProhibitedError(
                    "The module version already exists and re-indexing modules is disabled")

        with db.get_connection() as conn:
            # Insert new module into table
            insert_statement = db.module_version.insert().values(
                module_provider_id=self._module_provider.pk,
                version=self.version,
                published=False,
                beta=self._extracted_beta_flag,
                internal=False,
                extraction_complete=False
            )
            insert_res = conn.execute(insert_statement)

            # Set current module pk directly from ID and clear cache
            self._pk = insert_res.lastrowid
            self._cache_db_row = None

        return previous_version_published

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

    @property
    def graph_data_url(self):
        """Return URl for graph data"""
        return f"/v1/terrareg/modules/{self.module_version.id}/graph/data/{self.TYPE}/{self.path}"

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
    def git_path(self):
        """Return path of module within git"""
        # Join git path for root module to path of submodule
        root_module_path = self.module_version.git_path
        if root_module_path:
            return safe_join_paths('/', root_module_path, self.path)[1:]
        else:
            return self.path

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

    def get_terrareg_api_details(self, request_domain):
        """Return dict of submodule details with additional attributes used by terrareg UI."""
        api_details = self.get_api_module_specs()
        source_browse_url = self.get_source_browse_url()
        tfsec_failures = self.get_tfsec_failures()
        terraform_version_constraint = self.get_terraform_version_constraints()
        api_details.update({
            "display_source_url": source_browse_url if source_browse_url else self._module_version.get_source_base_url(),
            "security_failures": len(tfsec_failures) if tfsec_failures is not None else 0,
            "security_results": tfsec_failures,
            "graph_url": f"/modules/{self.module_version.id}/graph/{self.TYPE}/{self.path}",
            "usage_example": self.get_usage_example(request_domain)
        })
        # Only update terraform version constraint if one is defined in the example,
        # otherwise default to root module's constraint
        if terraform_version_constraint:
            api_details["terraform_version_constraint"] = terraform_version_constraint
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

    def get_terrareg_api_details(self, *args, **kwargs):
        api_details = super(Example, self).get_terrareg_api_details(*args, **kwargs)
        yearly_cost = self.module_details.infracost.get('totalMonthlyCost', None)
        if yearly_cost:
            yearly_cost = "{:.2f}".format(round((float(yearly_cost) * 12), 2))
        api_details['cost_analysis'] = {
            'yearly_cost': yearly_cost
        }
        return api_details


class FileObject:
    """Base file object for example/module file in DB"""

    @staticmethod
    def get_db_table():
        """Return DB table for class"""
        raise NotImplementedError

    @property
    def file_name(self):
        """Return name of file"""
        return self._path.split('/')[-1]

    @property
    def path(self):
        """Return path of example file."""
        return self._path

    def _get_db_row(self):
        """Method to obtain row from database"""
        raise NotImplementedError

    @property
    def pk(self):
        """Get ID from DB row"""
        return self._get_db_row()['id']

    def get_content(self, sanitise=True):
        """Return content of example file."""
        content = Database.decode_blob(self._get_db_row()["content"])
        if content and sanitise:
            # Add pre tags before/after to allow for broken tags
            # inside content, e.g. for heredocs
            content = f"<pre>{content}</pre>"

            # Sanitise content
            content = sanitise_html_content(content)

            # Remove encoded 'pre' tags -
            # "&lt;pre&gt;" at start and "&lt;/pre&gt;" after
            content = content[11:][:-12]

        return content

    def __init__(self, path: str):
        """Store identifying data."""
        self._path = path
        self._cache_db_row = None

    def update_attributes(self, **kwargs):
        """Update DB row."""
        # Encode columns that are binary blobs in the database
        for kwarg in kwargs:
            if kwarg in ['content']:
                kwargs[kwarg] = Database.encode_blob(kwargs[kwarg])

        db = Database.get()
        update = self.get_db_table().update().where(
            self.get_db_table().c.id == self.pk
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


class ExampleFile(FileObject):

    @staticmethod
    def get_db_table():
        """Return DB table for class"""
        db = Database.get()
        return db.example_file

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

    def __init__(self, example: Example, path: str):
        """Store identifying data."""
        self._example = example
        super(ExampleFile, self).__init__(path)

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

    def get_content(self, server_hostname):
        """Return content with source replaced"""
        # Replace source lines that use relative paths
        return self._example.replace_source_in_file(
            content=super(ExampleFile, self).get_content(),
            server_hostname=server_hostname)


class ModuleVersionFile(FileObject):
    """File associated with module version"""

    @classmethod
    def get(cls, module_version: ModuleVersion, path: str):
        """Obtain instance of object, if it exists in the database"""
        module_version_file = cls(module_version=module_version, path=path)
        if module_version_file._get_db_row() is None:
            return None
        return module_version_file

    @staticmethod
    def get_db_table():
        """Return DB table for class"""
        db = Database.get()
        return db.module_version_file

    @classmethod
    def create(cls, module_version: ModuleVersion, path: str):
        """Create instance of object in database."""
        # Insert module file into database
        db = Database.get()
        insert_statement = db.module_version_file.insert().values(
            module_version_id=module_version.pk,
            path=path
        )
        with db.get_connection() as conn:
            conn.execute(insert_statement)

        # Return instance of object
        return cls(module_version=module_version, path=path)

    def __init__(self, module_version: ModuleVersion, path: str):
        """Store identifying data."""
        self._module_version = module_version
        super(ModuleVersionFile, self).__init__(path)

    def _get_db_row(self):
        """Return DB row for git provider."""
        if self._cache_db_row is None:
            db = Database.get()
            select = db.module_version_file.select().where(
                db.module_version_file.c.module_version_id == self._module_version.pk,
                db.module_version_file.c.path == self._path
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                return res.fetchone()
        return self._cache_db_row

    def get_content(self):
        """Return content to be displayed in UI"""
        # Convert markdown files to HTML
        if self.path.lower().endswith('.md'):
            # Perform sanitisation of markdown after
            # conversion to HTML
            content = super(ModuleVersionFile, self).get_content(sanitise=False)
            content = convert_markdown_to_html(file_name=self.file_name, markdown_html=content)
            # return content
            content = sanitise_html_content(content, allow_markdown_html=True)
        else:
            content = super(ModuleVersionFile, self).get_content()
            content = '<pre>' + content + '</pre>'
        return content


class ModuleVersionExtractionStatus:

    @classmethod
    def get_by_request_id(cls, module_provider, request_id):
        """Get by request id"""
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    db.module_extraction_status
                ).join(
                    db.module_version,
                    db.module_extraction_status.c.module_version_id==db.module_version.c.id
                ).join(
                    db.module_provider,
                    db.module_version.c.module_provider_id==db.module_provider.c.id
                ).where(
                    db.module_extraction_status.c.request_id==request_id,
                    db.module_provider.c.id==module_provider.pk
                )).fetchone()
            if not res:
                return None
            return cls(res)

    def __init__(self, row):
        """Store member variables"""
        self._row = row

    @property
    def json(self):
        """Return JSON value for status"""
        return {
            # 'timestamp': self._row['timestamp'],
            # 'last_update': self._row['last_update'],
            'status': self._row['status'].value,
            'message': self._row['message']
        }
