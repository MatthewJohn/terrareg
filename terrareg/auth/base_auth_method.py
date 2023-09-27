

class BaseAuthMethod:
    """Base auth method"""

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
        raise NotImplementedError

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        raise NotImplementedError

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        return False

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        return False

    @classmethod
    def get_current_instance(cls):
        """Get instance of auth method, if user is authenticated"""
        return cls() if cls.check_auth_state() else None

    @classmethod
    def check_auth_state(cls):
        """Check whether user is logged in using this method and return instance of object"""
        raise NotImplementedError

    def check_namespace_access(self, permission_type, namespace):
        """Check level of access to namespace"""
        raise NotImplementedError

    def get_all_namespace_permissions(self):
        """Return all permissions by namespace"""
        return {}

    def get_username(self):
        """Get username of current user"""
        raise NotImplementedError

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        raise NotImplementedError

    def can_access_terraform_api(self):
        """Whether the user can access APIs used by terraform"""
        # Default to using 'read' API access
        return self.can_access_read_api()

    def should_record_terraform_analytics(self):
        """Whether Terraform downloads by the user should be recorded"""
        # Default to True for all users
        return True

    def get_terraform_auth_token(self):
        """Get terraform auth token from request"""
        # Default to return None
        return None
