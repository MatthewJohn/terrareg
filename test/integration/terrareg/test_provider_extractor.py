
import unittest.mock
import base64

from typing import Union, Tuple
import contextlib

import pytest

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
import terrareg.errors


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
