
from abc import abstractmethod, ABC
from typing import List, Optional
from flask import request

from .base_auth_method import BaseAuthMethod
import terrareg.api_key

class BaseApiKeyAuthMethod(ABC, BaseAuthMethod):
    """Base auth method for API key-based authentication"""

    key_type: Optional[terrareg.api_key.ApiKeyType]

    def __init__(self, matched_api_key: Optional[terrareg.api_key.ApiKey]=None):
        self._matched_api_key = matched_api_key

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    @classmethod
    def get_api_key_type(cls) -> terrareg.api_key.ApiKeyType:
        """Return API key type for class"""
        if cls.key_type is None:
            raise Exception("API key is not configured with a type")
        return cls.key_type

    @property
    def matched_api_key(self) -> Optional[terrareg.api_key.ApiKey]:
        """Return the DB-backed ApiKey that authenticated this request, or None for env-var keys."""
        return self._matched_api_key

    @abstractmethod
    @classmethod
    def get_static_keys(cls) -> List[str]:
        """Return API keys configured directly for this auth method."""
        ...

    @classmethod
    def check_auth_state(cls) -> bool:
        """Check if admin API key is provided"""
        return cls._check_api_key(cls.get_valid_keys())

    @classmethod
    def is_enabled(cls) -> bool:
        """Whether admin API key auth is configured."""
        return bool(cls.get_static_keys() or terrareg.api_key.ApiKey.has_active_keys(cls.api_key_type()))


    @classmethod
    def get_current_instance(cls):
        """Get an auth instance with any matched API key attached to it."""
        valid, matched_api_key = cls._check_api_key_with_matched_key(cls.get_valid_keys())
        if not valid:
            return None

        return cls(matched_api_key=matched_api_key)

    @classmethod
    def _check_api_key_with_matched_key(cls, valid_keys) -> tuple[bool, 'terrareg.api_key.ApiKey | None']:
        """Whether whether API key is valid"""
        if not isinstance(valid_keys, list):
            valid_keys = []

        # Obtain API key from request, ensuring that it is
        # not empty
        actual_key = request.headers.get('X-Terrareg-ApiKey', '')
        if not actual_key:
            return False, None
    
        # Iterate through API keys, ensuring 
        for valid_key in valid_keys:
            # Ensure the valid key is not empty:
            if actual_key and actual_key == valid_key:
                return True, None

        if cls.key_type is not None:
            stored_api_key = terrareg.api_key.ApiKey.verify_key(actual_key, cls.key_type)
            if stored_api_key is not None:
                stored_api_key.mark_used()
                return True, stored_api_key
        return False, None

    @classmethod
    def _check_api_key(cls, valid_keys):
        """Whether whether API key is valid"""
        return cls._check_api_key_with_matched_key(valid_keys)[0]

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # API keys can only access APIs that use different auth check methods
        return False
