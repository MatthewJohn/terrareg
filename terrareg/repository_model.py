
from typing import Union, List

import terrareg.provider_source
import terrareg.audit
import terrareg.audit_action
import terrareg.database
import terrareg.repository_kind


class Repository:

    @classmethod
    def create(cls, provider_source: 'terrareg.provider_source.BaseProviderSource',
               provider_id: str, name: str, owner: str) -> Union[None, 'Repository']:
        """Create user group"""
        # Check if repository exists by provider source and ID
        if cls.get_by_provider_source_and_provider_id(provider_source=provider_source, provider_id=provider_id):
            return None

        pk = cls._insert_into_database(
            provider_source=provider_source,
            provider_id=provider_id,
            name=name, owner=owner
        )

        obj = cls(pk=pk)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.REPOSITORY_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def _insert_into_database(cls, provider_source: 'terrareg.provider_source.BaseProviderSource',
                              provider_id: str, name: str, owner: str) -> int:
        """Insert new user group into database."""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.repository.insert().values(
                provider_source_name=provider_source.name,
                provider_id=provider_id,
                name=name,
                owner=owner
            ))
            return res.lastrowid

    @classmethod
    def get_repositories_by_owner_list(cls, owners: List[str]) -> List['Repository']:
        """Return list of repositories matching a list of owners"""
        db = terrareg.database.Database.get()
        select = db.repository.select().where(
            db.repository.c.owner.in_(owners)
        )
        with db.get_connection() as conn:
            res = conn.execute(select).all()
            return [
                cls(pk=row['id'])
                for row in res
            ]

    @classmethod
    def get_by_provider_source_and_provider_id(cls, provider_source: 'terrareg.provider_source.BaseProviderSource', provider_id: str) -> Union[None, 'Repository']:
        """Get repository by provider source and provider ID"""
        db = terrareg.database.Database.get()
        select = db.repository.select().where(
            db.repository.c.provider_source_name==provider_source.name,
            db.repository.c.provider_id==provider_id
        )
        with db.get_connection() as conn:
            row = conn.execute(select).fetchone()

        if row:
            return cls(pk=row['id'])

    @property
    def pk(self) -> int:
        """Return DB ID"""
        return self._pk

    @property
    def id(self) -> str:
        """Return ID of repository"""
        return f"{self.owner}/{self.name}"

    @property
    def owner(self) -> str:
        """Return owner of repository"""
        return self._get_db_row()['owner']

    @property
    def name(self) -> str:
        """Return name of repository"""
        return self._get_db_row()['name']

    @property
    def provider_id(self) -> str:
        """Return provider ID of repository"""
        return self._get_db_row()['provider_id']

    @property
    def kind(self) -> Union[None, terrareg.repository_kind.RepositoryKind]:
        """Return repository kind"""
        if self.name.startswith("terraform-provider-"):
            return terrareg.repository_kind.RepositoryKind.PROVIDER
        elif self.name.startswith("terraform-"):
            return terrareg.repository_kind.RepositoryKind.MODULE
        return None

    def __init__(self, pk: int):
        """Store member variables"""
        self._pk = pk
        self._row_cache = None

    def _get_db_row(self):
        """Return DB row for user group."""
        if self._row_cache is None:
            db = terrareg.database.Database.get()
            # Obtain row from user group table.
            select = db.repository.select().where(
                db.repository.c.id==self.pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._row_cache = res.fetchone()
        return self._row_cache
