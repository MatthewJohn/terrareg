
import json
import pytest

from test.integration.terrareg import TerraregIntegrationTest
from test.integration.terrareg.models.provider_source import test_provider_source
import terrareg.database
import terrareg.provider_source
import terrareg.provider_source_type


class TestBaseProviderSource(TerraregIntegrationTest):

    _CLASS = terrareg.provider_source.BaseProviderSource
    ADDITIONAL_CONFIG = {
        "test_config": "test_value",
    }

    @property
    def class_(self):
        """Return test class"""
        if self._CLASS is None:
            raise NotImplementedError
        return self._CLASS

    def test_name(self, test_provider_source):
        """Test name property"""
        # Ensure value from DB is correctly obtained from class initialisation
        assert test_provider_source.name == "Test Provider Source"
        # Ensure the value is obtained from member variable
        test_provider_source._name = "New Provider Name"
        assert test_provider_source.name == "New Provider Name"

    def test_api_name(self, test_provider_source):
        """Test api_name property"""
        # Ensure value from DB is correctly obtained
        assert test_provider_source.api_name == "test-provider-source"
        # Ensure the value is obtained from the cached DB row
        test_provider_source._cache_db_row = {"api_name": "mock-api-name"}
        assert test_provider_source.api_name == "mock-api-name"

    def test__config(self, test_provider_source):
        """Test _config property"""
        assert test_provider_source._config == self.ADDITIONAL_CONFIG

    def test___init__(self):
        """Test class initialisation"""
        instance = self.class_(name="Test Name")
        assert instance._name == "Test Name"
        assert instance._cache_db_row is None

    def test__get_db_row(self, test_provider_source):
        """Test _get_db_row."""
        assert test_provider_source._cache_db_row is None
        assert dict(test_provider_source._get_db_row()) == {
            'api_name': 'test-provider-source',
            'config': terrareg.database.Database.encode_blob(json.dumps(self.ADDITIONAL_CONFIG)),
            'name': 'Test Provider Source',
            'provider_source_type': terrareg.provider_source_type.ProviderSourceType.GITHUB,
        }

        # Ensure cached row has been updated
        assert dict(test_provider_source._cache_db_row) == {
            'api_name': 'test-provider-source',
            'config': terrareg.database.Database.encode_blob(json.dumps(self.ADDITIONAL_CONFIG)),
            'name': 'Test Provider Source',
            'provider_source_type': terrareg.provider_source_type.ProviderSourceType.GITHUB,
        }

        # Update cache DB row to ensure the cached version is
        # returned
        test_provider_source._cache_db_row = {"new_dict": "true"}
        assert test_provider_source._get_db_row() == {"new_dict": "true"}
