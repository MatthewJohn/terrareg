
from typing import Dict

from terrareg.errors import InvalidProviderSourceConfigError

from .base import BaseProviderSource
import terrareg.provider_source_type

class GithubProviderSource(BaseProviderSource):

    TYPE = terrareg.provider_source_type.ProviderSourceType.GITHUB

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, str]:
        """Generate DB config from config"""
        db_config = {}
        for required_attr in ["base_url", "api_url", "client_id", "client_secret"]:
            if not (val := config.get(required_attr)) or not isinstance(val, str):
                raise InvalidProviderSourceConfigError(f"Missing required Github provider source config: {required_attr}")

            db_config[required_attr] = val
        return db_config
