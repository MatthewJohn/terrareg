
import subprocess
import contextlib
from io import BytesIO
import os
import tempfile
import tarfile

import terrareg.provider_version_model
import terrareg.repository_model
import terrareg.provider_source.repository_release_metadata
import terrareg.models
import terrareg.config
from terrareg.errors import MissingSignureArtifactError, UnableToObtainReleaseSourceError


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

    def __init__(self, provider_version: 'terrareg.provider_version_model.ProviderVersion',
                       release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata'):
        """Store member variables"""
        self._provider_version = provider_version
        self._release_metadata = release_metadata

    def process_version(self):
        """Perform extraction"""
        self.extract_documentation()
        raise Exception('adg')

    @contextlib.contextmanager
    def _obtain_source_code(self):
        """Obtain source code and extract into temporary location"""
        with tempfile.TemporaryDirectory() as temp_directory:
            # Create child directory for the provider name
            provider_name = self._provider_version.provider.name
            source_dir = os.path.join(temp_directory, provider_name)
            os.mkdir(source_dir)

            # Obtain archive of release
            archive_data, extract_subdirectory = self._provider_version.provider.repository.get_release_archive(
                release_metadata=self._release_metadata
            )

            if not archive_data:
                raise UnableToObtainReleaseSourceError("Unable to obtain release source for provider release")

            # Extract archive
            archive_fh = BytesIO(archive_data)
            with tarfile.open(fileobj=archive_fh, mode="r:gz") as tar:
                tar.extractall(path=source_dir)

            # If the repository provider uses a sub-directory for the source,
            # obtain this
            if extract_subdirectory:
                source_dir = os.path.join(source_dir, extract_subdirectory)

            # Check if source directory is named after then provider
            # (apparently this is important for tfplugindocs)
            # and if not, rename it
            if os.path.dirname(source_dir) != provider_name:
                new_source_dir = os.path.abspath(os.path.join(source_dir, "..", provider_name))
                os.rename(
                    source_dir,
                    new_source_dir
                )
                source_dir = new_source_dir

            yield source_dir

    def extract_documentation(self):
        """Extract documentation from release"""
        with self._obtain_source_code() as source_dir:
            with tempfile.TemporaryDirectory() as go_path:
                go_env = os.environ.copy()
                go_env["GOROOT"] = "/usr/local/go"
                go_env["GOPATH"] = go_path
                # Run get get to initialise repository
                try:
                    subprocess.check_output(
                        ['go', 'get'],
                        cwd=source_dir,
                        env=go_env,
                    )
                except subprocess.CalledProcessError as exc:
                    print(
                        "An error occurred whilst getting 'go get': " +
                        (f": {str(exc)}: {exc.output.decode('utf-8')}" if terrareg.config.Config().DEBUG else "")
                    )
                    return

                # Run go module for extractings docs
                try:
                    subprocess.check_output(
                        ['go', 'get', 'github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs'],
                        cwd=source_dir,
                        env=go_env,
                    )
                except subprocess.CalledProcessError as exc:
                    print(
                        "An error occurred whilst getting tfplugindocs: " +
                        (f": {str(exc)}: {exc.output.decode('utf-8')}" if terrareg.config.Config().DEBUG else "")
                    )
                    return

                # Run go module for extractings docs
                try:
                    subprocess.check_output(
                        ['go', 'run', 'github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs'],
                        cwd=source_dir,
                        env=go_env,
                    )
                except subprocess.CalledProcessError as exc:
                    print(
                        "An error occurred whilst extracting terraform provider docs: " +
                        (f": {str(exc)}: {exc.output.decode('utf-8')}" if terrareg.config.Config().DEBUG else "")
                    )
                    return
