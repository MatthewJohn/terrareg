
import json
from re import L
from typing import Dict, Union, Type, List
import unittest.mock
import pytest


from test.integration.terrareg import TerraregIntegrationTest
import terrareg.config
import terrareg.database
import terrareg.errors
import terrareg.provider_source_type
import terrareg.provider_source
import terrareg.provider_source.factory


class TestProviderSourceFactory(TerraregIntegrationTest):
    """Test ProviderSourceFactory class"""

    def test_get(self):
        """Test get method"""
        terrareg.provider_source.factory.ProviderSourceFactory._INSTANCE = None
        instance = terrareg.provider_source.factory.ProviderSourceFactory.get()
        assert isinstance(instance, terrareg.provider_source.factory.ProviderSourceFactory)
        assert terrareg.provider_source.factory.ProviderSourceFactory._INSTANCE is instance

        # Ensure subsequent calls returns cached version
        assert terrareg.provider_source.factory.ProviderSourceFactory.get() is instance

    def test_get_provider_classes(self) -> Dict[str, Type['terrareg.provider_source.BaseProviderSource']]:
        """Return all provider classes"""
        factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        factory._CLASS_MAPPING = None

        mappings = factory.get_provider_classes()
        assert list(mappings.keys()) == [terrareg.provider_source_type.ProviderSourceType.GITHUB]
        assert mappings[terrareg.provider_source_type.ProviderSourceType.GITHUB] is terrareg.provider_source.GithubProviderSource

        assert factory._CLASS_MAPPING is mappings

        # Ensure subsequent calls returns cached version
        assert factory.get_provider_classes() is mappings

    @pytest.mark.parametrize('type_, expected_result', [
        (terrareg.provider_source_type.ProviderSourceType.GITHUB, terrareg.provider_source.GithubProviderSource),
        ("does_not_exist", None),
        (None, None),
    ])
    def test_get_provider_source_class_by_type(self, type_, expected_result):
        """Test get_provider_source_class_by_type"""
        factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        assert factory.get_provider_source_class_by_type(type_=type_) is expected_result

    @pytest.mark.parametrize('call_name, provider_name, provider_type, expected_class_type', [
        ('test-provider-source', 'test-provider-source', terrareg.provider_source_type.ProviderSourceType.GITHUB, terrareg.provider_source.GithubProviderSource),
        # Invalid name
        ('does-not-exist', 'test-provider-source', terrareg.provider_source_type.ProviderSourceType.GITHUB, None),
    ])
    def test_get_provider_source_by_name(self, call_name, provider_name, provider_type, expected_class_type):
        """Test get_provider_source_by_name"""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name=provider_name,
                api_name="ut-name",
                provider_source_type=provider_type,
                config=db.encode_blob(json.dumps({}))
            ))

        try:
            factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
            res = factory.get_provider_source_by_name(name=call_name)
            if expected_class_type is None:
                assert res is None
            else:
                assert isinstance(res, expected_class_type)

        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.api_name=="ut-name"))

    @pytest.mark.parametrize('call_api_name, provider_api_name, provider_type, expected_class_type', [
        ('test-provider-source', 'test-provider-source', terrareg.provider_source_type.ProviderSourceType.GITHUB, terrareg.provider_source.GithubProviderSource),
        # Invalid name
        ('does-not-exist', 'test-provider-source', terrareg.provider_source_type.ProviderSourceType.GITHUB, None),
    ])
    def test_get_provider_source_by_api_name(self, call_api_name, provider_api_name, provider_type, expected_class_type):
        """Test get_provider_source_by_api_name"""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name="Provider Name",
                api_name=provider_api_name,
                provider_source_type=provider_type,
                config=db.encode_blob(json.dumps({}))
            ))

        try:
            factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
            res = factory.get_provider_source_by_api_name(api_name=call_api_name)
            if expected_class_type is None:
                assert res is None
            else:
                assert isinstance(res, expected_class_type)

        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name=="Provider Name"))

    def test_get_all_provider_sources(self):
        """Test get_all_provider_sources"""
        db = terrareg.database.Database.get()
        # Create two test providers
        with db.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name="Test Provider 1",
                api_name="prov-1",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({}))
            ))
            conn.execute(db.provider_source.insert().values(
                name="Test Provider 2",
                api_name="prov-2",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob(json.dumps({}))
            ))
        try:
            factory = terrareg.provider_source.factory.ProviderSourceFactory.get()

            res = factory.get_all_provider_sources()
            assert len(res) == 2

            for itx in res:
                assert isinstance(itx, terrareg.provider_source.GithubProviderSource)

            assert sorted([itx.name for itx in res]) == ['Test Provider 1', 'Test Provider 2']
            assert sorted([itx.api_name for itx in res]) == ["prov-1", "prov-2"]

        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.api_name.in_(["prov-1", "prov-2"])))

    @pytest.mark.parametrize('name, expected_api_name', [
        # Empty values
        (None, None),
        ('', None),
        # Unchanged
        ('test-name', 'test-name'),
        # Lower case
        ('TestName', 'testname'),
        # Space replacement
        ('test name', 'test-name'),
        # Special chars
        ('test@name', 'testname'),
        # Numberes
        ('testname1234', 'testname1234'),
        # Leading/trailing space
        (' test', 'test'),
        ('test ', 'test'),
        # Leading/trailing dash
        ('-test', 'test'),
        ('test-', 'test'),
        # Combined test
        (' -This 15 a Sp3c!@L Test CASE+=  ', 'this-15-a-sp3cl-test-case'),
    ])
    def test__name_to_api_name(self, name, expected_api_name):
        """Test _name_to_api_name method"""
        factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
        assert factory._name_to_api_name(name) == expected_api_name

    def test_initialise_from_config_create(self):
        """Test initialise_from_config."""

        with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES',
                                 '[{"name": "Test Create 1", "type": "github", "config": {"some_test": "config"}}]'), \
                unittest.mock.patch(
                    'terrareg.provider_source.GithubProviderSource.generate_db_config_from_source_config',
                    unittest.mock.MagicMock()) as mock_generate_db_config_from_source_config:

            try:

                mock_generate_db_config_from_source_config.return_value = {"test_db_config": "here"}

                factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
                factory.initialise_from_config()

                # Ensure generate_db_config_from_source_config was called
                mock_generate_db_config_from_source_config.assert_called_once_with({"name": "Test Create 1", "type": "github", "config": {"some_test": "config"}})

                result = factory.get_provider_source_by_name("Test Create 1")
                assert isinstance(result, terrareg.provider_source.GithubProviderSource)
                assert result.name == "Test Create 1"
                assert result.api_name == "test-create-1"
                assert result._config == {"test_db_config": "here"}

            finally:
                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    conn.execute(db.provider_source.delete(db.provider_source.c.name=="Test Create 1"))

    @pytest.mark.parametrize('config, expected_error', [
        # Invalid type
        ('[{"name": "Test Provider 1", "type": "invalid Type", "config": {}}]', "Invalid provider source type. Valid types: github"),
        # No type
        ('[{"name": "Test Provider 1", "config": {}}]', "Provider source config does not contain required attribute: type"),
        # Missing config
        ('[{"name": "Test Provider 1", "type": "github"}]', "Missing required Github provider source config: base_url"),
        # Invalid name
        ('[{"name": "  --  ", "type": "github"}]', 'Invalid provider source config: Name must contain some alphanumeric characters'),
        # Mising name
        ('[{"type": "gihtub", "config": {}}]', "Provider source config does not contain required attribute: name"),
        # Invalid JSON
        ('{"invalid JSON', "Provider source config is not a valid JSON list of objects"),
        ('{}', "Provider source config is not a valid JSON list of objects"),
        ('[["Hi"]]', "Provider source config is not a valid JSON list of objects"),
    ])
    def test_initialise_from_config_invalid_config(self, config, expected_error):
        """Test initialise_from_config with invalid configurations."""

        with unittest.mock.patch('terrareg.config.Config.PROVIDER_SOURCES', config):
            factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
            with pytest.raises(terrareg.errors.InvalidProviderSourceConfigError) as err:
                factory.initialise_from_config()

            assert err.value.args[0] == expected_error

    def test_initialise_from_config_update_existing(self):
        """Test initialise_from_config updating pre-existing provider source."""
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            conn.execute(db.provider_source.insert().values(
                name="Test Pre-existing",
                api_name="test-pre-existing",
                provider_source_type=terrareg.provider_source_type.ProviderSourceType.GITHUB,
                config=db.encode_blob('{"old_config": "value"}')
            ))

        try:
            with unittest.mock.patch(
                        'terrareg.config.Config.PROVIDER_SOURCES',
                        '[{"name": "Test Pre-existing", "type": "github", "config": {"new": "config"}}]'), \
                    unittest.mock.patch(
                        'terrareg.provider_source.GithubProviderSource.generate_db_config_from_source_config',
                        unittest.mock.MagicMock()) as mock_generate_db_config_from_source_config:

                mock_generate_db_config_from_source_config.return_value = {"new": "generated_config"}

                factory = terrareg.provider_source.factory.ProviderSourceFactory.get()
                factory.initialise_from_config()

                mock_generate_db_config_from_source_config.assert_called_once_with({
                    "name": "Test Pre-existing", "type": "github", "config": {"new": "config"}
                })

                with db.get_connection() as conn:
                    res = conn.execute(db.provider_source.select(db.provider_source.c.api_name=="test-pre-existing")).all()
                    assert len(res) == 1
                    assert dict(res[0]) == {
                        "name": "Test Pre-existing",
                        "api_name": "test-pre-existing",
                        "provider_source_type": terrareg.provider_source_type.ProviderSourceType.GITHUB,
                        "config": db.encode_blob('{"new": "generated_config"}')
                    }
        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name=="Test Pre-existing"))

    def test_initialise_from_config_duplicate(self):
        """Test initialise_from_config with a duplicate provider defined in config."""
        db = terrareg.database.Database.get()

        try:
            with unittest.mock.patch(
                        'terrareg.config.Config.PROVIDER_SOURCES',
                        '[{"name": "Test Duplicate", "type": "github", "config": {"first": "config"}}, {"name": "Test Duplicate", "type": "github", "config": {"second": "config"}}]'), \
                    unittest.mock.patch(
                        'terrareg.provider_source.GithubProviderSource.generate_db_config_from_source_config',
                        unittest.mock.MagicMock()) as mock_generate_db_config_from_source_config:

                def generate_db_config(config):
                    """Return orignal config dict as DB config"""
                    return config["config"]
                mock_generate_db_config_from_source_config.side_effect = generate_db_config

                factory = terrareg.provider_source.factory.ProviderSourceFactory.get()

                with pytest.raises(terrareg.errors.InvalidProviderSourceConfigError):
                    factory.initialise_from_config()

                mock_generate_db_config_from_source_config.assert_called_once_with({
                    "name": "Test Duplicate", "type": "github", "config": {"first": "config"}
                })

                with db.get_connection() as conn:
                    res = conn.execute(db.provider_source.select(db.provider_source.c.api_name=="test-duplicate")).all()
                    assert len(res) == 1
                    assert dict(res[0]) == {
                        "name": "Test Duplicate",
                        "api_name": "test-duplicate",
                        "provider_source_type": terrareg.provider_source_type.ProviderSourceType.GITHUB,
                        "config": db.encode_blob('{"first": "config"}')
                    }
        finally:
            with db.get_connection() as conn:
                conn.execute(db.provider_source.delete(db.provider_source.c.name=="Test Duplicate"))
