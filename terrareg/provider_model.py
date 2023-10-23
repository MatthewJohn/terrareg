"""
Module for Provider model.

@TODO Split models from models.py into seperate modules in new package.
"""

from distutils.version import LooseVersion
import os
import re
from typing import Union, List

import sqlalchemy

import terrareg.database
from terrareg.errors import CouldNotFindGpgKeyForProviderVersionError, DuplicateProviderError, InvalidModuleProviderNameError, InvalidRepositoryNameError, ProviderNameNotPermittedError
import terrareg.config
import terrareg.provider_tier
import terrareg.audit
import terrareg.audit_action
import terrareg.models
import terrareg.provider_source
import terrareg.repository_model
import terrareg.provider_category_model
import terrareg.provider_version_model
import terrareg.provider_extractor
import terrareg.utils


class Provider:

    @classmethod
    def repository_name_to_provider_name(cls, repository_name: str) -> Union[None, str]:
        """Convert repository name to provider name"""
        if repository_name and repository_name.startswith("terraform-provider-"):
            return re.sub(r"^terraform-provider-", "", repository_name)
        return None

    @classmethod
    def create(cls, repository: 'terrareg.repository_model.Repository',
               provider_category: 'terrareg.provider_category_model.ProviderCategory') -> 'Provider':
        """Create instance of object in database."""
        # Ensure that there is not already a provider that exists
        duplicate_provider = Provider.get_by_repository(repository=repository)
        if duplicate_provider:
            # @TODO TEMP
            return duplicate_provider
            raise DuplicateProviderError("A duplicate provider exists with the same name in the namespace")

        # Obtain namespace based on repository owner
        namespace = terrareg.models.Namespace.get(name=repository.owner, create=False, include_redirect=False, case_insensitive=True)

        # Create provider
        db = terrareg.database.Database.get()

        if not (provider_name := cls.repository_name_to_provider_name(repository_name=repository.name)):
            raise InvalidRepositoryNameError("Invalid repository name")

        insert = db.provider.insert().values(
            namespace_id=namespace.pk,
            name=provider_name,
            description=db.encode_blob(repository.description),
            tier=terrareg.provider_tier.ProviderTier.COMMUNITY,
            repository_id=repository.pk,
            provider_category_id=provider_category.pk
        )
        with db.get_connection() as conn:
            conn.execute(insert)

        obj = cls(namespace=namespace, name=provider_name)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.PROVIDER_CREATE,
            object_type=obj.__class__.__name__,
            object_id=obj.id,
            old_value=None, new_value=None
        )

        return obj

    @classmethod
    def get_by_pk(cls, pk: int) -> Union[None, 'Provider']:
        """Obtain provider by primary key"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider.c.namespace_id,
            db.provider.c.name
        ).select_from(
            db.provider
        ).where(
            db.provider.c.id==pk
        )

        with db.get_connection() as conn:
            row = conn.execute(select).first()
        if not row:
            return None

        namespace = terrareg.models.Namespace.get_by_pk(pk=row["namespace_id"])
        if namespace is None:
            return None

        return cls(namespace=namespace, name=row["name"])

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

    @property
    def full_name(self) -> str:
        """Return full name, i.e. terraform-provider-name"""
        return f"terraform-provider-{self.name}"

    @property
    def pk(self) -> int:
        """Return DB pk for provider"""
        return self._get_db_row()["id"]

    @property
    def tier(self) -> 'terrareg.provider_tier.ProviderTier':
        """Return provider tier"""
        return self._get_db_row()["tier"]

    @property
    def base_directory(self) -> str:
        """Return base directory."""
        return terrareg.utils.safe_join_paths(self._namespace.base_provider_directory, self._name)

    @property
    def repository(self) -> 'terrareg.repository_model.Repository':
        """Return repository for provider"""
        return terrareg.repository_model.Repository.get_by_pk(self._get_db_row()["repository_id"])

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

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._namespace.base_provider_directory):
            self._namespace.create_provider_data_directory()
        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    def get_latest_version(self) -> Union[None, 'terrareg.provider_version_model.ProviderVersion']:
        """Return latest version of module provider"""
        if provider_version_pk := self._get_db_row()["latest_version_id"]:
            return terrareg.provider_version_model.ProviderVersion.get_by_pk(provider_version_pk)
        return None

    def get_all_versions(self) -> List['terrareg.provider_version_model.ProviderVersion']:
        """Return list of all provider versions"""
        db = terrareg.database.Database.get()
        select = db.provider_version.select(
            db.provider_version.c.version
        ).where(
            db.provider_version.c.provider_id==self.pk
        )
        with db.get_connection() as conn:
            rows = conn.execute(select).all()
        return [
            terrareg.provider_version_model.ProviderVersion(provider=self, version=row["version"])
            for row in rows
        ]


    def calculate_latest_version(self):
        """Obtain all versions of provider and sort by semantic version numbers to obtain latest version."""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version.c.version
        ).join(
            db.provider,
            db.provider_version.c.provider_id==db.provider.c.id
        ).where(
            db.provider.c.id==self.pk,
            db.provider_version.c.beta==False
        )
        with db.get_connection() as conn:
            res = conn.execute(select)

            # Convert to list
            rows = [r for r in res]

        # Sort rows by semantic versioning
        rows.sort(key=lambda x: LooseVersion(x['version']), reverse=True)

        # Ensure at least one row
        if not rows:
            return None

        # Obtain latest row
        return terrareg.provider_version_model.ProviderVersion(provider=self, version=rows[0]['version'])

    def refresh_versions(self) -> List['terrareg.provider_version_model.ProviderVersion']:
        """Refresh versions from provider source and create new provider versions"""
        repository = self.repository

        releases_metadata = repository.get_new_releases(provider=self)

        provider_versions = []

        for release_metadata in releases_metadata:
            provider_version = terrareg.provider_version_model.ProviderVersion(provider=self, version=release_metadata.version)

            gpg_key = terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                repository=repository, release_metadata=release_metadata,
                namespace=self.namespace
            )
            if not gpg_key:
                raise CouldNotFindGpgKeyForProviderVersionError(f"Could not find a valid GPG key to verify the signature of the release: {release_metadata.name}")

            with provider_version.create_extraction_wrapper(git_tag=release_metadata.tag, gpg_key=gpg_key):
                provider_extractor = terrareg.provider_extractor.ProviderExtractor(
                    provider_version=provider_version,
                    release_metadata=release_metadata
                )
                provider_extractor.process_version()

            provider_versions.append(provider_version)

        return provider_versions

    def update_attributes(self, **kwargs: dict) -> None:
        """Update DB row."""
        db = terrareg.database.Database.get()
        update = sqlalchemy.update(db.provider).where(
            db.provider.c.namespace_id==self.namespace.pk,
            db.provider.c.name==self.name
        ).values(**kwargs)
        with db.get_connection() as conn:
            conn.execute(update)

        # Remove cached DB row
        self._cache_db_row = None
