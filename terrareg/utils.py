
import datetime
import os
import glob
import urllib.parse

import bleach
from terrareg.markdown_link_modifier import markdown

from terrareg.errors import TerraregError
import terrareg.config


class PathDoesNotExistError(TerraregError):
    """Path does not exist."""

    pass


class PathIsNotWithinBaseDirectoryError(TerraregError):
    """Sub path is not within base directory."""
    
    pass


def safe_join_paths(base_dir, *sub_paths, is_dir=False, is_file=False, allow_same_directory=False):
    """Combine base_dir and sub_path and ensure directory """

    # Ensure all of the sub_paths start with a relative path, if they start with a slash
    sub_paths = list(sub_paths)
    for itx, sub_path in enumerate(sub_paths):
        if sub_path.startswith('/'):
            sub_paths[itx] = '.{sub_path}'.format(sub_path=sub_paths[itx])

    joined_path = os.path.join(base_dir, *sub_paths)

    return check_subdirectory_within_base_dir(
        base_dir=base_dir, sub_dir=joined_path,
        is_dir=is_dir, is_file=is_file,
        allow_same_directory=allow_same_directory)

def safe_iglob(base_dir, pattern, recursive, is_file=False, is_dir=False):
    """Perform iglob, ensuring that each of the returned values is within the base directory."""
    results = []
    for res in glob.iglob('{base_dir}/{pattern}'.format(base_dir=base_dir, pattern=pattern),
                          recursive=recursive):
        results.append(check_subdirectory_within_base_dir(
            base_dir=base_dir, sub_dir=res,
            is_file=is_file, is_dir=is_dir
        ))
    return results

def check_subdirectory_within_base_dir(
    base_dir,
    sub_dir,
    is_dir=False,
    is_file=False,
    allow_same_directory=False,
):
    """
    Security boundary check:
    - Resolves symlinks via realpath() and ensures the resolved sub path is within the resolved base path.
    - Optionally asserts the resolved path is a directory or file.
    - Returns the resolved absolute path (real path).
    """
    if is_dir and is_file:
        raise Exception("Cannot expect object to be file and directory.")

    # Resolve both paths (dereference symlinks) and normalize.
    real_base = os.path.realpath(base_dir)
    real_sub = os.path.realpath(sub_dir)

    # Ensure the target exists (note: broken symlinks will fail here).
    # Commented out to allow for checking paths that may not yet exist (e.g. shared).
    # if not os.path.exists(real_sub):
    #     raise PathDoesNotExistError("File/directory does not exist")

    # Enforce containment using commonpath (safe vs prefix tricks).
    base_norm = os.path.normpath(real_base)
    sub_norm = os.path.normpath(real_sub)

    same = (sub_norm == base_norm)
    if same and not allow_same_directory:
        raise PathIsNotWithinBaseDirectoryError("Sub path is not within base directory")

    if os.path.commonpath([base_norm, sub_norm]) != base_norm:
        raise PathIsNotWithinBaseDirectoryError("Sub path is not within base directory")

    # Type checks on the resolved path.
    if is_dir and not os.path.isdir(sub_norm):
        raise PathDoesNotExistError("Directory does not exist")
    if is_file and not os.path.isfile(sub_norm):
        raise PathDoesNotExistError("File does not exist")

    return sub_norm


def sanitise_html_content(text, allow_markdown_html=False):
    """Sanitise HTML content to be returned via API to be displayed in UI"""
    kwargs = {}
    if allow_markdown_html:
        kwargs['tags'] = frozenset({
            # Original upstream configuration
            'a', 'abbr', 'acronym', 'b', 'blockquote', 'code', 'em', 'i', 'li', 'ol', 'strong', 'ul',
            # Custom allowed tags
            'p', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'table', 'thead', 'tbody', 'th', 'tr', 'td', 'pre', 'img', 'br'
        })
        kwargs['attributes'] = {
            # Original upstream configuration
            'a': [
                'href',
                'title',
                'name',  # Custom allowed attribute for anchors
                'id',  # Custom allowed attribute for anchors
            ],
            'acronym': ['title'],
            'abbr': ['title'],
            # Custom allowed attributes
            'h1': ['id'], 'h2': ['id'], 'h3': ['id'],
            'h4': ['id'], 'h5': ['id'], 'h6': ['id'],
            'img': ['src'],
            'code': ['class']
        }

    return (
        bleach.clean(
            text, **kwargs
        )
        if text else
        text
    )


def convert_markdown_to_html(file_name, markdown_html):
    """Convert markdown to HTML"""
    return markdown(
        markdown_html,
        file_name=file_name,
        extensions=[
            'fenced_code',
            'tables',
            'mdx_truly_sane_lists',
            'terrareg.markdown_link_modifier']
    )

def get_public_url_details(fallback_domain=None):
    """Get protocol, domain and port used to access Terrareg."""
    config = terrareg.config.Config()

    # Set default values
    domain = config.DOMAIN_NAME or fallback_domain
    port = 443
    protocol = 'https'

    if config.PUBLIC_URL:
        parsed_url = urllib.parse.urlparse(config.PUBLIC_URL)
        # Only use values from parsed URL if it has a hostname,
        # otherwise it is invalid
        if parsed_url.hostname:
            protocol = parsed_url.scheme or 'https'
            port = parsed_url.port or (80 if protocol == 'http' else 443)
            domain = parsed_url.hostname

    return protocol, domain, port

def get_datetime_now():
    """Return datetime now"""
    return datetime.datetime.now()
