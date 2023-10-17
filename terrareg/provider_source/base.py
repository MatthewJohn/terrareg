
from typing import Dict, Union
import json

import terrareg.database

class BaseProviderSource:

    TYPE = None

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, Union[str, bool]]:
        """Validate user-provided config and generate configuration for database"""
        raise NotImplementedError

    @property
    def name(self) -> str:
        """Return name"""
        return self._name

    @property
    def api_name(self) -> str:
        """Return API name"""
        return self._get_db_row()["api_name"]

    @property
    def login_button_text(self):
        """Return login button text"""
        raise NotImplementedError

    @property
    def _config(self) -> Dict[str, Union[str, bool]]:
        """Return config for provider source"""
        return json.loads(terrareg.database.Database.decode_blob(self._get_db_row()['config']))

    def __init__(self, name: str):
        """Inialise member variables"""
        self._name = name
        self._cache_db_row: Union[None, Dict[str, Union[str, bool]]] = None

    def _get_db_row(self):
        """Return database row for module details."""
        if self._cache_db_row is None:
            db = terrareg.database.Database.get()
            select = db.provider_source.select(
            ).where(
                db.provider_source.c.name == self.name
            )
            with db.get_connection() as conn:
                res = conn.execute(select)
                self._cache_db_row = res.fetchone()

        return self._cache_db_row

    def update_repositories(self, access_token: str) -> None:
        """Refresh list of repositories"""
        raise NotImplementedError
