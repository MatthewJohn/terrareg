import datetime

import pytest

import terrareg.errors
import terrareg.api_key
from test.integration.terrareg import TerraregIntegrationTest


class TestApiKey(TerraregIntegrationTest):
    """Test ApiKey model."""

    def test_create_verify_mark_used_and_revoke(self):
        """Test API keys can be created, verified, marked used and revoked."""
        api_key, plaintext_key = terrareg.api_key.ApiKey.create(
            name='CI upload key',
            key_type=terrareg.api_key.ApiKeyType.UPLOAD,
            created_by='admin'
        )

        assert api_key.name == 'CI upload key'
        assert api_key.key_type == terrareg.api_key.ApiKeyType.UPLOAD.value
        assert api_key.key_prefix == plaintext_key[:terrareg.api_key.ApiKey.PREFIX_LENGTH]
        assert api_key.last_used_at is None

        verified_key = terrareg.api_key.ApiKey.verify_key(plaintext_key, terrareg.api_key.ApiKeyType.UPLOAD)
        assert verified_key.pk == api_key.pk

        verified_key.mark_used()
        assert terrareg.api_key.ApiKey.get(api_key.pk).last_used_at is not None

        verified_key.revoke()
        assert terrareg.api_key.ApiKey.verify_key(plaintext_key, terrareg.api_key.ApiKeyType.UPLOAD) is None

    def test_expired_keys_are_not_verified(self):
        """Test expired API keys no longer authenticate."""
        _, plaintext_key = terrareg.api_key.ApiKey.create(
            name='expired key',
            key_type=terrareg.api_key.ApiKeyType.PUBLISH,
            expires_at=datetime.datetime.now() - datetime.timedelta(minutes=1)
        )

        assert terrareg.api_key.ApiKey.verify_key(plaintext_key, terrareg.api_key.ApiKeyType.PUBLISH) is None

    def test_invalid_type_raises(self):
        """Test invalid API key types are rejected."""
        with pytest.raises(terrareg.errors.InvalidApiKeyTypeError):
            terrareg.api_key.ApiKey.create(name='bad key', key_type='invalid-type')