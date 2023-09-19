

from unittest import mock
import pytest

from terrareg.auth import AdminApiKeyAuthMethod, UserGroupNamespacePermissionType
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest
from test import BaseTest


class TestAdminApiKeyAuthMethod(BaseAuthMethodTest):
    """Test methods of AdminApiKeyAuthMethod auth method"""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.is_built_in_admin() is True

    def test_is_admin(self):
        """Test is_admin method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.is_admin() is True

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = AdminApiKeyAuthMethod()
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
            obj = AdminApiKeyAuthMethod()
            assert obj.is_enabled() is expected_result

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.requires_csrf_tokens is False

    @pytest.mark.parametrize('public_api_keys', [
        (None),
        ([]),
        (['publishapikey'])
    ])
    def test_can_publish_module_version(self, public_api_keys):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys):
            obj = AdminApiKeyAuthMethod()
            assert obj.can_publish_module_version(namespace='testnamespace') is True

    @pytest.mark.parametrize('upload_api_keys', [
        None,
        [],
        ['uploadapikey']
    ])
    def test_can_upload_module_version(self, upload_api_keys):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys):
            obj = AdminApiKeyAuthMethod()
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
    def test_check_auth_state(self, admin_authentication_config, api_key_header, expected_result):
        """Test check_auth_state method"""
        headers = {}
        # Add API key header to request
        if api_key_header is not None:
            headers['HTTP_X_TERRAREG_APIKEY'] = api_key_header

        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', admin_authentication_config), \
                BaseTest.get().SERVER._app.test_request_context(environ_base=headers):

            obj = AdminApiKeyAuthMethod()
            assert obj.check_auth_state() is expected_result

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, True),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, True)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.get_username() == 'Built-in admin'

    def test_can_access_read_api(self):
        """Test can_access_read_api method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.can_access_read_api() == True

    def test_can_access_terraform_api(self):
        """Test can_access_terraform_api method"""
        obj = AdminApiKeyAuthMethod()
        assert obj.can_access_terraform_api() == True
