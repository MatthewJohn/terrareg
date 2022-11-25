

from unittest import mock
import pytest

from terrareg.auth import AdminSessionAuthMethod, UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg.auth.base_session_auth_method_tests import BaseSessionAuthMethodTests

# Required as this is sued by BaseAdminSessionAuthMethod
from test import test_request_context


class TestAdminSessionAuthMethod(BaseSessionAuthMethodTests):
    """Test methods of AdminSessionAuthMethod auth method"""

    CLS = AdminSessionAuthMethod

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = AdminSessionAuthMethod()
        assert obj.is_built_in_admin() is True

    def test_is_admin(self):
        """Test is_admin method"""
        obj = AdminSessionAuthMethod()
        assert obj.is_admin() is True

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = AdminSessionAuthMethod()
        # @TODO Should this be True?
        assert obj.is_authenticated() is True

    @pytest.mark.parametrize('admin_authentication_token,expected_result', [
        (None, False),
        ('', False),
        ('ISAPASSWORD', True)
    ])
    def test_is_enabled(self, admin_authentication_token, expected_result):
        """Test is_enabled method"""
        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', admin_authentication_token):
            obj = AdminSessionAuthMethod()
            assert obj.is_enabled() is expected_result

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = AdminSessionAuthMethod()
        assert obj.requires_csrf_tokens is True

    @pytest.mark.parametrize('public_api_keys', [
        (None),
        ([]),
        (['publishapikey'])
    ])
    def test_can_publish_module_version(self, public_api_keys):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys):
            obj = AdminSessionAuthMethod()
            assert obj.can_publish_module_version(namespace='testnamespace') is True

    @pytest.mark.parametrize('upload_api_keys', [
        None,
        [],
        ['uploadapikey']
    ])
    def test_can_upload_module_version(self, upload_api_keys):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys):
            obj = AdminSessionAuthMethod()
            assert obj.can_upload_module_version(namespace='testnamespace') is True

    def test_check_session(self):
        """Test check_session method"""
        obj = AdminSessionAuthMethod()
        assert obj.check_session() is True

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, True),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, True)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = AdminSessionAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = AdminSessionAuthMethod()
        assert obj.get_username() == 'Built-in admin'
