

class UnknownFiletypeError(Exception):
    """Uploaded filetype is unknown."""

    pass


class NoModuleVersionAvailableError(Exception):
    """No version of this module available."""

    pass


class InvalidTerraregMetadataFileError(Exception):
    """Error whilst reading terrareg metadata file."""

    pass


class DatabaseMustBeIniistalisedError(Exception):
    """Database object must be initialised before accessing tables."""

    pass
