
import pytest

from test.integration.terrareg import TerraregIntegrationTest
import terrareg.provider_source.repository_release_metadata


class TestRepositoryReleaseMetadata(TerraregIntegrationTest):
    """Test RepositoryReleaseMetadata model"""

    def test_init(self):
        """Test __init__ method"""
        release_artifacts = [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
                name="unittest-release-art", provider_id="unittest-artifact-provider-id"
            )
        ]
        obj = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name="Unit Test Name",
            tag="v5.7.2",
            archive_url="https://example.com/unittest/example.zip",
            commit_hash="abcdefgunittesthash",
            provider_id="unittestproviderreleaseid",
            release_artifacts=release_artifacts
        )
        assert obj.name == "Unit Test Name"
        assert obj.tag == "v5.7.2"
        assert obj.archive_url == "https://example.com/unittest/example.zip"
        assert obj.commit_hash == "abcdefgunittesthash"
        assert obj.provider_id == "unittestproviderreleaseid"
        assert obj.release_artifacts == release_artifacts

    @pytest.mark.parametrize('overrides, should_match', [
        ({}, True),
        ({'name': 'other name'}, False),
        ({'tag': 'other-tag'}, False),
        ({'provider_id': 'other-provider-id'}, False),
        ({'archive_url': 'https://anotherurl.com'}, False),
        ({'commit_hash': 'zxcvbnn'}, False),
        ({'release_artifacts': [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(name='other-artifact', provider_id='other-provider-id')
         ]}, False),
    ])
    def test_eq_method(self, overrides, should_match):
        """Test __eq__ method"""
        release_artifacts = [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
                name="unittest-release-art", provider_id="unittest-artifact-provider-id"
            )
        ]
        params = {
            "name": "Unit Test Name",
            "tag": "v5.7.2",
            "archive_url": "https://example.com/unittest/example.zip",
            "commit_hash": "abcdefgunittesthash",
            "provider_id": "unittestproviderreleaseid",
            "release_artifacts": release_artifacts
        }
        other_release_artifacts = [
            terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
                name="unittest-release-art", provider_id="unittest-artifact-provider-id"
            )
        ]
        other_params = {
            "name": "Unit Test Name",
            "tag": "v5.7.2",
            "archive_url": "https://example.com/unittest/example.zip",
            "commit_hash": "abcdefgunittesthash",
            "provider_id": "unittestproviderreleaseid",
            "release_artifacts": other_release_artifacts
        }
        other_params.update(overrides)

        first = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(**params)
        other = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(**other_params)

        if should_match:
            assert first == other
        else:
            assert first != other

    @pytest.mark.parametrize('tag, expected_version', [
        ('v1.5.2', '1.5.2'),
        ('v0.0.0', '0.0.0'),
        ('v999.998.997', '999.998.997'),

        ('', None),
        ('1.2.', None),
        ('1.2', None),
        ('1', None),
        ('somethingelse', None),
    ])
    def test_tag_to_version(self, tag, expected_version):
        """Test tag_to_version method"""
        assert terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata.tag_to_version(tag) == expected_version

    @pytest.mark.parametrize('tag, expected_version', [
        ('v1.5.2', '1.5.2'),
        ('v0.0.0', '0.0.0'),
        ('v999.998.997', '999.998.997'),

        ('', None),
        ('1.2.', None),
        ('1.2', None),
        ('1', None),
        ('somethingelse', None),
    ])
    def test_version(self, tag, expected_version):
        """Test version property"""
        repository_metadata = terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata(
            name='test',
            tag=tag,
            archive_url='',
            commit_hash='',
            provider_id='',
            release_artifacts=[]
        )
        assert repository_metadata.version == expected_version


class TestReleaseArtifactMetadata(TerraregIntegrationTest):
    """Test ReleaseArtifactMetadata model"""

    def test_init(self):
        """Test __init__ method"""
        obj = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
            name='unittest-name',
            provider_id='unittest-provider-id'
        )
        assert obj.name == 'unittest-name'
        assert obj.provider_id == 'unittest-provider-id'

    @pytest.mark.parametrize('overrides, should_match', [
        ({}, True),
        ({'name': 'other name'}, False),
        ({'provider_id': 'other-provider-id'}, False)
    ])
    def test_eq_method(self, overrides, should_match):
        """Test __eq__ method"""
        first = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(
            name='first-name',
            provider_id='first-provider-id'
        )
        other_args = {
            'name': 'first-name',
            'provider_id': 'first-provider-id'
        }
        other_args.update(overrides)
        other = terrareg.provider_source.repository_release_metadata.ReleaseArtifactMetadata(**other_args)

        if should_match:
            assert first == other
        else:
            assert first != other
