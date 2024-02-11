from tempfile import TemporaryDirectory
import unittest.mock
import os

import pytest

from test.integration.terrareg.fixtures import (
    test_provider_version, test_provider, test_repository,
    test_gpg_key, test_namespace, mock_provider_source,
    mock_provider_source_class, test_provider_category
)
import terrareg.provider_version_binary_model
import terrareg.provider_binary_types
import terrareg.errors
import terrareg.models
import terrareg.provider_model
import terrareg.provider_version_model
import terrareg.database
import terrareg.provider_binary_types
from test.integration.terrareg import TerraregIntegrationTest


class TestProviderVersionBinary(TerraregIntegrationTest):
    """Test ProviderVersionBinary"""

    @pytest.mark.parametrize('name, expected_os, expected_arch', [
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_amd64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_arm.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM),
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_arm64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_386.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386),

        ("terraform-provider-unittest-create-provider-name_6.4.1_windows_amd64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_windows_arm.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM),
        ("terraform-provider-unittest-create-provider-name_6.4.1_windows_arm64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_windows_386.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386),

        ("terraform-provider-unittest-create-provider-name_6.4.1_darwin_amd64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.DARWIN,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_darwin_arm.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.DARWIN,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM),
        ("terraform-provider-unittest-create-provider-name_6.4.1_darwin_arm64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.DARWIN,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_darwin_386.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.DARWIN,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386),

        ("terraform-provider-unittest-create-provider-name_6.4.1_freebsd_amd64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_freebsd_arm.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM),
        ("terraform-provider-unittest-create-provider-name_6.4.1_freebsd_arm64.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64),
        ("terraform-provider-unittest-create-provider-name_6.4.1_freebsd_386.zip",
         terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
         terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386),
    ])
    def test_create(self, name, expected_os, expected_arch, test_provider_version):
        """Test create"""
        with TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):

            os.mkdir(os.path.join(temp_dir, "providers"))
    
            provider_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                provider_version=test_provider_version,
                name=name,
                checksum="c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1",
                content=b"Some test Content"
            )
            assert isinstance(provider_binary, terrareg.provider_version_binary_model.ProviderVersionBinary)

            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                rows = conn.execute(db.provider_version_binary.select().where(
                    db.provider_version_binary.c.name==name
                )).all()

                assert len(rows) == 1
                row = rows[0]

            assert row["name"] == name
            assert row["provider_version_id"] == test_provider_version.pk
            assert row["checksum"] == "c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1"
            assert row["operating_system"] is expected_os
            assert row["architecture"] is expected_arch

            expected_binary_path = os.path.join(temp_dir, "providers", "some-organisation", "unittest-create-provider-name", "6.4.1", name)
            assert os.path.isfile(expected_binary_path)
            with open(expected_binary_path, "r") as fh:
                assert fh.read() == "Some test Content"

    @pytest.mark.parametrize('name, error', [
        # Invalid full name
        ("terraform-provide-unittest-create-provider-name_6.4.1_linux_amd64.zip", terrareg.errors.InvalidProviderBinaryNameError),
        # Invalid repo name
        ("terraform-provider-unittest-creat-provider-name_6.4.1_linux_amd64.zip", terrareg.errors.InvalidProviderBinaryNameError),
        # Incorrect version
        ("terraform-provider-unittest-create-provider-name_6.4.2_linux_amd64.zip", terrareg.errors.InvalidProviderBinaryNameError),
        # Invalid OS
        ("terraform-provider-unittest-create-provider-name_6.4.1_linax_amd64.zip", terrareg.errors.InvalidProviderBinaryOperatingSystemError),
        # Invalid arch
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_284.zip", terrareg.errors.InvalidProviderBinaryArchitectureError),
        # Invalid suffix
        ("terraform-provider-unittest-create-provider-name_6.4.1_linux_amd64.tar", terrareg.errors.InvalidProviderBinaryNameError),
    ])
    def test_create_invalid_name(self, name, error, test_provider_version):
        """Test create with invalid binary names"""
        with TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):

            os.mkdir(os.path.join(temp_dir, "providers"))
    
            with pytest.raises(error):
                terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                    provider_version=test_provider_version,
                    name=name,
                    checksum="c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1",
                    content=b"Some test Content"
                )

    def test_create_duplicate(self, test_provider_version):
        """Create create with duplicate provider binary"""
        with TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):

            terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                provider_version=test_provider_version,
                name="terraform-provider-unittest-create-provider-name_6.4.1_linux_amd64.zip",
                checksum="c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1",
                content=b"Some original test Content"
            )

            with pytest.raises(terrareg.errors.ProviderVersionBinaryAlreadyExistsError):
                terrareg.provider_version_binary_model.ProviderVersionBinary.create(
                    provider_version=test_provider_version,
                    name="terraform-provider-unittest-create-provider-name_6.4.1_linux_amd64.zip",
                    checksum="someotherchecksum",
                    content=b"Some test duplicate content"
                )

            provider_version_binaries = terrareg.provider_version_binary_model.ProviderVersionBinary.get_by_provider_version(provider_version=test_provider_version)
            assert len(provider_version_binaries) == 1
            assert provider_version_binaries[0].checksum == "c27f1263ae06f263d59eb1f172c7fe39f6d7a06771544d869cc272d94ed301d1"

            with open(os.path.join(temp_dir, provider_version_binaries[0].local_file_path.lstrip(os.path.sep)), "r") as fh:
                assert fh.read() == "Some original test Content"

    def test__insert_db_row(self, test_provider_version):
        """Test _insert_db_row method"""
        pk = terrareg.provider_version_binary_model.ProviderVersionBinary._insert_db_row(
            provider_version=test_provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM,
            name="unittest-binary-provider",
            checksum="abcdefg0987654"
        )

        assert isinstance(pk, int)

        db = terrareg.database.Database.get()
        with db.get_connection() as conn:
            res = conn.execute(db.provider_version_binary.select().where(db.provider_version_binary.c.id==pk)).all()
            assert len(res) == 1
            assert dict(res[0]) == {
                "id": pk,
                "name": "unittest-binary-provider",
                "checksum": "abcdefg0987654",
                "operating_system": terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
                "architecture": terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM,
                "provider_version_id": test_provider_version.pk
            }

    def test_get(self):
        """Test get method"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        valid_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert isinstance(valid_binary, terrareg.provider_version_binary_model.ProviderVersionBinary)
        assert valid_binary.checksum == "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"

        non_existent = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64
        )
        assert non_existent is None

    def test_get_by_provider_version(self):
        """Test get_by_provider_version"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        binaries = terrareg.provider_version_binary_model.ProviderVersionBinary.get_by_provider_version(provider_version=provider_version)
        assert isinstance(binaries, list)
        assert len(binaries) == 4

        # Sort binaries by checksum and check values
        binaries.sort(key=lambda x: x.checksum)
        assert binaries[0].name == "terraform-provider-multiple-versions_1.5.0_linux_amd64.zip"
        assert binaries[0].checksum == "a26d0401981bf2749c129ab23b3037e82bd200582ff7489e0da2a967b50daa98"
        assert binaries[0].operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX
        assert binaries[0].architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        assert binaries[1].name == "terraform-provider-multiple-versions_1.5.0_linux_arm64.zip"
        assert binaries[1].checksum == "bda5d57cf68ab142f5d0c9a5a0739577e24444d4e8fe4a096ab9f4935bec9e9a"
        assert binaries[1].operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.LINUX
        assert binaries[1].architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.ARM64
        assert binaries[2].name == "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip"
        assert binaries[2].checksum == "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"
        assert binaries[2].operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS
        assert binaries[2].architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        assert binaries[3].name == "terraform-provider-multiple-versions_1.5.0_darwin_amd64.zip"
        assert binaries[3].checksum == "e8bc51e741c45feed8d9d7eb1133ac0107152cab3c1db12e74495d4b4ec75a0c"
        assert binaries[3].operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.DARWIN
        assert binaries[3].architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64

    def test_name(self):
        """Test name property"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert test_provider_version_binary.name == "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip"

        test_provider_version_binary._cache_db_row = {
            "name": "unittestname"
        }
        assert test_provider_version_binary.name == "unittestname"

    def test_architecture(self):
        """Test architecture property"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert test_provider_version_binary.architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64

        test_provider_version_binary._cache_db_row = {
            "architecture": terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386
        }
        assert test_provider_version_binary.architecture is terrareg.provider_binary_types.ProviderBinaryArchitectureType.I386

    def test_operating_system(self):
        """Test operating_system property"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert test_provider_version_binary.operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS

        test_provider_version_binary._cache_db_row = {
            "operating_system": terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD
        }
        assert test_provider_version_binary.operating_system is terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.FREEBSD

    def test_checksum(self):
        """Test checksum property"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert test_provider_version_binary.checksum == "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf"

        test_provider_version_binary._cache_db_row = {
            "checksum": "some-other-checksum-unittest"
        }
        assert test_provider_version_binary.checksum == "some-other-checksum-unittest"

    def test_local_file_path(self):
        """Test local_file_path property"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        with unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', '/some/test/directory'):
            assert test_provider_version_binary.local_file_path == "/some/test/directory/providers/initial-providers/multiple-versions/1.5.0/terraform-provider-multiple-versions_1.5.0_windows_amd64.zip"

    def test_provider_version(self):
        """Test provider_version attribute"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        assert isinstance(test_provider_version_binary.provider_version, terrareg.provider_version_model.ProviderVersion)
        assert test_provider_version_binary.provider_version.pk == provider_version.pk

    def test___init__(self):
        """Test __init__ method"""
        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary(pk=5341)
        assert test_provider_version_binary._pk == 5341
        assert test_provider_version_binary._cache_db_row is None

    def test__get_db_row(self):
        """Test _get_db_row method"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        pk = test_provider_version_binary._pk
        test_provider_version_binary._cache_db_row = None

        assert dict(test_provider_version_binary._get_db_row()) == {
            "id": pk,
            "provider_version_id": provider_version.pk,
            "name": "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip",
            "checksum": "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf",
            "operating_system": terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            "architecture": terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        }
        assert dict(test_provider_version_binary._cache_db_row) == {
            "id": pk,
            "provider_version_id": provider_version.pk,
            "name": "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip",
            "checksum": "c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf",
            "operating_system": terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            "architecture": terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        }

        test_provider_version_binary._cache_db_row = {"test": "row"}
        assert test_provider_version_binary._get_db_row() == {"test": "row"}

        # Test non-existent
        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary(pk=5341000)
        assert test_provider_version_binary._get_db_row() is None
        assert test_provider_version_binary._cache_db_row is None

    def test_create_local_binary(self):
        """Test create_local_binary"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )
        with TemporaryDirectory() as temp_dir, \
                unittest.mock.patch('terrareg.config.Config.DATA_DIRECTORY', temp_dir):

            test_provider_version_binary.create_local_binary(b"Some test local binary content\nisHere!")

            provider_binary_path = os.path.join(temp_dir, "providers", "initial-providers", "multiple-versions", "1.5.0", "terraform-provider-multiple-versions_1.5.0_windows_amd64.zip")
            assert os.path.isfile(provider_binary_path)
            with open(provider_binary_path, "r") as fh:
                assert fh.read() == "Some test local binary content\nisHere!"


    def test_get_api_outline(self) -> dict:
        """Test get_api_outline"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion.get(provider=provider, version="1.5.0")

        test_provider_version_binary = terrareg.provider_version_binary_model.ProviderVersionBinary.get(
            provider_version=provider_version,
            operating_system_type=terrareg.provider_binary_types.ProviderBinaryOperatingSystemType.WINDOWS,
            architecture_type=terrareg.provider_binary_types.ProviderBinaryArchitectureType.AMD64
        )

        assert test_provider_version_binary.get_api_outline() == {
            'arch': 'amd64',
            'download_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_windows_amd64.zip',
            'filename': 'terraform-provider-multiple-versions_1.5.0_windows_amd64.zip',
            'os': 'windows',
            'protocols': ['5.0'],
            'shasum': 'c2d859efacc3fbe1662bef92c80ce32c966834597625914592891eaf564af4bf',
            'shasums_signature_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_SHA256SUMS.sig',
            'shasums_url': 'https://github.example.com/initial-providers/terraform-provider-multiple-versions/releases/download/v1.5.0/terraform-provider-multiple-versions_1.5.0_SHA256SUMS',
            'signing_keys': {
                'gpg_public_keys': [
                    {
                        'ascii_armor': '-----BEGIN PGP PUBLIC KEY BLOCK-----\n'
                            '\n'
                            'mI0EZUHt7QEEAKgSXXCkqShvE54omLsE0Gzu/Es2Nelwnps8ETlcHPKag0VlZch/\n'
                            '0HPyF3hGsdZM7GB1il7fGCGw6Urkmci7XkRj2M09QtAvE2YPOqfNfMvHQrLIAkBV\n'
                            'lP/4xIBnGMmsUYVMAeo0DiDdFf3Q3pIbWDhd7+OCPKh80F/pYM1Rm4qnABEBAAG0\n'
                            'UVRlc3QgVGVycmFyZWcgVGVzdHMgKFRlc3QgS2V5IGZvciB0ZXJyYXJlZyBUZXN0\n'
                            'cykgPHRlcnJhcmVnLXRlc3RzQGNvbGFtYWlsLmNvLnVrPojOBBMBCgA4FiEEIadO\n'
                            'Tj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwMFCwkIBwIGFQoJCAsCBBYCAwECHgEC\n'
                            'F4AACgkQNN43SsNkDNtkywP/SR8U/c3gzAY4w0KF3ZG5sBJqrBfdA2d2R//Bsjvz\n'
                            'jRCpGdaXVBJG2FFyfl5QLLhC56rS6nsX6vcXkrRGQtYG6Bhroo6eWjVnyT1RMM+A\n'
                            'wD5uwCijPlSdl82q91aFQk3jwqNoe4/gr9ERHagx3MAgMTEhIzPaKpGHtL7TPM+B\n'
                            'nOi4jQRlQe3tAQQAxCeKNhBAv13aXeSvPI1JKW9pcg5g9Hfd4s/qj82/0hE/Kfgt\n'
                            '4u7RGOEe7q1WgKirtoiv/XSpwKMSlXtt9AH8lbgkveiJ3V+DqJxdzCm42Zlyvg9Z\n'
                            '9sqLz6XOAyMkv44U1x182KMipuuethRmSemN8jthc4Bh5iEM/l7460IyRk8AEQEA\n'
                            'AYi2BBgBCgAgFiEEIadOTj/f5DhTK9WENN43SsNkDNsFAmVB7e0CGwwACgkQNN43\n'
                            'SsNkDNtn+AP+Pm3+u+if0BExYTMKJ0/dU4ICWBkyuuMDkQlz8oOn9/w9EYvkqR/r\n'
                            'QypRou1K0KbLxBCz0vqAM7KLXe0rKwZZ3eWSThiwTJkFlkJsUgwMqqROteYmWm3S\n'
                            'MK0hMLszB/mfN0Q2DW4U0tWslehdEA+aaccwA5PVFKdkA12ImK500TY=\n'
                            '=EL4W\n'
                            '-----END PGP PUBLIC KEY BLOCK-----',
                        'key_id': '34DE374AC3640CDB',
                        'source': '',
                        'source_url': None,
                        'trust_signature': ''
                    }
                ]
            }
        }
