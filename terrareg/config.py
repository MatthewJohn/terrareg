
import os

class Config:

    @property
    def DATA_DIRECTORY(self):
        return os.path.join(os.environ.get('DATA_DIRECTORY', os.getcwd()), 'data')

    @property
    def DATABASE_URL(self):
        """
        URL for database.
        Defaults to local sqlite database.
        """
        return os.environ.get('DATABASE_URL', 'sqlite:///modules.db')

    @property
    def LISTEN_PORT(self):
        """
        Port for server to listen on.
        """
        return int(os.environ.get('LISTEN_PORT', 5000))

    @property
    def ALLOW_UNIDENTIFIED_DOWNLOADS(self):
        """
        Whether modules can be downloaded with terraform
        without specifying an identification string in
        the namespace
        """
        return False

    @property
    def DEBUG(self):
        """Whether flask and sqlalchemy is setup in debug mode."""
        return bool(os.environ.get('DEBUG', False))

    @property
    def ANALYTICS_TOKEN_PHRASE(self):
        """Name of analytics token to provide in responses (e.g. application name, team name etc.)"""
        return os.environ.get('ANALYTICS_TOKEN_PHRASE', 'analytics token')

    @property
    def EXAMPLE_ANALYTICS_TOKEN(self):
        """Example analytics token to provide in responses (e.g. my-tf-application, my-slack-channel etc.)"""
        return os.environ.get('EXAMPLE_ANALYTICS_TOKEN', 'my-tf-application')

    @property
    def ALLOWED_PROVIDERS(self):
        """
        Comma-seperated list of allowed providers.

        Leave empty to disable allow-list and allow all providers.
        """
        return [
            attr for attr in os.environ.get('ALLOWED_PROVIDERS', '').split(',') if attr
        ]

    @property
    def TRUSTED_NAMESPACES(self):
        """Comma-separated list of trusted namespaces."""
        return [
            attr for attr in os.environ.get('TRUSTED_NAMESPACES', '').split(',') if attr
        ]

    @property
    def VERIFIED_MODULE_NAMESPACES(self):
        """
        List of namespaces, who's modules will be automatically set to verified.
        """
        return [
            attr for attr in os.environ.get('VERIFIED_MODULE_NAMESPACES', '').split(',') if attr
        ]

    @property
    def DELETE_EXTERNALLY_HOSTED_ARTIFACTS(self):
        """
        Whether uploaded modules, that provide an external URL for the artifact,
        should be removed after analysis.
        If enabled, module versions with externally hosted artifacts cannot be re-analysed after upload. 
        """
        return os.environ.get('DELETE_EXTERNALLY_HOSTED_ARTIFACTS', 'False') == 'True'

    @property
    def ALLOW_MODULE_HOSTING(self):
        """
        Whether uploaded modules can be downloaded directly.
        If disabled, all modules must be configured with a git URL.
        """
        return os.environ.get('ALLOW_MODULE_HOSTING', 'True') == 'True'

    @property
    def REQUIRED_MODULE_METADATA_ATTRIBUTES(self):
        """
        Comma-seperated list of metadata attributes that each uploaded module _must_ contain, otherwise the upload is aborted.
        """
        return [
            attr for attr in os.environ.get('REQUIRED_MODULE_METADATA_ATTRIBUTES', '').split(',') if attr
        ]

    @property
    def APPLICATION_NAME(self):
        """Name of application to be displayed in web interface."""
        return os.environ.get('APPLICATION_NAME', 'Terrareg')

    @property
    def LOGO_URL(self):
        """URL of logo to be used in web interface."""
        return os.environ.get('LOGO_URL', '/static/images/logo.png')

    @property
    def ANALYTICS_AUTH_KEYS(self):
        """
        List of comma-separated values for terraform auth tokens for deployment environments.

        E.g. `xxxxxx.deploy1.xxxxxxxxxxxxx:dev,zzzzzz.deploy1.zzzzzzzzzzzzz:prod`
        In this example, in the 'dev' environment, the auth token for terraform would be: `xxxxxx.deploy1.xxxxxxxxxxxxx`
        and the auth token for terraform for prod would be: `zzzzzz.deploy1.zzzzzzzzzzzzz`.

        To disable auth tokens and to report all downloads, leave empty.

        To only record downloads in a single environment, specify a single auth token. E.g. `zzzzzz.deploy1.zzzzzzzzzzzzz`
        """
        return [
            token for token in os.environ.get('ANALYTICS_AUTH_KEYS', '').split(',') if token
        ]

    @property
    def ADMIN_AUTHENTICATION_TOKEN(self):
        """
        Token to use for authorisation to be able to modify modules in the user interface.
        """
        return os.environ.get('ADMIN_AUTHENTICATION_TOKEN', None)

    @property
    def SECRET_KEY(self):
        """
        Flask secret key used for encrypting sessions.

        Can be generated using: `python -c 'import secrets; print(secrets.token_hex())'`
        """
        return os.environ.get('SECRET_KEY', None)

    @property
    def ADMIN_SESSION_EXPIRY_MINS(self):
        """
        Session timeout for admin cookie sessions
        """
        return int(os.environ.get('ADMIN_SESSION_EXPIRY_MINS', 5))

    @property
    def AUTO_PUBLISH_MODULE_VERSIONS(self):
        """
        Whether new module versions (either via upload, import or hook) are automatically
        published and available.

        If this is disabled, the publish endpoint must be called before the module version
        is displayed in the list of module versions.

        NOTE: Even whilst in an unpublished state, the module version can still be accessed directly, but not used within terraform.
        """
        return os.environ.get('AUTO_PUBLISH_MODULE_VERSIONS', 'True') == 'True'

    @property
    def MODULES_DIRECTORY(self):
        """
        Directory with a module's source that contains sub-modules.

        submodules are expected to be within sub-directories of the submodule directory.

        E.g. If MODULES_DIRECTORY is set to `modules`, with the root module, the following would be expected for a submodule: `modules/submodulename/main.tf`.

        This can be set to an empty string, to expected submodules to be in the root directory of the parent module.
        """
        return os.environ.get('MODULES_DIRECTORY', 'modules')

    @property
    def EXAMPLES_DIRECTORY(self):
        """
        Directory with a module's source that contains examples.

        Examples are expected to be within sub-directories of the examples directory.

        E.g. If EXAMPLES_DIRECTORY is set to `examples`, with the root module, the following would be expected for an example: `examples/myexample/main.tf`.
        """
        return os.environ.get('MODULES_DIRECTORY', 'examples')

    @property
    def GIT_PROVIDER_CONFIG(self):
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
        ```
        [{"name": "Github", "base_url": "https://github.com/{namespace}/{module}", "clone_url": "ssh://git@github.com:{namespace}/{module}.git", "browse_url": "https://github.com/{namespace}/{module}/tree/{tag}/{path}"},
        {"name": "Bitbucket", "base_url": "https://bitbucket.org/{namespace}/{module}", "clone_url": "ssh://git@bitbucket.org:{namespace}/{module}-{provider}.git", "browse_url": "https://bitbucket.org/{namespace}/{module}-{provider}/src/{tag}/{path}"},
        {"name": "Gitlab", "base_url": "https://gitlab.com/{namespace}/{module}", "clone_url": "ssh://git@gitlab.com:{namespace}/{module}-{provider}.git", "browse_url": "https://gitlab.com/{namespace}/{module}-{provider}/-/tree/{tag}/{path}"}]
        ```
        """
        return os.environ.get('GIT_PROVIDER_CONFIG', '[]')

    @property
    def ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER(self):
        """
        Whether module providers can specify their own git repository source.
        """
        return os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', 'True') == 'True'

    @property
    def ALLOW_CUSTOM_GIT_URL_MODULE_VERSION(self):
        """
        Whether module versions can specify git repository in terrareg config.
        """
        return os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', 'True') == 'True'
