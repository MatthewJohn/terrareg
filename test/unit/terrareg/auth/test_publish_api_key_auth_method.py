

from unittest import mock
import pytest

from terrareg.auth import PublishApiKeyAuthMethod, UserGroupNamespacePermissionType
from test.unit.terrareg.auth.base_auth_method_test import BaseAuthMethodTest
from test import BaseTest


class TestPublishApiKeyAuthMethod(BaseAuthMethodTest):
    """Test methods of PublishApiKeyAuthMethod auth method"""

    def test_is_built_in_admin(self):
        """Test is_built_in_admin method"""
        obj = PublishApiKeyAuthMethod()
        assert obj.is_built_in_admin() is False

    def test_is_admin(self):
        """Test is_admin method"""
        obj = PublishApiKeyAuthMethod()
        assert obj.is_admin() is False

    def test_is_authenticated(self):
        """Test is_authenticated method"""
        obj = PublishApiKeyAuthMethod()
        # @TODO Should this be True?
        assert obj.is_authenticated() is True

    @pytest.mark.parametrize('publish_api_keys,expected_result', [
        (None, False),
        ([], False),
        (['uploadapikey', True])
    ])
    def test_is_enabled(self, publish_api_keys, expected_result):
        """Test is_enabled method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', publish_api_keys):
            obj = PublishApiKeyAuthMethod()
            assert obj.is_enabled() is expected_result

    def test_requires_csrf_tokens(self):
        """Test requires_csrf_token method"""
        obj = PublishApiKeyAuthMethod()
        assert obj.requires_csrf_tokens is False

    @pytest.mark.parametrize('public_api_keys,expected_result', [
        (None, True),
        ([], True),
        (['publishapikey'], True)
    ])
    def test_can_publish_module_version(self, public_api_keys, expected_result):
        """Test can_publish_module_version method"""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', public_api_keys):
            obj = PublishApiKeyAuthMethod()
            assert obj.can_publish_module_version(namespace='testnamespace') is expected_result

    @pytest.mark.parametrize('upload_api_keys,expected_result', [
        (None, False),
        ([], False),
        (['uploadapikey'], False)
    ])
    def test_can_upload_module_version(self, upload_api_keys, expected_result):
        """Test can_upload_module_version method"""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', upload_api_keys):
            obj = PublishApiKeyAuthMethod()
            assert obj.can_upload_module_version(namespace='testnamespace') is expected_result

    @pytest.mark.parametrize('api_key_config,api_key_header,expected_result', [
        ([], None, False),
        (['testApiKey1'], None, False),
        (['testApiKey1', ''], None, False),

        ([], '', False),
        (['testApiKey1'], '', False),
        (['testApiKey1', ''], '', False),

        ([], 'incorrect', False),
        (['testApiKey1'], 'incorrect', False),
        (['testApiKey1', ''], '', False),

        (['valid'], 'valid', True),
        (['', 'valid', '', 'another'], 'valid', True),
    ])
    def test_check_auth_state(self, api_key_config, api_key_header, expected_result):
        """Test check_auth_state method"""
        headers = {}
        # Add API key header to request
        if api_key_header is not None:
            headers['HTTP_X_TERRAREG_APIKEY'] = api_key_header

        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', api_key_config), \
                BaseTest.get().SERVER._app.test_request_context(environ_base=headers):

            obj = PublishApiKeyAuthMethod()
            assert obj.check_auth_state() is expected_result

    @pytest.mark.parametrize('namespace,access_type,expected_result', [
        ('testnamespace', UserGroupNamespacePermissionType.MODIFY, False),
        ('testnamespace', UserGroupNamespacePermissionType.FULL, False)
    ])
    def test_check_namespace_access(self, namespace, access_type, expected_result):
        """Test check_namespace_access method"""
        obj = PublishApiKeyAuthMethod()
        assert obj.check_namespace_access(access_type, namespace) is expected_result

