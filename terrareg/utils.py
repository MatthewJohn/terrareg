
import os
import glob

from terrareg.errors import TerraregError


class PathDoesNotExistError(TerraregError):
    """Path does not exist."""

    pass


class PathIsNotWithinBaseDirectoryError(TerraregError):
    """Sub path is not within base directory."""
    
    pass


def safe_join_paths(base_dir, *sub_paths, is_dir=False, is_file=False):
    """Combine base_dir and sub_path and ensure directory """

    joined_path = os.path.join(base_dir, *sub_paths)

    return check_subdirectory_within_base_dir(
        base_dir=base_dir, sub_dir=joined_path,
        is_dir=is_dir, is_file=is_file)

def safe_iglob(base_dir, pattern, recursive, relative_results, is_file=False, is_dir=False):
    """Perform iglob, ensuring that each of the returned values is within the base directory."""
    results = []
    iglob_args = {
        'pattern': pattern,
        'recursive': recursive
    }
    # If user expects relative results -
    # use root_dir parameter to iglob and use safe_join_paths
    # to check that result is within base directory
    if relative_results:
        for res in glob.iglob(root_dir=base_dir, **iglob_args):
            safe_join_paths(base_dir, res)
            results.append(res)

    # Otherwise, if full paths are expected, prepend base_dir to pattern
    # and use check_subdirectory_within_base_dir to check result is in
    # base directory.
    else:
        for res in glob.iglob('{base_dir}/{pattern}'.format(base_dir=base_dir, pattern=pattern),
                            recursive=recursive):
            results.append(check_subdirectory_within_base_dir(
                base_dir=base_dir, sub_dir=res,
                is_file=is_file, is_dir=is_dir
            ))
    return results

def check_subdirectory_within_base_dir(base_dir, sub_dir, is_dir=False, is_file=False):
    """
    Ensure directory is within base directory.
    Sub directory should be full paths - it should not be relative to the base_dir.

    Perform optional checks if sub_dir is a path or file.

    Return the real path of the sub directory.
    """

    if is_dir and is_file:
        raise Exception('Cannot expect object to be file and directory.')

    # Convert to real paths
    ## Since the base directory might not be a real path,
    ## convert it first
    real_base_path = os.path.realpath(base_dir)
    try:
        ## Convert sub-path to real path to compare
        real_sub_dir = os.path.realpath(sub_dir)
    except OSError:
        raise PathDoesNotExistError('File/directory does not exist')

    ## Ensure sub-path starts wtih base path.
    ## Append trailing slash to ensure, to avoid
    ## allowing /opt/test-this with a base directory of /opt/test
    real_base_path_trailing_slash = '{0}/'.format(real_base_path)
    if not real_sub_dir.startswith(real_base_path_trailing_slash):
        raise PathIsNotWithinBaseDirectoryError('Sub path is not within base directory')

    if is_dir:
        if not os.path.isdir(real_sub_dir):
            raise PathDoesNotExistError('Directory does not exist')
    if is_file:
        if not os.path.isfile(real_sub_dir):
            raise PathDoesNotExistError('File does not exist')

    # Return absolute path of joined paths
    return os.path.abspath(real_sub_dir)
