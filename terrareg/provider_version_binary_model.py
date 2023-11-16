
from glob import escape
import os
import re
from typing import Union, List

import sqlalchemy
from terrareg.errors import InvalidProviderBinaryArchitectureError, InvalidProviderBinaryNameError, InvalidProviderBinaryOperatingSystemError, ProviderVersionBinaryAlreadyExistsError

import terrareg.provider_version_model
import terrareg.provider_documentation_type
import terrareg.database
import terrareg.provider_binary_types


class ProviderVersionBinary:
    """Model for handling provider binaries"""

    @classmethod
    def create(cls,
               provider_version: 'terrareg.provider_version_model.ProviderVersion',
               name: str,
               checksum: str,
               content: bytes) -> Union[None, 'ProviderVersionBinary']:
        """Create provider version document"""
        # Extract OS and arch from name and ensure filename matches an expected type
        name_re = re.compile(
            # terraform-provider-jmon_2.1.1_linux_386.zip
            re.escape(f"{provider_version.provider.full_name}_{provider_version.version}_") + r"([0-9a-z]+)_([0-9a-z]+)\.zip"
        )

        if not (name_match := name_re.match(name)):
            raise InvalidProviderBinaryNameError("Provider binary is not a valid name")
        
        os_name = name_match.group(1)
        arch_name = name_match.group(2)

        try:
            os_type = terrareg.provider_binary_types.ProviderBinaryOperatingSystemType(os_name)
        except ValueError:
            raise InvalidProviderBinaryOperatingSystemError("Invalid operating system in provider binary")

        try:
            arch_type = terrareg.provider_binary_types.ProviderBinaryArchitectureType(arch_name)
        except ValueError:
            raise InvalidProviderBinaryArchitectureError("Invalid architecture in provider binary")


        # Check if provider binary already exists
        if cls.get(
                provider_version=provider_version,
                operating_system_type=os_type,
                architecture_type=arch_type):
            raise ProviderVersionBinaryAlreadyExistsError("Provider binary already exists")

        pk = cls._insert_db_row(
            provider_version=provider_version,
            operating_system_type=os_type,
            architecture_type=arch_type,
            name=name,
            checksum=checksum
        )

        # Store binary
        obj = cls(pk=pk)
        provider_version.create_data_directory()
        obj.create_local_binary(content=content)

        return obj

    @classmethod
    def _insert_db_row(cls,
                       provider_version: 'terrareg.provider_version_model.ProviderVersion',
                       operating_system_type: 'terrareg.provider_binary_types.ProviderBinaryOperatingSystemType',
                       architecture_type: 'terrareg.provider_binary_types.ProviderBinaryArchitectureType',
                       name: str,
                       checksum: str) -> int:
        """Insert database row for new binary"""
        db = terrareg.database.Database.get()
        insert = sqlalchemy.insert(db.provider_version_binary).values(
            provider_version_id=provider_version.pk,
            name=name,
            operating_system=operating_system_type,
            architecture=architecture_type,
            checksum=checksum
        )
        with db.get_connection() as conn:
            res = conn.execute(insert)
            return res.lastrowid

    @classmethod
    def get(cls,
            provider_version: 'terrareg.provider_version_model.ProviderVersion',
            operating_system_type: 'terrareg.provider_binary_types.ProviderBinaryOperatingSystemType',
            architecture_type: 'terrareg.provider_binary_types.ProviderBinaryArchitectureType') -> Union[None, 'ProviderVersionBinary']:
        """Obtain binary by provider version, OS and arch types"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version_binary.c.id
        ).select_from(
            db.provider_version_binary
        ).where(
            db.provider_version_binary.c.provider_version_id==provider_version.pk,
            db.provider_version_binary.c.operating_system==operating_system_type,
            db.provider_version_binary.c.architecture==architecture_type
        )

        with db.get_connection() as conn:
            row = conn.execute(select).first()

        if row:
            return cls(pk=row["id"])
        return None

    @classmethod
    def get_by_provider_version(cls, provider_version: 'terrareg.provider_version_model.ProviderVersion') -> List['ProviderVersionBinary']:
        """Obtain all binaries for provider version"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version_binary.c.id
        ).select_from(
            db.provider_version_binary
        ).where(
            db.provider_version_binary.c.provider_version_id==provider_version.pk,
        )
        with db.get_connection() as conn:
            rows = conn.execute(select).all()
        return [
            cls(pk=row["id"])
            for row in rows
        ]

    @property
    def name(self) -> str:
        """Return name of file"""
        return self._get_db_row()["name"]

    @property
    def architecture(self) -> 'terrareg.provider_binary_types.ProviderBinaryArchitectureType':
        """Retrun architecture"""
        return self._get_db_row()["architecture"]

    @property
    def operating_system(self) -> 'terrareg.provider_binary_types.ProviderBinaryOperatingSystemType':
        """Retrun operating system"""
        return self._get_db_row()["operating_system"]

    @property
    def checksum(self) -> str:
        """Return checksum of file"""
        return self._get_db_row()["checksum"]

    @property
    def local_file_path(self) -> str:
        """Return path of file on disk"""
        return os.path.join(self.provider_version.base_directory, self.name)

    @property
    def provider_version(self):
        """Return provider_version"""
        return terrareg.provider_version_model.ProviderVersion.get_by_pk(self._get_db_row()["provider_version_id"])

    def __init__(self, pk):
        """Store member variables"""
        self._pk = pk
        self._cache_db_row = None

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db.provider_version_binary.select().where(
                db.provider_version_binary.c.id == self._pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def create_local_binary(self, content: bytes):
        """Create local binary file"""
        with open(self.local_file_path, "wb") as fh:
            fh.write(content)

    def get_api_outline(self) -> dict:
        """Return API details"""
        gpg_key = self.provider_version.gpg_key
        return {
            "protocols": self.provider_version.protocols,
            "os": self.operating_system.value,
            "arch": self.architecture.value,
            "filename": self.name,
            "download_url": self.provider_version.provider.repository.provider_source.get_public_artifact_download_url(
                provider_version=self.provider_version,
                artifact_name=self.name
            ),
            "shasums_url": self.provider_version.provider.repository.provider_source.get_public_artifact_download_url(
                provider_version=self.provider_version,
                artifact_name=self.provider_version.checksum_file_name
            ),
            "shasums_signature_url": self.provider_version.provider.repository.provider_source.get_public_artifact_download_url(
                provider_version=self.provider_version,
                artifact_name=self.provider_version.checksum_signature_file_name
            ),
            "shasum": self.checksum,
            "signing_keys": {
                "gpg_public_keys": [
                    {
                        "key_id": gpg_key.key_id,
                        "ascii_armor": gpg_key.ascii_armor,
                        "trust_signature": "",
                        "source": "",
                        "source_url": None
                    }
                ]
            }
        }

