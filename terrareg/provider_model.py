"""
Module for Provider model.

@TODO Split models from models.py into seperate modules in new package.
"""

import re
from typing import Union

import terrareg.database
from terrareg.errors import DuplicateProviderError, InvalidModuleProviderNameError, ProviderNameNotPermittedError
import terrareg.config
import terrareg.provider_tier
import terrareg.audit
import terrareg.audit_action
import terrareg.models


class Provider:

    @classmethod
    def create(cls, namespace: 'terrareg.models.Namespace', name: str, description: str) -> 'Provider':
        """Create instance of object in database."""
        # Validate module provider name
        cls._validate_name(name)

        # Ensure that there is not already a provider that exists
        duplicate_provider = Provider.get(namespace=namespace, name=name)
        if duplicate_provider:
            raise DuplicateProviderError("A duplicate provider exists with the same name in the namespace")

        # Create module provider
        db = terrareg.database.Database.get()

        insert = db.provider.insert().values(
            namespace_id=namespace.pk,
            name=name,
            description=description,
            tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
            provider_source_id=None
        )
        with db.get_connection() as conn:
            conn.execute(insert)

        obj = cls(namespace=namespace, name=name)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.PROVIDER_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    def _validate_name(name: str) -> None:
        """Validate name of module"""
        if not re.match(r'^[0-9a-z]+$', name):
            raise InvalidModuleProviderNameError('Provider name is invalid')

    @classmethod
    def get(cls, namespace: 'terrareg.models.Namespace', name: str) -> Union['Provider', None]:
        """Create object and ensure the object exists."""
        obj = cls(namespace=namespace, name=name)

        # If there is no row, the module provider does not exist
        if obj._get_db_row() is None:
            return None

        # Otherwise, return object
        return obj

    def __init__(self, namespace: 'terrareg.models.Namespace', name: str):
        """Validate name and store member variables."""
        self._validate_name(name)
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
                db.namespace.c.id == self._module._namespace.pk,
                db.provider.c.name == self.name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row
