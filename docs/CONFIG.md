## Application environment variables

The following environment variables are available to configure the application



### ADDITIONAL_MODULE_TABS


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


Default: `[["Release Notes", ["RELEASE_NOTES.md", "CHANGELOG.md"]], ["License", ["LICENSE"]]]`


### ADMIN_AUTHENTICATION_TOKEN


Password/API key to for authentication as the built-in admin user.


Default: ``


### ADMIN_SESSION_EXPIRY_MINS


Session timeout for admin cookie sessions


Default: `60`


### ALLOWED_PROVIDERS


Comma-seperated list of allowed providers.

Leave empty to disable allow-list and allow all providers.


Default: ``


### ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER


Whether module providers can specify their own git repository source.


Default: `True`


### ALLOW_CUSTOM_GIT_URL_MODULE_VERSION


Whether module versions can specify git repository in terrareg config.


Default: `True`


### ALLOW_FORCEFUL_MODULE_PROVIDER_REDIRECT_DELETION


Whether to allow the force deletion of module provider redirects.

Force deletion is required if module calls are still using the redirect and analytics tokens indicate that
some have not used migrated to the new name.


Default: `False`


### ALLOW_MODULE_HOSTING


Whether uploaded modules can be downloaded directly.
If disabled, all modules must be configured with a git URL.


Default: `True`


### ALLOW_UNIDENTIFIED_DOWNLOADS


Whether modules can be downloaded with Terraform
without specifying an identification string in
the namespace


Default: `False`


### ANALYTICS_AUTH_KEYS


List of comma-separated values for Terraform auth tokens for deployment environments.

E.g. `xxxxxx.deploy1.xxxxxxxxxxxxx:dev,zzzzzz.deploy1.zzzzzzzzzzzzz:prod`
In this example, in the 'dev' environment, the auth token for Terraform would be: `xxxxxx.deploy1.xxxxxxxxxxxxx`
and the auth token for Terraform for prod would be: `zzzzzz.deploy1.zzzzzzzzzzzzz`.

To disable auth tokens and to report all downloads, leave empty.

To only record downloads in a single environment, specify a single auth token. E.g. `zzzzzz.deploy1.zzzzzzzzzzzzz`

For information on using these API keys, please see Terraform: https://docs.w3cub.com/terraform/commands/cli-config.html#credentials


Default: ``


### ANALYTICS_TOKEN_DESCRIPTION

Describe to be provided to user about analytics token (e.g. `The name of your application`)

Default: ``


### ANALYTICS_TOKEN_PHRASE

Name of analytics token to provide in responses (e.g. `application name`, `team name` etc.)

Default: `analytics token`


### APPLICATION_NAME

Name of application to be displayed in web interface.

Default: `Terrareg`


### AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION


Whether to automatically generate module provider descriptions, if they are not provided in terrareg metadata file of the module.


Default: `True`


### AUTOGENERATE_USAGE_BUILDER_VARIABLES


Whether to automatically generate usage builder variables from the required variables and their descriptions.
When disabled, the usage builder will only be displayed on a module when the "variable_template" section
of the terrareg.json metadata file is populated.


Default: `True`


### AUTO_CREATE_MODULE_PROVIDER


Whether to automatically create module providers when
uploading module versions, either from create endpoint or hooks.

If disabled, modules must be created using the module provider create endpoint (or via the web interface).


Default: `True`


### AUTO_CREATE_NAMESPACE


Whether to automatically create namespaces when
uploading a module version to a module provider in a non-existent namespace.

If disabled, namespaces must be created using the namespace create endpoint (or via web UI)


Default: `True`


### AUTO_PUBLISH_MODULE_VERSIONS


Whether new module versions (either via upload, import or hook) are automatically
published and available.

If this is disabled, the publish endpoint must be called before the module version
is displayed in the list of module versions.

NOTE: Even whilst in an unpublished state, the module version can still be accessed directly, but not used within Terraform.


Default: `True`


### CONTRIBUTED_NAMESPACE_LABEL

Custom name for 'contributed namespace' in UI.

Default: `Contributed`


### DATABASE_URL


URL for database.
Defaults to local sqlite database.

To setup SQLite datbase, use `sqlite:///<path to sqlite DB>`

To setup MySQL, use `mysql+mysqlconnector://<user>:<password>@<host>[:<port>]/<database>`


Default: `sqlite:///modules.db`


### DATA_DIRECTORY



Default: `./data`


### DEBUG

Whether flask and sqlalchemy is setup in debug mode.

Default: `False`


### DEFAULT_TERRAFORM_VERSION


Default version of Terraform that will be used to extract module, if terraform required_version has not been specified.


Default: `1.3.6`


### DELETE_EXTERNALLY_HOSTED_ARTIFACTS


Whether uploaded modules, that provide an external URL for the artifact,
should be removed after analysis.
If enabled, module versions with externally hosted artifacts cannot be re-analysed after upload.


Default: `False`


### DISABLE_ANALYTICS


Disable module download anaytics.

This stops analytics tokens being displayed in the UI.

This also sets `ALLOW_UNIDENTIFIED_DOWNLOADS` to True


Default: `False`


### DISABLE_TERRAREG_EXCLUSIVE_LABELS


Whether to disable 'terrareg exclusive' labels from feature tabs in UI.

Set to 'True' to disable the labels.


Default: `False`


### DOMAIN_NAME


Domain name that the system is hosted on.

This should be setup for all installations, but is required for infracost and OpenID authentication.

Note: The configuration is deprecated, please set the `PUBLIC_URL` configuration instead.


Default: ``


### ENABLE_ACCESS_CONTROLS


Enables role based access controls for SSO users.

Enabling this feature will restrict all SSO users from performing any admin tasks.
Group mappings can be setup in the settings page, providing SSO users and groups with global or namespace-based permissions.

When disabled, all SSO users will have global admin privileges.


Default: `False`


### ENABLE_SECURITY_SCANNING


Whether to perform security scans of uploaded modules and display them against the module, submodules and examples.


Default: `True`


### EXAMPLES_DIRECTORY


Directory with a module's source that contains examples.

Examples are expected to be within sub-directories of the examples directory.

E.g. If EXAMPLES_DIRECTORY is set to `examples`, with the root module, the following would be expected for an example: `examples/myexample/main.tf`.


Default: `examples`


### EXAMPLE_ANALYTICS_TOKEN


Example analytics token to provide in responses (e.g. my-tf-application, my-slack-channel etc.).

Note that, if this token is used in a module call, it will be ignored and treated as if
an analytics token has not been provided.
If analytics tokens are required, this stops users from accidently using the example placeholder in
Terraform projects.


Default: `my-tf-application`


### EXAMPLE_FILE_EXTENSIONS


Comma-separated list of file extensions to be extracted/shown in example file lists.

Example: `tf,sh,json`

Supported languages for syntax highlighting:
* HCL
* Javascript/JSON
* Bash
* Batch
* PL/SQL
* PowerShell
* Python
* Dockerfile/docker-compose

NOTE: For new file types to be shown module versions must be re-indexed


Default: `['tf', 'tfvars', 'sh', 'json']`


### GIT_CLONE_TIMEOUT


Timeout for git clone commands in seconds.

Leave empty to disable timeout for clone.

Warning - if a git clone is performed without a timeout set, the request may never complete and could leave locks on the database, requiring an application restart.


Default: `300`


### GIT_PROVIDER_CONFIG


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
If using SSH, the domain must be seperated by the path using a forward slash. Use a colon to specify a port (e.g. `ssh://gitlab.corp.com:7999/namespace/name.git`)

- browse_url - Formatted URL for user-viewable source code
(e.g. 'https://github.com/{namespace}/{module}-{provider}/tree/{tag}/{path}'
or 'https://bitbucket.org/{namespace}/{module}/src/{version}?at=refs%2Ftags%2F{tag_uri_encoded}').
Must include placeholdes:
- {path} (for source file/folder path)
- {tag} or {tag_uri_encoded} for the git tag

An example for public repositories, using SSH for cloning, might be:
```
[{"name": "Github", "base_url": "https://github.com/{namespace}/{module}", "clone_url": "ssh://git@github.com/{namespace}/{module}.git", "browse_url": "https://github.com/{namespace}/{module}/tree/{tag}/{path}"},
{"name": "Bitbucket", "base_url": "https://bitbucket.org/{namespace}/{module}", "clone_url": "ssh://git@bitbucket.org/{namespace}/{module}-{provider}.git", "browse_url": "https://bitbucket.org/{namespace}/{module}-{provider}/src/{tag}/{path}"},
{"name": "Gitlab", "base_url": "https://gitlab.com/{namespace}/{module}", "clone_url": "ssh://git@gitlab.com/{namespace}/{module}-{provider}.git", "browse_url": "https://gitlab.com/{namespace}/{module}-{provider}/-/tree/{tag}/{path}"}]
```


Default: `[]`


### IGNORE_ANALYTICS_TOKEN_AUTH_KEYS


A list of a Terraform auth keys that can be used to authenticate to the registry
to ignore check for valid analaytics token in module path.

It is recommended that examples in modules do not include analytics tokens in calls to other modules
hosted in the registry, to avoid having to add analytics tokens to test the examples during development
of the modules.

No analytics of the module usage are captured when this auth key is used.

The value should be a comma-seperated list of auth keys.

The auth key can be used by placing in a "credential" block in the user's .terraformrc file.


Default: ``


### INFRACOST_API_KEY


API key for Infracost.

Set this to enable cost-analysis of module examples.

To generate an API key:
Log in at https://dashboard.infracost.io > select your organization > Settings

For cost analysis to be performed on modules which utilise other modules from this registry, ensure `DOMAIN_NAME` is set.


Default: ``


### INFRACOST_PRICING_API_ENDPOINT


Self-hosted infracost pricing API endpoint.

For information on self-hosting the infracost pricing API, see https://www.infracost.io/docs/cloud_pricing_api/self_hosted/


Default: ``


### INFRACOST_TLS_INSECURE_SKIP_VERIFY


Whether to skip TLS verification for self-hosted pricing endpoints


Default: `False`


### INTERNAL_EXTRACTION_ANALYTICS_TOKEN


Analaytics token used by Terraform initialised by the registry.

This is used by the registry to call back to itself when analysing module examples.

The value should be changed if it might result in a conflict with a legitimate analytics token used in Terraform
that calls modules from the registry.


Default: `internal-terrareg-analytics-token`


### LISTEN_PORT


Port for server to listen on.


Default: `5000`


### LOGO_URL

URL of logo to be used in web interface.

Default: `/static/images/logo.png`


### MANAGE_TERRAFORM_RC_FILE


Whether terrareg with manage (overwrite) the terraform.rc file in the user's home directory.

This is required for terraform to authenticate to the registry to ignore any analytics when initialising Terraform modules during extraction.

When this is disabled, analytics will be recorded when Terrareg extracts modules that call to other modules in the registry.

This is disabled by default in the application, meaning that running terrareg locally, by default, will not manage this file.
The docker container, by default, overrides this to enable the functionality, since it is running in an isolated environment and unlikely to overwrite user's own configurations.


Default: `False`


### MODULES_DIRECTORY


Directory with a module's source that contains sub-modules.

submodules are expected to be within sub-directories of the submodule directory.

E.g. If MODULES_DIRECTORY is set to `modules`, with the root module, the following would be expected for a submodule: `modules/submodulename/main.tf`.

This can be set to an empty string, to expected submodules to be in the root directory of the parent module.


Default: `modules`


### MODULE_LINKS


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


Default: `[]`


### MODULE_VERSION_REINDEX_MODE


This configuration defines how re-indexes a module version, that already exists, behaves.

This can be set to one of:
* 'legacy' - The new module version will replace the old one. Until the version is re-published, it will not be available to Terraform. Analytics for the module version will be retained.
* 'auto-publish' - The new module version will replace the old one. If the previous version was published, the new version will be automatically published. Analytics for the module version will be retained.
* 'prohibit' - If a module version has already been indexed, it cannot be re-indexed via hooks/API calls without the version first being deleted.


Default: `legacy`


### OPENID_CONNECT_CLIENT_ID


Client ID for OpenID Conect authentication


Default: ``


### OPENID_CONNECT_CLIENT_SECRET


Client secret for OpenID Conect authentication


Default: ``


### OPENID_CONNECT_DEBUG


Enable debug of OpenID connect via stdout.

This should only be enabled for non-production environments.


Default: `False`


### OPENID_CONNECT_ISSUER


Base Issuer URL for OpenID Conect authentication.

A well-known URL will be expected at `${OPENID_CONNECT_ISSUER}/.well-known/openid-configuration`


Default: ``


### OPENID_CONNECT_LOGIN_TEXT


Text for sign in button for OpenID Connect authentication


Default: `Login using OpenID Connect`


### OPENID_CONNECT_SCOPES


Comma-seperated list of scopes to be included in OpenID authentication request.

The OpenID profile should provide a 'groups' attribute, containing a list of groups
that the user is a member of.

A common configuration may require a 'groups' scope to be added to the list of scopes.


Default: `['openid', 'profile']`


### PUBLIC_URL


The URL that is used for accessing Terrareg by end-users.

E.g.:
`https://terrareg.mycorp.com`
`https://terrareg.mycorg.com:8443`


Ensure the protocol is https if Terrareg is accessible over https.
Provide the port if necessary.

If left empty, the protocol will be assumed to be HTTPS, the port will be assumed to be 443
and the domain will fallback to be the value set in `DOMAIN_NAME`.


Default: ``


### PUBLISH_API_KEYS


List of comma-separated list of API keys to publish module versions.

To disable authentication for publish endpoint, leave empty.


Default: ``


### REDIRECT_DELETION_LOOKBACK_DAYS


Number of days' worth of analytics data to use to determine if a redirect is still in use.

For example, if set to 1, if a Terraform module was accessed via a redirect in the past 1 day, it will require
forceful deletion to delete (unless a more recent download of the module by the same analytics token no longer uses the redirect).

Value of `0` disables the lookback and redirects can always be removed without force

Value of `-1` will not limit the lookback period and all analytics will be used.


Default: `-1`


### REQUIRED_MODULE_METADATA_ATTRIBUTES


Comma-seperated list of metadata attributes that each uploaded module _must_ contain, otherwise the upload is aborted.


Default: ``


### SAML2_DEBUG


Enable debug of Saml2 via stdout.

This should only be enabled for non-production environments.


Default: `False`


### SAML2_ENTITY_ID


SAML2 provider entity ID of the application.


Default: ``


### SAML2_GROUP_ATTRIBUTE


SAML2 user data group attribute.


Default: `groups`


### SAML2_IDP_METADATA_URL


SAML2 provider metadata url


Default: ``


### SAML2_ISSUER_ENTITY_ID


SAML2 provider entity ID.

This is required if the SAML2 provider metadata endpoint contains multiple entities.


Default: ``


### SAML2_LOGIN_TEXT


Text for sign in button for SAML2 authentication


Default: `Login using SAML`


### SAML2_PRIVATE_KEY


SAML2 private key for this application.

To generate, run:
```
openssl genrsa -out private.key 4096
openssl req -new -x509 -key private.key -out publickey.cer -days 365
# Export values to environment variables
export SAML2_PRIVATE_KEY="$(cat private.key)"
export SAML2_PUBLIC_KEY="$(cat publickey.cer)"
```


Default: ``


### SAML2_PUBLIC_KEY


SAML2 public key for this application.

To generate, see SAML2_PRIVATE_KEY


Default: ``


### SECRET_KEY


Flask secret key used for encrypting sessions.

Can be generated using: `python -c 'import secrets; print(secrets.token_hex())'`


Default: ``


### SENTRY_DSN

DSN Integration URL for sentry

Default: ``


### SENTRY_TRACES_SAMPLE_RATE


Sample rate for capturing traces in sentry.

Must be a number between 0.0 and 1.0


Default: `1.0`


### SERVER


Set the server application used for running the application. Set the `SERVER` environment variable to one of the following options:

* `builtin` - Use the default built-in flask web server. This is less performant and is no longer recommended for production use-cases.
* `waitress` - Uses [waitress](https://docs.pylonsproject.org/projects/waitress/en/latest/index.html) for running the application. This does not support SSL offloading, meaning that it must be used behind a reverse proxy that performs SSL-offloading.


Default: `builtin`


### SSL_CERT_PRIVATE_KEY


Path to SSL private certificate key.

If running in a container, the key must be mounted inside the container.
This value must be set to the path of the key within the container.

This must be set in accordance with SSL_CERT_PUBLIC_KEY - both must either be
set or left empty.


Default: ``


### SSL_CERT_PUBLIC_KEY


Path to SSL public key.

If running in a container, the key must be mounted inside the container.
This value must be set to the path of the key within the container.

This must be set in accordance with SSL_CERT_PRIVATE_KEY - both must either be
set or left empty.


Default: ``


### TERRAFORM_ARCHIVE_MIRROR


Mirror for obtaining version list and downloading Terraform


Default: `https://releases.hashicorp.com/terraform`


### TERRAFORM_EXAMPLE_VERSION_TEMPLATE


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



Default: `{major}.{minor}.{patch}`


### THREADED

Whether flask is configured to enable threading

Default: `True`


### TRUSTED_NAMESPACES

Comma-separated list of trusted namespaces.

Default: ``


### TRUSTED_NAMESPACE_LABEL

Custom name for 'trusted namespace' in UI.

Default: `Trusted`


### UPLOAD_API_KEYS


List of comma-separated list of API keys to upload/import new module versions.

For bitbucket hooks, one of these keys must be provided as the 'secret' to the webhook.

To disable authentication for upload endpoint, leave empty.


Default: ``


### VERIFIED_MODULE_LABEL

Custom name for 'verified module' in UI.

Default: `Verified`


### VERIFIED_MODULE_NAMESPACES


List of namespaces, who's modules will be automatically set to verified.


Default: ``

