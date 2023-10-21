
from typing import Union

import sqlalchemy

import terrareg.provider_version_model
import terrareg.provider_documentation_type
import terrareg.database


class ProviderVersionDocumentation:
    """Interface for creating and managing provider version documentation files"""

    @classmethod
    def create(cls,
               provider_version: 'terrareg.provider_version_model.ProviderVersion',
               documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
               name: str,
               filename: str,
               content: str) -> Union[None, 'ProviderVersionDocumentation']:
        """Create provider version document"""
        # Check if document already exists
        if cls.get(provider_version=provider_version, documentation_type=documentation_type, name=name):
            return None

        pk = cls._insert_db_row(
            provider_version=provider_version,
            documentation_type=documentation_type,
            name=name,
            filename=filename,
            content=content
        )

        return cls(pk=pk)

    @classmethod
    def _insert_db_row(cls,
                       provider_version: 'terrareg.provider_version_model.ProviderVersion',
                       documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
                       name: str,
                       filename: str,
                       content: str) -> int:
        """Insert database row for new document"""
        db = terrareg.database.Database.get()
        insert = sqlalchemy.insert(db.provider_version_documentation).values(
            provider_version_id=provider_version.pk,
            documentation_type=documentation_type,
            name=name,
            filename=filename,
            content=db.encode_blob(content)
        )
        with db.get_connection() as conn:
            res = conn.execute(insert)
            return res.lastrowid

    @classmethod
    def get(cls,
            provider_version: 'terrareg.provider_version_model.ProviderVersion',
            documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
            name: str) -> Union[None, 'ProviderVersionDocumentation']:
        """Obtain document by provider version, type and name"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version_documentation.c.id
        ).select_from(
            db.provider_version_documentation
        ).where(
            db.provider_version_documentation.c.provider_version_id==provider_version.pk,
            db.provider_version_documentation.c.documentation_type==documentation_type,
            db.provider_version_documentation.c.name==name
        )

        with db.get_connection() as conn:
            row = conn.execute(select).first()

        if row:
            return cls(pk=row["id"])
        return None

    def __init__(self, pk):
        """Store member variables"""
        self._pk = pk
