
import re
from typing import Union, List, Tuple

import sqlalchemy

import terrareg.provider_source
import terrareg.provider_source.factory
import terrareg.audit
import terrareg.audit_action
import terrareg.database
import terrareg.repository_kind
import terrareg.provider_source.repository_release_metadata
import terrareg.provider_model


class Repository:

    @classmethod
    def create(cls, provider_source: 'terrareg.provider_source.BaseProviderSource',
               provider_id: str, name: str, description: str, owner: str,
               clone_url: str, logo_url: Union[str, None]) -> Union[None, 'Repository']:
        """Create user group"""
        # Check if repository exists by provider source and ID
        if cls.get_by_provider_source_and_provider_id(provider_source=provider_source, provider_id=provider_id):
            return None

        pk = cls._insert_into_database(
            provider_source=provider_source,
            provider_id=provider_id,
            name=name,
            description=description,
            owner=owner,
            clone_url=clone_url,
            logo_url=logo_url,
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
                              provider_id: str, name: str, description: str, owner: str,
                              clone_url: str, logo_url: Union[str, None]) -> int:
        """Insert new user group into database."""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.repository.insert().values(
                provider_source_name=provider_source.name,
                provider_id=provider_id,
                name=name,
                description=terrareg.database.Database.encode_blob(description),
                owner=owner,
                clone_url=clone_url,
                logo_url=logo_url
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

    @classmethod
    def get_by_pk(cls, pk: int) -> Union[None, 'Repository']:
        """Get repository by ID"""
        db = terrareg.database.Database.get()
        select = db.repository.select().where(
            db.repository.c.id==pk
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
    def provider_source(self) -> 'terrareg.provider_source.BaseProviderSource':
        """Return provider source for repository"""
        return terrareg.provider_source.factory.ProviderSourceFactory.get().get_provider_source_by_name(
            self._get_db_row()["provider_source_name"]
        )

    @property
    def kind(self) -> Union[None, 'terrareg.repository_kind.RepositoryKind']:
        """Return repository kind"""
        if self.name.startswith("terraform-provider-"):
            return terrareg.repository_kind.RepositoryKind.PROVIDER
        elif self.name.startswith("terraform-"):
            return terrareg.repository_kind.RepositoryKind.MODULE
        return None

    @property
    def description(self) -> Union[None, str]:
        """Return description"""
        if description_blob := self._get_db_row()["description"]:
            return terrareg.database.Database.decode_blob(description_blob)
        return None

    @property
    def logo_url(self) -> Union[str, None]:
        """Return logo URL"""
        return self._get_db_row()["logo_url"]

    @property
    def clone_url(self) -> str:
        """Return clone URL"""
        return self._get_db_row()["clone_url"]

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

    def update_attributes(self, **kwargs):
        """Update DB row."""
        # Check for any blob and encode the values
        for kwarg in kwargs:
            if kwarg in ['description']:
                kwargs[kwarg] = terrareg.database.Database.encode_blob(kwargs[kwarg])

        db = terrareg.database.Database.get()
        update = sqlalchemy.update(
            db.repository
        ).where(
            db.repository.c.id==self.pk
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None

    def get_new_releases(self, provider: 'terrareg.provider_model.Provider') -> List['terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata']:
        """Obtain all repository releases that aren't associated with a pre-existing release"""
        return self.provider_source.get_new_releases(
            provider=provider
        )

    def get_release_artifact(self,
                             provider: 'terrareg.provider_model.Provider',
                             artifact_metadata: 'terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata',
                             release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata'):
        """Return release artifact file content from provider source, injecting access token"""
        return self.provider_source.get_release_artifact(
            provider=provider,
            artifact_metadata=artifact_metadata,
            release_metadata=release_metadata
        )

    def get_release_archive(self,
                            provider: 'terrareg.provider_model.Provider',
                            release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> Tuple[bytes, Union[None, str]]:
        """Obtain release archive using provider source, injecting access token, returning bytes of archive and the sub-directory used for extraction"""
        return self.provider_source.get_release_archive(
            provider=provider,
            release_metadata=release_metadata,
        )
