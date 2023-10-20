
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
        signature_file_name = cls.generate_artifact_name(repository=repository, release_metadata=release_metadata, file_suffix="SHA256SUMS.sig")
        for release_artifact in release_metadata.release_artifacts:
            if release_artifact.name == signature_file_name:
                signature_artifact = release_artifact
                break
        else:
            raise MissingSignureArtifactError("There is no release artifact for signature file")

        release_signature = repository.get_release_artifact(
            artifact_metadata=signature_artifact,
            release_metadata=release_metadata
        )
        if not release_signature:
            raise MissingSignureArtifactError("Failed to download signature artifact file")

        for gpg_key in terrareg.models.GpgKey.get_by_namespace(namespace=namespace):
            if gpg_key.match_signure(release_signature):
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
