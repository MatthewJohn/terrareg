

from unittest import mock
import pytest

from terrareg.auth import NotAuthenticated, UserGroupNamespacePermissionType
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

    @pytest.mark.parametrize('public_api_keys,expected_result', [
        (None, True),
        ([], True),
        (['publishapikey', False])
    ])
    def test_can_publish_module_version(self, public_api_keys, expected_result):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys):
            obj = NotAuthenticated()
            assert obj.can_publish_module_version(namespace='testnamespace') is expected_result

    @pytest.mark.parametrize('upload_api_keys,expected_result', [
        (None, True),
        ([], True),
        (['publishapikey', False])
    ])
    def test_can_upload_module_version(self, upload_api_keys, expected_result):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys):
            obj = NotAuthenticated()
            assert obj.can_upload_module_version(namespace='testnamespace') is expected_result

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

