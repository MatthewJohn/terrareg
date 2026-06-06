from unittest import mock

import pytest

from terrareg.auth import UploadAndPublishApiKeyAuthMethod
from terrareg.user_group_namespace_permission_type import UserGroupNamespacePermissionType
from test import BaseTest
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest


class TestUploadAndPublishApiKeyAuthMethod(BaseAuthMethodTest):
    """Test methods of UploadAndPublishApiKeyAuthMethod auth method."""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.is_built_in_admin() is False

    def test_is_admin(self):
        """Test is_admin method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.is_admin() is False

    def test_is_authenticated(self):
        """Test is_authenticated method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.is_authenticated() is True

    def test_is_enabled(self):
        """Test DB-backed upload-and-publish API keys enable the auth method."""
        with mock.patch('terrareg.api_key.ApiKey.has_active_keys', return_value=True) as mock_has_active_keys:
            obj = UploadAndPublishApiKeyAuthMethod()
            assert obj.is_enabled() is True

        mock_has_active_keys.assert_called_once()

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.requires_csrf_tokens is False

    def test_can_publish_module_version(self):
        """Test can_publish_module_version method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.can_publish_module_version(namespace='testnamespace') is True

    def test_can_publish_module_version_restricted_namespace(self):
        """Test namespace-restricted keys only publish within their namespace."""
        obj = UploadAndPublishApiKeyAuthMethod()
        obj.matched_api_key = mock.MagicMock(namespace='allowed')
        assert obj.can_publish_module_version(namespace='allowed') is True
        assert obj.can_publish_module_version(namespace='blocked') is False

    def test_can_upload_module_version(self):
        """Test can_upload_module_version method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.can_upload_module_version(namespace='testnamespace') is True

    def test_can_upload_module_version_restricted_namespace(self):
        """Test namespace-restricted keys only upload within their namespace."""
        obj = UploadAndPublishApiKeyAuthMethod()
        obj.matched_api_key = mock.MagicMock(namespace='allowed')
        assert obj.can_upload_module_version(namespace='allowed') is True
        assert obj.can_upload_module_version(namespace='blocked') is False

    @pytest.mark.parametrize('api_key_header,expected_result', [
        (None, False),
        ('', False),
        ('incorrect', False),
        ('valid', True),
    ])
    def test_check_auth_state(self, api_key_header, expected_result):
        """Test check_auth_state method."""
        headers = {}
        if api_key_header is not None:
            headers['HTTP_X_TERRAREG_APIKEY'] = api_key_header

        mock_api_key = mock.MagicMock()
        with mock.patch(
            'terrareg.api_key.ApiKey.verify_key',
            return_value=mock_api_key if api_key_header == 'valid' else None
        ) as mock_verify_key, BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            obj = UploadAndPublishApiKeyAuthMethod()
            assert obj.check_auth_state() is expected_result

        if api_key_header:
            mock_verify_key.assert_called_once_with(api_key_header, 'upload_and_publish')
            if expected_result:
                mock_api_key.mark_used.assert_called_once_with()
        else:
            mock_verify_key.assert_not_called()

    def test_get_current_instance_with_db_api_key(self):
        """Test upload-and-publish auth stores matched DB-backed API keys on the instance."""
        headers = {'HTTP_X_TERRAREG_APIKEY': 'db-valid'}
        mock_api_key = mock.MagicMock()
        with mock.patch('terrareg.api_key.ApiKey.verify_key', return_value=mock_api_key), \
                BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            obj = UploadAndPublishApiKeyAuthMethod.get_current_instance()

        assert obj is not None
        assert obj.matched_api_key is mock_api_key

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, False),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, False)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

    def test_get_username(self):
        """Test get_username method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.get_username() == 'Upload and Publish API Key'

    def test_can_access_read_api(self):
        """Test can_access_read_api method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.can_access_read_api() == False

    def test_can_access_terraform_api(self):
        """Test can_access_terraform_api method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.can_access_terraform_api() == False

    def test_should_record_terraform_analytics(self):
        """Test should_record_terraform_analytics method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.should_record_terraform_analytics() is True

    def test_get_terraform_auth_token(self):
        """Test get_terraform_auth_token method."""
        obj = UploadAndPublishApiKeyAuthMethod()
        assert obj.get_terraform_auth_token() is None