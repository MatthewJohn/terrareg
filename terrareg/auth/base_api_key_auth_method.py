
from flask import request

from .base_auth_method import BaseAuthMethod
import terrareg.models

class BaseApiKeyAuthMethod(BaseAuthMethod):
    """Base auth method for API key-based authentication"""

    key_type = None

    def __init__(self, matched_api_key=None):
        self._matched_api_key = matched_api_key

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    @property
    def matched_api_key(self):
        """Return the DB-backed ApiKey that authenticated this request, or None for env-var keys."""
        return self._matched_api_key

    @classmethod
    def get_valid_keys(cls):
        """Return API keys configured directly for this auth method."""
        return []

    @classmethod
    def get_current_instance(cls):
        """Get an auth instance with any matched API key attached to it."""
        valid, matched_api_key = cls._check_api_key_with_matched_key(cls.get_valid_keys())
        if not valid:
            return None

        return cls(matched_api_key=matched_api_key)

    @classmethod
    def _check_api_key_with_matched_key(cls, valid_keys) -> tuple[bool, terrareg.models.ApiKey | None]:
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
            stored_api_key = terrareg.models.ApiKey.verify_key(actual_key, cls.key_type)
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
