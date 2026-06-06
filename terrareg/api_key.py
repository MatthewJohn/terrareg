
import datetime
import hashlib
import secrets
from enum import Enum
from typing import Optional, Tuple, List

import sqlalchemy

from terrareg.errors import InvalidApiKeyTypeError
from terrareg.database import Database


class ApiKeyType(Enum):
    """Supported API key types."""

    ADMIN = 'admin'
    UPLOAD = 'upload'
    PUBLISH = 'publish'
    UPLOAD_AND_PUBLISH = 'upload_and_publish'


class ApiKey:
    """Interface to create and validate UI-managed API keys."""

    TOKEN_BYTES = 32
    PREFIX_LENGTH = 8
    PBKDF2_ITERATIONS = 600000

    @classmethod
    def _normalise_key_type(cls, key_type):
        """Convert key type to a validated string value."""
        if isinstance(key_type, ApiKeyType):
            return key_type.value

        if isinstance(key_type, str):
            key_type = key_type.lower()
            if key_type == 'module_full':
                key_type = ApiKeyType.UPLOAD_AND_PUBLISH.value
            valid_key_types = [api_key_type.value for api_key_type in ApiKeyType]
            if key_type in valid_key_types:
                return key_type

        raise InvalidApiKeyTypeError('Invalid API key type: {}'.format(key_type))

    @classmethod
    def _generate_key_hash(cls, api_key, salt):
        """Generate a stable PBKDF2 hash for an API key."""
        return hashlib.pbkdf2_hmac(
            'sha256',
            api_key.encode('utf-8'),
            salt.encode('utf-8'),
            cls.PBKDF2_ITERATIONS
        ).hex()

    @classmethod
    def create(cls, name, key_type, created_by=None, expires_at=None, namespace=None):
        """Create an API key and return the model plus the plaintext secret."""
        key_type = cls._normalise_key_type(key_type)
        plaintext_key = secrets.token_urlsafe(cls.TOKEN_BYTES)
        salt = secrets.token_hex(16)
        key_hash = cls._generate_key_hash(plaintext_key, salt)

        db = Database.get()
        insert = db.api_key.insert().values(
            name=name,
            key_type=key_type,
            key_prefix=plaintext_key[:cls.PREFIX_LENGTH],
            key_hash=key_hash,
            key_salt=salt,
            created_at=datetime.datetime.now(),
            created_by=created_by,
            last_used_at=None,
            expires_at=expires_at,
            is_active=True,
            namespace=namespace or None,
        )

        with db.get_connection() as conn:
            res = conn.execute(insert)

        return cls(id=res.inserted_primary_key[0]), plaintext_key

    @classmethod
    def get(cls, id):
        """Get API key by primary key."""
        api_key = cls(id=id)
        if api_key._get_db_row() is None:
            return None
        return api_key

    @classmethod
    def get_all(cls, key_type=None):
        """Return all API keys, optionally filtered by type."""
        db = Database.get()
        select = db.api_key.select()
        if key_type is not None:
            select = select.where(db.api_key.c.key_type == cls._normalise_key_type(key_type))

        with db.get_connection() as conn:
            return [
                cls(id=row['id'])
                for row in conn.execute(select)
            ]

    @classmethod
    def has_active_keys(cls, key_type):
        """Whether any active, non-expired API keys exist for the type."""
        db = Database.get()
        select = sqlalchemy.select(db.api_key.c.id).where(
            db.api_key.c.key_type == cls._normalise_key_type(key_type),
            db.api_key.c.is_active == True,
            sqlalchemy.or_(
                db.api_key.c.expires_at.is_(None),
                db.api_key.c.expires_at > datetime.datetime.now()
            )
        ).limit(1)

        with db.get_connection() as conn:
            return conn.execute(select).fetchone() is not None

    @classmethod
    def verify_key(cls, api_key, key_type):
        """Validate a plaintext API key against stored hashes."""
        key_type = cls._normalise_key_type(key_type)
        db = Database.get()
        select = db.api_key.select().where(
            db.api_key.c.key_type == key_type,
            db.api_key.c.is_active == True,
            sqlalchemy.or_(
                db.api_key.c.expires_at.is_(None),
                db.api_key.c.expires_at > datetime.datetime.now()
            )
        )

        with db.get_connection() as conn:
            for row in conn.execute(select):
                candidate_hash = cls._generate_key_hash(api_key, row['key_salt'])
                if hmac.compare_digest(candidate_hash, row['key_hash']):
                    return cls(id=row['id'])
        return None

    @property
    def pk(self):
        """Return DB ID for API key."""
        return self._get_db_row()['id']

    @property
    def name(self):
        """Return API key name."""
        return self._get_db_row()['name']

    @property
    def key_type(self):
        """Return API key type."""
        return self._get_db_row()['key_type']

    @property
    def key_prefix(self):
        """Return API key prefix."""
        return self._get_db_row()['key_prefix']

    @property
    def created_at(self):
        """Return API key creation timestamp."""
        return self._get_db_row()['created_at']

    @property
    def created_by(self):
        """Return API key creator."""
        return self._get_db_row()['created_by']

    @property
    def last_used_at(self):
        """Return API key last usage timestamp."""
        return self._get_db_row()['last_used_at']

    @property
    def expires_at(self):
        """Return API key expiry timestamp."""
        return self._get_db_row()['expires_at']

    @property
    def is_active(self):
        """Return whether API key is active."""
        return self._get_db_row()['is_active']

    @property
    def namespace(self):
        """Return optional namespace restriction for this API key."""
        return self._get_db_row()['namespace']

    def __init__(self, id):
        """Store member variable for ID."""
        self._id = id
        self._row_cache = None

    def mark_used(self):
        """Update last-used time for the API key."""
        db = Database.get()
        update = db.api_key.update().where(
            db.api_key.c.id == self.pk
        ).values(last_used_at=datetime.datetime.now())
        with db.get_connection() as conn:
            conn.execute(update)
        self._row_cache = None

    def revoke(self):
        """Deactivate the API key."""
        db = Database.get()
        update = db.api_key.update().where(
            db.api_key.c.id == self.pk
        ).values(is_active=False)
        with db.get_connection() as conn:
            conn.execute(update)
        self._row_cache = None

    def delete(self):
        """Delete the API key."""
        db = Database.get()
        delete = db.api_key.delete().where(db.api_key.c.id == self.pk)
        with db.get_connection() as conn:
            conn.execute(delete)
        self._row_cache = None

    def _get_db_row(self):
        """Return DB row for API key."""
        if self._row_cache is None:
            db = Database.get()
            select = db.api_key.select().where(db.api_key.c.id == self._id)
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._row_cache = res.fetchone()
        return self._row_cache