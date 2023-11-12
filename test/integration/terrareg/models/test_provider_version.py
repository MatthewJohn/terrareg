
import contextlib
from datetime import datetime
import json
from typing import Union, List
import os
import re
import unittest.mock

import pytest
import sqlalchemy
import semantic_version
from terrareg.constants import PROVIDER_EXTRACTION_VERSION

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
