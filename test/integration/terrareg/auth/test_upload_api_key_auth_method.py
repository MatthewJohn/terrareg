import datetime

from unittest import mock

import pytest

from terrareg.auth import UploadApiKeyAuthMethod
import terrareg.api_key
from test.integration.terrareg import TerraregIntegrationTest
from test import BaseTest


class TestUploadApiKeyAuthMethod(TerraregIntegrationTest):

    def test_authentication_with_db_key(self):
        """Test upload auth works with DB-backed API key."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Test upload key',
            key_type=terrareg.api_key.ApiKeyType.UPLOAD,
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert UploadApiKeyAuthMethod.check_auth_state() is True

            instance = UploadApiKeyAuthMethod.get_current_instance()
            assert instance is not None
            assert instance.matched_api_key.pk == api_key.pk

    def test_is_enabled_with_db_key(self):
        """Test DB-backed keys enable auth method."""
        with mock.patch('terrareg.config.Config.UPLOAD_API_KEYS', []):
            api_key, _ = terrareg.api_key.ApiKey.create(
                name='Test upload key',
                key_type=terrareg.api_key.ApiKeyType.UPLOAD,
                created_by='test'
            )
            assert UploadApiKeyAuthMethod.is_enabled() is True

    def test_namespace_restricted_key(self):
        """Test namespace-restricted keys only work for their namespace."""
        api_key, _ = terrareg.api_key.ApiKey.create(
            name='Restricted upload key',
            key_type=terrareg.api_key.ApiKeyType.UPLOAD,
            namespace='allowed-namespace',
            created_by='test'
        )

        instance = UploadApiKeyAuthMethod(matched_api_key=api_key)

        assert instance.can_upload_module_version('allowed-namespace') is True
        assert instance.can_upload_module_version('other-namespace') is False

    def test_unrestricted_key_all_namespaces(self):
        """Test unrestricted keys work for all namespaces."""
        api_key, _ = terrareg.api_key.ApiKey.create(
            name='Unrestricted upload key',
            key_type=terrareg.api_key.ApiKeyType.UPLOAD,
            created_by='test'
        )

        instance = UploadApiKeyAuthMethod(matched_api_key=api_key)

        assert instance.can_upload_module_version('namespace1') is True
        assert instance.can_upload_module_version('namespace2') is True

    def test_expired_key_fails(self):
        """Test expired keys don't authenticate."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Expired upload key',
            key_type=terrareg.api_key.ApiKeyType.UPLOAD,
            expires_at=datetime.datetime.now() - datetime.timedelta(minutes=1),
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert UploadApiKeyAuthMethod.check_auth_state() is False