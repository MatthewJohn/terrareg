
import contextlib
from datetime import datetime
import json
from tempfile import TemporaryDirectory
from typing import Union, List
import os
import re
import unittest.mock

import pytest
import sqlalchemy
import semantic_version
from terrareg.constants import PROVIDER_EXTRACTION_VERSION
from test import mock_create_audit_event

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.errors
import terrareg.utils
import terrareg.provider_model
import terrareg.database
import terrareg.audit
import terrareg.audit_action
import terrareg.models
import terrareg.provider_search
import terrareg.provider_version_documentation_model
import terrareg.provider_version_binary_model
import terrareg.analytics
import terrareg.provider_version_model


class TestProviderVersion(TerraregIntegrationTest):

    @pytest.mark.parametrize('version', [
        'astring',
        '',
        '1',
        '1.1',
        '.23.1',
        '1.1.1.1',
        '1.1.1.',
        '1.2.3-dottedsuffix1.2',
        '1.2.3-invalid-suffix',
        '1.0.9-'
    ])
    def test___validate_version_invalid(self, version):
        """Test invalid provider versions"""
        with pytest.raises(terrareg.errors.InvalidVersionError):
            terrareg.provider_version_model.ProviderVersion._validate_version(version=version)

    @pytest.mark.parametrize('version,beta', [
        ('1.1.1', False),
        ('13.14.16', False),
        ('1.10.10', False),
        ('01.01.01', False),  # @TODO Should this be allowed?
        ('1.2.3-alpha', True),
        ('1.2.3-beta', True),
        ('1.2.3-anothersuffix1', True),
        ('1.2.2-123', True)
    ])
    def test___validate_version_valid(self, version, beta):
        """Test valid provider versions"""
        assert terrareg.provider_version_model.ProviderVersion._validate_version(version=version) == beta

    def test_get(cls):
        """Test get method"""
        # Get valid version
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert isinstance(version_obj, terrareg.provider_version_model.ProviderVersion)
        assert version_obj._version == "1.5.0"
        assert version_obj._provider is provider_obj

        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.9.0")
        assert version_obj is None

        # Test with invalid version number
        with pytest.raises(terrareg.errors.InvalidVersionError):
            terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.9")

    def test_get_by_pk(self):
        """Test get_by_pk"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")
        pk = version_obj.pk

        version_obj = terrareg.provider_version_model.ProviderVersion.get_by_pk(pk=pk)
        assert isinstance(version_obj, terrareg.provider_version_model.ProviderVersion)
        assert version_obj._version == "1.5.0"

        # Ensure parent objects have been created correctly
        assert version_obj._provider.name == "test-initial"
        assert version_obj._provider._namespace._name == "initial-providers"

    @pytest.mark.parametrize('published_at, expected_value', [
        (datetime(year=2023, month=10, day=10, hour=6, minute=23, second=0), 'October 10, 2023'),
        (None, None),
    ])
    def test_publish_date_display(self, published_at, expected_value):
        """Return display view of date of provider published."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert isinstance(version_obj._get_db_row()["published_at"], datetime)

        version_obj._cache_db_row = {
            "published_at": published_at
        }
        assert version_obj.publish_date_display == expected_value

    def test_version(self):
        """Test version property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj._version == "1.5.0"
        assert version_obj.version == "1.5.0"

        version_obj._version = "2.3.2"
        assert version_obj.version == "2.3.2"

    def test_git_tag(self):
        """Test git_tag property."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.git_tag == "v1.5.0"

        # Modify DB column and ensure it is used
        version_obj._cache_db_row = {
            "git_tag": "v56.21.32"
        }
        assert version_obj.git_tag == "v56.21.32"

    def test_base_directory(self):
        """Test base_directory property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        with unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', '/tmp/some-dir'):
            assert version_obj.base_directory == "/tmp/some-dir/providers/initial-providers/test-initial/1.5.0"

    @pytest.mark.parametrize('version, expected', [
        ('1.1.0', False),
        ('1.1.0-beta', True),
    ])
    def test_beta(self, version, expected):
        """Test beta flag."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version=version)

        assert version_obj.beta is expected

    def test_pk(self):
        """Test pk property."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert isinstance(version_obj.pk, int)

        version_obj._cache_db_row = {"id": 123}
        assert version_obj.pk == 123

    def test_id(self):
        """Test id property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.id == "initial-providers/test-initial/1.5.0"

    @pytest.mark.parametrize('version, exists', [
        ('1.0.0', True),
        ('1.2.3', False),
    ])
    def test_exists(self, version, exists):
        """Test exists property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="to-delete")
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version=version)

        assert version_obj.exists is exists

        if exists:
            # Delete from database
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete(db.provider_version.c.id==version_obj.pk))
            
            version_obj._cache_db_row = None
            assert version_obj.exists is False

    def test_provider(self):
        """Test provider property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.provider is provider_obj

    @pytest.mark.parametrize('extraction_version, up_to_date', [
        (PROVIDER_EXTRACTION_VERSION, True),
        (PROVIDER_EXTRACTION_VERSION - 1, False),
        (PROVIDER_EXTRACTION_VERSION + 1, False)
    ])
    def test_provider_extraction_up_to_date(self, extraction_version, up_to_date):
        """Test provider_extraction_up_to_date"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

        version_obj._cache_db_row = {
            "extraction_version": extraction_version
        }
        assert version_obj.provider_extraction_up_to_date is up_to_date


    @pytest.mark.parametrize('version, expected_latest_version', [
        ('2.0.1', True),
        ('2.0.0', False),
        ('1.1.0', False),
    ])
    def test_is_latest_version(self, version, expected_latest_version):
        """Test is_latest_version property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version=version)
        assert version_obj.is_latest_version is expected_latest_version

    @pytest.mark.parametrize('version, expected_gpg_key_fingerprint', [
        ('2.0.0', '21A74E4E3FDFE438532BD58434DE374AC3640CDB'),
        ('2.0.1', '94CA72B7A2F4606A6C18211AE94A4F2AD628D926'),
    ])
    def test_gpg_key(self, version, expected_gpg_key_fingerprint):
        """Test gpg_key property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version=version)
        gpg_key = version_obj.gpg_key
        assert isinstance(gpg_key, terrareg.models.GpgKey)
        assert gpg_key.fingerprint == expected_gpg_key_fingerprint

    def test_checksum_file_name(self):
        """Test checksum_file_name property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.checksum_file_name == "terraform-provider-test-initial_1.5.0_SHA256SUMS"
    
    def test_checksum_signature_file_name(self):
        """Test checksum_signature_file_name property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.checksum_signature_file_name == "terraform-provider-test-initial_1.5.0_SHA256SUMS.sig"

    def test_manifest_file_name(self):
        """Test manifest_file_name property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        assert version_obj.manifest_file_name == "terraform-provider-test-initial_1.5.0_manifest.json"

    def test___init__(self):
        """Test __init__ method."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="1.5.0")

        assert version_obj._extracted_beta_flag == False
        assert version_obj._provider is provider_obj
        assert version_obj._version == "1.5.0"
        assert version_obj._cache_db_row is None

        beta_version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="1.5.0-beta")
        assert beta_version_obj._extracted_beta_flag is True

    def test___eq__(self):
        """Test __eq__ method"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        first_version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        second_version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="2.0.1")
        second_provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        different_provider_version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=second_provider_obj, version="1.5.0")

        assert first_version_obj == terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

        assert first_version_obj != second_version_obj
        assert first_version_obj != different_provider_version_obj

    def test___gt__(self):
        """Test __gt__ method"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        lower = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="2.0.0")
        upper = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="2.0.1")

        assert upper > lower
        assert not (lower > upper)

    def test___lt__(self):
        """Test __lt__ method"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        lower = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="2.0.0")
        upper = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="2.0.1")

        assert lower < upper
        assert not (upper < lower)

    def test__get_db_row(self):
        """Test _get_db_row method"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="1.5.0")

        assert version_obj._cache_db_row is None

        db_row = version_obj._get_db_row()
        assert version_obj._cache_db_row is db_row

        db_row = dict(db_row)
        assert isinstance(db_row["id"], int)
        assert isinstance(db_row["gpg_key_id"], int)
        assert isinstance(db_row["provider_id"], int)
        assert isinstance(db_row["published_at"], datetime)

        db_row["id"] = 55
        db_row["gpg_key_id"] = 1
        db_row["provider_id"] = 1
        db_row["published_at"] = datetime(2023, 11, 13, 5, 43, 30, 897287)
        assert db_row == {
            'beta': False,
            'extraction_version': None,
            'git_tag': 'v1.5.0',
            'gpg_key_id': 1,
            'id': 55,
            'protocol_versions': None,
            'provider_id': 1,
            'published_at': datetime(2023, 11, 13, 5, 43, 30, 897287),
            'version': '1.5.0',
        }

    def test_generate_file_name_from_suffix(self):
        """Test generate_file_name_from_suffix method"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

        assert version_obj.generate_file_name_from_suffix("test_file_suffix") == "terraform-provider-test-initial_1.5.0_test_file_suffix"

    def test_create_data_directory(self):
        """Test create_data_directory."""
        with TemporaryDirectory() as temp_data_dir, \
                unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_data_dir):

            os.mkdir(os.path.join(temp_data_dir, "providers"))
            
            namespace_obj = terrareg.models.Namespace.get("initial-providers")
            provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
            version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
            version_obj.create_data_directory()

            assert os.path.isdir(os.path.join(temp_data_dir, "providers", "initial-providers", "test-initial", "1.5.0"))

    def test_create_extraction_wrapper(self):
        """Test create_extraction_wrapper"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="to-delete")
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="300.6.3")
        gpg_key = terrareg.models.GpgKey.get_by_fingerprint("21A74E4E3FDFE438532BD58434DE374AC3640CDB")

        assert version_obj._get_db_row() is None

        with version_obj.create_extraction_wrapper(git_tag="v3.6.3", gpg_key=gpg_key):

            db_row = version_obj._get_db_row()
            assert db_row is not None
            assert db_row["git_tag"] == "v3.6.3"
            assert db_row["gpg_key_id"] == gpg_key.pk
            assert db_row["published_at"] is None

            # Force refresh of provider data and ensure the new version is not marked as a new versino
            provider_obj._cache_db_row = None
            assert provider_obj._get_db_row()["latest_version_id"] != db_row["id"]

        # Ensure that once outside of context, the new version has published_at and latest_version_id
        # has been updated
        version_obj._cache_db_row = None
        assert isinstance(version_obj._get_db_row()["published_at"], datetime)
        provider_obj._cache_db_row = None
        assert provider_obj._get_db_row()["latest_version_id"] == db_row["id"]

    def test_prepare_version(self):
        """Test prepare_version"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="to-delete")
        gpg_key = terrareg.models.GpgKey.get_by_fingerprint("21A74E4E3FDFE438532BD58434DE374AC3640CDB")

        with unittest.mock.patch("terrareg.provider_version_model.ProviderVersion.create_data_directory", unittest.mock.MagicMock()) as mock_create_data_directory, \
                unittest.mock.patch("terrareg.provider_version_model.ProviderVersion._create_db_row", unittest.mock.MagicMock()) as mock_create_db_row:
            
            version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="1.21.3")
            version_obj.prepare_version(git_tag="vunittest-git-tag", gpg_key=gpg_key)

            mock_create_data_directory.assert_called_once_with()
            mock_create_db_row.assert_called_once_with(git_tag="vunittest-git-tag", gpg_key=gpg_key)

        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            audit_row = dict(conn.execute(db.audit_history.select().order_by(db.audit_history.c.timestamp.desc()).limit(1)).first())
            assert audit_row["action"] == terrareg.audit_action.AuditAction.PROVIDER_VERSION_INDEX
            assert audit_row["object_id"] == "initial-providers/to-delete/1.21.3"
            assert audit_row["object_type"] == "ProviderVersion"

    @pytest.mark.parametrize('provider, version, latest', [
        ('to-delete', '523.2.1', True),
        ('to-delete', '5.2.0', False),
        ('empty-provider-publish', '5.2.0', True),
    ])
    def test_publish(self, provider, version, latest):
        """Test publish."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name=provider)
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version=version)
        gpg_key = terrareg.models.GpgKey.get_by_fingerprint("21A74E4E3FDFE438532BD58434DE374AC3640CDB")

        version_obj.prepare_version(git_tag=f"v{version}", gpg_key=gpg_key)

        # Obtain previous latest version from provider
        previous_latest_version_id = provider_obj._get_db_row()["latest_version_id"]

        version_obj.publish()

        provider_obj._cache_db_row = None
        new_latest_version_id = provider_obj._get_db_row()["latest_version_id"]

        if latest:
            assert new_latest_version_id != previous_latest_version_id
            assert new_latest_version_id == version_obj.pk
        else:
            assert new_latest_version_id == previous_latest_version_id
            assert new_latest_version_id != version_obj.pk

    def test_get_total_downloads(self):
        """Test get_total_downloads"""
        with unittest.mock.patch("terrareg.analytics.ProviderAnalytics.get_provider_total_downloads", unittest.mock.MagicMock(return_value=12345)) as mock_get_provider_total_downloads:
            namespace_obj = terrareg.models.Namespace.get("initial-providers")
            provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
            version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

            assert version_obj.get_total_downloads() == 12345
            mock_get_provider_total_downloads.assert_called_once_with(provider=provider_obj)

    def test_get_downloads(self):
        """Test get_downloads."""
        with unittest.mock.patch("terrareg.analytics.ProviderAnalytics.get_provider_version_total_downloads", unittest.mock.MagicMock(return_value=12345)) as mock_get_provider_version_total_downloads:
            namespace_obj = terrareg.models.Namespace.get("initial-providers")
            provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
            version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

            assert version_obj.get_downloads() == 12345
            mock_get_provider_version_total_downloads.assert_called_once_with(provider_version=version_obj)

    @pytest.mark.parametrize("db_row_value, expected_value", [
        (None, ["5.0"]),
        ('["5.0", "6.0"]', ["5.0", "6.0"]),
    ])
    def test_protocols(self, db_row_value, expected_value):
        """Test protocols property"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        version_obj._cache_db_row = {
            "protocol_versions": terrareg.database.Database.encode_blob(db_row_value)
        }

        assert version_obj.protocols == expected_value

    def test_update_attributes(self):
        """Test update_attributes"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="update-attributes")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.0.0")

        new_gpg_key = terrareg.models.GpgKey.get_by_fingerprint("94CA72B7A2F4606A6C18211AE94A4F2AD628D926")

        version_obj.update_attributes(
            protocol_versions=json.dumps(["5.2.1", "1.2.3"]),
            gpg_key_id=new_gpg_key.pk,
            extraction_version=23,
            published_at=datetime(year=2023, month=2, day=3, hour=23, minute=0, second=6),
            git_tag="v5.4.3",
            beta=True
        )

        assert dict(version_obj._get_db_row()) == {
            'beta': True,
            'extraction_version': 23,
            'git_tag': 'v5.4.3',
            'gpg_key_id': new_gpg_key.pk,
            'id': version_obj.pk,
            'protocol_versions': b'["5.2.1", "1.2.3"]',
            'provider_id': provider_obj.pk,
            'published_at': datetime(2023, 2, 3, 23, 0, 6),
            'version': '1.0.0',
        }

    def test_get_api_binaries_outline(self):
        """Test get_api_binaries_outline"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

        assert version_obj.get_api_binaries_outline() == {
            'platforms': [
                {'arch': 'amd64', 'os': 'linux'},
                {'arch': 'arm64', 'os': 'linux'},
                {'arch': 'amd64', 'os': 'windows'},
                {'arch': 'amd64', 'os': 'darwin'}
            ],
            'protocols': ['5.0'],
            'version': '1.5.0',
        }

    def test_get_api_outline(self):
        """Test get_api_outline"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="test-initial")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        db_row = dict(version_obj._get_db_row())
        db_row["published_at"] = datetime(year=2023, month=10, day=12, hour=2, minute=42, second=12)
        version_obj._cache_db_row = db_row

        assert version_obj.get_api_outline() == {
            'alias': None,
            'description': 'Test Initial Provider',
            'downloads': 0,
            'id': 'initial-providers/test-initial/1.5.0',
            'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
            'name': 'test-initial',
            'namespace': 'initial-providers',
            'owner': 'initial-providers',
            'published_at': '2023-10-12T02:42:12',
            'source': 'https://github.example.com/initial-providers/terraform-provider-test-initial',
            'tag': 'v1.5.0',
            'tier': 'community',
            'version': '1.5.0'
        }

    def test_get_v2_include(self):
        """Test get_v2_include"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")
        db_row = dict(version_obj._get_db_row())
        db_row["published_at"] = datetime(year=2023, month=10, day=12, hour=2, minute=42, second=12)
        db_row["id"] = 23
        version_obj._cache_db_row = db_row

        assert version_obj.get_v2_include() == {
            'type': 'provider-versions',
            'id': '23',
            'attributes': {
                'description': 'Test Multiple Versions',
                'downloads': 0,
                'published-at': '2023-10-12T02:42:12',
                'tag': 'v1.5.0',
                'version': '1.5.0'
            },
            'links': {'self': '/v2/provider-versions/23'}
        }

    def test_get_api_details(self) -> dict:
        """Test get_api_details."""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="multiple-versions")
        version_obj = terrareg.provider_version_model.ProviderVersion.get(provider=provider_obj, version="1.5.0")

        db_row = dict(version_obj._get_db_row())
        db_row["published_at"] = datetime(year=2023, month=10, day=12, hour=2, minute=42, second=12)
        version_obj._cache_db_row = db_row

        assert version_obj.get_api_details() == {
            'alias': None,
            'description': 'Test Multiple Versions',
            'docs': [
                {
                    'category': 'overview',
                    'id': '6344',
                    'language': 'hcl',
                    'path': 'index.md',
                    'slug': 'overview',
                    'subcategory': None,
                    'title': 'Overview'
                },
                {
                    'category': 'resources',
                    'id': '6345',
                    'language': 'hcl',
                    'path': 'data-sources/thing.md',
                    'slug': 'some_resource',
                    'subcategory': 'some-subcategory',
                    'title': 'multiple_versions_thing'
                },
                {
                    'category': 'resources',
                    'id': '6346',
                    'language': 'python',
                    'path': 'data-sources/thing.md',
                    'slug': 'some_resource',
                    'subcategory': 'some-subcategory',
                    'title': 'multiple_versions_thing'
                },
                {
                    'category': 'resources',
                    'id': '6347',
                    'language': 'hcl',
                    'path': 'resources/new-thing.md',
                    'slug': 'some_new_resource',
                    'subcategory': 'some-second-subcategory',
                    'title': 'multiple_versions_thing_new'
                }
            ],
            'downloads': 0,
            'id': 'initial-providers/multiple-versions/1.5.0',
            'logo_url': 'https://git.example.com/initalproviders/terraform-provider-test-initial.png',
            'name': 'multiple-versions',
            'namespace': 'initial-providers',
            'owner': 'initial-providers',
            'published_at': '2023-10-12T02:42:12',
            'source': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions',
            'tag': 'v1.5.0',
            'tier': 'community',
            'version': '1.5.0',
            'versions': [
                '2.0.1',
                '2.0.0',
                '1.5.0',
                '1.1.0',
                '1.1.0-beta',
                '1.0.0'
            ]
        }

    def test__create_db_row(self):
        """Test _create_db_row"""
        namespace_obj = terrareg.models.Namespace.get("initial-providers")
        provider_obj = terrareg.provider_model.Provider.get(namespace=namespace_obj, name="to-delete")
        version_obj = terrareg.provider_version_model.ProviderVersion(provider=provider_obj, version="10.20.30")
        gpg_key = terrareg.models.GpgKey.get_by_fingerprint("94CA72B7A2F4606A6C18211AE94A4F2AD628D926")

        assert version_obj._get_db_row() is None

        # Ensure row doesn't exist in DB
        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            assert conn.execute(db.provider_version.select().where(db.provider_version.c.version=="10.20.30")).first() is None

        version_obj._create_db_row(gpg_key=gpg_key, git_tag="v2.3.4-unittest")

        with db.get_connection() as conn:
            row = conn.execute(db.provider_version.select().where(db.provider_version.c.version=="10.20.30")).first()

        assert row is not None
        row = dict(row)
        row["id"] = 55
        assert row == {
            'beta': False,
            'extraction_version': None,
            'git_tag': 'v2.3.4-unittest',
            'gpg_key_id': gpg_key.pk,
            'id': 55,
            'protocol_versions': None,
            'provider_id': provider_obj.pk,
            'published_at': None,
            'version': '10.20.30',
        }
