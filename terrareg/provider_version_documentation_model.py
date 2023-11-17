
from typing import Union, List

import sqlalchemy

import terrareg.provider_version_model
import terrareg.provider_documentation_type
import terrareg.database
import terrareg.utils


class ProviderVersionDocumentation:
    """Interface for creating and managing provider version documentation files"""

    @classmethod
    def generate_slug_from_name(cls, name: str) -> str:
        """Generate slug from name"""
        for extension_to_remove in ['.md', '.markdown', '.html']:
            if name.endswith(extension_to_remove):
                name = name[:(0-len(extension_to_remove))]
        
        return name

    @classmethod
    def create(cls,
               provider_version: 'terrareg.provider_version_model.ProviderVersion',
               documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
               name: str,
               title: Union[str, None],
               description: Union[str, None],
               filename: str,
               language: str,
               subcategory: Union[str, None],
               content: str) -> Union[None, 'ProviderVersionDocumentation']:
        """Create provider version document"""
        # Generate slug from name
        slug = cls.generate_slug_from_name(name=name)

        # Check if document already exists
        if cls.get(provider_version=provider_version, language=language, documentation_type=documentation_type, slug=slug):
            return None

        pk = cls._insert_db_row(
            provider_version=provider_version,
            documentation_type=documentation_type,
            name=name,
            slug=slug,
            title=title,
            description=description,
            language=language,
            subcategory=subcategory,
            filename=filename,
            content=content
        )

        return cls(pk=pk)

    @classmethod
    def _insert_db_row(cls,
                       provider_version: 'terrareg.provider_version_model.ProviderVersion',
                       documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
                       name: str,
                       slug: str,
                       title: Union[str, None],
                       description: Union[str, None],
                       language: str,
                       subcategory: Union[str, None],
                       filename: str,
                       content: str) -> int:
        """Insert database row for new document"""
        db = terrareg.database.Database.get()
        insert = sqlalchemy.insert(db.provider_version_documentation).values(
            provider_version_id=provider_version.pk,
            documentation_type=documentation_type,
            name=name,
            slug=slug,
            title=title,
            description=db.encode_blob(description),
            language=language,
            subcategory=subcategory,
            filename=filename,
            content=db.encode_blob(content)
        )
        with db.get_connection() as conn:
            res = conn.execute(insert)
            return res.lastrowid

    @classmethod
    def get_by_pk(cls, pk: int):
        """Return instance of Provider Version documentation by PK"""
        obj = cls(pk=pk)
        if obj.exists:
            return obj
        return None

    @classmethod
    def get(cls,
            provider_version: 'terrareg.provider_version_model.ProviderVersion',
            documentation_type: 'terrareg.provider_documentation_type.ProviderDocumentationType',
            slug: str,
            language: str) -> Union[None, 'ProviderVersionDocumentation']:
        """Obtain document by provider version, type and slug"""
        db = terrareg.database.Database.get()
        select = sqlalchemy.select(
            db.provider_version_documentation.c.id
        ).select_from(
            db.provider_version_documentation
        ).where(
            db.provider_version_documentation.c.provider_version_id==provider_version.pk,
            db.provider_version_documentation.c.documentation_type==documentation_type,
            db.provider_version_documentation.c.language==language,
            db.provider_version_documentation.c.slug==slug
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

    @classmethod
    def search(cls, provider_version: 'terrareg.provider_version_model.ProviderVersion',
               category: 'terrareg.provider_documentation_type.ProviderDocumentationType',
               language: str,
               slug: str) -> List['ProviderVersionDocumentation']:
        """Search for provider version documentation using query filters"""
        db = terrareg.database.Database.get()

        select = sqlalchemy.select(
            db.provider_version_documentation.c.id
        ).select_from(
            db.provider_version_documentation
        ).where(
            db.provider_version_documentation.c.provider_version_id==provider_version.pk,
            db.provider_version_documentation.c.documentation_type==category,
            db.provider_version_documentation.c.slug==slug,
            db.provider_version_documentation.c.language==language,
        )
        with db.get_connection() as conn:
            rows = conn.execute(select).all()
        return [
            cls(pk=row["id"])
            for row in rows
        ]

    @property
    def exists(self):
        """Return whether document exists"""
        return bool(self._get_db_row())

    @property
    def title(self) -> str:
        """Return title"""
        # Attempt to return title
        if title := self._get_db_row()["title"]:
            return title

        # Fallback to returning name
        name = self._get_db_row()["name"]
        if name.endswith(".md"):
            name = name[:-3]
        return name

    @property
    def pk(self) -> int:
        """Return PK"""
        return self._pk

    @property
    def category(self) -> 'terrareg.provider_documentation_type.ProviderDocumentationType':
        """Return category"""
        return self._get_db_row()["documentation_type"]

    @property
    def language(self) -> str:
        """Return language"""
        return self._get_db_row()["language"]

    @property
    def filename(self) -> str:
        """Return filename"""
        return self._get_db_row()["filename"]

    @property
    def slug(self) -> str:
        """Return slug"""
        return self._get_db_row()["slug"]

    @property
    def subcategory(self) -> str:
        """Return subcategory"""
        return self._get_db_row()["subcategory"]

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
        return {
            "id": str(self._pk),
            "title": self.title,
            "path": self.filename,
            "slug": self.slug,
            "category": self.category.value,
            "subcategory": self.subcategory,
            "language": self.language
        }

    def get_v2_api_outline(self) -> dict:
        """Return API outline for v2 API"""
        return {
            "type": "provider-docs",
            "id": str(self.pk),
            "attributes": {
                "category": self.category.value,
                "language": self.language,
                "path": self.filename,
                "slug": self.slug,
                "subcategory": self.subcategory,
                "title": self.title,
                "truncated": False
            },
            "links": {
                "self": f"/v2/provider-docs/{self.pk}"
            }
        }

    def get_v2_api_details(self, html: bool=False):
        """Return V2 spec API details"""
        response = self.get_v2_api_outline()
        response["attributes"]["content"] = self.get_content(html=html)
        return response

    def get_content(self, html=False):
        """Return content of documentation"""
        content = terrareg.database.Database.decode_blob(self._get_db_row()["content"])
        if html:
            content = terrareg.utils.convert_markdown_to_html(file_name=self.filename, markdown_html=content)
            content = terrareg.utils.sanitise_html_content(content, allow_markdown_html=True)
        return content
