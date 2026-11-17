import datetime

from unittest import mock

import pytest

from terrareg.auth import AdminApiKeyAuthMethod
import terrareg.api_key
from test.integration.terrareg import TerraregIntegrationTest
from test import BaseTest


class TestAdminApiKeyAuthMethod(TerraregIntegrationTest):

    def test_authentication_with_db_key(self):
        """Test admin auth works with DB-backed API key."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Test admin key',
            key_type=terrareg.api_key.ApiKeyType.ADMIN,
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert AdminApiKeyAuthMethod.check_auth_state() is True

            instance = AdminApiKeyAuthMethod.get_current_instance()
            assert instance is not None
            assert instance.matched_api_key.pk == api_key.pk
            assert instance.is_admin() is True

            updated_key = terrareg.api_key.ApiKey.get(api_key.pk)
            assert updated_key.last_used_at is not None

    def test_is_enabled_with_db_key(self):
        """Test DB-backed keys enable auth method."""
        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', None):
            api_key, _ = terrareg.api_key.ApiKey.create(
                name='Test admin key',
                key_type=terrareg.api_key.ApiKeyType.ADMIN,
                created_by='test'
            )
            assert AdminApiKeyAuthMethod.is_enabled() is True

    def test_expired_key_fails(self):
        """Test expired keys don't authenticate."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Expired key',
            key_type=terrareg.api_key.ApiKeyType.ADMIN,
            expires_at=datetime.datetime.now() - datetime.timedelta(minutes=1),
            created_by='test'
        )

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert AdminApiKeyAuthMethod.check_auth_state() is False

    def test_revoked_key_fails(self):
        """Test revoked keys don't authenticate."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='Revoked key',
            key_type=terrareg.api_key.ApiKeyType.ADMIN,
            created_by='test'
        )

        api_key.revoke()

        headers = {'HTTP_X_TERRAREG_APIKEY': plaintext_key}
        with BaseTest.get().SERVER._app.test_request_context(environ_base=headers):
            assert AdminApiKeyAuthMethod.check_auth_state() is False

    def test_static_key_and_db_key(self):
        """Test both static and DB keys work."""
        static_key = 'test-static-key'
        with mock.patch('terrareg.config.Config.ADMIN_AUTHENTICATION_TOKEN', static_key):
            api_key, db_key = terrareg.api_key.ApiKey.create(
                name='DB key',
                key_type=terrareg.api_key.ApiKeyType.ADMIN,
                created_by='test'
            )

            assert AdminApiKeyAuthMethod.is_enabled() is True

            headers_static = {'HTTP_X_TERRAREG_APIKEY': static_key}
            with BaseTest.get().SERVER._app.test_request_context(environ_base=headers_static):
                assert AdminApiKeyAuthMethod.check_auth_state() is True
                instance = AdminApiKeyAuthMethod.get_current_instance()
                assert instance is not None
                assert instance.matched_api_key is None  # Static keys don't have DB backing

            headers_db = {'HTTP_X_TERRAREG_APIKEY': db_key}
            with BaseTest.get().SERVER._app.test_request_context(environ_base=headers_db):
                assert AdminApiKeyAuthMethod.check_auth_state() is True
                instance = AdminApiKeyAuthMethod.get_current_instance()
                assert instance is not None
                assert instance.matched_api_key.pk == api_key.pk