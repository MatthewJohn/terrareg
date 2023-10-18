"""
Module for Provider model.

@TODO Split models from models.py into seperate modules in new package.
"""

import re
from typing import Union

import sqlalchemy

import terrareg.database
from terrareg.errors import DuplicateProviderError, InvalidModuleProviderNameError, ProviderNameNotPermittedError
import terrareg.config
import terrareg.provider_tier
import terrareg.audit
import terrareg.audit_action
import terrareg.models
import terrareg.provider_source
import terrareg.repository_model
import terrareg.provider_category_model


class Provider:

    @classmethod
    def create(cls, repository: 'terrareg.repository_model.Repository',
               provider_category: 'terrareg.provider_category_model.ProviderCategory') -> 'Provider':
        """Create instance of object in database."""
        # Ensure that there is not already a provider that exists
        duplicate_provider = Provider.get_by_repository(repository=repository)
        if duplicate_provider:
            raise DuplicateProviderError("A duplicate provider exists with the same name in the namespace")

        # Obtain namespace based on repository owner
        namespace = terrareg.models.Namespace.get(name=repository.owner, create=False, include_redirect=False, case_insensitive=True)

        # Create provider
        db = terrareg.database.Database.get()

        insert = db.provider.insert().values(
            namespace_id=namespace.pk,
            name=repository.name,
            description=None,
            tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
            repository_id=repository.pk,
            provider_category_id=provider_category.pk
        )
        with db.get_connection() as conn:
            conn.execute(insert)

        obj = cls(namespace=namespace, name=repository.name)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.PROVIDER_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def get_by_repository(cls, repository: 'terrareg.repository_model.Repository') -> Union[None, 'Provider']:
        """Obtain provider by repository"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider.c.namespace_id,
            db.provider.c.name
        ).select_from(
            db.provider
        ).where(
            db.provider.c.repository_id==repository.pk
        )
        with db.get_connection() as conn:
            row = conn.execute(select).fetchone()
        if not row:
            return None

        return cls(namespace=terrareg.models.Namespace.get_by_pk(row["namespace_id"]), name=row["name"])

    @classmethod
    def get(cls, namespace: 'terrareg.models.Namespace', name: str) -> Union['Provider', None]:
        """Create object and ensure the object exists."""
        obj = cls(namespace=namespace, name=name)

        # If there is no row, the module provider does not exist
        if obj._get_db_row() is None:
            return None

        # Otherwise, return object
        return obj

    @property
    def id(self) -> int:
        """Obtain id"""
        return f"{self.namespace.name}/{self.name}"

    @property
    def namespace(self) -> 'terrareg.models.Namespace':
        """Return namespace for provider"""
        return terrareg.models.Namespace.get_by_pk(pk=self._get_db_row()["namespace_id"])

    @property
    def name(self) -> str:
        """Return provider name"""
        return self._name

    def __init__(self, namespace: 'terrareg.models.Namespace', name: str):
        """Validate name and store member variables."""
        self._namespace = namespace
        self._name = name
        self._cache_db_row = None

    def _get_db_row(self) -> Union[dict, None]:
        """Return database row for module provider."""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db.provider.select(
            ).join(
                db.namespace,
                db.provider.c.namespace_id==db.namespace.c.id
            ).where(
                db.namespace.c.id == self._namespace.pk,
                db.provider.c.name == self.name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row
