

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
