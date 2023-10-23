
from distutils.dir_util import copy_tree
from glob import glob
import re
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
import terrareg.provider_documentation_type
import terrareg.provider_version_documentation_model
import terrareg.module_extractor
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
        # raise Exception('adg')

    @contextlib.contextmanager
    def _obtain_source_code(self):
        """Obtain source code and extract into temporary location"""
        with tempfile.TemporaryDirectory() as temp_directory:
            # Create child directory for the provider name
            provider = self._provider_version.provider
            repository = provider.repository
            provider_name = provider.name
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
            if os.path.basename(source_dir) != repository.name:
                new_source_dir = os.path.abspath(os.path.join(source_dir, "..", repository.name))
                os.rename(
                    source_dir,
                    new_source_dir
                )
                source_dir = new_source_dir

            # Setup git repository inside directory
            git_env = os.environ.copy()
            git_env["HOME"] = temp_directory
            subprocess.check_output(["git", "init"], cwd=source_dir, env=git_env)
            # Setup fake git user to avoid errors when committing
            subprocess.check_output(["git", "config", "user.email", "terrareg@localhost"], cwd=source_dir, env=git_env)
            subprocess.check_output(["git", "config", "user.name", "Terrareg"], cwd=source_dir, env=git_env)
            subprocess.check_output(["git", "add", "*"], cwd=source_dir, env=git_env)
            subprocess.check_output(["git", "commit", "-m", "Initial commit"], cwd=source_dir, env=git_env)
            clone_url = self._provider_version.provider.repository.clone_url
            if clone_url.endswith(".git"):
                clone_url = re.sub(r"\.git$", "", clone_url)
            subprocess.check_output(["git", "remote", "add", "origin", clone_url], cwd=source_dir, env=git_env)

            yield source_dir

    def extract_documentation(self):
        """Extract documentation from release"""
        with self._obtain_source_code() as source_dir:
            with tempfile.TemporaryDirectory() as temp_go_package_cache:
                with terrareg.module_extractor.ModuleExtractor._switch_terraform_versions(source_dir):
                    go_env = os.environ.copy()
                    go_env["GOROOT"] = "/usr/local/go"
                    go_env["GOPATH"] = temp_go_package_cache

                    documentation_directory = os.path.join(source_dir, "docs")

                    # Create documentation directory, if it does not exist
                    if not os.path.isdir(documentation_directory):
                        os.mkdir(documentation_directory)

                    try:
                        subprocess.call(
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
                        subprocess.call(
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

                documentation_directory = os.path.join(source_dir, "docs")
                self._collect_markdown_documentation(
                    source_directory=source_dir,
                    documentation_directory=documentation_directory,
                    documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.OVERVIEW,
                    file_filter="index.md"
                )
                self._collect_markdown_documentation(
                    source_directory=source_dir,
                    documentation_directory=os.path.join(documentation_directory, "resources"),
                    documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.RESOURCE
                )
                self._collect_markdown_documentation(
                    source_directory=source_dir,
                    documentation_directory=os.path.join(documentation_directory, "data-sources"),
                    documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.DATA_SOURCE
                )

    def _collect_markdown_documentation(self, source_directory: str, documentation_directory: str, documentation_type: terrareg.provider_documentation_type.ProviderDocumentationType, file_filter=None):
        """Collect markdown documentation from directory and store in database"""
        # If a file filter has not been provided, use all markdown files
        if file_filter is None:
            file_filter = "*.md"

        for file_path in glob(os.path.join(documentation_directory, file_filter), recursive=False):
            with open(file_path, "r") as document_fh:
                content = document_fh.read()

            # Remove source directory from start of file path
            # so the path is relative to the root of the repository.
            # Add 1 to length of source directory to handle the trailing slash
            filename = file_path[(len(source_directory) + 1):]

            terrareg.provider_version_documentation_model.ProviderVersionDocumentation.create(
                provider_version=self._provider_version,
                documentation_type=documentation_type,
                name=os.path.basename(filename),
                filename=filename,
                content=content
            )
