
import datetime
from enum import Enum

import sqlalchemy
from flask import g, request, session
import flask
import pyop.exceptions

import terrareg.config
from terrareg.database import Database
import terrareg.models
import terrareg.models
import terrareg.openid_connect
import terrareg.saml
from terrareg.terraform_idp import TerraformIdp
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType


class AuthenticationType(Enum):
    """Determine the method of authentication."""
    NOT_CHECKED = 0
    NOT_AUTHENTICATED = 1
    AUTHENTICATION_TOKEN = 2
    SESSION_PASSWORD = 3
    SESSION_OPENID_CONNECT = 4
    SESSION_SAML = 5


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
                    TerraformOidcAuthMethod,
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

    def get_all_namespace_permissions(self):
        """Return all permissions by namespace"""
        return {}

    def get_username(self):
        """Get username of current user"""
        raise NotImplementedError

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        raise NotImplementedError


class NotAuthenticated(BaseAuthMethod):
    """Base auth method for unauthenticated users"""

    @property
    def requires_csrf_tokens(self):
        """Whether auth type requires CSRF tokens"""
        return False

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
        # If API key authentication is not configured for publishing modules and
        # RBAC is not enabled, allow unauthenticated access
        if (not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not PublishApiKeyAuthMethod.is_enabled()):
            return True
        return False

    def can_upload_module_version(self, namespace):
        """Whether user can upload/index module version within a namespace."""
        # If API key authentication is not configured for uploading modules and
        # RBAC is not enabled, allow unauthenticated access
        if (not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not UploadApiKeyAuthMethod.is_enabled()):
            return True
        return False

    def get_username(self):
        """Get username of current user"""
        return 'Unauthenticated User'

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # Unauthenticated users can only access 'read' APIs
        # if global anonymous access is allowed
        return terrareg.config.Config().ALLOW_UNAUTHENTICATED_ACCESS


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

    def get_username(self):
        """Get username of current user"""
        return 'Built-in admin'

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        return True


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

        return flask.session.get(cls.SESSION_AUTH_TYPE_KEY, None) == cls.SESSION_AUTH_TYPE_VALUE.value

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
                not terrareg.models.Session.check_session(flask.session.get('session_id', None)) or
                not flask.session.get('is_admin_authenticated', False)):
            return False

        # Ensure session type is set to the current session and session is valid
        return cls.check_session_auth_type() and cls.check_session()

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        # Logged in SSO users can access any 'read' APIs
        return True


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

    def get_username(self):
        """Get username of current user"""
        return 'Upload API Key'


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

    def get_username(self):
        """Get username of current user"""
        return 'Publish API Key'


class BaseSsoAuthMethod(BaseSessionAuthMethod):
    """Base methods for SSO based authentication"""

    def get_group_memberships(self):
        """Return list of groups that the user is a member of"""
        raise NotImplementedError

    def is_admin(self):
        """Check if user is an admin"""
        # Check if RBAC is enabled, if not, all authenticated users
        # are treated as admins
        if not terrareg.config.Config().ENABLE_ACCESS_CONTROLS:
            return True

        # Obtain list of user's groups
        for group in self.get_group_memberships():
            user_group = terrareg.models.UserGroup.get_by_group_name(group)
            if user_group is not None and user_group.site_admin:
                return True
        return False

    def can_publish_module_version(self, namespace):
        """Determine if user can publish a module version to given namespace."""
        return (
            # If PUBLISH API keys have not been enabled and
            # RBAC has not been enabled,
            # allow user to publish module versions, as this
            # can be performed without authentication
            ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not PublishApiKeyAuthMethod.is_enabled())) or
            # Otherwise, check for MODIFY namespace access
            self.check_namespace_access(namespace=namespace, permission_type=UserGroupNamespacePermissionType.MODIFY)
        )

    def can_upload_module_version(self, namespace):
        """Determine if user can upload a module version to given namespace."""
        return (
            # If UPLOAD API keys have not been enabled and
            # RBAC has not been enabled,
            # allow user to publish module versions, as this
            # can be performed without authentication
            ((not terrareg.config.Config().ENABLE_ACCESS_CONTROLS) and (not UploadApiKeyAuthMethod.is_enabled())) or
            # Otherwise, check for MODIFY namespace access
            self.check_namespace_access(namespace=namespace, permission_type=UserGroupNamespacePermissionType.MODIFY)
        )

    def get_all_namespace_permissions(self):
        """Obtain all namespace permissions for user."""
        # Obtain list of user's groups
        groups = self.get_group_memberships()

        # Find any permissions
        db = Database.get()
        with db.get_connection() as conn:
            res = conn.execute(
                sqlalchemy.select(
                    db.user_group_namespace_permission.c.permission_type,
                    db.namespace.c.namespace
                ).join(
                    db.user_group,
                    db.user_group_namespace_permission.c.user_group_id==db.user_group.c.id
                ).join(
                    db.namespace,
                    db.user_group_namespace_permission.c.namespace_id==db.namespace.c.id
                ).where(
                    db.user_group.c.name.in_(groups)
                )
            )
            return {
                terrareg.models.Namespace(row['namespace']): row['permission_type']
                for row in res
            }

    def check_namespace_access(self, permission_type, namespace):
        """Check access level to a given namespace."""
        # Check admin access
        if self.is_admin():
            return True

        namespace_obj = terrareg.models.Namespace.get(namespace)
        if not namespace_obj:
            return False

        # Obtain list of user's groups
        user_groups = []
        for group in self.get_group_memberships():
            user_group = terrareg.models.UserGroup.get_by_group_name(group)
            # If user group object was found for SSO group,
            # add to list
            if user_group:
                user_groups.append(user_group)

        user_group_permissions = terrareg.models.UserGroupNamespacePermission.get_permissions_by_user_groups_and_namespace(
            user_groups=user_groups,
            namespace=namespace_obj
        )
        for user_group_permission in user_group_permissions:
            if (user_group_permission.permission_type == permission_type or
                    user_group_permission.permission_type == UserGroupNamespacePermissionType.FULL):
                return True
        return False


class SamlAuthMethod(BaseSsoAuthMethod):
    """Auth method for SAML authentication"""

    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_SAML

    @classmethod
    def check_session(cls):
        """Check SAML session is valid"""
        return bool(flask.session.get('samlUserdata'))

    @classmethod
    def is_enabled(cls):
        return terrareg.saml.Saml2.is_enabled()

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        user_data_groups = flask.session.get('samlUserdata', None)
        if user_data_groups and isinstance(user_data_groups, dict):
            groups = user_data_groups.get(terrareg.config.Config().SAML2_GROUP_ATTRIBUTE)
            if isinstance(groups, list):
                return groups
        return []

    def get_username(self):
        """Get username of current user"""
        return flask.session.get('samlNameId')


class OpenidConnectAuthMethod(BaseSsoAuthMethod):
    """Auth method for OpenID authentication"""
    
    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_OPENID_CONNECT
    
    @classmethod
    def check_session(cls):
        """Check OpenID session"""
        # Check OpenID connect expiry time
        try:
            session_timestamp = float(flask.session.get('openid_connect_expires_at', 0))
        except ValueError:
            return False

        if datetime.datetime.now() >= datetime.datetime.fromtimestamp(session_timestamp):
            return False

        try:
            terrareg.openid_connect.OpenidConnect.validate_session_token(flask.session.get('openid_connect_id_token'))
        except:
            # Catch any exceptions when validating session and return False
            return False
        # If validate did not raise errors, return True
        return True

    def get_group_memberships(self):
        """Return list of groups that the user a member of"""
        return flask.session.get('openid_groups', []) or []

    @classmethod
    def is_enabled(cls):
        return terrareg.openid_connect.OpenidConnect.is_enabled()

    def get_username(self):
        """Get username of current user"""
        return flask.session.get('openid_username')


class AdminApiKeyAuthMethod(BaseAdminAuthMethod, BaseApiKeyAuthMethod):
    """Auth method for admin API key"""

    @classmethod
    def check_auth_state(cls):
        """Check if admin API key is provided"""
        return cls._check_api_key([terrareg.config.Config().ADMIN_AUTHENTICATION_TOKEN])


class AdminSessionAuthMethod(BaseAdminAuthMethod, BaseSessionAuthMethod):
    """Auth method for admin session"""

    SESSION_AUTH_TYPE_VALUE = AuthenticationType.SESSION_PASSWORD

    @classmethod
    def check_session(cls):
        """Check admin session"""
        # There are no additional attributes to check
        return True


class TerraformOidcAuthMethod(BaseAuthMethod):

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
        return TerraformIdp.get().is_enabled

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
        print(request.headers)
        print(request.data)
        if 'Authorization' in request.headers:
            # Check header with OpenIDC
            try:
                res = TerraformIdp.get().provider.handle_userinfo_request(request.data, request.headers)
                print(res)
                return True
            except pyop.exceptions.InvalidAccessToken:
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
        raise "Terraform CLI User"

    def can_access_read_api(self):
        """Whether the user can access 'read' APIs"""
        return True
