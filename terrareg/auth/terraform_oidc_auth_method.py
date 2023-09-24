
from flask import request
import pyop.exceptions

from .base_auth_method import BaseAuthMethod
import terrareg.terraform_idp


class TerraformOidcAuthMethod(BaseAuthMethod):
    """"Auth method for Terraform OIDC"""

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
        return terrareg.terraform_idp.TerraformIdp.get().is_enabled

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
    def check_auth_state(cls):
        """Check whether user is logged in using this method and return instance of object"""
        if 'Authorization' in request.headers:
            # Check header with OpenIDC
            try:
                res = terrareg.terraform_idp.TerraformIdp.get().provider.handle_userinfo_request(request.data, request.headers)
                return True
            except (pyop.exceptions.InvalidAccessToken, pyop.exceptions.BearerTokenError):
                return False

        return False

    def check_namespace_access(self, permission_type, namespace):
        """Check level of access to namespace"""
        return False

    def get_all_namespace_permissions(self):
        """Return all permissions by namespace"""
        return {}

    def get_username(self):
        """Get username of current user"""
        return "Terraform CLI User"

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        return False

    def can_access_terraform_api(self):
        """Terraform can only access those APIs used by Terraform"""
        return True
