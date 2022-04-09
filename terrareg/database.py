
import sqlalchemy


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

