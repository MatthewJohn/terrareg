import datetime

import pytest

import terrareg.errors
import terrareg.models
from test.integration.terrareg import TerraregIntegrationTest


class TestApiKey(TerraregIntegrationTest):
    """Test ApiKey model."""

    def test_create_verify_mark_used_and_revoke(self):
        """Test API keys can be created, verified, marked used and revoked."""
        api_key, plaintext_key = terrareg.models.ApiKey.create(
            name='CI upload key',
            key_type=terrareg.models.ApiKeyType.UPLOAD,
            created_by='admin'
        )

        assert api_key.name == 'CI upload key'
        assert api_key.key_type == terrareg.models.ApiKeyType.UPLOAD.value
        assert api_key.key_prefix == plaintext_key[:terrareg.models.ApiKey.PREFIX_LENGTH]
        assert api_key.last_used_at is None

        verified_key = terrareg.models.ApiKey.verify_key(plaintext_key, terrareg.models.ApiKeyType.UPLOAD)
        assert verified_key.pk == api_key.pk

        verified_key.mark_used()
        assert terrareg.models.ApiKey.get(api_key.pk).last_used_at is not None

        verified_key.revoke()
        assert terrareg.models.ApiKey.verify_key(plaintext_key, terrareg.models.ApiKeyType.UPLOAD) is None

    def test_expired_keys_are_not_verified(self):
        """Test expired API keys no longer authenticate."""
        _, plaintext_key = terrareg.models.ApiKey.create(
            name='expired key',
            key_type=terrareg.models.ApiKeyType.PUBLISH,
            expires_at=datetime.datetime.now() - datetime.timedelta(minutes=1)
        )

        assert terrareg.models.ApiKey.verify_key(plaintext_key, terrareg.models.ApiKeyType.PUBLISH) is None

    def test_invalid_type_raises(self):
        """Test invalid API key types are rejected."""
        with pytest.raises(terrareg.errors.InvalidApiKeyTypeError):
            terrareg.models.ApiKey.create(name='bad key', key_type='invalid-type')