
import contextlib
from typing import Union
import os
import re

import sqlalchemy
import semantic_version
from terrareg.constants import PROVIDER_EXTRACTION_VERSION

from terrareg.errors import InvalidVersionError, ReindexingExistingProviderVersionsIsProhibitedError
import terrareg.utils
import terrareg.provider_model
import terrareg.database
import terrareg.audit
import terrareg.audit_action
import terrareg.models
import terrareg.provider_version_documentation_model


class ProviderVersion:

    @staticmethod
    def _validate_version(version):
        """Validate version, checking if version is a beta version."""
        match = re.match(r'^[0-9]+\.[0-9]+\.[0-9]+((:?-[a-z0-9]+)?)$', version)
        if not match:
            raise InvalidVersionError('Version is invalid')
        return bool(match.group(1))

    @classmethod
    def get(cls, provider: 'terrareg.provider_model.Provider', version: str) -> Union[None, 'ProviderVersion']:
        """Get provider version"""
        obj = cls(provider=provider, version=version)
        if obj._get_db_row():
            return obj
        return None

    @classmethod
    def get_by_pk(cls, pk: int) -> Union[None, 'ProviderVersion']:
        """Obtain provider version by primary key"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version.c.provider_id,
            db.provider_version.c.version
        ).select_from(
            db.provider_version
        ).where(
            db.provider_version.c.id==pk
        )
        with db.get_connection() as conn:
            row = conn.execute(select).first()
        if not row:
            return None
        provider = terrareg.provider_model.Provider.get_by_pk(pk=row["provider_id"])
        if provider is None:
            return None

        return cls(provider=provider, version=row["version"])

    @property
    def publish_date_display(self):
        """Return display view of date of module published."""
        published_at = self._get_db_row()['published_at']
        if published_at:
            return published_at.strftime('%B %d, %Y')
        return None

    @property
    def published(self):
        """Return whether module is published"""
        return bool(self._get_db_row()['published'])

    @property
    def version(self):
        """Return version."""
        return self._version

    @property
    def source_git_tag(self):
        """Return git tag used for extraction clone"""
        tag = semantic_version.Version(version_string=self._version)
        return self._module_provider.git_tag_format.format(
            version=self._version,
            major=tag.major,
            minor=tag.minor,
            patch=tag.patch,
            build=tag.build
        )

    @property
    def git_tag_ref(self):
        """Return git tag ref for extraction."""
        tag = semantic_version.Version(version_string=self._version)
        return self._module_provider.git_ref_format.format(
            version=self._version,
            major=tag.major,
            minor=tag.minor,
            patch=tag.patch,
            build=tag.build
        )

    @property
    def base_directory(self) -> str:
        """Return base directory."""
        return terrareg.utils.safe_join_paths(self._provider.base_directory, self._version)

    @property
    def beta(self) -> bool:
        """Return whether module version is a beta version."""
        return self._get_db_row()['beta']

    @property
    def pk(self) -> bool:
        """Return DB ID of provider version."""
        return self._get_db_row()['id']

    @property
    def id(self) -> str:
        """Return ID in form of namespace/provider/version"""
        return '{provider_id}/{version}'.format(
            provider_id=self._provider.id,
            version=self.version
        )

    @property
    def exists(self):
        """Determine if provider version exists"""
        return bool(self._get_db_row())

    @property
    def provider(self) -> 'terrareg.provider_model.Provider':
        """Return provider"""
        return self._provider

    @property
    def provider_extraction_up_to_date(self):
        """Whether the extracted version data is up-to-date"""
        return self._get_db_row()["extraction_version"] == PROVIDER_EXTRACTION_VERSION

    @property
    def is_latest_version(self):
        """Return whether the version is the latest version for the provider"""
        return self._provider.get_latest_version() == self

    def __init__(self, provider: 'terrareg.provider_model.Provider', version: str):
        """Setup member variables."""
        self._extracted_beta_flag = self._validate_version(version)
        self._provider = provider
        self._version = version
        self._cache_db_row = None

    def __eq__(self, __o):
        """Check if two module versions are the same"""
        if isinstance(__o, self.__class__):
            return self.pk == __o.pk
        return super(ProviderVersion, self).__eq__(__o)

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db._provider_version.select().join(
                db.provider,
                db.provider_version.c.provider_id == db.provider.c.id
            ).where(
                db.provider.c.id == self._provider.pk,
                db.provider_version.c.version == self.version
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def create_data_directory(self):
        """Create data directory and data directories of parents."""
        # Check if parent exists
        if not os.path.isdir(self._provider.base_directory):
            self._provider.create_data_directory()

        # Check if data directory exists
        if not os.path.isdir(self.base_directory):
            os.mkdir(self.base_directory)

    @contextlib.contextmanager
    def create_extraction_wrapper(self, git_tag: str, gpg_key: 'terrareg.models.GpgKey'):
        """Handle module creation with yield for extraction"""
        self.prepare_version(gpg_key=gpg_key, git_tag=git_tag)

        yield

        self.publish()

    def prepare_version(self, git_tag: str, gpg_key: 'terrareg.models.GpgKey'):
        """
        Handle file upload of provider version.
        """
        self.create_data_directory()
        self._create_db_row(gpg_key=gpg_key, git_tag=git_tag)

        terrareg.audit.AuditEvent.create_audit_event(
            action=terrareg.audit_action.AuditAction.PROVIDER_VERSION_INDEX,
            object_type=self.__class__.__name__,
            object_id=self.id,
            old_value=None,
            new_value=None
        )

    def publish(self):
        """Publish module version."""
        # Calculate latest version will take beta flag into account and will only match
        # the current version if the current version is latest and is capable of being the
        # latest version.
        if (self._provider.calculate_latest_version() is not None and
                self._provider.calculate_latest_version().version == self.version):
            self._provider.update_attributes(latest_version_id=self.pk)

    def get_api_outline(self, target_terraform_version=None):
        """Return dict of basic version details for API response."""
        raise NotImplementedError
        row = self._get_db_row()
        api_outline = self._module_provider.get_api_outline()
        api_outline.update({
            "id": self.id,
            "owner": row['owner'],
            "version": self.version,
            "description": row['description'],
            "source": self.get_source_base_url(),
            "published_at": row['published_at'].isoformat() if row['published_at'] else None,
            "downloads": self.get_total_downloads(),
            "internal": self._get_db_row()['internal']
        })

        if target_terraform_version is not None:
            api_outline['version_compatibility'] = terrareg.version_constraint.VersionConstraint.is_compatible(
                constraint=self.get_terraform_version_constraints(),
                target_version=target_terraform_version
            ).value
        return api_outline

    def get_total_downloads(self):
        """Obtain total number of downloads for module version."""
        raise NotImplementedError
        return terrareg.analytics.AnalyticsEngine.get_module_version_total_downloads(
            module_version=self
        )

    def get_api_details(self) -> dict:
        """Return dict of version details for API response."""
        return {
            "id": self.id,
            "owner": self._provider.repository.owner,
            "namespace": self._provider.namespace.name,
            "name": self._provider.name,
            "alias": None,
            "version": self.version,
            "tag": None,
            "description": self.provider.repository.description,
            "source": None,
            "published_at": (self._get_db_row()["published_at"].toisoformat() if self._get_db_row()["published_at"] else None),
            "downloads": 0,
            "tier": self.provider.tier.value,
            "logo_url": None,
            "versions": [
                version.version
                for version in self.provider.get_all_versions()
            ],
            "docs": [
                doc.get_api_outline()
                for doc in terrareg.provider_version_documentation_model.ProviderVersionDocumentation.get_by_provider_version(self)
            ]
        }

    def get_db_where(self, db, statement):
        """Filter DB query by where for current object."""
        raise NotImplementedError
        return statement.where(
            db.module_version.c.module_provider_id == self._module_provider.pk,
            db.module_version.c.version == self.version
        )

    def _create_db_row(self, git_tag: str, gpg_key: 'terrareg.models.GpgKey') -> None:
        """
        Insert into database, removing any existing duplicate versions.

        Returns boolean whether the new version should be published
        (depending on previous DB row (if exists) was published or if auto publish is enabled.
        """
        db = terrareg.database.Database.get()

        # Check if a pre-existing version is present in database
        if self._get_db_row():
            raise ReindexingExistingProviderVersionsIsProhibitedError(
                "The provider version already exists and re-indexing modules is disabled")

        with db.get_connection() as conn:
            # Insert new module into table
            insert_statement = db.provider_version.insert().values(
                provider_id=self._provider.pk,
                version=self.version,
                git_tag=git_tag,
                beta=self._extracted_beta_flag,
                gpg_key_id=gpg_key.pk
            )
            conn.execute(insert_statement)
