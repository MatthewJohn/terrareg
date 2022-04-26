
import os

DATA_DIRECTORY = os.path.join(os.environ.get('DATA_DIRECTORY', os.getcwd()), 'data')

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
