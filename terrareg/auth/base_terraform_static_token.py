
import re
from flask import request

import terrareg.config
from .base_auth_method import BaseAuthMethod

class BaseTerraformStaticToken(BaseAuthMethod):
    """Base class for handling static terraform authentication tokens"""

    TOKEN_RE_MATCH = re.compile(r'Bearer (.*)')

    def is_built_in_admin(self):
        """Whether user is the built-in admin"""
        return False

    def is_admin(self):
        """Whether user is an admin"""
        return False

    def is_authenticated(self):
        """Whether user is authenticated"""
        return True

    @classmethod
    def is_enabled(cls):
        """Whether authentication method is enabled"""
        return bool(cls.get_valid_terraform_tokens())

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        return False

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        return False

    @classmethod
    def get_valid_terraform_tokens(cls):
        """Obtain list of valid tokens"""
        raise NotImplementedError

    @classmethod
    def get_terraform_auth_token(cls):
        """Get terraform auth token for current user"""
        auth_token_match = cls.TOKEN_RE_MATCH.match(request.headers.get('Authorization', ''))
        if auth_token_match and auth_token_match.group(1):
            return auth_token_match.group(1)
        return None

    @classmethod
    def check_auth_state(cls):
        """Check whether user is logged in using this method and return instance of object"""
        if (auth_token := cls.get_terraform_auth_token()) and auth_token in cls.get_valid_terraform_tokens():
            return True

        return False

    def check_namespace_access(self, permission_type, namespace):
        """Check level of access to namespace"""
        return False

    def get_all_namespace_permissions(self):
        """Return all permissions by namespace"""
        return {}

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        return False

    def can_access_terraform_api(self):
        """Terraform can only access those APIs used by Terraform"""
        return True
