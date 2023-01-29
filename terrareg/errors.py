

class TerraregError(Exception):
    """Base terrareg exception."""

    pass


class UnknownFiletypeError(TerraregError):
    """Uploaded filetype is unknown."""

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
    """Invalid module provder name."""

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


class RepositoryDoesNotContainPathError(RepositoryUrlParseError):
    """Repository URL does not contain path."""

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
    """Unable to aquire thread lock whilst switching Terraform"""

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
