"""Provide database class."""

import sqlalchemy

import terrareg.config
from terrareg.errors import DatabaseMustBeIniistalisedError


class Database():
    """Handle database connection and settng up database schema"""

    _META = None
    _ENGINE = None
    _INSTANCE = None

    def __init__(self):
        """Setup member variables."""
        self._git_provider = None
        self._module_provider = None
        self._module_version = None
        self._sub_module = None
        self._analytics = None

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

        self._git_provider = sqlalchemy.Table(
            'git_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('name', sqlalchemy.String, unique=True),
            sqlalchemy.Column('base_url_template', sqlalchemy.String),
            sqlalchemy.Column('clone_url_template', sqlalchemy.String),
            sqlalchemy.Column('browse_url_template', sqlalchemy.String)
        )

        self._module_provider = sqlalchemy.Table(
            'module_provider', meta,
            sqlalchemy.Column('id', sqlalchemy.Integer, primary_key = True),
            sqlalchemy.Column('namespace', sqlalchemy.String),
            sqlalchemy.Column('module', sqlalchemy.String),
            sqlalchemy.Column('provider', sqlalchemy.String),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String),
            sqlalchemy.Column('git_tag_format', sqlalchemy.String),
            sqlalchemy.Column('verified', sqlalchemy.Boolean),
            sqlalchemy.Column(
                'git_provider_id',
                sqlalchemy.ForeignKey(
                    'git_provider.id',
                    onupdate='CASCADE',
                    ondelete='SET NULL'),
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
            sqlalchemy.Column('version', sqlalchemy.String),
            sqlalchemy.Column('owner', sqlalchemy.String),
            sqlalchemy.Column('description', sqlalchemy.String),
            sqlalchemy.Column('repo_base_url_template', sqlalchemy.String),
            sqlalchemy.Column('repo_clone_url_template', sqlalchemy.String),
            sqlalchemy.Column('repo_browse_url_template', sqlalchemy.String),
            sqlalchemy.Column('published_at', sqlalchemy.DateTime),
            sqlalchemy.Column('readme_content', sqlalchemy.String),
            sqlalchemy.Column('module_details', sqlalchemy.String),
            sqlalchemy.Column('variable_template', sqlalchemy.String),
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
            sqlalchemy.Column('type', sqlalchemy.String),
            sqlalchemy.Column('path', sqlalchemy.String),
            sqlalchemy.Column('name', sqlalchemy.String),
            sqlalchemy.Column('readme_content', sqlalchemy.String),
            sqlalchemy.Column('module_details', sqlalchemy.String)
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
            sqlalchemy.Column('timestamp', sqlalchemy.String),
            sqlalchemy.Column('terraform_version', sqlalchemy.String),
            sqlalchemy.Column('analytics_token', sqlalchemy.String),
            sqlalchemy.Column('auth_token', sqlalchemy.String),
            sqlalchemy.Column('environment', sqlalchemy.String)
        )

    def select_module_version_joined_module_provider(self, *args, **kwargs):
        """Perform select on module_version, joined to module_provider table."""
        return sqlalchemy.select(self.module_version.join(
            self.module_provider, self.module_version.c.module_provider_id == self.module_provider.c.id
        ))
