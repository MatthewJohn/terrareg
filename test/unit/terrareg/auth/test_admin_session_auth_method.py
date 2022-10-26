

from unittest import mock
import pytest

from terrareg.auth import AdminSessionAuthMethod, UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg.auth.test_base_session_auth_method import BaseAdminSessionAuthMethod

# Required as this is sued by BaseAdminSessionAuthMethod
from test import test_request_context


class TestAdminSessionAuthMethod(BaseAdminSessionAuthMethod):
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

    @pytest.mark.parametrize('admin_authentication_config,api_key_header,expected_result', [
        (None, None, False),
        ('', None, False),
        ('testApiKey1', None, False),

        (None, '', False),
        ('', '', False),
        ('testAdminPass', '', False),

        (None, 'incorrect', False),
        ('', 'incorrect', False),
        ('testAdminPass', 'incorrect', False),

        ('valid', 'valid', True)
    ])
    def test_check_session(self, admin_authentication_config, api_key_header, expected_result):
        """Test check_session method"""
        headers = {}
        # Add API key header to request
        if api_key_header is not None:
            headers['HTTP_X_TERRAREG_APIKEY'] = api_key_header

        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', admin_authentication_config), \
                BaseTest.get().SERVER._app.test_request_context(environ_base=headers):

            obj = AdminSessionAuthMethod()
            assert obj.check_session() is expected_result

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, True),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, True)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = AdminSessionAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result
