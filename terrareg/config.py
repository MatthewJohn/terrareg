
import os

DATA_DIRECTORY = os.path.join(os.environ.get('DATA_DIRECTORY', os.getcwd()), 'data')

"""
Port for server to listen on.
"""
LISTEN_PORT = int(os.environ.get('LISTEN_PORT', 5000))

"""
Whether modules can be downloaded with terraform
without specifying an identification string in
the namespace
"""
ALLOW_UNIDENTIFIED_DOWNLOADS = False

"""Whether flask and sqlalchemy is setup in debug mode."""
DEBUG = bool(os.environ.get('DEBUG', False))

"""Name of analytics token to provide in responses (e.g. application name, team name etc.)"""
ANALYTICS_TOKEN_PHRASE = os.environ.get('ANALYTICS_TOKEN_PHRASE', 'analytics token')

"""Example analytics token to provide in responses (e.g. my-tf-application, my-slack-channel etc.)"""
EXAMPLE_ANALYTICS_TOKEN = os.environ.get('EXAMPLE_ANALYTICS_TOKEN', 'my-tf-application')

"""Comma-separated list of trusted namespaces."""
TRUSTED_NAMESPACES = [
    attr for attr in os.environ.get('TRUSTED_NAMESPACES', '').split(',') if attr
]

"""
List of namespaces, who's modules will be automatically set to verified.
"""
VERIFIED_MODULE_NAMESPACES = [
    attr for attr in os.environ.get('VERIFIED_MODULE_NAMESPACES', '').split(',') if attr
]

"""
Whether uploaded modules, that provide an external URL for the artifact,
should be removed after analysis.
If enabled, module versions with externally hosted artifacts cannot be re-analysed after upload. 
"""
DELETE_EXTERNALLY_HOSTED_ARTIFACTS = os.environ.get('DELETE_EXTERNALLY_HOSTED_ARTIFACTS', 'False') == 'True'

"""
Whether uploaded modules can be downloaded directly.
If disabled, all modules must be configured with a git URL.
"""
ALLOW_MODULE_HOSTING = os.environ.get('ALLOW_MODULE_HOSTING', 'True') == 'True'

"""
Comma-seperated list of metadata attributes that each uploaded module _must_ contain, otherwise the upload is aborted.
"""
REQUIRED_MODULE_METADATA_ATTRIBUTES = [
    attr for attr in os.environ.get('REQUIRED_MODULE_METADATA_ATTRIBUTES', '').split(',') if attr
]

"""Name of application to be displayed in web interface."""
APPLICATION_NAME = os.environ.get('APPLICATION_NAME', 'Terrareg')

"""URL of logo to be used in web interface."""
LOGO_URL = os.environ.get('LOGO_URL', '/static/images/logo.png')

"""
List of comma-separated values for terraform auth tokens for deployment environments.

E.g. xxxxxx.deploy1.xxxxxxxxxxxxx:dev,zzzzzz.deploy1.zzzzzzzzzzzzz:prod
In this example, in the 'dev' environment, the auth token for terraform would be: xxxxxx.deploy1.xxxxxxxxxxxxx
and the auth token for terraform for prod would be: zzzzzz.deploy1.zzzzzzzzzzzzz.

To disable auth tokens and to report all downloads, leave empty.

To only record downloads in a single environment, specify a single auth token. E.g. 'zzzzzz.deploy1.zzzzzzzzzzzzz'
"""
ANALYTICS_AUTH_KEYS = [
    token for token in os.environ.get('ANALYTICS_AUTH_KEYS', '').split(',') if token
]

"""
Token to use for authorisation to be able to modify modules in the user interface.
"""
ADMIN_AUTHENTICATION_TOKEN = os.environ.get('ADMIN_AUTHENTICATION_TOKEN', None)

"""
Flask secret key used for encrypting sessions.

Can be generated using: python -c 'import secrets; print(secrets.token_hex())'
"""
SECRET_KEY = os.environ.get('SECRET_KEY', None)

"""
Session timeout for admin cookie sessions
"""
ADMIN_SESSION_EXPIRY_MINS = int(os.environ.get('ADMIN_SESSION_EXPIRY_MINS', 5))

"""
Whether new module versions (either via upload, import or hook) are automatically
published and available.
If this is disabled, the publish endpoint must be called before the module version
is displayed in the list of module versions.
NOTE: Even whilst in an unpublished state, the module version can still be accessed directly, but not used within terraform.
"""
AUTO_PUBLISH_MODULE_VERSIONS = os.environ.get('AUTO_PUBLISH_MODULE_VERSIONS', 'True') == 'True'

"""
Directory with a module's source that contains sub-modules.

submodules are expected to be within sub-directories of the submodule directory.

E.g. If MODULES_DIRECTORY is set to 'modules', with the root module, the following would be expected for a submodule: 'modules/submodulename/main.tf'.

This can be set to an empty string, to expected submodules to be in the root directory of the parent module.
"""
MODULES_DIRECTORY = os.environ.get('MODULES_DIRECTORY', 'modules')

"""
Git provider config.
JSON list of known git providers.
Each item in the list should contain the following attributes:
 - name - Name of the git provider (e.g. 'Corporate Gitlab')

 - base_url - Formatted base URL for project's repo.
              (e.g. 'https://github.com/{namespace}/{module}'
                 or 'https://gitlab.corporate.com/{namespace}/{module}')
 - clone_url - Formatted clone URL for modules.
               (e.g. 'ssh://gitlab.corporate.com/scm/{namespace}/{module}.git'
                  or 'https://github.com/{namespace}/{module}-{provider}.git')
               Note: Do not include '{version}' placeholder in the URL -
               the git tag will be automatically provided.

 - browse_url - Formatted URL for user-viewable source code
                (e.g. 'https://github.com/{namespace}/{module}-{provider}/tree'
                   or 'https://bitbucket.org/{namespace}/{module}/src/{version}')

An example for public repositories might be:
[{"name": "Github", "base_url": "https://github.com/{namespace}/{module}", "clone_url": "ssh://git@github.com:{namespace}/{module}.git", "browse_url": "https://github.com/{namespace}/{module}/tree/{tag}/{path}"},
 {"name": "Bitbucket", "base_url": "https://bitbucket.org/{namespace}/{module}", "clone_url": "ssh://git@bitbucket.org:{namespace}/{module}-{provider}.git", "browse_url": "https://bitbucket.org/{namespace}/{module}-{provider}/src/{tag}/{path}"},
 {"name": "Gitlab", "base_url": "https://gitlab.com/{namespace}/{module}", "clone_url": "ssh://git@gitlab.com:{namespace}/{module}-{provider}.git", "browse_url": "https://gitlab.com/{namespace}/{module}-{provider}/-/tree/{tag}/{path}"}]
"""
GIT_PROVIDER_CONFIG = os.environ.get('GIT_PROVIDER_CONFIG', '[]')

"""
Whether module providers can specify their own git repository source.
"""
ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER = os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', 'True') == 'True'

"""
Whether module versions can specify git repository in terrareg config.
"""
ALLOW_CUSTOM_GIT_URL_MODULE_VERSION = os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', 'True') == 'True'
