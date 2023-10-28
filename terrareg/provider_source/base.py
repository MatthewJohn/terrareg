
from typing import Dict, Union, List, Tuple
import json

import terrareg.database
import terrareg.provider_source.repository_release_metadata
import terrareg.repository_model
import terrareg.provider_model
import terrareg.provider_version_model


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
        """Initialise member variables"""
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

    def get_user_access_token(self, code: str) -> Union[None, str]:
        """Obtain user access token from code"""
        raise NotImplementedError

    def update_repositories(self, access_token: str) -> None:
        """Refresh list of repositories"""
        raise NotImplementedError

    def get_new_releases(self, provider: 'terrareg.provider_model.Provider') -> List['terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata']:
        """Obtain all repository releases that aren't associated with a pre-existing release"""
        raise NotImplementedError

    def get_release_artifact(self,
                             provider: 'terrareg.provider_model.Provider',
                             artifact_metadata: 'terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata',
                             release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata'):
        """Return release artifact file content"""
        raise NotImplementedError

    def get_release_archive(self,
                            provider: 'terrareg.provider_model.Provider',
                            release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> Tuple[bytes, Union[None, str]]:
        """Obtain release archive"""
        raise NotImplementedError

    def get_public_source_url(self, repository: 'terrareg.repository_model.Repository'):
        """Return public URL for source"""
        raise NotImplementedError

    def get_public_artifact_download_url(self,
                                         provider_version: 'terrareg.provider_version_model.ProviderVersion',
                                         artifact_name: str):
        """Return public URL for source"""
        raise NotImplementedError
