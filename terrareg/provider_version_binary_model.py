
from glob import escape
import os
import re
from typing import Union, List

import sqlalchemy
from terrareg.errors import InvalidProviderBinaryArchitectureError, InvalidProviderBinaryNameError, InvalidProviderBinaryOperatingSystemError

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
            return None

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
    def get_by_provider_version(cls, provider_version: 'terrareg.provider_version_model.ProviderVersion') -> List['ProviderVersionDocumentation']:
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
        db_row = self._get_db_row()
        return {
            "protocols": ["4.0", "5.1"],
            "os": self.operating_system.value,
            "arch": self.architecture.value,
            "filename": self.name,
            "download_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_linux_amd64.zip",
            "shasums_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS",
            "shasums_signature_url": "https://releases.hashicorp.com/terraform-provider-random/2.0.0/terraform-provider-random_2.0.0_SHA256SUMS.sig",
            "shasum": self.checksum,
            "signing_keys": {
                "gpg_public_keys": [
                {
                    "key_id": "51852D87348FFC4C",
                    "ascii_armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----\nVersion: GnuPG v1\n\nmQENBFMORM0BCADBRyKO1MhCirazOSVwcfTr1xUxjPvfxD3hjUwHtjsOy/bT6p9f\nW2mRPfwnq2JB5As+paL3UGDsSRDnK9KAxQb0NNF4+eVhr/EJ18s3wwXXDMjpIifq\nfIm2WyH3G+aRLTLPIpscUNKDyxFOUbsmgXAmJ46Re1fn8uKxKRHbfa39aeuEYWFA\n3drdL1WoUngvED7f+RnKBK2G6ZEpO+LDovQk19xGjiMTtPJrjMjZJ3QXqPvx5wca\nKSZLr4lMTuoTI/ZXyZy5bD4tShiZz6KcyX27cD70q2iRcEZ0poLKHyEIDAi3TM5k\nSwbbWBFd5RNPOR0qzrb/0p9ksKK48IIfH2FvABEBAAG0K0hhc2hpQ29ycCBTZWN1\ncml0eSA8c2VjdXJpdHlAaGFzaGljb3JwLmNvbT6JATgEEwECACIFAlMORM0CGwMG\nCwkIBwMCBhUIAgkKCwQWAgMBAh4BAheAAAoJEFGFLYc0j/xMyWIIAIPhcVqiQ59n\nJc07gjUX0SWBJAxEG1lKxfzS4Xp+57h2xxTpdotGQ1fZwsihaIqow337YHQI3q0i\nSqV534Ms+j/tU7X8sq11xFJIeEVG8PASRCwmryUwghFKPlHETQ8jJ+Y8+1asRydi\npsP3B/5Mjhqv/uOK+Vy3zAyIpyDOMtIpOVfjSpCplVRdtSTFWBu9Em7j5I2HMn1w\nsJZnJgXKpybpibGiiTtmnFLOwibmprSu04rsnP4ncdC2XRD4wIjoyA+4PKgX3sCO\nklEzKryWYBmLkJOMDdo52LttP3279s7XrkLEE7ia0fXa2c12EQ0f0DQ1tGUvyVEW\nWmJVccm5bq25AQ0EUw5EzQEIANaPUY04/g7AmYkOMjaCZ6iTp9hB5Rsj/4ee/ln9\nwArzRO9+3eejLWh53FoN1rO+su7tiXJA5YAzVy6tuolrqjM8DBztPxdLBbEi4V+j\n2tK0dATdBQBHEh3OJApO2UBtcjaZBT31zrG9K55D+CrcgIVEHAKY8Cb4kLBkb5wM\nskn+DrASKU0BNIV1qRsxfiUdQHZfSqtp004nrql1lbFMLFEuiY8FZrkkQ9qduixo\nmTT6f34/oiY+Jam3zCK7RDN/OjuWheIPGj/Qbx9JuNiwgX6yRj7OE1tjUx6d8g9y\n0H1fmLJbb3WZZbuuGFnK6qrE3bGeY8+AWaJAZ37wpWh1p0cAEQEAAYkBHwQYAQIA\nCQUCUw5EzQIbDAAKCRBRhS2HNI/8TJntCAClU7TOO/X053eKF1jqNW4A1qpxctVc\nz8eTcY8Om5O4f6a/rfxfNFKn9Qyja/OG1xWNobETy7MiMXYjaa8uUx5iFy6kMVaP\n0BXJ59NLZjMARGw6lVTYDTIvzqqqwLxgliSDfSnqUhubGwvykANPO+93BBx89MRG\nunNoYGXtPlhNFrAsB1VR8+EyKLv2HQtGCPSFBhrjuzH3gxGibNDDdFQLxxuJWepJ\nEK1UbTS4ms0NgZ2Uknqn1WRU1Ki7rE4sTy68iZtWpKQXZEJa0IGnuI2sSINGcXCJ\noEIgXTMyCILo34Fa/C6VCm2WBgz9zZO8/rHIiQm1J5zqz0DrDwKBUM9C\n=LYpS\n-----END PGP PUBLIC KEY BLOCK-----",
                    "trust_signature": "",
                    "source": "HashiCorp",
                    "source_url": "https://www.hashicorp.com/security.html"
                }
                ]
            }
        }

