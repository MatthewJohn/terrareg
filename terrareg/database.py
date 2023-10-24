"""Provide database class."""

import sqlalchemy
import sqlalchemy.dialects.mysql

from flask import has_request_context
import flask
from terrareg.audit_action import AuditAction

import terrareg.config
from terrareg.errors import DatabaseMustBeIniistalisedError
from terrareg.provider_tier import ProviderTier
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.namespace_type import NamespaceType
from terrareg.provider_source_type import ProviderSourceType
import terrareg.provider_documentation_type
import terrareg.provider_binary_types


class Database():
    """Handle database connection and setting up database schema"""

    _META = None
    _ENGINE = None
    _INSTANCE = None

    blob_encoding_format = 'utf-8'
    MEDIUM_BLOB_SIZE = ((2 ** 24) - 1)

    @staticmethod
    def encode_blob(value):
        """Encode string as a blog value"""
        # Convert any untruthful values to empty string
        if not value:
            value = ''
        return value.encode(Database.blob_encoding_format)

    @staticmethod
    def decode_blob(value):
        """Decode blob as a string."""
        if value is None:
            return None
        return value.decode(Database.blob_encoding_format)

    @staticmethod
    def medium_blob():
        """Return column type for medium blob."""
        return sqlalchemy.LargeBinary(
                length=Database.MEDIUM_BLOB_SIZE).with_variant(
                    sqlalchemy.dialects.mysql.MEDIUMBLOB(), "mysql")

    def __init__(self):
        """Setup member variables."""
        self._session = None
        self._terraform_idp_authorization_code = None
        self._terraform_idp_access_token = None
        self._terraform_idp_subject_identifier = None
        self._user_group = None
        self._user_group_namespace_permission = None
        self._git_provider = None
        self._namespace_redirect = None
        self._namespace = None
        self._module_provider_redirect = None
        self._module_provider = None
        self._module_details = None
        self._module_version = None
        self._sub_module = None
        self._gpg_key = None
        self._provider_category = None
        self._provider_source = None
        self._repository = None
        self._provider = None
        self._provider_version = None
        self._provider_version_documentation = None
        self._provider_version_binary = None
        self._analytics = None
        self._example_file = None
        self._module_version_file = None
        self.transaction_connection = None

    @property
    def session(self):
        """Return session table"""
        if self._session is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._session

    @property
    def terraform_idp_authorization_code(self):
        """Return terraform_idp_authorization_code table"""
        if self._terraform_idp_authorization_code is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._terraform_idp_authorization_code

    @property
    def terraform_idp_access_token(self):
        """Return terraform_idp_access_token table"""
        if self._terraform_idp_access_token is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._terraform_idp_access_token

    @property
    def terraform_idp_subject_identifier(self):
        """Return terraform_idp_subject_identifier table"""
        if self._terraform_idp_subject_identifier is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._terraform_idp_subject_identifier

    @property
    def user_group(self):
        """Return user_group table."""
        if self._user_group is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._user_group

    @property
    def user_group_namespace_permission(self):
        """Return user_group_namespace_permission table."""
        if self._user_group_namespace_permission is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._user_group_namespace_permission

    @property
    def git_provider(self):
        """Return git_provider table."""
        if self._git_provider is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._git_provider

    @property
    def namespace_redirect(self):
        """Return namespace_redirect redirect table."""
        if self._namespace_redirect is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._namespace_redirect

    @property
    def namespace(self):
        """Return namespace table."""
        if self._namespace is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._namespace

    @property
    def module_provider_redirect(self):
        """Return module provider redirect table."""
        if self._module_provider_redirect is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_provider_redirect

    @property
    def module_provider(self):
        """Return module_provider table."""
        if self._module_provider is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_provider

    @property
    def module_details(self):
        """Return module_details table."""
        if self._module_details is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_details

    @property
    def module_version(self):
        """Return module_version table."""
        if self._module_version is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_version

    @property
    def sub_module(self):
        """Return submodule table."""
        if self._sub_module is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._sub_module

    @property
    def analytics(self):
        """Return analytics table."""
        if self._analytics is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._analytics

    @property
    def example_file(self):
        """Return analytics table."""
        if self._example_file is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._example_file

    @property
    def module_version_file(self):
        """Return analytics table."""
        if self._module_version_file is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_version_file

    @property
    def gpg_key(self):
        """Return gpg_key table."""
        if self._gpg_key is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._gpg_key

    @property
    def provider_category(self):
        """Return provider_category table."""
        if self._provider_category is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider_category

    @property
    def provider_source(self):
        """Return provider_source table."""
        if self._provider_source is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider_source

    @property
    def repository(self):
        """Return provider_source table."""
        if self._repository is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._repository

    @property
    def provider(self):
        """Return provider table."""
        if self._provider is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider

    @property
    def provider_version(self):
        """Return provider_version table."""
        if self._provider_version is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider_version

    @property
    def provider_version_documentation(self):
        """Return provider_version_documentation table."""
        if self._provider_version_documentation is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider_version_documentation

    @property
    def provider_version_binary(self):
        """Return provider_version_binary table."""
        if self._provider_version_binary is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._provider_version_binary

    @property
    def audit_history(self):
        """Audit history table."""
        if self._audit_history is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._audit_history

    @classmethod
    def reset(cls):
        """Reset database connections."""
        cls._INSTANCE = None
        cls._META = None
        cls._ENGINE = None

    @classmethod
    def get(cls):
        """Get singleton instance of class."""
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
        """Get singleton instance of engine."""
        if cls._ENGINE is None:
            cls._ENGINE = sqlalchemy.create_engine(
                terrareg.config.Config().DATABASE_URL,
                echo=terrareg.config.Config().DEBUG,
                pool_pre_ping=True,
                pool_recycle=300
            )
        return cls._ENGINE

    def initialise(self):
        """Initialise database schema."""
        meta = self.get_meta()
        engine = self.get_engine()

        GENERAL_COLUMN_SIZE = 128
        LARGE_COLUMN_SIZE = 1024
        URL_COLUMN_SIZE = 1024

        self._session = sqlalchemy.Table(
            'session', meta,
            sqlalchemy.Column('id', sqlalchemy.String(128), primary_key=True),
            sqlalchemy.Column('expiry', sqlalchemy.DateTime, nullable=False),
            sqlalchemy.Column('provider_source_auth', Database.medium_blob())
        )

        self._terraform_idp_authorization_code = sqlalchemy.Table(
            'terraform_idp_authorization_code', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('key', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False, unique=True),
            sqlalchemy.Column('data', Database.medium_blob()),
            sqlalchemy.Column('expiry', sqlalchemy.DateTime, nullable=False)
        )

        self._terraform_idp_access_token = sqlalchemy.Table(
            'terraform_idp_access_token', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('key', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False, unique=True),
            sqlalchemy.Column('data', Database.medium_blob()),
            sqlalchemy.Column('expiry', sqlalchemy.DateTime, nullable=False)
        )

        self._terraform_idp_subject_identifier = sqlalchemy.Table(
            'terraform_idp_subject_identifier', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('key', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False, unique=True),
            sqlalchemy.Column('data', Database.medium_blob()),
            sqlalchemy.Column('expiry', sqlalchemy.DateTime, nullable=False)
        )

        self._user_group = sqlalchemy.Table(
            'user_group', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False, unique=True),
            sqlalchemy.Column('site_admin', sqlalchemy.Boolean, default=False, nullable=False)
        )

        self._user_group_namespace_permission = sqlalchemy.Table(
            'user_group_namespace_permission', meta,
            sqlalchemy.Column(
                'user_group_id',
                sqlalchemy.ForeignKey(
                    'user_group.id',
                    name='fk_user_group_namespace_permission_user_group_id_user_group_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False,
                primary_key=True
            ),
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_user_group_namespace_permission_namespace_id_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False,
                primary_key=True
            ),
            sqlalchemy.Column('permission_type', sqlalchemy.Enum(UserGroupNamespacePermissionType))
        )

        self._git_provider = sqlalchemy.Table(
            'git_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), unique=True),
            sqlalchemy.Column('base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE))
        )

        self._namespace = sqlalchemy.Table(
            'namespace', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('namespace', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('display_name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('namespace_type', sqlalchemy.Enum(NamespaceType), nullable=False, default=NamespaceType.NONE)
        )

        self._namespace_redirect = sqlalchemy.Table(
            'namespace_redirect', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_namespace_redirect_namespace_id_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            )
        )

        self._module_provider_redirect = sqlalchemy.Table(
            'module_provider_redirect', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            # Original module name/provider
            sqlalchemy.Column('module', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('provider', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            # Original namespace ID
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_module_provider_redirect_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            # Target module provider
            sqlalchemy.Column(
                'module_provider_id',
                sqlalchemy.ForeignKey(
                    'module_provider.id',
                    name='fk_module_provider_redirect_module_provider_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            )
        )

        self._module_provider = sqlalchemy.Table(
            'module_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_module_provider_namespace_id_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('module', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('provider', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('git_tag_format', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('git_path', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('verified', sqlalchemy.Boolean),
            sqlalchemy.Column(
                'git_provider_id',
                sqlalchemy.ForeignKey(
                    'git_provider.id',
                    name='fk_module_provider_git_provider_id_git_provider_id',
                    onupdate='CASCADE',
                    ondelete='SET NULL'),
                nullable=True
            ),
            sqlalchemy.Column(
                'latest_version_id',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    name='fk_module_provider_latest_version_id_module_version_id',
                    onupdate='CASCADE',
                    ondelete='SET NULL',
                    use_alter=True
                ),
                nullable=True
            )
        )

        self._module_details = sqlalchemy.Table(
            'module_details', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('readme_content', Database.medium_blob()),
            sqlalchemy.Column('terraform_docs', Database.medium_blob()),
            sqlalchemy.Column('tfsec', Database.medium_blob()),
            sqlalchemy.Column('infracost', Database.medium_blob()),
            sqlalchemy.Column('terraform_graph', Database.medium_blob()),
            sqlalchemy.Column('terraform_modules', Database.medium_blob()),
            sqlalchemy.Column('terraform_version', Database.medium_blob())
        )

        self._module_version = sqlalchemy.Table(
            'module_version', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'module_provider_id',
                sqlalchemy.ForeignKey(
                    'module_provider.id',
                    name='fk_module_version_module_provider_id_module_provider_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('version', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column(
                'module_details_id',
                sqlalchemy.ForeignKey(
                    'module_details.id',
                    name='fk_module_version_module_details_id_module_details_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=True
            ),
            # Whether the module version is a beta version
            sqlalchemy.Column('beta', sqlalchemy.BOOLEAN, nullable=False),
            sqlalchemy.Column('owner', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('description', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('published_at', sqlalchemy.DateTime),
            sqlalchemy.Column('variable_template', Database.medium_blob()),
            sqlalchemy.Column('internal', sqlalchemy.Boolean, nullable=False),
            sqlalchemy.Column('published', sqlalchemy.Boolean),
            sqlalchemy.Column('extraction_version', sqlalchemy.Integer)
        )

        self._sub_module = sqlalchemy.Table(
            'submodule', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'parent_module_version',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    name='fk_submodule_parent_module_version_module_version_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column(
                'module_details_id',
                sqlalchemy.ForeignKey(
                    'module_details.id',
                    name='fk_submodule_module_details_id_module_details_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=True
            ),
            sqlalchemy.Column('type', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('path', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE))
        )

        self._analytics = sqlalchemy.Table(
            'analytics', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('parent_module_version', sqlalchemy.Integer, index=True, nullable=False),
            sqlalchemy.Column('timestamp', sqlalchemy.DateTime),
            sqlalchemy.Column('terraform_version', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('analytics_token', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('auth_token', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('environment', sqlalchemy.String(GENERAL_COLUMN_SIZE)),

            # Columns for providing redirect deletion protection
            sqlalchemy.Column('namespace_name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('module_name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('provider_name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
        )

        self._example_file = sqlalchemy.Table(
            'example_file', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'submodule_id',
                sqlalchemy.ForeignKey(
                    'submodule.id',
                    name='fk_example_file_submodule_id_submodule_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('path', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('content', Database.medium_blob())
        )

        # Additional files for module provider (e.g. additional README files)
        self._module_version_file = sqlalchemy.Table(
            'module_version_file', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column(
                'module_version_id',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    name='fk_module_version_file_module_version_id_module_version_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('path', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('content', Database.medium_blob())
        )

        self._gpg_key = sqlalchemy.Table(
            'gpg_key', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_gpg_key_namespace_id_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('ascii_armor', Database.medium_blob()),
            sqlalchemy.Column('key_id', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('fingerprint', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('source', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('source_url', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('created_at', sqlalchemy.DateTime),
            sqlalchemy.Column('updated_at', sqlalchemy.DateTime),
        )

        self._provider_source = sqlalchemy.Table(
            'provider_source', meta,
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), primary_key=True),
            sqlalchemy.Column('api_name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('provider_source_type', sqlalchemy.Enum(ProviderSourceType)),
            sqlalchemy.Column('config', Database.medium_blob()),
        )

        self._provider_category = sqlalchemy.Table(
            'provider_category', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('slug', sqlalchemy.String(GENERAL_COLUMN_SIZE), unique=True),
            sqlalchemy.Column('user_selectable', sqlalchemy.Boolean, default=True),
        )

        self._repository = sqlalchemy.Table(
            'repository', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('provider_id', sqlalchemy.String),
            sqlalchemy.Column('owner', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('description', Database.medium_blob()),
            sqlalchemy.Column('authentication_key', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('clone_url', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column(
                'provider_source_name',
                sqlalchemy.ForeignKey(
                    'provider_source.name',
                    name='fk_repository_provider_source_name_provider_source_name',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
        )

        self._provider = sqlalchemy.Table(
            'provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'namespace_id',
                sqlalchemy.ForeignKey(
                    'namespace.id',
                    name='fk_provider_namespace_id_namespace_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('description', Database.medium_blob()),
            sqlalchemy.Column('tier', sqlalchemy.Enum(ProviderTier)),
            sqlalchemy.Column(
                'provider_category_id',
                sqlalchemy.ForeignKey(
                    'provider_category.id',
                    name="fk_provider_provider_category_id_provider_category_id",
                    onupdate="CASCADE",
                    ondelete="SET NULL"
                )
            ),
            sqlalchemy.Column(
                'repository_id',
                sqlalchemy.ForeignKey(
                    'repository.id',
                    name='fk_provider_repository_id_repository_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=True
            ),
            sqlalchemy.Column(
                'latest_version_id',
                sqlalchemy.ForeignKey(
                    'provider_version.id',
                    name='fk_provider_latest_version_id_provider_version_id',
                    onupdate='CASCADE',
                    ondelete='SET NULL',
                    use_alter=True
                ),
                nullable=True
            )
        )

        self._provider_version = sqlalchemy.Table(
            'provider_version', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'provider_id',
                sqlalchemy.ForeignKey(
                    'provider.id',
                    name='fk_provider_version_provider_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
            sqlalchemy.Column(
                'gpg_key_id',
                sqlalchemy.ForeignKey(
                    'gpg_key.id',
                    name='fk_provider_version_gpg_key_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
            sqlalchemy.Column('version', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('git_tag', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('beta', sqlalchemy.BOOLEAN, nullable=False),
            sqlalchemy.Column('published_at', sqlalchemy.DateTime),
            sqlalchemy.Column('extraction_version', sqlalchemy.Integer),
            sqlalchemy.Column('protocol_versions', self.medium_blob()),
        )

        self._provider_version_documentation = sqlalchemy.Table(
            'provider_version_documentation', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'provider_version_id',
                sqlalchemy.ForeignKey(
                    'provider_version.id',
                    name='fk_provider_version_documentation_provider_version_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('filename', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column(
                'documentation_type',
                sqlalchemy.Enum(terrareg.provider_documentation_type.ProviderDocumentationType),
                nullable=False
            ),
            sqlalchemy.Column('content', Database.medium_blob())
        )

        self._provider_version_binary = sqlalchemy.Table(
            "provider_version_binary", meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'provider_version_id',
                sqlalchemy.ForeignKey(
                    'provider_version.id',
                    name='fk_provider_version_documentation_provider_version_id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'
                ),
                nullable=False
            ),
            sqlalchemy.Column("name", sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column("operating_system", sqlalchemy.Enum(terrareg.provider_binary_types.ProviderBinaryOperatingSystemType), nullable=False),
            sqlalchemy.Column("architecture", sqlalchemy.Enum(terrareg.provider_binary_types.ProviderBinaryArchitectureType), nullable=False),
            sqlalchemy.Column("checksum", sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
        )

        self._audit_history = sqlalchemy.Table(
            'audit_history', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key=True),
            sqlalchemy.Column('timestamp', sqlalchemy.DateTime),
            sqlalchemy.Column('username', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('action', sqlalchemy.Enum(AuditAction)),
            sqlalchemy.Column('object_type', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('object_id', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('old_value', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('new_value', sqlalchemy.String(GENERAL_COLUMN_SIZE))
        )

    def select_module_version_joined_module_provider(self, *select_args):
        """Perform select on module_version, joined to module_provider table."""
        return sqlalchemy.select(
            *select_args
        ).select_from(self.module_version).join(
            self.module_provider, self.module_version.c.module_provider_id==self.module_provider.c.id
        ).join(
            self.namespace, self.module_provider.c.namespace_id==self.namespace.c.id
        )

    def select_module_provider_joined_latest_module_version(self, *select_args):
        """Perform select on module_provider, joined to latest version from module_version table."""
        return sqlalchemy.select(
            *select_args
        ).select_from(self.module_provider).join(
            self.module_version, self.module_provider.c.latest_version_id==self.module_version.c.id
        ).join(
            self.namespace, self.module_provider.c.namespace_id==self.namespace.c.id
        )

    @classmethod
    def get_current_transaction(cls):
        """Check if currently in transaction."""
        if has_request_context():
            if cls.get().transaction_connection is not None:
                raise Exception('Global database transaction is present whilst in request context!')

            if flask.g.get('database_transaction_connection', None):
                return flask.g.get('database_transaction_connection', None)
        else:
            if cls.get().transaction_connection is not None:
                return cls.get().transaction_connection

        return None

    @classmethod
    def start_transaction(cls):
        """Start DB transaction, store in current context and return"""
        # Check if currently in transaction
        if cls.get_current_transaction():
            raise Exception('Already within database transaction')
        conn = Database.get().get_connection()
        return Transaction(conn)

    @classmethod
    def get_connection(cls):
        """Get connection, checking for transaction and returning it."""
        current_transaction = cls.get_current_transaction()
        if current_transaction is not None:
            # Wrap current transaction in fake 'with' wrapper,
            # to handle 'with get_connection():'
            return TransactionConnectionWrapper(current_transaction)

        # If transaction is not currently active, return database connection
        return cls.get().get_engine().connect()


class TransactionConnectionWrapper:

    def __init__(self, transaction):
        """Store transaction"""
        self._transaction = transaction

    def __enter__(self):
        """On enter, return transaction."""
        return self._transaction

    def __exit__(self, *args, **kwargs):
        """Do nothing on exit"""
        self._transaction = None


class Transaction:
    """Custom wrapper for database transaction."""

    @property
    def transaction(self):
        """Return database transaction object."""
        return self._transaction_outer

    @property
    def connection(self):
        """Return database connection object."""
        return self._connection

    def __init__(self, connection):
        """Store database connection."""
        self._connection = connection
        self._transaction_outer = None
    
    def __enter__(self):
        """Start transaction and store in current context."""
        self._transaction_outer = self._connection.begin()

        self._transaction_outer.__enter__()

        # Store current transaction in context, so it is
        # returned by any get_connection methods
        if has_request_context():
            flask.g.database_transaction_connection = self._connection
        else:
            Database.get().transaction = self._connection

        return self

    def __exit__(self, *args, **kwargs):
        """End transaction and remove from current context."""
        if has_request_context():
            flask.g.database_transaction_connection = None
        else:
            Database.get().transaction = None

        self._transaction_outer.__exit__(*args, **kwargs)

