
from asyncio import subprocess
from io import BytesIO
import json
import os
from re import L
from subprocess import check_output
import tarsafe
from tempfile import TemporaryDirectory
import unittest.mock
import base64

from typing import ContextManager
import contextlib

import pytest
from terrareg.provider_binary_types import ProviderBinaryArchitectureType, ProviderBinaryOperatingSystemType

import terrareg.provider_version_model
import terrareg.repository_model
import terrareg.provider_source.repository_release_metadata
import terrareg.models
import terrareg.config
import terrareg.provider_documentation_type
import terrareg.provider_version_documentation_model
import terrareg.module_extractor
import terrareg.provider_model
import terrareg.provider_version_binary_model
from terrareg.errors import (
    InvalidChecksumFileError, InvalidProviderManifestFileError, InvalidReleaseArtifactChecksumError, MissingReleaseArtifactError, MissingSignureArtifactError,
    UnableToObtainReleaseSourceError
)

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.provider_extractor
import terrareg.database
import terrareg.errors


@pytest.fixture
def test_provider_version_wrapper():
    """Create test provider version, create_extraction_wrapper and provider extractor instance"""
    namespace = terrareg.models.Namespace.get("initial-providers")
    provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
    provider_version = terrareg.provider_version_model.ProviderVersion(provider=provider, version='1.9.4')
    gpg_key = terrareg.models.GpgKey.get_by_fingerprint(fingerprint="A0FC4319ABAF9C28A16821DF4F3072E58D16FF6D")

    @contextlib.contextmanager
    def inner() -> ContextManager[terrareg.provider_extractor.ProviderExtractor]:
        provider_version_id = None
        try:
            release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="Release v1.9.4",
                tag="v1.9.4",
                archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
                commit_hash="abcdefg123455",
                provider_id="unittest-release-id",
                release_artifacts=[
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="terraform-provider-multiple-versions_1.9.4_windows_arm64.zip", provider_id="previous-provider-id"),
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="terraform-provider-multiple-versions_1.9.4_manifest.json", provider_id="metadata-provider-id"),
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="terraform-provider-multiple-versions_1.9.4_linux_amd64.zip", provider_id="another-provider-id-"),
                ]
            )

            with provider_version.create_extraction_wrapper(git_tag="1.9.4", gpg_key=gpg_key):
                provider_version_id = provider_version.pk
                provider_extractor = terrareg.provider_extractor.ProviderExtractor(
                    provider_version=provider_version,
                    release_metadata=release_metadata
                )
                yield provider_extractor

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                if provider_version_id:
                    conn.execute(db.provider_version_binary.delete().where(
                        db.provider_version_binary.c.provider_version_id==provider_version_id
                    ))
                    conn.execute(db.provider_version_documentation.delete().where(
                        db.provider_version_documentation.c.provider_version_id==provider_version_id
                    ))
                conn.execute(db.provider_version.delete().where(
                    db.provider_version.c.provider_id==provider.pk,
                    db.provider_version.c.version=="1.9.4"
                ))
    return inner


class TestProviderExtractor(TerraregIntegrationTest):
    """Test ProviderExtractor"""

    def test_obtain_gpg_key(self):
        """"Test obtain_gpg_key"""

        # To generate these, place the source file (shasums) into shasums file
        # Import GPG key from test_data (provider_extractor test key)
        # Generate signature using (--detach-sig vs --sign is important!):
        # gpg -u A0FC4319ABAF9C28A16821DF4F3072E58D16FF6D --output shasums.sig --detach-sig ./shasums
        # base64 ./shasums.sig
        artifacts = [
            """4e13e517a3b6f474b734559c96f4fc01678ea5299b5c61844a2747727a52e80f  ./requirements-dev.txt
aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  ./requirements.txt
""".encode('utf-8'),
            base64.b64decode("""
iLMEAAEKAB0WIQSg/EMZq6+cKKFoId9PMHLljRb/bQUCZVsR+gAKCRBPMHLljRb/bXmpA/9Ycl/a
9ZKFevCamJLjMxw2K7OV12hWdR5X5pZ/Rse1gAOYQNaSbKwchM0ChDh/nrFMYzvErHsw/he8OjOK
G3KtIxGITPvTgjL7Zj0OxJSQAAgQN/bmDNM/jxhYevNsJjqnHeSBHm7U6IsLHFKNiSDj1c2yom4p
UnkCiCt3juqNNA==
""".strip())
        ]

        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        mock_download_artifact = unittest.mock.MagicMock(side_effect=artifacts)

        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="", provider_id="unittest-shasum-id")
            ]
        )

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact):
            gpg_key = terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                provider=provider, namespace=namespace, release_metadata=release_metadata)

        assert gpg_key is not None
        assert isinstance(gpg_key, terrareg.models.GpgKey)
        assert gpg_key.fingerprint == "A0FC4319ABAF9C28A16821DF4F3072E58D16FF6D"

        mock_download_artifact.assert_has_calls(calls=[
            unittest.mock.call(provider=provider, release_metadata=release_metadata, file_name='terraform-provider-multiple-versions_1.9.4_SHA256SUMS'),
            unittest.mock.call(provider=provider, release_metadata=release_metadata, file_name='terraform-provider-multiple-versions_1.9.4_SHA256SUMS.sig'),
        ])

    def test_obtain_gpg_key_signature_mismatch(self):
        """"Test obtain_gpg_key with data not matching signature"""

        artifacts = [
            """Some other data""".encode('utf-8'),
            base64.b64decode("""
iLMEAAEKAB0WIQSg/EMZq6+cKKFoId9PMHLljRb/bQUCZVsR+gAKCRBPMHLljRb/bXmpA/9Ycl/a
9ZKFevCamJLjMxw2K7OV12hWdR5X5pZ/Rse1gAOYQNaSbKwchM0ChDh/nrFMYzvErHsw/he8OjOK
G3KtIxGITPvTgjL7Zj0OxJSQAAgQN/bmDNM/jxhYevNsJjqnHeSBHm7U6IsLHFKNiSDj1c2yom4p
UnkCiCt3juqNNA==
""".strip())
        ]

        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        mock_download_artifact = unittest.mock.MagicMock(side_effect=artifacts)

        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="", provider_id="unittest-shasum-id")
            ]
        )

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact):
            gpg_key = terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                provider=provider, namespace=namespace, release_metadata=release_metadata)
            
            assert gpg_key is None

    def test_obtain_gpg_key_non_existent_key(self):
        """"Test obtain_gpg_key with data not matching signature"""

        artifacts = [
            """4e13e517a3b6f474b734559c96f4fc01678ea5299b5c61844a2747727a52e80f  ./requirements-dev.txt
aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  ./requirements.txt
""".encode('utf-8'),
            base64.b64decode("""
iQIzBAABCgAdFiEEFuiKD2WrkvH6VtkWAFU051tdoBYFAmVbihgACgkQAFU051tdoBbgqRAA46KT
ZNsU0Alr1eKly71gf1HkLN5hwv7Ru6u8cYHaIZ3hhUsEOzFuTR0ljX6sIiS6VnDGEu1lfxYYgymT
ePnQ9A7yzXWzRt5pCTT0ClLDdZdfpVJEeKkUD5jVwD1bd03wJ1AEx0wese8lZJwa7y1t0qSbRIQP
+WD8rCCgWi+JlKL2DlOClLEeTgc9005g88Z2v7CMSJ7qTkxSEwUOU8Bcm5ZKf+7CCE4mWgajoj9Q
mAv553bKoyVlM5cu+INWXrJsdqabEVzKeyQUT9vXkFDB1LPqlDRgcuRtROKFKfgY82/vIG3D6zng
eNvVZJmx+5nIuGGQRFsD3R9AQKFg/EPmRXDMoVnXr5RMnhLMvmi9JS9Xk2VwXXhJZAWOHzeXhFjJ
1D50L92yVjsAU+NUqf10X+BIfwhlhvw/AAo9a/4oIkBfIuPjps6EYHY1mOQ1A6ly8g96FtYl5D9E
OeNQBNdbIYZ3n0LbwC2CN35lUwFlfwkURDi6N/u5wvaYoRv+YyXmnjS1m2KP347EotLzEkMwt3e7
Iplvutu10Yry2IQJXQBZCsUxlKVDez7vYZudmpcSAK7UGIHtCvsDEI6s3FVx6kMW0BupzktZjufh
IsGEwymvzTwuBgzvVEPffRE6wS15ejnYZFR5kCxER8qqNmQ2XsmSZ9U8TX4P3McoZ4efs2M=
""".strip())
        ]

        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        mock_download_artifact = unittest.mock.MagicMock(side_effect=artifacts)

        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="", provider_id="unittest-shasum-id")
            ]
        )

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact):
            gpg_key = terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                provider=provider, namespace=namespace, release_metadata=release_metadata)
            
            assert gpg_key is None

    def test_obtain_gpg_key_no_sig_file(self):
        """"Test obtain_gpg_key when sig file does not exist"""
        artifacts = [
            """4e13e517a3b6f474b734559c96f4fc01678ea5299b5c61844a2747727a52e80f  ./requirements-dev.txt
aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  ./requirements.txt
""".encode('utf-8'),
            terrareg.errors.MissingReleaseArtifactError
        ]

        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        mock_download_artifact = unittest.mock.MagicMock(side_effect=artifacts)

        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="", provider_id="unittest-shasum-id")
            ]
        )

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact):
            with pytest.raises(terrareg.errors.MissingSignureArtifactError):
                terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                    provider=provider, namespace=namespace, release_metadata=release_metadata)

    def test_obtain_gpg_key_no_sha_file(self):
        """"Test obtain_gpg_key when not shasums file exists"""

        artifacts = [
            terrareg.errors.MissingReleaseArtifactError,
        ]

        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        mock_download_artifact = unittest.mock.MagicMock(side_effect=artifacts)

        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="", provider_id="unittest-shasum-id")
            ]
        )

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact):
            with pytest.raises(terrareg.errors.MissingSignureArtifactError):
                terrareg.provider_extractor.ProviderExtractor.obtain_gpg_key(
                    provider=provider, namespace=namespace, release_metadata=release_metadata)

    def test__download_artifact(self):
        """Test _download_artifact"""
        mock_get_release_artifact = unittest.mock.MagicMock(return_value=b"Some binary value of artifact file")

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact):

            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

            artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-test-file", provider_id="provider-id-of-test-file")
            release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="Release v1.9.4",
                tag="v1.9.4",
                archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
                commit_hash="abcdefg123455",
                provider_id="unittest-release-id",
                release_artifacts=[
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-first-test-file", provider_id="previous-provider-id-"),
                    artifact_metadata,
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-other-test-file", provider_id="another-provider-id-"),
                ]
            )

            res = terrareg.provider_extractor.ProviderExtractor._download_artifact(
                provider=provider,
                release_metadata=release_metadata,
                file_name="some-test-file"
            )
            assert res == b"Some binary value of artifact file"

            mock_get_release_artifact.assert_called_once_with(
                provider=provider,
                artifact_metadata=artifact_metadata,
                release_metadata=release_metadata
            )

    def test__download_artifact_unable_to_download(self):
        """Test _download_artifact with being unable to download file"""
        mock_get_release_artifact = unittest.mock.MagicMock(return_value=None)

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact):

            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

            artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-test-file", provider_id="provider-id-of-test-file")
            release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="Release v1.9.4",
                tag="v1.9.4",
                archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
                commit_hash="abcdefg123455",
                provider_id="unittest-release-id",
                release_artifacts=[
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-first-test-file", provider_id="previous-provider-id-"),
                    artifact_metadata,
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-other-test-file", provider_id="another-provider-id-"),
                ]
            )

            with pytest.raises(terrareg.errors.MissingReleaseArtifactError):
                terrareg.provider_extractor.ProviderExtractor._download_artifact(
                    provider=provider,
                    release_metadata=release_metadata,
                    file_name="some-test-file"
                )

            mock_get_release_artifact.assert_called_once_with(
                provider=provider,
                artifact_metadata=artifact_metadata,
                release_metadata=release_metadata
            )

    def test__download_artifact_non_existent_release_artifact(self):
        """Test _download_artifact with a file that doesn't exist in the release"""
        mock_get_release_artifact = unittest.mock.MagicMock(return_value=None)

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact):

            namespace = terrareg.models.Namespace.get("initial-providers")
            provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

            artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-test-file", provider_id="provider-id-of-test-file")
            release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="Release v1.9.4",
                tag="v1.9.4",
                archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
                commit_hash="abcdefg123455",
                provider_id="unittest-release-id",
                release_artifacts=[
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-first-test-file", provider_id="previous-provider-id-"),
                    artifact_metadata,
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-other-test-file", provider_id="another-provider-id-"),
                ]
            )

            with pytest.raises(terrareg.errors.MissingReleaseArtifactError):
                terrareg.provider_extractor.ProviderExtractor._download_artifact(
                    provider=provider,
                    release_metadata=release_metadata,
                    file_name="a-none-existent-file"
                )

            mock_get_release_artifact.assert_not_called()

    def test_generate_artifact_name(self):
        """Test generate_artifact_name"""
        
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")

        artifact_metadata = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-test-file", provider_id="provider-id-of-test-file")
        release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Release v1.9.4",
            tag="v1.9.4",
            archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
            commit_hash="abcdefg123455",
            provider_id="unittest-release-id",
            release_artifacts=[
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-first-test-file", provider_id="previous-provider-id-"),
                artifact_metadata,
                terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-other-test-file", provider_id="another-provider-id-"),
            ]
        )
        assert terrareg.provider_extractor.ProviderExtractor.generate_artifact_name(
            repository=provider.repository,
            release_metadata=release_metadata,
            file_suffix="unittest-file-suffix"
        ) == "terraform-provider-multiple-versions_1.9.4_unittest-file-suffix"

    def test___init__(self):
        """Test __init__"""
        namespace = terrareg.models.Namespace.get("initial-providers")
        provider = terrareg.provider_model.Provider.get(namespace=namespace, name="multiple-versions")
        provider_version = terrareg.provider_version_model.ProviderVersion(provider=provider, version='1.9.4')
        gpg_key = terrareg.models.GpgKey.get_by_fingerprint(fingerprint="A0FC4319ABAF9C28A16821DF4F3072E58D16FF6D")
        try:
            release_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
                name="Release v1.9.4",
                tag="v1.9.4",
                archive_url="https://git.example.com/artifacts/downloads/v1.9.4.tar.gz",
                commit_hash="abcdefg123455",
                provider_id="unittest-release-id",
                release_artifacts=[
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-first-test-file", provider_id="previous-provider-id-"),
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="some-other-test-file", provider_id="another-provider-id-"),
                ]
            )

            with provider_version.create_extraction_wrapper(git_tag="1.9.4", gpg_key=gpg_key):
                provider_extractor = terrareg.provider_extractor.ProviderExtractor(
                    provider_version=provider_version,
                    release_metadata=release_metadata
                )
                assert provider_extractor._provider_version is provider_version
                assert provider_extractor._provider is provider

                assert isinstance(provider_extractor._repository, terrareg.repository_model.Repository)
                assert provider_extractor._repository.pk == provider.repository.pk

                assert provider_extractor._release_metadata is release_metadata

        finally:
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                conn.execute(db.provider_version.delete().where(
                    db.provider_version.c.provider_id==provider.pk,
                    db.provider_version.c.version=="1.9.4"
                ))

    @pytest.mark.parametrize('extract_manifest_file_raises, extract_binaries_raises, extract_documentation_raises, should_call_extract_binaries, should_call_extract_documentation, should_raise', [
        # Good run
        (False, False, False, True, True, False),

        # Cxtract manifest file raises
        (True, False, False, False, False, True),
        # Extract binaries raises
        (False, True, False, True, False, True),
        # Extract documentation raises
        (False, False, True, True, True, True),
        
    ])
    def test_process_version(self, extract_manifest_file_raises, extract_binaries_raises, extract_documentation_raises,
                             should_call_extract_binaries, should_call_extract_documentation, should_raise,
                             test_provider_version_wrapper):
        """Perform extraction"""
        mock_extract_manifest_file = unittest.mock.MagicMock(side_effect=[Exception if extract_manifest_file_raises else None])
        mock_extract_binaries = unittest.mock.MagicMock(side_effect=[Exception if extract_binaries_raises else None])
        mock_extract_documentation = unittest.mock.MagicMock(side_effect=[Exception if extract_documentation_raises else None])

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.extract_manifest_file', mock_extract_manifest_file), \
                unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.extract_binaries', mock_extract_binaries), \
                unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor.extract_documentation', mock_extract_documentation):
            
            with test_provider_version_wrapper() as provider_extractor:
                if should_raise:
                    with pytest.raises(Exception):
                        provider_extractor.process_version()
                else:
                    provider_extractor.process_version()

        mock_extract_manifest_file.assert_called_once_with()
        if should_call_extract_binaries:
            mock_extract_binaries.assert_called_once_with()
        else:
            mock_extract_binaries.assert_not_called()
        if should_call_extract_documentation:
            mock_extract_documentation.assert_called_once_with()
        else:
            mock_extract_documentation.assert_not_called()


    def test__obtain_source_code(self, test_provider_version_wrapper):
        """Obtain source code and extract into temporary location"""

        test_tar_gz = BytesIO()
        archive_id = "1234abcef-main"
        with TemporaryDirectory() as temp_dir:
            base_archive_dir = os.path.join(temp_dir, archive_id)
            os.mkdir(base_archive_dir)
            with open(os.path.join(base_archive_dir, "test_file.txt"), "w") as fh:
                fh.write("Test file")
            os.mkdir(os.path.join(base_archive_dir, "subdir"))
            with open(os.path.join(base_archive_dir, "subdir", "test_subdir_file.txt"), "w") as fh:
                fh.write("Test subdir file")
            
            with tarsafe.open(fileobj=test_tar_gz, mode="w:gz") as tar:
                tar.add(temp_dir, recursive=True, arcname="")

        test_tar_gz.seek(0)
        mock_get_release_archive = unittest.mock.MagicMock(return_value=(test_tar_gz.read(), archive_id))

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_archive', mock_get_release_archive):
            with test_provider_version_wrapper() as provider_extractor:

                with provider_extractor._obtain_source_code() as source_code_dir:
                    assert isinstance(source_code_dir, str)

                    assert os.path.isdir(source_code_dir)
                    assert os.path.isfile(os.path.join(source_code_dir, "test_file.txt"))
                    assert os.path.isdir(os.path.join(source_code_dir, "subdir"))
                    assert os.path.isfile(os.path.join(source_code_dir, "subdir", "test_subdir_file.txt"))

                    with open(os.path.join(source_code_dir, "test_file.txt"), "r") as fh:
                        assert fh.read() == "Test file"
                    with open(os.path.join(source_code_dir, "subdir", "test_subdir_file.txt"), "r") as fh:
                        assert fh.read() == "Test subdir file"

                    # Ensure git has been setup correctly
                    git_log = check_output(["git", "log"], cwd=source_code_dir).decode('utf-8')
                    assert "Author: Terrareg <terrareg@localhost>" in git_log
                    assert "Initial commit" in git_log
                    assert "nothing to commit, working tree clean" in check_output(["git", "status"], cwd=source_code_dir).decode('utf-8')

                    # Check git origin
                    git_remote = check_output(["git", "remote", "get-url", "origin"], cwd=source_code_dir).decode('utf-8')
                    assert "https://git.example.com/initalproviders/terraform-provider-multiple-versions" in git_remote

            # Ensure directory is deleted afterwards
            assert not os.path.isdir(source_code_dir)

    @pytest.mark.parametrize('pre_create_docs_dir', [
        True,
        False
    ])
    def test_extract_documentation(self, pre_create_docs_dir, test_provider_version_wrapper):
        """Test extract_documentation"""

        with TemporaryDirectory() as source_dir:

            @contextlib.contextmanager
            def mock_obtain_source_code_side_effect():
                yield source_dir

            mock_obtain_source_code = unittest.mock.MagicMock(side_effect=mock_obtain_source_code_side_effect)

            mock_switch_terraform_versions = unittest.mock.MagicMock()

            mock_subprocess = unittest.mock.MagicMock()
            mock_collect_markdown_documentation = unittest.mock.MagicMock()

            # Optionally pre-create docs directory
            if pre_create_docs_dir:
                os.mkdir(os.path.join(source_dir, "docs"))

            with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._obtain_source_code', mock_obtain_source_code), \
                    unittest.mock.patch('terrareg.module_extractor.ModuleExtractor._switch_terraform_versions', mock_switch_terraform_versions), \
                    unittest.mock.patch('terrareg.provider_extractor.subprocess', mock_subprocess), \
                    unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._collect_markdown_documentation', mock_collect_markdown_documentation), \
                    test_provider_version_wrapper() as provider_extractor:

                provider_extractor.extract_documentation()

                mock_obtain_source_code.assert_called_once_with()
                assert os.path.isdir(os.path.join(source_dir, "docs"))

                if not pre_create_docs_dir:
                    mock_subprocess.call.assert_called_once_with(
                        ['tfplugindocs', 'generate'],
                        cwd=source_dir,
                        env=unittest.mock.ANY
                    )
                    # Get env and ensure the required variables are there
                    env_vars = mock_subprocess.call.call_args_list[0].kwargs["env"]
                    # Ensure parent env variables are passed
                    assert "PATH" in env_vars
                    # Ensure GO env vars are injected
                    assert "GOROOT" in env_vars
                    assert env_vars["GOROOT"] == "/usr/local/go"
                    assert "GOPATH" in env_vars
                    assert env_vars["GOPATH"].startswith("/tmp/")

                else:
                    mock_subprocess.assert_not_called()

                mock_collect_markdown_documentation.assert_has_calls(calls=[
                    unittest.mock.call(
                        source_directory=source_dir,
                        documentation_directory=os.path.join(source_dir, "docs"),
                        documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
                        file_filter='index.md'
                    ),
                    unittest.mock.call(
                        source_directory=source_dir,
                        documentation_directory=os.path.join(source_dir, "docs", "resources"),
                        documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
                    ),
                    unittest.mock.call(
                        source_directory=source_dir,
                        documentation_directory=os.path.join(source_dir, "docs", "data-sources"),
                        documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
                    ),
                    unittest.mock.call(
                        source_directory=source_dir,
                        documentation_directory=os.path.join(source_dir, "docs", "guides"),
                        documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE,
                    )
                ], any_order=False)

    @pytest.mark.parametrize('content, expected_title, expected_subcategory, expected_description, expected_content', [
        # Test without content
        ("", None, None, None, ""),
        (None, None, None, None, None),

        ("""
# Some Markdown here

Some content
""".strip(), None, None, None, """
# Some Markdown here

Some content
""".strip()),

        # With metadata
        ("""
---
subcategory: "random/v1"
page_title: "Test Provider: a_random_resource"
description: |-
  This creates a random resource using random data
---

# Some markdown header

Some content here

## Another header

Bottom
""".strip(), "Test Provider: a_random_resource", "random/v1", "This creates a random resource using random data", """
# Some markdown header

Some content here

## Another header

Bottom
""".strip()),

        # With unrelated metadata
        ("""
---
something: "random/v1"
somethingelse: "Test Provider: a_random_resource"
athirdthing: |-
  This creates a random resource using random data
---

# Some markdown header

Some content here

## Another header

Bottom
""".strip(), None, None, None, """
# Some markdown header

Some content here

## Another header

Bottom
""".strip())
    ])
    def test__extract_markdown_metadata(self, content, expected_title, expected_subcategory,
                                        expected_description, expected_content,
                                        test_provider_version_wrapper):
        """Test _extract_markdown_metadata"""

        with test_provider_version_wrapper() as provider_extractor:
            assert provider_extractor._extract_markdown_metadata(content) == (expected_title, expected_subcategory, expected_description, expected_content)

    @pytest.mark.parametrize('documentation_type', [
        terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE,
        terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE,
        terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
        terrareg.provider_documentation_type.ProviderDocumentationType.PROVIDER,
        terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE,
    ])
    @pytest.mark.parametrize('directory, file_filter, expected_files', [
        ('', 'overview.md', {'docs/overview.md': {"file_id": "overview.md", "slug": "overview", "name": "overview.md"}}),
        ('guides', None, {'docs/guides/guide-1.md': {"file_id": "guides/guide-1.md", "slug": "guide-1", "name": "guide-1.md"},
                          'docs/guides/test_guide 2.md': {"file_id": "guides/test_guide 2.md", "slug": "test_guide_2", "name": "test_guide 2.md"}}),
        ('empty-dir', None, {}),
        ('recursive', None, {'docs/recursive/test-recursive-base.md': {
            "file_id": "recursive/test-recursive-base.md", "slug": "test-recursive-base", "name": "test-recursive-base.md"}}),

    ])
    def test__collect_markdown_documentation(self, documentation_type, directory, file_filter, expected_files,
                                             test_provider_version_wrapper):
        """Test _collect_markdown_documentation"""

        # Copy expected files dict, to avoid manipulations breaking
        # parameterised tests
        expected_files = dict(expected_files)

        with TemporaryDirectory() as temp_dir:
            os.mkdir(os.path.join(temp_dir, 'docs'))
            for dir in ['guides', 'empty-dir', 'recursive', os.path.join('recursive', 'subdir')]:
                os.mkdir(os.path.join(temp_dir, 'docs', dir))

            for file_ in ['overview.md', os.path.join('guides', 'guide-1.md'), os.path.join('guides', 'test_guide 2.md'),
                        os.path.join('recursive', 'test-recursive-base.md'),
                        os.path.join('recursive', 'subdir', 'test-recursive-sub.md')]:
                with open(os.path.join(temp_dir, 'docs', file_), "w") as fh:
                    fh.write(f"""
---
subcategory: "{documentation_type.value}/{file_}"
page_title: "Test Provider: a_random_resource_{file_}"
description: |-
  This is a test description for {file_}
---

Test Markdown content: {file_}""")

            with test_provider_version_wrapper() as provider_extractor:
                provider_extractor._collect_markdown_documentation(
                    source_directory=temp_dir,
                    documentation_directory=os.path.join(temp_dir, 'docs', directory),
                    documentation_type=documentation_type,
                    file_filter=file_filter
                )

                db = terrareg.database.Database.get()
                with db.get_connection() as conn:
                    res = conn.execute(db.provider_version_documentation.select().where(
                        db.provider_version_documentation.c.provider_version_id==provider_extractor._provider_version.pk
                    )).all()
                    assert len(res) == len(expected_files)

                    for row in res:
                        assert row['filename'] in expected_files
                        expected_file_content = expected_files[row['filename']]
                        del expected_files[row['filename']]

                        assert dict(row) == {
                            'id': row["id"],
                            'content': f'Test Markdown content: {expected_file_content["file_id"]}'.encode('utf-8'),
                            'description': f'This is a test description for {expected_file_content["file_id"]}'.encode('utf-8'),
                            'documentation_type': documentation_type,
                            'filename': f'docs/{expected_file_content["file_id"]}',
                            'language': 'hcl',
                            'name': expected_file_content["name"],
                            'provider_version_id': provider_extractor._provider_version.pk,
                            'slug': expected_file_content["slug"],
                            'subcategory': f'{documentation_type.value}/{expected_file_content["file_id"]}',
                            'title': f'Test Provider: a_random_resource_{expected_file_content["file_id"]}',
                        }


    def test__process_release_file(self, test_provider_version_wrapper):
        """Test _process_release_file"""

        mock_get_release_artifact = unittest.mock.MagicMock(return_value=b"Some binary value of artifact file")

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact), \
                test_provider_version_wrapper() as provider_extractor:

            provider_extractor._process_release_file(
                checksum="a41a58bd5ac74aabbe95b33909aa3fb5bca17efb9825f3924cf4ccfe393a6abc",
                file_name="terraform-provider-multiple-versions_1.9.4_linux_amd64.zip"
            )

            # Obtain DB row for release binary
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.provider_version_binary.select().where(
                    db.provider_version_binary.c.provider_version_id==provider_extractor._provider_version.pk
                )).all()
                assert len(res) == 1
                row = dict(res[0])
                assert row['architecture'] == ProviderBinaryArchitectureType.AMD64
                assert row['checksum'] == 'a41a58bd5ac74aabbe95b33909aa3fb5bca17efb9825f3924cf4ccfe393a6abc'
                assert row['name'] == 'terraform-provider-multiple-versions_1.9.4_linux_amd64.zip'
                assert row['operating_system'] == ProviderBinaryOperatingSystemType.LINUX

    def test__process_release_file_invalid_checksum(self, test_provider_version_wrapper):
        """Test _process_release_file with invalid checksum"""

        mock_get_release_artifact = unittest.mock.MagicMock(return_value=b"Some other binary value that does not match")

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact), \
                test_provider_version_wrapper() as provider_extractor:

            with pytest.raises(terrareg.errors.InvalidReleaseArtifactChecksumError):
                provider_extractor._process_release_file(
                    checksum="a41a58bd5ac74aabbe95b33909aa3fb5bca17efb9825f3924cf4ccfe393a6abc",
                    file_name="terraform-provider-multiple-versions_1.9.4_linux_amd64.zip"
                )

            # Obtain DB row for release binary
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.provider_version_binary.select().where(
                    db.provider_version_binary.c.provider_version_id==provider_extractor._provider_version.pk
                )).all()
                assert len(res) == 0

    def test__process_release_file_non_existent_file(self, test_provider_version_wrapper):
        """Test _process_release_file with non-existent file"""

        mock_get_release_artifact = unittest.mock.MagicMock(return_value=None)

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact), \
                test_provider_version_wrapper() as provider_extractor:

            with pytest.raises(terrareg.errors.MissingReleaseArtifactError):
                provider_extractor._process_release_file(
                    checksum="a41a58bd5ac74aabbe95b33909aa3fb5bca17efb9825f3924cf4ccfe393a6abc",
                    file_name="terraform-provider-multiple-versions_1.9.4_linux_amd64.zip"
                )

            # Obtain DB row for release binary
            db = terrareg.database.Database.get()
            with db.get_connection() as conn:
                res = conn.execute(db.provider_version_binary.select().where(
                    db.provider_version_binary.c.provider_version_id==provider_extractor._provider_version.pk
                )).all()
                assert len(res) == 0

    def test_extract_binaries(self, test_provider_version_wrapper):
        """Test extract_binaries"""

        # Return checksum file, including manifest file and some empty lines
        mock_download_artifact = unittest.mock.MagicMock(return_value=b"""

aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  terraform-provider-multiple-versions_1.9.4_linux_arm64.zip
       

5bf710f5427bafae2a01103ebb48271fedd0ab784e04d11ef95bc057dce8cf7f  terraform-provider-multiple-versions_1.9.4_windows_arm64.zip
8720fafebf4e5ab4affc7426d78b23ce2f2b54a8dfbda4d45abf8051b4558b51  terraform-provider-multiple-versions_1.9.4_manifest.json
0671886887347330e64e584e2e7f02d96b705ff5aad8341a05c9881ecd3ea58d  terraform-provider-multiple-versions_1.9.4_windows_amd64.zip
""")
        mock_process_release_file = unittest.mock.MagicMock()

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact), \
                unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._process_release_file', mock_process_release_file), \
                test_provider_version_wrapper() as provider_extractor:

            provider_extractor.extract_binaries()

            mock_download_artifact.assert_called_once_with(
                provider=provider_extractor._provider,
                release_metadata=provider_extractor._release_metadata,
                file_name="terraform-provider-multiple-versions_1.9.4_SHA256SUMS"
            )

            mock_process_release_file.assert_has_calls(calls=[
                unittest.mock.call(checksum='aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657', file_name='terraform-provider-multiple-versions_1.9.4_linux_arm64.zip'),
                unittest.mock.call(checksum='5bf710f5427bafae2a01103ebb48271fedd0ab784e04d11ef95bc057dce8cf7f', file_name='terraform-provider-multiple-versions_1.9.4_windows_arm64.zip'),
                unittest.mock.call(checksum='0671886887347330e64e584e2e7f02d96b705ff5aad8341a05c9881ecd3ea58d', file_name='terraform-provider-multiple-versions_1.9.4_windows_amd64.zip')
            ], any_order=False)

    def test_extract_binaries_invalid_checksum(self, test_provider_version_wrapper):
        """Test extract_binaries with invalid checksum file"""

        mock_download_artifact = unittest.mock.MagicMock(return_value=b"""
aec01bca39c7f614bc263e299a1fcdd09da3073369756efa6bced80531a45657  terraform-provider-multiple-versions_1.9.4_linux_arm64.zip
5bf710f5427bafae2a01103ebb48271fedd0ab784e04d11ef95bc057dce8cf7f  terraform-provider-multiple-versions_1.9.4_windows_arm64.zip
this is a random line
0671886887347330e64e584e2e7f02d96b705ff5aad8341a05c9881ecd3ea58d  terraform-provider-multiple-versions_1.9.4_windows_amd64.zip
""".strip())
        mock_process_release_file = unittest.mock.MagicMock()

        with unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._download_artifact', mock_download_artifact), \
                unittest.mock.patch('terrareg.provider_extractor.ProviderExtractor._process_release_file', mock_process_release_file), \
                test_provider_version_wrapper() as provider_extractor:

            with pytest.raises(terrareg.errors.InvalidChecksumFileError):
                provider_extractor.extract_binaries()

    @pytest.mark.parametrize('file_content, expected_value, expected_property, expected_error', [
        # Handle default value, ensuring None is stored in database
        (None, None, ["5.0"], None),

        # handle custom value
        ('{"metadata": {"protocol_versions": ["6.0"]}, "version": 1}', ["6.0"], ["6.0"], None),
        ('{"metadata": {"protocol_versions": ["6.0", "23.1"]}, "version": 1}', ["6.0", "23.1"], ["6.0", "23.1"], None),

        # Handle Invalid JSON
        ('{"invalid JSON', None, None, "Could not read manifests file"),
        # Not a dict
        ('["hi"]', None, None, "Manifest file did not contain valid object"),

        # Without version
        ('{}', None, None, "Invalid manifest version. Only version 1 is supported"),
        # Invalid version
        ('{"version": 2}', None, None, "Invalid manifest version. Only version 1 is supported"),
        ('{"version": 0}', None, None, "Invalid manifest version. Only version 1 is supported"),

        # Protocol versions undefined
        ('{"version": 1, "metadata": null}', None, None, "metadata.procotol_versions is not valid in manifest"),
        ('{"version": 1, "metadata": []}', None, None, "metadata.procotol_versions is not valid in manifest"),
        ('{"version": 1, "metadata": {}}', None, None, "metadata.procotol_versions is not valid in manifest"),
        ('{"version": 1, "metadata": {"protocol_versions": {}}}', None, None, "metadata.procotol_versions is not valid in manifest"),
        ('{"version": 1, "metadata": {"protocol_versions": "adg"}}', None, None, "metadata.procotol_versions is not valid in manifest"),
        # Wrong protocol version type
        ('{"version": 1, "metadata": {"protocol_versions": ["adg"]}}', None, None, "Invalid protocol version found"),
    ])
    def test_extract_manifest_file(self, file_content, expected_value, expected_property, expected_error, test_provider_version_wrapper):
        """Test extract_manifest_file"""

        mock_get_release_artifact = unittest.mock.MagicMock(return_value=file_content.encode('utf-8') if file_content else file_content)

        with unittest.mock.patch('terrareg.provider_source.github.GithubProviderSource.get_release_artifact', mock_get_release_artifact):
            with test_provider_version_wrapper() as provider_extractor:

                if expected_error:
                    with pytest.raises(terrareg.errors.InvalidProviderManifestFileError) as err:
                        provider_extractor.extract_manifest_file()
                    print(dir(err))
                    assert str(err.value) == expected_error
                else:
                    provider_extractor.extract_manifest_file()

                    # Ensure value stored in database matches the value from the file
                    db = terrareg.database.Database.get()
                    with db.get_connection() as conn:
                        res = conn.execute(db.provider_version.select().where(db.provider_version.c.id==provider_extractor._provider_version.pk)).all()
                        if expected_value:
                            assert res[0]["protocol_versions"] is not None
                            assert json.loads(res[0]["protocol_versions"]) == expected_value
                        else:
                            assert res[0]["protocol_versions"] is None

                    # Ensure property matches expected value
                    assert provider_extractor._provider_version.protocols == expected_property

                mock_get_release_artifact.assert_called_once_with(
                    provider=provider_extractor._provider,
                    artifact_metadata=terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
                        name="terraform-provider-multiple-versions_1.9.4_manifest.json",
                        provider_id="metadata-provider-id"
                    ),
                    release_metadata=provider_extractor._release_metadata
                )
