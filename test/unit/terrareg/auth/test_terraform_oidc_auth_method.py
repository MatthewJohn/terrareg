

from unittest import mock

import pytest
import pyop.exceptions
from werkzeug.datastructures import Headers, EnvironHeaders

from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from terrareg.auth import TerraformOidcAuthMethod
from test import BaseTest, test_request_context
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest
import terrareg.terraform_idp


class TestTerraformOidcAuthMethod(BaseAuthMethodTest):
    """Test methods of TerraformOidcAuthMethod auth method"""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = TerraformOidcAuthMethod()
        assert obj.is_built_in_admin() is False

    def test_is_admin(self):
        """Test is_admin method"""
        obj = TerraformOidcAuthMethod()
        assert obj.is_admin() is False

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = TerraformOidcAuthMethod()
        assert obj.is_authenticated() is True

    @pytest.mark.parametrize('terraform_idp_is_enabled', [
        True,
        False
    ])
    def test_is_enabled(self, terraform_idp_is_enabled):
        """Test is_enabled method"""
        obj = TerraformOidcAuthMethod()
        with mock.patch('terrareg.terraform_idp.TerraformIdp.is_enabled', terraform_idp_is_enabled):
            assert obj.is_enabled() is terraform_idp_is_enabled

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = TerraformOidcAuthMethod()
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
            obj = TerraformOidcAuthMethod()
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
            obj = TerraformOidcAuthMethod()
            assert obj.can_upload_module_version(namespace='testnamespace') is False

    @pytest.mark.parametrize('headers, raises_exception, expected_call, expected_result', [
        # No header
        ({}, False, False, False),
        # Invalid token
        ({'Authorization': 'Unittest authorization header'}, True, True, False),
        # Working header
        ({'Authorization': 'Unittest authorization header'}, False, True, True),
    ])
    def test_check_auth_state(self, headers, raises_exception, expected_call, expected_result, test_request_context):
        """Test check_auth_state method"""
        terrareg.terraform_idp.TerraformIdp._INSTANCE = None

        obj = TerraformOidcAuthMethod()

        mock_provider = mock.MagicMock()
        def raise_exception(*args, **kwargs):
            if raises_exception:
                raise pyop.exceptions.InvalidAccessToken('Unittest exception')

        mock_provider.handle_userinfo_request = mock.MagicMock()
        mock_provider.handle_userinfo_request.side_effect = raise_exception

        with mock.patch('terrareg.terraform_idp.TerraformIdp.provider', mock_provider), \
                BaseTest.get().SERVER._app.test_request_context(headers=headers) as request_context:
            
            assert obj.check_auth_state() is expected_result

            if expected_call:
                mock_provider.handle_userinfo_request.assert_called_once_with(b'', EnvironHeaders(request_context.request.environ))
            else:
                mock_provider.handle_userinfo_request.assert_not_called()

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, False),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, False)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = TerraformOidcAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

    def test_get_username(self):
        """Test get_username method"""
        obj = TerraformOidcAuthMethod()
        assert obj.get_username() == "Terraform CLI User"

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_read_api(self, allow_unauthenticated_access):
        """Test can_access_read_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = TerraformOidcAuthMethod()
            assert obj.can_access_read_api() is False

    @pytest.mark.parametrize('allow_unauthenticated_access', [
        True,
        False
    ])
    def test_can_access_terraform_api(self, allow_unauthenticated_access):
        """Test can_access_terraform_api method"""
        with mock.patch('terrareg.config.Config.ALLOW_UNAUTHENTICATED_ACCESS', allow_unauthenticated_access):
            obj = TerraformOidcAuthMethod()
            assert obj.can_access_terraform_api() == True

