
from typing import Union, List

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

    @classmethod
    def get_by_provider_version(cls, provider_version: 'terrareg.provider_version_model.ProviderVersion') -> List['ProviderVersionDocumentation']:
        """Obtain all provider documentation for provider version"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version_documentation.c.id
        ).select_from(
            db.provider_version_documentation
        ).where(
            db.provider_version_documentation.c.provider_version_id==provider_version.pk,
        )
        with db.get_connection() as conn:
            rows = conn.execute(select).all()
        return [
            cls(pk=row["id"])
            for row in rows
        ]

    @property
    def title(self) -> str:
        """Return title"""
        name = self._get_db_row()["name"]
        if name.endswith(".md"):
            return name[:-3]

    def __init__(self, pk):
        """Store member variables"""
        self._pk = pk
        self._cache_db_row = None

    def _get_db_row(self):
        """Get object from database"""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db.provider_version_documentation.select().where(
                db.provider_version_documentation.c.id == self._pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()
        return self._cache_db_row

    def get_api_outline(self) -> dict:
        """Return API details"""
        db_row = self._get_db_row()
        return {
            "id": str(self._pk),
            "title": self.title,
            "path": db_row["filename"],
            # @TODO Generate Slug in extraction
            "slug": db_row["name"],
            "category": db_row["documentation_type"].value,
            "subcategory": "",
            "language": "hcl"
        }
