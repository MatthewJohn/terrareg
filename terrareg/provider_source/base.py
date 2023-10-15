
from typing import Dict

class BaseProviderSource:

    TYPE = None

    @classmethod
    def generate_db_config_from_source_config(cls, config: Dict[str, str]) -> Dict[str, str]:
        """Validate user-provided config and generate configuration for database"""
        raise NotImplementedError

    def __init__(self, name: str):
        """Inialise member variables"""
        self._name = name
