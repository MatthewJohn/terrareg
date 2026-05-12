
from flask import g, request

from .base_auth_method import BaseAuthMethod
import terrareg.models

_MATCHED_API_KEY_G_KEY = '_matched_api_key'


class BaseApiKeyAuthMethod(BaseAuthMethod):
    """Base auth method for API key-based authentication"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    @property
    def matched_api_key(self):
        """Return the DB-backed ApiKey that authenticated this request, or None for env-var keys."""
        return g.get(_MATCHED_API_KEY_G_KEY, None)

    @classmethod
    def _check_api_key(cls, valid_keys, key_type=None):
        """Whether whether API key is valid"""
        if not isinstance(valid_keys, list):
            valid_keys = []

        # Obtain API key from request, ensuring that it is
        # not empty
        actual_key = request.headers.get('X-Terrareg-ApiKey', '')
        if not actual_key:
            return False
    
        # Iterate through API keys, ensuring 
        for valid_key in valid_keys:
            # Ensure the valid key is not empty:
            if actual_key and actual_key == valid_key:
                return True

        if key_type is not None:
            stored_api_key = terrareg.models.ApiKey.verify_key(actual_key, key_type)
            if stored_api_key is not None:
                stored_api_key.mark_used()
                setattr(g, _MATCHED_API_KEY_G_KEY, stored_api_key)
                return True
        return False

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # API keys can only access APIs that use different auth check methods
        return False
