
from asyncio import subprocess
from io import BytesIO
import os
from re import L
from subprocess import check_output
import tarfile
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
                    terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name="terraform-provider-multiple-versions_1.9.4_windows_arm64.zip", provider_id="previous-provider-id-"),
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
            
            with tarfile.open(fileobj=test_tar_gz, mode="w:gz") as tar:
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
