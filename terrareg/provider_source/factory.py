
import json
from typing import Dict, Union, Type, List

import sqlalchemy

import terrareg.config
import terrareg.database
from terrareg.errors import InvalidProviderSourceConfigError
import terrareg.provider_source_type
import terrareg.provider_source


class ProviderSourceFactory:
    """Factory class for generating and getting provider sources"""

    _CLASS_MAPPING: Union[None, Dict['terrareg.provider_source_type.ProviderSourceType', Type['terrareg.provider_source.BaseProviderSource']]] = None
    _INSTANCE: Union[None, 'ProviderSourceFactory'] = None

    @classmethod
    def get(cls) -> 'ProviderSourceFactory':
        """Get instance of Provider Source Factory"""
        if cls._INSTANCE is None:
            cls._INSTANCE = cls()
        return cls._INSTANCE

    def get_provider_classes(self) -> Dict[str, Type['terrareg.provider_source.BaseProviderSource']]:
        """Return all provider classes"""
        if self._CLASS_MAPPING is None:
            self._CLASS_MAPPING = {
                provider_source_class.TYPE: provider_source_class
                for provider_source_class in terrareg.provider_source.BaseProviderSource.__subclasses__()
            }
        return self._CLASS_MAPPING

    def get_provider_source_class_by_type(self, type_: 'terrareg.provider_source_type.ProviderSourceType') -> Union[Type['terrareg.provider_source.BaseProviderSource'], None]:
        """Obtain provider source class by name"""
        # Ensure type if valid
        if not type_:
            return None
        return self.get_provider_classes().get(type_)

    def get_provider_source_by_name(self, name: str) -> Union['terrareg.provider_source.BaseProviderSource', None]:
        """Obtain instance of provider source by name"""
        # Obtain row from DB, to determine provider source type
        database = terrareg.database.Database.get()
        select = sqlalchemy.select(
            database.provider_source.c.name,
            database.provider_source.c.provider_source_type
        ).select_from(
            database.provider_source
        ).where(
            database.provider_source.c.name==name
        )
        with database.get_connection() as conn:
            res = conn.execute(select).first()

        # If there are no matching rows, return None
        if res is None:
            return None

        # Obtain class of provider source
        class_ = self.get_provider_source_class_by_type(res['provider_source_type'])
        if class_ is None:
            return None

        # Return instance of provider source class
        return class_(name=res['name'])

    def get_provider_source_by_api_name(self, api_name: str) -> Union['terrareg.provider_source.BaseProviderSource', None]:
        """Obtain instance of provider source by API name"""
        # Obtain row from DB, to determine provider source type
        database = terrareg.database.Database.get()
        select = sqlalchemy.select(
            database.provider_source.c.name,
            database.provider_source.c.provider_source_type
        ).select_from(
            database.provider_source
        ).where(
            database.provider_source.c.api_name==api_name
        )
        with database.get_connection() as conn:
            res = conn.execute(select).first()

        # If there are no matching rows, return None
        if res is None:
            return None

        # Obtain class of provider source
        class_ = self.get_provider_source_class_by_type(res['provider_source_type'])
        if class_ is None:
            return None

        # Return instance of provider source class
        return class_(name=res['name'])

    def get_all_provider_sources(self) -> List['terrareg.provider_source.BaseProviderSource']:
        """Return all provider sources"""
        database = terrareg.database.Database.get()
        select = sqlalchemy.select(
            database.provider_source.c.name,
            database.provider_source.c.provider_source_type
        ).select_from(
            database.provider_source
        )
        with database.get_connection() as conn:
            res = conn.execute(select).all()
        return [
            self.get_provider_source_class_by_type(row['provider_source_type'])(name=row['name'])
            for row in res
        ]

    def initialise_from_config(self) -> None:
        """Load provider sources from config into database."""
        provider_source_configs = json.loads(terrareg.config.Config().PROVIDER_SOURCES)
        db = terrareg.database.Database.get()

        names = []
        for provider_source_config in provider_source_configs:
            # Validate provider config
            for attr in ['name', 'type']:
                if attr not in provider_source_config:
                    raise InvalidProviderSourceConfigError(
                        'Git provider config does not contain required attribute: {}'.format(attr))

            # Check name validity
            name: str = provider_source_config.get("name")
            if type(name) is not str or not name:
                raise InvalidProviderSourceConfigError("Provider source name is empty")
            if name.lower() in names:
                raise InvalidProviderSourceConfigError(f"Duplicate Provider Source name found: {name}")
            names.append(name.lower())

            # Obtain type of provider source
            type_name = provider_source_config.get("type")
            try:
                type_ = terrareg.provider_source_type.ProviderSourceType(type_name)
            except ValueError:
                valid_types_string = ", ".join([
                    type_itx.value
                    for type_itx in terrareg.provider_source_type.ProviderSourceType
                ])
                raise InvalidProviderSourceConfigError(f"Invalid provider source type. Valid types: {valid_types_string}")

            provider_source_class = self.get_provider_source_class_by_type(type_)
            if not provider_source_class:
                raise Exception(f'Internal Exception, could not find class for {type_}')

            provider_db_config = provider_source_class.generate_db_config_from_source_config(provider_source_config)

            # Check if git provider exists in DB
            existing_provider_source = self.get_provider_source_by_name(name=name)
            fields = {
                'provider_source_type': type_,
                'config': terrareg.database.Database.encode_blob(json.dumps(provider_db_config))
            }
            if existing_provider_source:
                # Update existing row
                upsert = db.provider_source.update().where(
                    db.provider_source.c.name==name
                ).values(
                    **fields
                )
            else:
                upsert = db.provider_source.insert().values(
                    name=name,
                    api_name=name.lower(),
                    **fields
                )
            with db.get_connection() as conn:
                conn.execute(upsert)
