

from unittest import mock

import pytest

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test import BaseTest, test_request_context
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest


class BaseTerraformStaticTokenTests(BaseAuthMethodTest):
    """Test methods of self.CLS auth method"""

    CLS = None

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = self.CLS()
        assert obj.is_built_in_admin() is False

    def test_is_admin(self):
        """Test is_admin method"""
        obj = self.CLS()
        assert obj.is_admin() is False

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = self.CLS()
        assert obj.is_authenticated() is True

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = self.CLS()
        assert obj.requires_csrf_tokens is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    @pytest.mark.parametrize('public_api_keys', [
        (None),
        ([]),
        (['publishapikey'])
    ])
    def test_can_publish_module_version(self, public_api_keys, allow_unauthenticated_access):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys), \
                mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = self.CLS()
            assert obj.can_publish_module_version(namespace='testnamespace') is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    @pytest.mark.parametrize('upload_api_keys', [
        (None),
        ([]),
        (['publishapikey'])
    ])
    def test_can_upload_module_version(self, upload_api_keys, allow_unauthenticated_access):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys), \
                mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = self.CLS()
            assert obj.can_upload_module_version(namespace='testnamespace') is False


    @pytest.mark.parametrize('namespace,access_type', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY),
        ('testnamespace', UserGroupNamespacePermissionType.FULL)
    ])
    def test_check_namespace_access(self, namespace, access_type):
        """Test check_namespace_access method"""
        obj = self.CLS()
        assert obj.check_namespace_access(access_type, namespace) is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_read_api(self, allow_unauthenticated_access):
        """Test can_access_read_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = self.CLS()
            assert obj.can_access_read_api() is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_terraform_api(self, allow_unauthenticated_access):
        """Test can_access_terraform_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = self.CLS()
            assert obj.can_access_terraform_api() == True
