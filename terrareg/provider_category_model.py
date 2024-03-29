
import json
import re
from typing import Union, List, Dict

import sqlalchemy

import terrareg.database
import terrareg.config
from terrareg.errors import InvalidProviderCategoryConfigError



class ProviderCategory:
    """Model for provider category"""

    @classmethod
    def get_by_pk(cls, pk: int) -> Union[None, 'ProviderCategory']:
        """Return provider category by pk"""
        obj = cls(pk=pk)
        if obj.exists:
            return obj
        return None

    @property
    def name(self) -> str:
        """Return name"""
        return self._get_db_row()["name"]

    @property
    def slug(self) -> str:
        """Return slug"""
        return self._get_db_row()["slug"]

    @property
    def pk(self) -> int:
        """Return slug"""
        return self._pk

    @property
    def exists(self):
        """Return whether the provider category exists"""
        return bool(self._get_db_row())

    @property
    def user_selectable(self):
        """Return whether provider category is user selectable"""
        return self._get_db_row()["user_selectable"]

    def __eq__(self, __o):
        """Check if two provider categories are the same"""
        if isinstance(__o, self.__class__):
            return self.pk == __o.pk
        return super(ProviderCategory, self).__eq__(__o)

    def __init__(self, pk: int):
        """Initialise member variables"""
        self._pk = pk
        self._cache_db_row: Union[None, Dict[str, Union[str, bool]]] = None

    def _get_db_row(self):
        """Return database row for module details."""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db.provider_category.select(
            ).where(
                db.provider_category.c.id == self._pk
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row

    def get_v2_include(self):
        """Return API response from v2 include"""
        return {
            "type": "categories",
            "id": str(self.pk),
            "attributes": {
                "name": self.name,
                "slug": self.slug,
                "user-selectable": self.user_selectable,
            },
            "links": {
                "self": f"/v2/categories/{self.pk}"
            }
        }


class ProviderCategoryFactory:

    _INSTANCE: Union[None, 'ProviderCategoryFactory'] = None

    @classmethod
    def get(cls) -> 'ProviderCategoryFactory':
        """Get instance of Provider Source Factory"""
        if cls._INSTANCE is None:
            cls._INSTANCE = cls()
        return cls._INSTANCE

    def get_provider_category_by_slug(self, slug: str) -> Union[ProviderCategory, None]:
        """Obtain instance of provider category by API name"""
        # Obtain row from DB, to determine provider category type
        database = terrareg.database.Database.get()
        select = sqlalchemy.select(
            database.provider_category.c.id
        ).select_from(
            database.provider_category
        ).where(
            database.provider_category.c.slug==slug
        )
        with database.get_connection() as conn:
            res = conn.execute(select).first()

        # If there are no matching rows, return None
        if res is None:
            return None

        # Return instance of provider category class
        return ProviderCategory(pk=res['id'])

    def get_provider_category_by_pk(self, pk: int) -> Union[None, ProviderCategory]:
        """Return instance of provider category by pk"""
        instance = ProviderCategory(pk=pk)
        if instance.exists:
            return instance
        return None

    def get_all_provider_categories(self) -> List[ProviderCategory]:
        """Return all provider categories"""
        database = terrareg.database.Database.get()
        select = sqlalchemy.select(
            database.provider_category.c.id
        ).select_from(
            database.provider_category
        )
        with database.get_connection() as conn:
            res = conn.execute(select).all()
        return [
            ProviderCategory(pk=row['id'])
            for row in res
        ]

    def name_to_slug(self, name: str):
        """Convert name to slug"""
        name = name.lower()
        name = re.sub(r'[^a-z0-9\-]', '-', name)
        name = re.sub(r'--+', '-', name)
        if name.startswith('-'):
            name = name[1:]
        if name.endswith('-'):
            name = name[:-1]
        return name

    def initialise_from_config(self) -> None:
        """Load provider categories from config into database."""
        try:
            provider_category_configs = json.loads(terrareg.config.Config().PROVIDER_CATEGORIES)
        except:
            raise InvalidProviderCategoryConfigError("Provider categories configuration is not valid JSON")
        db = terrareg.database.Database.get()

        with db.start_transaction():

            for provider_category_config in provider_category_configs:
                # Validate provider config
                for attr in ['name', 'id']:
                    if attr not in provider_category_config:
                        raise InvalidProviderCategoryConfigError(
                            'Provider Category config does not contain required attribute: {}'.format(attr))

                pk = provider_category_config.get("id")
                try:
                    pk = int(pk)
                except ValueError:
                    raise InvalidProviderCategoryConfigError("ID for provider is invalid - must be a valid integer")

                name = provider_category_config.get("name")
                if not isinstance(name, str):
                    raise InvalidProviderCategoryConfigError("Provider category name is invalid - must be a valid string")

                user_selectable = provider_category_config.get("user-selectable", True)
                if not isinstance(user_selectable, bool):
                    raise InvalidProviderCategoryConfigError("Provider Category config 'user-selectable' field must be a boolean value")

                slug = self.name_to_slug(provider_category_config.get("slug", name))
                if not isinstance(slug, str):
                    raise InvalidProviderCategoryConfigError("Provider category slug is invalid - must be a valid string")

                # Check if git provider exists in DB
                existing_provider_category = self.get_provider_category_by_pk(pk=pk)
                fields = {
                    'name': name,
                    'slug': slug,
                    'user_selectable': user_selectable
                }
                if existing_provider_category:
                    # Update existing row
                    upsert = db.provider_category.update().where(
                        db.provider_category.c.id==pk
                    ).values(
                        **fields
                    )
                else:
                    upsert = db.provider_category.insert().values(
                        id=pk,
                        **fields
                    )
                with db.get_connection() as conn:
                    conn.execute(upsert)
