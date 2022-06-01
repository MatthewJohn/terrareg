"""Provide database class."""

import sqlalchemy
import sqlalchemy.dialects.mysql

import terrareg.config
from terrareg.errors import DatabaseMustBeIniistalisedError


class Database():
    """Handle database connection and settng up database schema"""

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
        self._git_provider = None
        self._module_provider = None
        self._module_version = None
        self._sub_module = None
        self._analytics = None
        self._example_file = None

    @property
    def git_provider(self):
        """Return git_provider table."""
        if self._git_provider is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._git_provider

    @property
    def module_provider(self):
        """Return module_provider table."""
        if self._module_provider is None:
            raise DatabaseMustBeIniistalisedError('Database class must be initialised.')
        return self._module_provider

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
                echo=terrareg.config.Config().DEBUG)
        return cls._ENGINE

    def initialise(self):
        """Initialise database schema."""
        meta = self.get_meta()
        engine = self.get_engine()

        GENERAL_COLUMN_SIZE = 128
        LARGE_COLUMN_SIZE = 1024
        URL_COLUMN_SIZE = 1024

        self._git_provider = sqlalchemy.Table(
            'git_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE), unique=True),
            sqlalchemy.Column('base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE))
        )

        self._module_provider = sqlalchemy.Table(
            'module_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('namespace', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('module', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('provider', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('git_tag_format', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('verified', sqlalchemy.Boolean),
            sqlalchemy.Column(
                'git_provider_id',
                sqlalchemy.ForeignKey(
                    'git_provider.id',
                    onupdate='CASCADE',
                    ondelete='SET NULL'),
                nullable=True
            ),
            sqlalchemy.Column(
                'latest_version_id',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    onupdate='CASCADE',
                    ondelete='SET NULL',
                    use_alter=True
                ),
                nullable=True
            )
        )

        self._module_version = sqlalchemy.Table(
            'module_version', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'module_provider_id',
                sqlalchemy.ForeignKey(
                    'module_provider.id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('version', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            # Whether the module version is a beta version
            sqlalchemy.Column('beta', sqlalchemy.BOOLEAN, nullable=False),
            sqlalchemy.Column('owner', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('description', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String(URL_COLUMN_SIZE)),
            sqlalchemy.Column('published_at', sqlalchemy.DateTime),
            sqlalchemy.Column('readme_content', Database.medium_blob()),
            sqlalchemy.Column('module_details', Database.medium_blob()),
            sqlalchemy.Column('variable_template', Database.medium_blob()),
            sqlalchemy.Column('internal', sqlalchemy.Boolean, nullable=False),
            sqlalchemy.Column('published', sqlalchemy.Boolean)
        )

        self._sub_module = sqlalchemy.Table(
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
            sqlalchemy.Column('type', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('path', sqlalchemy.String(LARGE_COLUMN_SIZE)),
            sqlalchemy.Column('name', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('readme_content', Database.medium_blob()),
            sqlalchemy.Column('module_details', Database.medium_blob())
        )

        self._analytics = sqlalchemy.Table(
            'analytics', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'parent_module_version',
                sqlalchemy.ForeignKey(
                    'module_version.id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('timestamp', sqlalchemy.DateTime),
            sqlalchemy.Column('terraform_version', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('analytics_token', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('auth_token', sqlalchemy.String(GENERAL_COLUMN_SIZE)),
            sqlalchemy.Column('environment', sqlalchemy.String(GENERAL_COLUMN_SIZE))
        )

        self._example_file = sqlalchemy.Table(
            'example_file', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column(
                'submodule_id',
                sqlalchemy.ForeignKey(
                    'submodule.id',
                    onupdate='CASCADE',
                    ondelete='CASCADE'),
                nullable=False
            ),
            sqlalchemy.Column('path', sqlalchemy.String(GENERAL_COLUMN_SIZE), nullable=False),
            sqlalchemy.Column('content', Database.medium_blob())
        )

    def select_module_version_joined_module_provider(self, *select_args):
        """Perform select on module_version, joined to module_provider table."""
        return sqlalchemy.select(
            *select_args
        ).select_from(self.module_version).join(
            self.module_provider, self.module_version.c.module_provider_id == self.module_provider.c.id
        )

    def select_module_provider_joined_latest_module_version(self, *select_args):
        """Perform select on module_provider, joined to latest version from module_version table."""
        return sqlalchemy.select(
            *select_args
        ).select_from(self.module_provider).join(
            self.module_version, self.module_provider.c.latest_version_id == self.module_version.c.id
        )
