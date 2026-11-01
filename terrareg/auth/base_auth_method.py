
from abc import abstractmethod
from typing import Optional, Self, Dict

import flask

import terrareg.models
import terrareg.user_group_namespace_permission_type


class BaseAuthMethod:
    """Base auth method"""

    def is_built_in_admin(self) -> bool:
        """Whether user is the built-in admin"""
        return False

    def is_admin(self) -> bool:
        """Whether user is an admin"""
        return False

    def is_authenticated(self) -> bool:
        """Whether user is authenticated"""
        return True

    @classmethod
    @abstractmethod
    def is_enabled(cls) -> bool:
        """Whether authentication method is enabled"""
        raise NotImplementedError

    @property
    @abstractmethod
    def requires_csrf_tokens(self) -> bool:
        """Whether auth type requires CSRF tokens"""
        ...

    def can_publish_module_version(self, namespace: Optional[str]) -> bool:
        """Whether user can publish module version within a namespace."""
        return False

    def can_upload_module_version(self, namespace: Optional[str]) -> bool:
        """Whether user can upload/index module version within a namespace."""
        return False

    @classmethod
    def get_current_instance(cls) -> Optional[Self]:
        """Get instance of auth method, if user is authenticated"""
        return cls() if cls.check_auth_state() else None

    @classmethod
    @abstractmethod
    def check_auth_state(cls) -> bool:
        """Check whether user is logged in using this method and return instance of object"""
        ...

    @abstractmethod
    def check_namespace_access(self, permission_type: terrareg.user_group_namespace_permission_type.UserGroupNamespacePermissionType, namespace: Optional[str]) -> bool:
        """Check level of access to namespace"""
        ...

    def get_all_namespace_permissions(self) -> Dict[str, bool]:
        """Return all permissions by namespace"""
        return {}

    @abstractmethod
    def get_username(self) -> str:
        """Get username of current user"""
        ...

    @abstractmethod
    def can_access_read_api(self) -> bool:
        """Whether the user can access 'read' APIs"""
        ...

    def can_access_terraform_api(self) -> bool:
        """Whether the user can access APIs used by terraform"""
        # Default to using 'read' API access
        return self.can_access_read_api()

    def should_record_terraform_analytics(self) -> bool:
        """Whether Terraform downloads by the user should be recorded"""
        # Default to True for all users
        return True

    def get_terraform_auth_token(self) -> Optional[str]:
        """Get terraform auth token from request"""
        # Default to return None
        return None
