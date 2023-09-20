

from unittest import mock
import pytest

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.auth import NotAuthenticated
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest


class TestNotAuthenticated(BaseAuthMethodTest):
    """Test methods of NotAuthenticated auth method"""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = NotAuthenticated()
        assert obj.is_built_in_admin() is False

    def test_is_admin(self):
        """Test is_admin method"""
        obj = NotAuthenticated()
        assert obj.is_admin() is False

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = NotAuthenticated()
        assert obj.is_authenticated() is False

    def test_is_enabled(self):
        """Test is_enabled method"""
        obj = NotAuthenticated()
        assert obj.is_enabled() is True

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = NotAuthenticated()
        assert obj.requires_csrf_tokens is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    @pytest.mark.parametrize('public_api_keys,expected_result_api_result', [
        (None, True),
        ([], True),
        (['publishapikey', False])
    ])
    def test_can_publish_module_version(self, public_api_keys, allow_unauthenticated_access, expected_result_api_result):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys), \
                mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = NotAuthenticated()
            assert obj.can_publish_module_version(namespace='testnamespace') is (expected_result_api_result and allow_unauthenticated_access)

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    @pytest.mark.parametrize('upload_api_keys,expected_result_api_result', [
        (None, True),
        ([], True),
        (['publishapikey', False])
    ])
    def test_can_upload_module_version(self, upload_api_keys, allow_unauthenticated_access, expected_result_api_result):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys), \
                mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = NotAuthenticated()
            assert obj.can_upload_module_version(namespace='testnamespace') is (expected_result_api_result and allow_unauthenticated_access)

    def test_check_auth_state(self):
        """Test check_auth_state method"""
        obj = NotAuthenticated()
        assert obj.check_auth_state() is True

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, False),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, False)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = NotAuthenticated()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = NotAuthenticated()
        assert obj.get_username() == 'Unauthenticated User'

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_read_api(self, allow_unauthenticated_access):
        """Test can_access_read_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = NotAuthenticated()
            assert obj.can_access_read_api() == allow_unauthenticated_access

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_terraform_api(self, allow_unauthenticated_access):
        """Test can_access_terraform_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = NotAuthenticated()
            assert obj.can_access_terraform_api() == allow_unauthenticated_access

    def test_should_record_terraform_analytics(self):
        """Test should_record_terraform_analytics method"""
        obj = NotAuthenticated()
        assert obj.should_record_terraform_analytics() is True

    def test_get_terraform_auth_token(self):
        """Test get_terraform_auth_token method"""
        obj = NotAuthenticated()
        assert obj.get_terraform_auth_token() is None
