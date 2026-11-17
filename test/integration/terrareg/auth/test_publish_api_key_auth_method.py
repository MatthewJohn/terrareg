import datetime

from unittest import mock

import pytest

from terrareg.auth import PublishApiKeyAuthMethod
import terrareg.api_key
from test.integration.terrareg import TerraregIntegrationTest
from test import BaseTest


class TestPublishApiKeyAuthMethod(TerraregIntegrationTest):

    def test_authentication_with_db_key(self):
        """Test publish auth works with DB-backed API key."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Test publish key',
            key_type=terrareg.api_key.ApiKeyType.PUBLISH,
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert PublishApiKeyAuthMethod.check_auth_state() is True

            instance = PublishApiKeyAuthMethod.get_current_instance()
            assert instance is not None
            assert instance.matched_api_key.pk == api_key.pk
            assert instance.matched_api_key.name == 'Test publish key'

    def test_is_enabled_with_db_key(self):
        """Test DB-backed keys enable auth method."""
        with mock.patch('terrareg.config.Config.PUBLISH_API_KEYS', []):
            api_key, _ = terrareg.api_key.ApiKey.create(
                name='Test publish key',
                key_type=terrareg.api_key.ApiKeyType.PUBLISH,
                created_by='test'
            )
            assert PublishApiKeyAuthMethod.is_enabled() is True

    def test_namespace_restricted_key(self):
        """Test namespace-restricted keys only work for their namespace."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Restricted publish key',
            key_type=terrareg.api_key.ApiKeyType.PUBLISH,
            namespace='allowed-namespace',
            created_by='test'
        )

        instance = PublishApiKeyAuthMethod(matched_api_key=api_key)

        assert instance.can_publish_module_version('allowed-namespace') is True
        assert instance.can_publish_module_version('other-namespace') is False

    def test_unrestricted_key_all_namespaces(self):
        """Test unrestricted keys work for all namespaces."""
        api_key, _ = terrareg.api_key.ApiKey.create(
            name='Unrestricted publish key',
            key_type=terrareg.api_key.ApiKeyType.PUBLISH,
            created_by='test'
        )

        instance = PublishApiKeyAuthMethod(matched_api_key=api_key)

        assert instance.can_publish_module_version('namespace1') is True
        assert instance.can_publish_module_version('namespace2') is True

    def test_expired_key_fails(self):
        """Test expired keys don't authenticate."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Expired publish key',
            key_type=terrareg.api_key.ApiKeyType.PUBLISH,
            expires_at=datetime.datetime.now() - datetime.timedelta(minutes=1),
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert PublishApiKeyAuthMethod.check_auth_state() is False