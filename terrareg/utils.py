
import os

from terrareg.errors import TerraregError


class PathDoesNotExistError(TerraregError):
    """Path does not exist."""

    pass


class PathIsNotWithinBaseDirectoryError(TerraregError):
    """Sub path is not within base directory."""
    
    pass


def safe_join_paths(base_dir, *sub_paths, is_dir=False, is_file=False):
    """Combine base_dir and sub_path and ensure directory """

    if is_dir and is_file:
        raise Exception('Cannot expect object to be file and directory.')

    joined_path = os.path.join(base_dir, *sub_paths)

    # Convert to real paths
    ## Since the base directory might not be a real path,
    ## convert it first
    real_base_path = os.path.realpath(base_dir)
    try:
        ## Convert sub-path to real path to compare
        real_joined_path = os.path.realpath(joined_path)
    except OSError:
        raise PathDoesNotExistError('File/directory does not exist')

    ## Ensure sub-path starts wtih base path.
    ## Append trailing slash to ensure, to avoid
    ## allowing /opt/test-this with a base directory of /opt/test
    real_base_path_trailing_slash = '{0}/'.format(real_base_path)
    if not real_joined_path.startswith(real_base_path_trailing_slash):
        raise PathIsNotWithinBaseDirectoryError('Sub path is not within base directory')

    if is_dir:
        if not os.path.isdir(real_joined_path):
            raise PathDoesNotExistError('Directory does not exist')
    if is_file:
        if not os.path.isfile(real_joined_path):
            raise PathDoesNotExistError('File does not exist')

    # Return absolute path of joined paths
    return os.path.abspath(joined_path)
