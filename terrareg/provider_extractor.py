
import terrareg.provider_version_model
import terrareg.repository_model
import terrareg.provider_source.repository_release_metadata
import terrareg.models
from terrareg.errors import MissingSignureArtifactError


class ProviderExtractor:
    """Handle extracting data for provider version"""

    @classmethod
    def obtain_gpg_key(cls, repository: 'terrareg.repository_model.Repository',
                       namespace: 'terrareg.models.Namespace',
                       release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> 'terrareg.models.GpgKey':
        """"Obtain GPG key for signature of release"""
        shasum_file_name = cls.generate_artifact_name(repository=repository, release_metadata=release_metadata, file_suffix="SHA256SUMS")
        shasum_signature_file_name = cls.generate_artifact_name(repository=repository, release_metadata=release_metadata, file_suffix="SHA256SUMS.sig")

        shasum_signature_artifact = None
        shasum_artifact = None
        for release_artifact in release_metadata.release_artifacts:
            if release_artifact.name == shasum_file_name:
                shasum_artifact = release_artifact
            elif release_artifact.name == shasum_signature_file_name:
                shasum_signature_artifact = release_artifact

            # Once the shasum and signature file have been found, exit
            if shasum_signature_artifact and shasum_artifact:
                break
        else:
            raise MissingSignureArtifactError("Could not find SHA or SHA signature file for release")

        shasums = repository.get_release_artifact(
            artifact_metadata=shasum_artifact,
            release_metadata=release_metadata
        )
        if not shasums:
            raise MissingSignureArtifactError("Failed to download SHASUMS artifact file")

        shasums_signature = repository.get_release_artifact(
            artifact_metadata=shasum_signature_artifact,
            release_metadata=release_metadata
        )
        if not shasums_signature:
            raise MissingSignureArtifactError("Failed to download SHASUMS signature artifact file")

        for gpg_key in terrareg.models.GpgKey.get_by_namespace(namespace=namespace):
            if gpg_key.verify_data_signature(signature=shasums_signature, data=shasums):
                return gpg_key

        return None

    @classmethod
    def generate_artifact_name(cls,
                               repository: 'terrareg.repository_model.Repository',
                               release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata',
                               file_suffix: str):
        """Generate artifact file name"""
        return f"{repository.name}_{release_metadata.version}_{file_suffix}"

    def __init__(self, provider_version: 'terrareg.provider_version_model.ProviderVersion'):
        """Store member variables"""
        self._provider_version = provider_version

    def process_version(self):
        """Perform extraction"""
        raise Exception('adg')
