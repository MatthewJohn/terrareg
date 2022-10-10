
import datetime

import sqlalchemy
from flask import g, request, session

import terrareg.config
from terrareg.database import Database
from terrareg.models import Namespace, Session
import terrareg.openid_connect
import terrareg.saml
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType


class AuthFactory:
    """
    Factory for obtaining current user authentication method
    and authentication providing wrappers
    """

    FLASK_GLOBALS_AUTH_KEY = 'user_auth'

    def get_current_auth_method(self):
        """Obtain user's current login state"""
        # Check if current authenticate type has been determined
        if (current_auth_method := g.get(AuthFactory.FLASK_GLOBALS_AUTH_KEY, None)) is not None:
            return current_auth_method

        # Iterate through auth methods, checking if user is authenticated
        for cls in [AdminApiKeyAuthMethod,
                    AdminSessionAuthMethod,
                    UploadApiKeyAuthMethod,
                    PublishApiKeyAuthMethod,
                    SamlAuthMethod,
                    OpenidConnectAuthMethod,
                    NotAuthenticated]:
            if not cls.is_enabled():
                continue

            if auth_method := cls.get_current_instance():
                setattr(g, AuthFactory.FLASK_GLOBALS_AUTH_KEY, auth_method)
                return auth_method

        raise Exception('Unable to determine current auth type - not caught by NotAuthenticated')


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


class NotAuthenticated(BaseAuthMethod):
    """Base auth method for unauthenticated users"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return True

    def is_authenticated(self):
        """Whether user is authenticated"""
        return False

    def check_namespace_access(self, permission_type, namespace):
        """Unauthenticated users have no namespace access."""
        return False

    @classmethod
    def check_auth_state(cls):
        """Always return True as a last-catch auth mechanism"""
        return True

    @classmethod
    def is_enabled(cls):
        return True

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        # If API key authentication is not configured for publishing modules,
        # allow unauthenticated access
        if not PublishApiKeyAuthMethod.is_enabled():
            return True
        return False

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        # If API key authentication is not configured for uploading modules,
        # allow unauthenticated access
        if not UploadApiKeyAuthMethod.is_enabled():
            return True
        return False


class BaseAdminAuthMethod(BaseAuthMethod):
    """Base auth method admin authentication"""

    def is_admin(self):
        """Return whether user is an admin"""
        return True

    def is_built_in_admin(self):
        """Whether user is the built-in admin"""
        return True

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        return True

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        # Allow full access to all namespaces
        return True

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN)


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


class BaseSessionAuthMethod(BaseAuthMethod):
    """Base auth method for session-based authentication"""

    SESSION_AUTH_TYPE_KEY = 'authentication_type'
    SESSION_AUTH_TYPE_VALUE = None

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return True

    @classmethod
    def check_session_auth_type(cls):
        """Check if the current type of authenticate is set in session."""
        # Check if auth type value has been overriden
        if cls.SESSION_AUTH_TYPE_VALUE is None:
            raise NotImplementedError

        return session.get(cls.SESSION_AUTH_TYPE_KEY, None) == cls.SESSION_AUTH_TYPE_VALUE

    @classmethod
    def check_session(cls):
        """Check if auth-specific session is valid."""
        raise NotImplementedError

    @classmethod
    def check_auth_state(cls):
        """Check if session is valid."""
        # Ensure session secret key is set,
        # session ID is present and valid and
        # is_admin_authenticated session is set
        if (not terrareg.config.Config().SECRET_KEY or
                not Session.check_session(session.get('session_id', None)) or
                not session.get('is_admin_authenticated', False)):
            return False

        # Ensure session type is set to the current session and session is valid
        return cls.check_session_auth_type() and cls.check_session()


class UploadApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for upload API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if upload API key is provided"""
        return cls._check_api_key(terrareg.config.Config().UPLOAD_API_KEYS)

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().UPLOAD_API_KEYS)

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        return False


class PublishApiKeyAuthMethod(BaseApiKeyAuthMethod):
    """Auth method for publish API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if upload API key is provided"""
        return cls._check_api_key(terrareg.config.Config().PUBLISH_API_KEYS)

    @classmethod
    def is_enabled(cls):
        return bool(terrareg.config.Config().PUBLISH_API_KEYS)

    def can_publish_module_version(self, namespace):
        """Whether user can publish module version within a namespace."""
        return True

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        return False


class BaseSsoAuthMethod(BaseSessionAuthMethod):
    """Base methods for SSO based authentication"""

    def get_group_memberships(self):
        """Return list of groups that the user is a member of"""
        raise NotImplementedError

    def is_admin(self):
        """Check if user is an admin"""
        # Obtain list of user's groups
        groups = self.get_group_memberships()

        # Find any user groups that the user is a member of
        # that has admin permissions
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    db.user_group
                ).where(
                    db.user_group.c.name.in_(groups),
                    db.user_group.c.site_admin==True
                )
            )
            if res.fetchone():
                return True

            return False

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        namespace_obj = Namespace.get(namespace)

        # Obtain list of user's groups
        groups = self.get_group_memberships()

        # Find any permissions
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    db.user_group_namespace_permission
                ).join(
                    db.user_group,
                    db.user_group_namespace_permission.c.user_group_id==db.user_group.c.id
                ).where(
                    db.user_group_namespace_permission.c.namespace_id==namespace_obj.pk,
                    db.user_group.c.name.in_(groups)
                )
            )
            # Check each permission mapping that the user has for the namespace
            # and return True is the permission type matches the required permission,
            # or the permission is FULL access.
            for row in res:
                if (row['permission_type'] == permission_type or
                        row['permission_type'] == UserGroupNamespacePermissionType.FULL):
                    return True

            return False


class SamlAuthMethod(BaseSsoAuthMethod):
    """Auth method for SAML authentication"""

    SESSION_AUTH_TYPE_VALUE = 5

    @classmethod
    def check_session(cls):
        """Check SAML session is valid"""
        return session.get('samlUserdata')

    @classmethod
    def is_enabled(cls):
        return terrareg.saml.Saml2.is_enabled()

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return session.get('openidsamlUserdata_groups', {}).get('groups', [])


class OpenidConnectAuthMethod(BaseSsoAuthMethod):
    """Auth method for OpenID authentication"""
    
    SESSION_AUTH_TYPE_VALUE = 4
    
    @classmethod
    def check_session(cls):
        """Check OpenID session"""
        # Check OpenID connect expiry time
        if datetime.datetime.now() >= datetime.datetime.fromtimestamp(
                session.get('openid_connect_expires_at', 0)):
            return False

        try:
            terrareg.openid_connect.OpenidConnect.validate_session_token(session.get('openid_connect_id_token'))
        except:
            # Catch any exceptions when validating session and return False
            return False
        # If validate did not raise errors, return True
        return True

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return session.get('openid_groups', [])

    @classmethod
    def is_enabled(cls):
        return terrareg.openid_connect.OpenidConnect.is_enabled()


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if admin API key is provided"""
        return cls._check_api_key([terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN])


class AdminSessionAuthMethod(BaseAdminAuthMethod, BaseSessionAuthMethod):
    """Auth method for admin session"""

    SESSION_AUTH_TYPE_VALUE = 3

    @classmethod
    def check_session(cls):
        """Check admin session"""
        # There are no additional attributes to check
        return True
