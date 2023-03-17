
from enum import Enum
import os

from terrareg.errors import InvalidBooleanConfigurationError


class ModuleVersionReindexMode(Enum):
    """Module version re-indexing modes"""
    LEGACY = "legacy"
    AUTO_PUBLISH = "auto-publish"
    PROHIBIT = "prohibit"


class Config:

    @property
    def INTERNAL_EXTRACTION_ANALYITCS_TOKEN(self):
        """
        Analaytics token used by Terraform initialised by the registry.

        This is used by the registry to call back to itself when analysing module examples.

        The value should be changed if it might result in a conflict with a legitimate analytics token used in Terraform
        that calls modules from the registry.
        """
        return os.environ.get('INTERNAL_EXTRACTION_ANALYITCS_TOKEN', 'internal-terrareg-analytics-token')

    @property
    def IGNORE_ANALYTICS_TOKEN_AUTH_KEYS(self):
        """
        A list of a Terraform auth keys that can be used to authenticate to the registry
        to ignore check for valid analaytics token in module path.

        It is recommended that examples in modules do not include analytics tokens in calls to other modules
        hosted in the registry, to avoid having to add analytics tokens to test the examples during development
        of the modules.

        No analytics of the module usage are captured when this auth key is used.

        The value should be a comma-seperated list of auth keys.

        The auth key can be used by placing in a "credential" block in the user's .terraformrc file.
        """
        return [
            auth_key
            for auth_key in os.environ.get("IGNORE_ANALYTICS_TOKEN_AUTH_KEYS", "").split(",")
            if auth_key
        ]

    @property
    def PUBLIC_URL(self):
        """
        The URL that is used for accessing Terrareg by end-users.

        E.g.:
        `https://terrareg.mycorp.com`
        `https://terrareg.mycorg.com:8443`


        Ensure the protocol is https if Terrareg is accessible over https.
        Provide the port if necessary.

        If left empty, the protocol will be assumed to be HTTPS, the port will be assumed to be 443
        and the domain will fallback to be the value set in `DOMAIN_NAME`.
        """
        return os.environ.get('PUBLIC_URL', None)

    @property
    def DOMAIN_NAME(self):
        """
        Domain name that the system is hosted on.

        This should be setup for all installations, but is required for infracost and OpenID authentication.

        Note: The configuration is deprecated, please set the `PUBLIC_URL` configuration instead.
        """
        return os.environ.get('DOMAIN_NAME', None)

    @property
    def DATA_DIRECTORY(self):
        return os.path.join(os.environ.get('DATA_DIRECTORY', os.getcwd()), 'data')

    @property
    def DATABASE_URL(self):
        """
        URL for database.
        Defaults to local sqlite database.

        To setup SQLite datbase, use `sqlite:///<path to sqlite DB>`

        To setup MySQL, use `mysql+mysqlconnector://<user>:<password>@<host>[:<port>]/<database>`
        """
        return os.environ.get('DATABASE_URL', 'sqlite:///modules.db')

    @property
    def LISTEN_PORT(self):
        """
        Port for server to listen on.
        """
        return int(os.environ.get('LISTEN_PORT', 5000))

    @property
    def SSL_CERT_PRIVATE_KEY(self):
        """
        Path to SSL private certificate key.

        If running in a container, the key must be mounted inside the container.
        This value must be set to the path of the key within the container.

        This must be set in accordance with SSL_CERT_PUBLIC_KEY - both must either be
        set or left empty.
        """
        return os.environ.get('SSL_CERT_PRIVATE_KEY', None)

    @property
    def SSL_CERT_PUBLIC_KEY(self):
        """
        Path to SSL public key.

        If running in a container, the key must be mounted inside the container.
        This value must be set to the path of the key within the container.

        This must be set in accordance with SSL_CERT_PRIVATE_KEY - both must either be
        set or left empty.
        """
        return os.environ.get('SSL_CERT_PUBLIC_KEY', None)

    @property
    def ALLOW_UNIDENTIFIED_DOWNLOADS(self):
        """
        Whether modules can be downloaded with Terraform
        without specifying an identification string in
        the namespace
        """
        return self.convert_boolean(os.environ.get('ALLOW_UNIDENTIFIED_DOWNLOADS', 'False'))

    @property
    def DEBUG(self):
        """Whether flask and sqlalchemy is setup in debug mode."""
        return self.convert_boolean(os.environ.get('DEBUG', 'False'))

    @property
    def THREADED(self):
        """Whether flask is configured to enable threading"""
        return self.convert_boolean(os.environ.get('THREADED', 'True'))

    @property
    def ANALYTICS_TOKEN_PHRASE(self):
        """Name of analytics token to provide in responses (e.g. `application name`, `team name` etc.)"""
        return os.environ.get('ANALYTICS_TOKEN_PHRASE', 'analytics token')

    @property
    def ANALYTICS_TOKEN_DESCRIPTION(self):
        """Describe to be provided to user about analytics token (e.g. `The name of your application`)"""
        return os.environ.get('ANALYTICS_TOKEN_DESCRIPTION', '')

    @property
    def EXAMPLE_ANALYTICS_TOKEN(self):
        """
        Example analytics token to provide in responses (e.g. my-tf-application, my-slack-channel etc.).

        Note that, if this token is used in a module call, it will be ignored and treated as if
        an analytics token has not been provided.
        If analytics tokens are required, this stops users from accidently using the example placeholder in
        Terraform projects.
        """
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
    def TRUSTED_NAMESPACE_LABEL(self):
        """Custom name for 'trusted namespace' in UI."""
        return os.environ.get('TRUSTED_NAMESPACE_LABEL', 'Trusted')

    @property
    def CONTRIBUTED_NAMESPACE_LABEL(self):
        """Custom name for 'contributed namespace' in UI."""
        return os.environ.get('CONTRIBUTED_NAMESPACE_LABEL', 'Contributed')

    @property
    def VERIFIED_MODULE_NAMESPACES(self):
        """
        List of namespaces, who's modules will be automatically set to verified.
        """
        return [
            attr for attr in os.environ.get('VERIFIED_MODULE_NAMESPACES', '').split(',') if attr
        ]

    @property
    def VERIFIED_MODULE_LABEL(self):
        """Custom name for 'verified module' in UI."""
        return os.environ.get('VERIFIED_MODULE_LABEL', 'Verified')

    @property
    def DISABLE_TERRAREG_EXCLUSIVE_LABELS(self):
        """
        Whether to disable 'terrareg exclusive' labels from feature tabs in UI.

        Set to 'True' to disable the labels.
        """
        return self.convert_boolean(os.environ.get('DISABLE_TERRAREG_EXCLUSIVE_LABELS', 'False'))

    @property
    def DELETE_EXTERNALLY_HOSTED_ARTIFACTS(self):
        """
        Whether uploaded modules, that provide an external URL for the artifact,
        should be removed after analysis.
        If enabled, module versions with externally hosted artifacts cannot be re-analysed after upload.
        """
        return self.convert_boolean(os.environ.get('DELETE_EXTERNALLY_HOSTED_ARTIFACTS', 'False'))

    @property
    def ALLOW_MODULE_HOSTING(self):
        """
        Whether uploaded modules can be downloaded directly.
        If disabled, all modules must be configured with a git URL.
        """
        return self.convert_boolean(os.environ.get('ALLOW_MODULE_HOSTING', 'True'))

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
        List of comma-separated values for Terraform auth tokens for deployment environments.

        E.g. `xxxxxx.deploy1.xxxxxxxxxxxxx:dev,zzzzzz.deploy1.zzzzzzzzzzzzz:prod`
        In this example, in the 'dev' environment, the auth token for Terraform would be: `xxxxxx.deploy1.xxxxxxxxxxxxx`
        and the auth token for Terraform for prod would be: `zzzzzz.deploy1.zzzzzzzzzzzzz`.

        To disable auth tokens and to report all downloads, leave empty.

        To only record downloads in a single environment, specify a single auth token. E.g. `zzzzzz.deploy1.zzzzzzzzzzzzz`

        For information on using these API keys, please see Terraform: https://docs.w3cub.com/terraform/commands/cli-config.html#credentials
        """
        return [
            token for token in os.environ.get('ANALYTICS_AUTH_KEYS', '').split(',') if token
        ]

    @property
    def UPLOAD_API_KEYS(self):
        """
        List of comma-separated list of API keys to upload/import new module versions.

        For bitbucket hooks, one of these keys must be provided as the 'secret' to the webhook.

        To disable authentication for upload endpoint, leave empty.
        """
        return [
            token
            for token in os.environ.get('UPLOAD_API_KEYS', '').split(',')
            if token
        ]

    @property
    def PUBLISH_API_KEYS(self):
        """
        List of comma-separated list of API keys to publish module versions.

        To disable authentication for publish endpoint, leave empty.
        """
        return [
            token
            for token in os.environ.get('PUBLISH_API_KEYS', '').split(',')
            if token
        ]

    @property
    def ADMIN_AUTHENTICATION_TOKEN(self):
        """
        Password/API key to for authentication as the built-in admin user.
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
        return int(os.environ.get('ADMIN_SESSION_EXPIRY_MINS', 60))

    @property
    def AUTO_PUBLISH_MODULE_VERSIONS(self):
        """
        Whether new module versions (either via upload, import or hook) are automatically
        published and available.

        If this is disabled, the publish endpoint must be called before the module version
        is displayed in the list of module versions.

        NOTE: Even whilst in an unpublished state, the module version can still be accessed directly, but not used within Terraform.
        """
        return self.convert_boolean(os.environ.get('AUTO_PUBLISH_MODULE_VERSIONS', 'True'))

    @property
    def MODULE_VERSION_REINDEX_MODE(self):
        """
        This configuration defines how re-indexes a module version, that already exists, behaves.

        This can be set to one of:
         * 'legacy' - The new module version will replace the old one. Until the version is re-published, it will not be available to Terraform. Analytics for the module version will be retained.
         * 'auto-publish' - The new module version will replace the old one. If the previous version was published, the new version will be automatically published. Analytics for the module version will be retained.
         * 'prohibit' - If a module version has already been indexed, it cannot be re-indexed via hooks/API calls without the version first being deleted.
        """
        return ModuleVersionReindexMode(os.environ.get('MODULE_VERSION_REINDEX_MODE', 'legacy'))

    @property
    def AUTO_CREATE_MODULE_PROVIDER(self):
        """
        Whether to automatically create module providers when
        uploading module versions, either from create endpoint or hooks.

        If disabled, modules must be created using the module provider create endpoint (or via the web interface).
        """
        return self.convert_boolean(os.environ.get('AUTO_CREATE_MODULE_PROVIDER', 'True'))

    @property
    def AUTO_CREATE_NAMESPACE(self):
        """
        Whether to automatically create namespaces when
        uploading a module version to a module provider in a non-existent namespace.

        If disabled, namespaces must be created using the namespace create endpoint (or via web UI)
        """
        return self.convert_boolean(os.environ.get('AUTO_CREATE_NAMESPACE', 'True'))

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
        return os.environ.get('EXAMPLES_DIRECTORY', 'examples')

    @property
    def GIT_CLONE_TIMEOUT(self):
        """
        Timeout for git clone commands in seconds.

        Leave empty to disable timeout for clone.

        Warning - if a git clone is performed without a timeout set, the request may never complete and could leave locks on the database, requiring an application restart.
        """
        val = os.environ.get('GIT_CLONE_TIMEOUT', '300')
        return None if val is None else int(val)

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
                        (e.g. 'https://github.com/{namespace}/{module}-{provider}/tree/{tag}/{path}'
                        or 'https://bitbucket.org/{namespace}/{module}/src/{version}?at=refs%2Ftags%2F{tag_uri_encoded}').
                        Must include placeholdes:
                         - {path} (for source file/folder path)
                         - {tag} or {tag_uri_encoded} for the git tag

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
        return self.convert_boolean(os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER', 'True'))

    @property
    def ALLOW_CUSTOM_GIT_URL_MODULE_VERSION(self):
        """
        Whether module versions can specify git repository in terrareg config.
        """
        return self.convert_boolean(os.environ.get('ALLOW_CUSTOM_GIT_URL_MODULE_VERSION', 'True'))

    @property
    def TERRAFORM_EXAMPLE_VERSION_TEMPLATE(self):
        """
        Template of version number string to be used in Terraform examples in the UI.
        This is used by the snippet example of a Terraform module and the 'resource builder' example.

        The template can contain the following placeholders:
         * `{major}`, `{minor}`, `{patch}`
         * `{major_minus_one}`, `{minor_minus_one}`, `{patch_minus_one}`
         * `{major_plus_one}`, `{minor_plus_one}`, `{patch_plus_one}`

        Some examples:
         * `>= {major}.{minor}.{patch}, < {major_plus_one}.0.0`
         * `~> {major}.{minor}.{patch}`

        For more information, see Terraform documentation: https://www.terraform.io/language/expressions/version-constraints

        """
        return os.environ.get('TERRAFORM_EXAMPLE_VERSION_TEMPLATE', '{major}.{minor}.{patch}')

    @property
    def AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION(self):
        """
        Whether to automatically generate module provider descriptions, if they are not provided in terrareg metadata file of the module.
        """
        return self.convert_boolean(os.environ.get('AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION', 'True'))

    @property
    def AUTOGENERATE_USAGE_BUILDER_VARIABLES(self):
        """
        Whether to automatically generate usage builder variables from the required variables and their descriptions.
        When disabled, the usage builder will only be displayed on a module when the "variable_template" section
        of the terrareg.json metadata file is populated.
        """
        return self.convert_boolean(os.environ.get('AUTOGENERATE_USAGE_BUILDER_VARIABLES', 'True'))

    @property
    def ENABLE_SECURITY_SCANNING(self):
        """
        Whether to perform security scans of uploaded modules and display them against the module, submodules and examples.
        """
        return self.convert_boolean(os.environ.get('ENABLE_SECURITY_SCANNING', 'True'))

    @property
    def INFRACOST_API_KEY(self):
        """
        API key for Infracost.

        Set this to enable cost-analysis of module examples.

        To generate an API key:
        Log in at https://dashboard.infracost.io > select your organization > Settings

        For cost analysis to be performed on modules which utilise other modules from this registry, ensure `DOMAIN_NAME` is set.
        """
        return os.environ.get('INFRACOST_API_KEY', None)

    @property
    def INFRACOST_PRICING_API_ENDPOINT(self):
        """
        Self-hosted infracost pricing API endpoint.

        For information on self-hosting the infracost pricing API, see https://www.infracost.io/docs/cloud_pricing_api/self_hosted/
        """
        return os.environ.get('INFRACOST_PRICING_API_ENDPOINT', None)

    @property
    def INFRACOST_TLS_INSECURE_SKIP_VERIFY(self):
        """
        Whether to skip TLS verification for self-hosted pricing endpoints
        """
        return self.convert_boolean(os.environ.get('INFRACOST_TLS_INSECURE_SKIP_VERIFY', 'False'))

    @property
    def ADDITIONAL_MODULE_TABS(self):
        """
        Set additional markdown files from a module to be displayed in the UI.

        Value must be a JSON array of objects.
        Each object of the array defines an additional tab in the module.
        The object defines the name of the tab and a list of files in the repository.
        e.g. `[["Tab 1 Name", ["file-to-use.md", "alternate-file-to-use.md"]], ["tab 2", ["tab_2_file.md"]]]`

        The tabs will be displayed in order of their placement in the outer list.
        If multiple files are provided, the first file found in the repository will be used for the tab content.

        Filenames with an extension `.md` will be treated as markdown. All other files will be treated as plain-text.

        E.g.
        ```
        [["Release Notes": ["RELEASE_NOTES.md", "CHANGELOG.md"]], ["Development Guidelines", ["CONTRIBUTING.md"]], ["License", ["LICENSE"]]]
        ```
        """
        return os.environ.get('ADDITIONAL_MODULE_TABS', '[["Release Notes", ["RELEASE_NOTES.md", "CHANGELOG.md"]], ["License", ["LICENSE"]]]')

    @property
    def MODULE_LINKS(self):
        """
        List of custom links to display on module provides.

        Each link must contain a display and and link URL.
        These can contain placeholders, such as:
         * namespace
         * module
         * provider
         * version

        The links can be provided with a list of namespaces to limit the link to only modules within those namespaces.

        The format should be similar to this example:
        [
            {"text": "Text for the link e.g. Github Issues for {module}",
             "url": "https://github.com/{namespace}/{module}-{provider}/issues"},
            {"text": "Second link limited to two namespaces",
             "url": "https://mydomain.example.com/",
             "namespaces": ["namespace1", "namespace2"]}
        ]
        """
        return os.environ.get('MODULE_LINKS', '[]')

    @property
    def ENABLE_ACCESS_CONTROLS(self):
        """
        Enables role based access controls for SSO users.

        Enabling this feature will restrict all SSO users from performing any admin tasks.
        Group mappings can be setup in the settings page, providing SSO users and groups with global or namespace-based permissions.

        When disabled, all SSO users will have global admin privileges.
        """
        return self.convert_boolean(os.environ.get('ENABLE_ACCESS_CONTROLS', 'False'))

    @property
    def OPENID_CONNECT_LOGIN_TEXT(self):
        """
        Text for sign in button for OpenID Connect authentication
        """
        return os.environ.get('OPENID_CONNECT_LOGIN_TEXT', 'Login using OpenID Connect')

    @property
    def OPENID_CONNECT_CLIENT_ID(self):
        """
        Client ID for OpenID Conect authentication
        """
        return os.environ.get('OPENID_CONNECT_CLIENT_ID', None)

    @property
    def OPENID_CONNECT_CLIENT_SECRET(self):
        """
        Client secret for OpenID Conect authentication
        """
        return os.environ.get('OPENID_CONNECT_CLIENT_SECRET', None)

    @property
    def OPENID_CONNECT_ISSUER(self):
        """
        Base Issuer URL for OpenID Conect authentication.

        A well-known URL will be expected at `${OPENID_CONNECT_ISSUER}/.well-known/openid-configuration`
        """
        return os.environ.get('OPENID_CONNECT_ISSUER', None)

    @property
    def OPENID_CONNECT_DEBUG(self):
        """
        Enable debug of OpenID connect via stdout.

        This should only be enabled for non-production environments.
        """
        return self.convert_boolean(os.environ.get('OPENID_CONNECT_DEBUG', 'False'))

    @property
    def SAML2_LOGIN_TEXT(self):
        """
        Text for sign in button for SAML2 authentication
        """
        return os.environ.get('SAML2_LOGIN_TEXT', 'Login using SAML')

    @property
    def SAML2_IDP_METADATA_URL(self):
        """
        SAML2 provider metadata url
        """
        return os.environ.get('SAML2_IDP_METADATA_URL', None)

    @property
    def SAML2_ISSUER_ENTITY_ID(self):
        """
        SAML2 provider entity ID.
        
        This is required if the SAML2 provider metadata endpoint contains multiple entities.
        """
        return os.environ.get('SAML2_ISSUER_ENTITY_ID', None)

    @property
    def SAML2_ENTITY_ID(self):
        """
        SAML2 provider entity ID of the application.
        """
        return os.environ.get('SAML2_ENTITY_ID', None)

    @property
    def SAML2_PUBLIC_KEY(self):
        """
        SAML2 public key for this application.

        To generate, see SAML2_PRIVATE_KEY
        """
        return os.environ.get('SAML2_PUBLIC_KEY', None)

    @property
    def SAML2_PRIVATE_KEY(self):
        """
        SAML2 private key for this application.
        
        To generate, run:
        ```
        openssl genrsa -out private.key 4096
        openssl req -new -x509 -key private.key -out publickey.cer -days 365
        # Export values to environment variables
        export SAML2_PRIVATE_KEY="$(cat private.key)"
        export SAML2_PUBLIC_KEY="$(cat publickey.cer)"
        ```
        """
        return os.environ.get('SAML2_PRIVATE_KEY', None)

    @property
    def SAML2_GROUP_ATTRIBUTE(self):
        """
        SAML2 user data group attribute.
        """
        return os.environ.get('SAML2_GROUP_ATTRIBUTE', 'groups')

    @property
    def SAML2_DEBUG(self):
        """
        Enable debug of Saml2 via stdout.

        This should only be enabled for non-production environments.
        """
        return self.convert_boolean(os.environ.get('SAML2_DEBUG', 'False'))

    @property
    def DEFAULT_TERRAFORM_VERSION(self):
        """
        Default version of Terraform that will be used to extract module, if terraform required_version has not been specified.
        """
        return os.environ.get("DEFAULT_TERRAFORM_VERSION", "1.3.6")

    @property
    def TERRAFORM_ARCHIVE_MIRROR(self):
        """
        Mirror for obtaining version list and downloading Terraform
        """
        return os.environ.get("TERRAFORM_ARCHIVE_MIRROR", "https://releases.hashicorp.com/terraform")

    @property
    def MANAGE_TERRAFORM_RC_FILE(self):
        """
        Whether terrareg with manage (overwrite) the terraform.rc file in the user's home directory.

        This is required for terraform to authenticate to the registry to ignore any analytics when initialising Terraform modules during extraction.

        When this is disabled, analytics will be recorded when Terrareg extracts modules that call to other modules in the registry.

        This is disabled by default in the application, meaning that running terrareg locally, by default, will not manage this file.
        The docker container, by default, overrides this to enable the functionality, since it is running in an isolated environment and unlikely to overwrite user's own configurations.
        """
        return self.convert_boolean(os.environ.get("MANAGE_TERRAFORM_RC_FILE", "False"))

    @property
    def SENTRY_DSN(self):
        """DSN Integration URL for sentry"""
        return os.environ.get("SENTRY_DSN", "")

    @property
    def SENTRY_TRACES_SAMPLE_RATE(self):
        """
        Sample rate for capturing traces in sentry.

        Must be a number between 0.0 and 1.0
        """
        return float(os.environ.get("SENTRY_TRACES_SAMPLE_RATE", "1.0"))

    def convert_boolean(self, string):
        """Convert boolean environment variable to boolean."""
        if string.lower() in ['true', 'yes', '1']:
            return True
        elif string.lower() in ['false', 'no', '0']:
            return False

        raise InvalidBooleanConfigurationError('Boolean config value not valid. Must be one of: true, yes, 1, false, no, 0')
