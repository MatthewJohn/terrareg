
from flask import request

from .base_auth_method import BaseAuthMethod


class BaseApiKeyAuthMethod(BaseAuthMethod):
    """Base auth method for API key-based authentication"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    @classmethod
    def _check_api_key(cls, valid_keys):
        """Whether whether API key is valid"""
        if not isinstance(valid_keys, list):
            return False

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
        return False

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # API keys can only access APIs that use different auth check methods
        return False
