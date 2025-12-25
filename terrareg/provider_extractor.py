
from typing import Union, Tuple
from glob import glob
import json
import re
import subprocess
import contextlib
from io import BytesIO
import os
import tempfile
import tarfile
import hashlib

import frontmatter

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


class ProviderExtractor:
    """Handle extracting data for provider version"""

    @classmethod
    def obtain_gpg_key(cls,
                       provider: 'terrareg.provider_model.Provider',
                       namespace: 'terrareg.models.Namespace',
                       release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata') -> 'terrareg.models.GpgKey':
        """"Obtain GPG key for signature of release"""
        repository = provider.repository
        # @TODO How to handle providers that do not publish signed SHA files
        try:
            shasums = cls._download_artifact(
                provider=provider,
                release_metadata=release_metadata,
                file_name=cls.generate_artifact_name(repository=repository, release_metadata=release_metadata, file_suffix="SHA256SUMS")
            )
        except MissingReleaseArtifactError as exc:
            raise MissingSignureArtifactError(f"Could not obtain shasums file: {exc}")
        try:
            shasums_signature = cls._download_artifact(
                provider=provider,
                release_metadata=release_metadata,
                file_name=cls.generate_artifact_name(
                    repository=repository,
                    release_metadata=release_metadata,
                    file_suffix="SHA256SUMS.sig")
            )
        except MissingReleaseArtifactError as exc:
            raise MissingSignureArtifactError(f"Could not obtain shasums signature file: {exc}")

        for gpg_key in terrareg.models.GpgKey.get_by_namespace(namespace=namespace):
            if gpg_key.verify_data_signature(signature=shasums_signature, data=shasums):
                return gpg_key

        return None

    @classmethod
    def _download_artifact(cls,
                           provider: 'terrareg.provider_model.Provider',
                           release_metadata: 'terrareg.provider_source.repository_release_metadata.RepositoryReleaseMetadata',
                           file_name: str) -> bytes:
        """Obtain SHA file content"""
        artifact = None
        for release_artifact in release_metadata.release_artifacts:
            if release_artifact.name == file_name:
                artifact = release_artifact
                break

        else:
            raise MissingReleaseArtifactError(f"Could not find artifact in metadata: {file_name}")

        content = provider.repository.get_release_artifact(
            provider=provider,
            artifact_metadata=artifact,
            release_metadata=release_metadata
        )
        if not content:
            raise MissingReleaseArtifactError(f"Failed to download artifact file: {file_name}")

        return content

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
        self._provider = self._provider_version.provider
        self._repository = self._provider.repository
        self._release_metadata = release_metadata

    def process_version(self):
        """Perform extraction"""
        self.extract_manifest_file()
        self.extract_binaries()
        self.extract_documentation()

    @contextlib.contextmanager
    def _obtain_source_code(self):
        """Obtain source code and extract into temporary location"""
        with tempfile.TemporaryDirectory() as temp_directory:
            # Create child directory for the provider name
            provider_name = self._provider.name
            source_dir = os.path.join(temp_directory, provider_name)
            os.mkdir(source_dir)

            # Obtain archive of release
            archive_data, extract_subdirectory = self._provider_version.provider.repository.get_release_archive(
                provider=self._provider,
                release_metadata=self._release_metadata
            )

            if not archive_data:
                raise UnableToObtainReleaseSourceError("Unable to obtain release source for provider release")

            # Extract archive
            archive_fh = BytesIO(archive_data)
            with tarfile.open(fileobj=archive_fh, mode="r:gz") as tar:
                for entry in tar:
                #GOOD: Check that entry is safe
                    if os.path.isabs(entry.name) or ".." in entry.name:
                        raise ValueError("Illegal tar archive entry")
                    tar.extract(entry, path=source_dir)

            # If the repository provider uses a sub-directory for the source,
            # obtain this
            if extract_subdirectory:
                source_dir = os.path.join(source_dir, extract_subdirectory)

            # Check if source directory is named after then provider
            # (apparently this is important for tfplugindocs)
            # and if not, rename it
            if os.path.basename(source_dir) != self._repository.name:
                new_source_dir = os.path.abspath(os.path.join(source_dir, "..", self._repository.name))
                os.rename(
                    source_dir,
                    new_source_dir
                )
                source_dir = new_source_dir

            # Setup git repository inside directory
            git_env = {
                key: value
                for key, value in dict(os.environ.copy()).items()
                # Remove any environment variables for git commit username
                if not key.lower().startswith("git_")
            }
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
                documentation_directory = os.path.join(source_dir, "docs")

                # If documentation directory does not exist,
                # create it and use tfplugindocs to generate documentation
                if not os.path.isdir(documentation_directory):
                    os.mkdir(documentation_directory)

                    with terrareg.module_extractor.ModuleExtractor._switch_terraform_versions(source_dir):
                        go_env = os.environ.copy()
                        go_env["GOROOT"] = "/usr/local/go"
                        go_env["GOPATH"] = temp_go_package_cache

                        # Create documentation directory, if it does not exist
                        if not os.path.isdir(documentation_directory):
                            os.mkdir(documentation_directory)

                        # Run go module for extracting docs
                        try:
                            subprocess.call(
                                ['tfplugindocs', 'generate'],
                                cwd=source_dir,
                                env=go_env,
                            )
                        except subprocess.CalledProcessError as exc:
                            print(
                                "An error occurred whilst extracting terraform provider docs: " +
                                (f": {str(exc)}: {exc.output.decode('utf-8')}" if terrareg.config.Config().DEBUG else "")
                            )
                            return

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
                self._collect_markdown_documentation(
                    source_directory=source_dir,
                    documentation_directory=os.path.join(documentation_directory, "guides"),
                    documentation_type=terrareg.provider_documentation_type.ProviderDocumentationType.GUIDE
                )

    @classmethod
    def _extract_markdown_metadata(cls, content: str) -> Union[Tuple[str, str, str, str], Tuple[None, None, None, str]]:
        """
        Extract metadata from markdown content.
        Returns: title, subcategory, description and content (stripped of metadata)"""
        try:
            front_matter_obj = frontmatter.loads(content)
        except:
            # Handle failure to load and return original content
            return None, None, None, content

        metadata = front_matter_obj.metadata
        return (
            metadata.get("page_title", None),
            metadata.get("subcategory", None),
            metadata.get("description", None),
            front_matter_obj.content
        )

    def _collect_markdown_documentation(self, source_directory: str, documentation_directory: str, documentation_type: terrareg.provider_documentation_type.ProviderDocumentationType, file_filter=None):
        """Collect markdown documentation from directory and store in database"""
        # If a file filter has not been provided, use all markdown files
        if file_filter is None:
            file_filter = "*.md"

        for file_path in glob(os.path.join(documentation_directory, file_filter), recursive=False):
            with open(file_path, "r") as document_fh:
                content = document_fh.read()

            title, subcategory, description, content = self._extract_markdown_metadata(content)

            # Remove source directory from start of file path
            # so the path is relative to the root of the repository.
            # Add 1 to length of source directory to handle the trailing slash
            filename = file_path[(len(source_directory) + 1):]

            terrareg.provider_version_documentation_model.ProviderVersionDocumentation.create(
                provider_version=self._provider_version,
                documentation_type=documentation_type,
                name=os.path.basename(filename),
                title=title,
                description=description,
                language="hcl",
                subcategory=subcategory,
                filename=filename,
                content=content
            )

    def _process_release_file(self, checksum: str, file_name: str) -> None:
        """Process file in release"""
        # Download file
        content = self._download_artifact(
            provider=self._provider,
            release_metadata=self._release_metadata,
            file_name=file_name
        )
        if not content:
            raise MissingReleaseArtifactError("Invalid artifact file")

        # Verify content against checksum
        if hashlib.sha256(content).hexdigest() != checksum:
            raise InvalidReleaseArtifactChecksumError(f"Invalid checksum for {file_name}")

        # Create binary object
        terrareg.provider_version_binary_model.ProviderVersionBinary.create(
            provider_version=self._provider_version,
            name=file_name,
            checksum=checksum,
            content=content
        )

    def extract_binaries(self) -> None:
        """Obtain checksum file and download/validate/process each binary"""
        shasums = self._download_artifact(
            provider=self._provider,
            release_metadata=self._release_metadata,
            file_name=self.generate_artifact_name(
                repository=self._repository,
                release_metadata=self._release_metadata,
                file_suffix="SHA256SUMS"
            )
        )

        shasum_line_re = re.compile(r"^([a-z0-9]{64})[\t ]+(.*)$")

        manifest_file_name = f"{self._provider.full_name}_{self._provider_version.version}_manifest.json"
        for line in shasums.decode('utf-8').split("\n"):
            line = line.strip()

            # Ignore empty lines
            if not line:
                continue

            if not (match := shasum_line_re.match(line)):
                raise InvalidChecksumFileError("Invalid checksum file found")
            
            checksum = match.group(1)
            file_name = match.group(2)

            if file_name != manifest_file_name:
                self._process_release_file(checksum=checksum, file_name=file_name)

    def extract_manifest_file(self):
        """Extract manifest file"""
        try:
            manifest_file_content = self._download_artifact(
                provider=self._provider,
                release_metadata=self._release_metadata,
                file_name=self._provider_version.manifest_file_name
            )
        except MissingReleaseArtifactError:
            manifest_file_content = None

        # Handle empty file
        if not manifest_file_content:
            # If artifact cannot be found, return default manifests file,
            # as per https://developer.hashicorp.com/terraform/registry/providers/publishing#terraform-registry-manifest-file
            return {
                "metadata": {
                    "protocol_versions": ["5.0"]
                },
                "version": 1
            }

        try:
            manifest_content = json.loads(manifest_file_content)
        except Exception as exc:
            print(str(exc))
            raise InvalidProviderManifestFileError("Could not read manifests file")

        if not isinstance(manifest_content, dict):
            raise InvalidProviderManifestFileError("Manifest file did not contain valid object")

        # Ensure version is 1
        if manifest_content.get("version") != 1:
            raise InvalidProviderManifestFileError("Invalid manifest version. Only version 1 is supported")
        
        if (not (metadata := manifest_content.get("metadata", {})) or
                not isinstance(metadata, dict) or
                not (protocol_versions := metadata.get("protocol_versions", [])) or
                not isinstance(protocol_versions, list)):
            raise InvalidProviderManifestFileError("metadata.procotol_versions is not valid in manifest")

        # Ensure protocol versions is valid
        for protocol_version in protocol_versions:
            try:
                float(protocol_version)
            except ValueError:
                raise InvalidProviderManifestFileError("Invalid protocol version found")

        # Update provider version protocol versions
        self._provider_version.update_attributes(protocol_versions=json.dumps(protocol_versions))
