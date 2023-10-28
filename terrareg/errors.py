

class TerraregError(Exception):
    """Base terrareg exception."""

    pass


class UnknownFiletypeError(TerraregError):
    """Uploaded file-type is unknown."""

    pass


class NoModuleVersionAvailableError(TerraregError):
    """No version of this module available."""

    pass


class InvalidTerraregMetadataFileError(TerraregError):
    """Error whilst reading terrareg metadata file."""

    pass


class DatabaseMustBeIniistalisedError(TerraregError):
    """Database object must be initialised before accessing tables."""

    pass


class MetadataDoesNotContainRequiredAttributeError(TerraregError):
    """Module metadata does not contain required metadata attribute."""

    pass


class UploadError(TerraregError):
    """Error with file upload."""

    pass


class NoSessionSetError(TerraregError):
    """No session has been setup and required."""

    pass


class IncorrectCSRFTokenError(TerraregError):
    """CSRF token is not correct."""

    pass


class InvalidGitTagFormatError(TerraregError):
    """Invalid git tag format provided."""

    pass


class InvalidNamespaceNameError(TerraregError):
    """Invalid namespace name."""

    pass


class InvalidModuleNameError(TerraregError):
    """Invalid module name."""

    pass


class InvalidModuleProviderNameError(TerraregError):
    """Invalid module provider name."""

    pass


class InvalidVersionError(TerraregError):
    """Invalid version."""

    pass


class RepositoryUrlParseError(TerraregError):
    """An invalid repository URL has been provided."""

    pass


class RepositoryUrlDoesNotContainValidSchemeError(RepositoryUrlParseError):
    """Repository URL does not contain a scheme."""

    pass


class RepositoryUrlContainsInvalidSchemeError(RepositoryUrlParseError):
    """Repository URL contains an unknown scheme."""

    pass


class RepositoryUrlDoesNotContainHostError(RepositoryUrlParseError):
    """Repository URL does not contain a host/domain."""

    pass


class RepositoryUrlDoesNotContainPathError(RepositoryUrlParseError):
    """Repository URL does not contain path."""

    pass


class RepositoryUrlContainsInvalidPortError(RepositoryUrlParseError):
    """Repository URL contains a invalid port."""

    pass


class RepositoryUrlContainsInvalidTemplateError(RepositoryUrlParseError):
    """Repository URL contains invalid template."""

    pass


class InvalidGitProviderConfigError(TerraregError):
    """Invalid git provider config has been passed."""

    pass


class ModuleProviderCustomGitRepositoryUrlNotAllowedError(TerraregError):
    """Module provider cannot set custom git URL."""

    pass


class ModuleVersionCustomGitRepositoryUrlNotAllowedError(TerraregError):
    """Module provider cannot set custom git URL."""

    pass


class NoModuleDownloadMethodConfiguredError(TerraregError):
    """Module is not configured with a git URL and direct downloads are disabled"""

    pass


class ProviderNameNotPermittedError(TerraregError):
    """Provider name not in list of allowed providers."""

    pass


class UnableToProcessTerraformError(TerraregError):
    """An error occurred whilst attempting to process terraform."""

    pass


class GitCloneError(TerraregError):
    """An error occurred during git clone."""

    pass


class InvalidBooleanConfigurationError(TerraregError):
    """Invalid boolean environment variable."""

    pass


class NamespaceAlreadyExistsError(TerraregError):
    """A namespace already exists with the provided."""

    pass


class InvalidUserGroupNameError(TerraregError):
    """User group name is invalid."""

    pass


class UnableToGetGlobalTerraformLockError(TerraregError):
    """Unable to acquire thread lock whilst switching Terraform"""

    pass


class TerraformVersionSwitchError(TerraregError):
    """An error occurred whilst switching Terraform versions"""

    pass


class ReindexingExistingModuleVersionsIsProhibitedError(TerraregError):
    """Attempting to re-index a module version when re-indexing module versions is disabled."""

    pass


class InvalidNamespaceDisplayNameError(TerraregError):
    """Namespace display name is invalid"""

    pass


class DuplicateNamespaceDisplayNameError(TerraregError):
    """A namespace already exists with this display name"""

    pass


class NamespaceNotEmptyError(TerraregError):
    """Namespace cannot be deleted due to containing modules"""

    pass


class DuplicateModuleProviderError(TerraregError):
    """A module/provider already exists with the same name"""

    pass


class NonExistentModuleProviderRedirectError(TerraregError):
    """Module provider redirect does not exist"""

    pass


class NonExistentNamespaceRedirectError(TerraregError):
    """Namespace redirect does not exist"""

    pass


class ModuleProviderRedirectInUseError(TerraregError):
    """Module provider redirect is in use."""

    pass


class ModuleProviderRedirectForceDeletionNotAllowedError(TerraregError):
    """Force deletion of module provider redirects is not allowed"""

    pass


class InvalidPresignedUrlKeyError(TerraregError):
    """Invalid pre-signed URL key"""

    pass


class PresignedUrlsNotConfiguredError(TerraregError):
    """Missing configurations for pre-signed URLs"""

    pass


class InvalidGpgKeyError(TerraregError):
    """Invalid GPG Key"""

    pass


class DuplicateGpgKeyError(TerraregError):
    """"Duplicate GPG key exists"""

    pass


class DuplicateProviderError(TerraregError):
    """A provider already exists with the same name"""

    pass


class InvalidProviderSourceConfigError(TerraregError):
    """An invalid provider source config is present"""

    pass


class InvalidProviderCategoryConfigError(TerraregError):
    """An invalid provider category config is present"""

    pass


class ReindexingExistingProviderVersionsIsProhibitedError(TerraregError):
    """Cannot reindex a module provider"""

    pass


class MissingSignureArtifactError(TerraregError):
    """Missing signature error for artifacts"""

    pass


class CouldNotFindGpgKeyForProviderVersionError(TerraregError):
    """Could not find valid GPG key for release"""

    pass


class InvalidRepositoryNameError(TerraregError):
    """Repository name is invalid for a the given type of object"""

    pass


class UnableToObtainReleaseSourceError(TerraregError):
    """Unable to obtain release source for provider version"""

    pass


class MissingReleaseArtifactError(TerraregError):
    """Artifact is missing from release"""

    pass


class InvalidChecksumFileError(TerraregError):
    """Invalid line found in checksum file"""

    pass


class InvalidProviderBinaryNameError(TerraregError):
    """Invalid binary file name"""

    pass


class InvalidProviderBinaryOperatingSystemError(TerraregError):
    """Invalid operating system in provider binary name"""

    pass


class InvalidProviderBinaryArchitectureError(TerraregError):
    """Invalid architecture in provider binary name"""

    pass


class InvalidProviderManifestFileError(TerraregError):
    """Invalid manifests file found"""

    pass


class NoGithubAppInstallationError(TerraregError):
    """Github app is not installed in target org/user"""

    pass
