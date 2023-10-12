# User Guide

## Contents

* [Deployment](#deployment)
  * [Docker Environment Variables](#docker-environment-variables)
  * [Application environment variables](#application-environment-variables)
  * [Database Migrations](#database-migrations)
* [Security](#security)
   * [Single Single-On](#single-sign-on)
   * [Github Authentication](#github-authentication)
   * [Disable Public Access](#disable-public-access)
* [Module best practices](#module-best-practices)
* [Uploading Modules](#uploading-modules)
* [Security Scanning](#security-scanning)
* [Cost Analysis](#cost-analysis)
* [Module storage](#module-storage)
* [Git Providers](#git-providers)
* [Create modules in the registry](#create-modules-in-the-registry)
* [Module Usage Analytics](#module-usage-analytics)
* [Customising Terrareg UI](#customising-terrareg-ui)
* [API endpoints](./API.md)


# Deployment

## Docker environment variables

The following environment variables are available to configure the docker container

### MIGRATE_DATABASE

Whether to perform a database migration on container startup.

Set to `True` to enable database migration

*Note:* Be very careful when scaling the application. There should never be more than one instance of Terrareg running with `MIGRATE_DATABASE` set to `True` during an upgrade.

When upgrading, scale the application to a single instance before upgrading to a newer version.

Alternatively, set `MIGRATE_DATABASE` to `False` and run a dedicated instance for performing database upgrades.
Use `MIGRATE_DATABASE_ONLY` to run an instance that will perform the necessary database migrations and immediately exit.

Default: `False`

### MIGRATE_DATABASE_ONLY

Whether to perform database migration and exit immediately.

This is useful for scheduling migrations by starting a 'migration' instance of the application.

Set to `True` to exit after migration.

The `MIGRATE_DATABASE` environment variable must also be set to `True` to perform the migration, otherwise nothing will be performed and the container will exit.

Default: `False`

### SSH_PRIVATE_KEY

Provide the contents of the SSH key to perform git clones.

This is an alternative to mounting the '.ssh' directory of the root user.

Default: ''


## Application environment variables

For a list of available application configuration environment variables, please see [docs/CONFIG.md](./CONFIG.md)


## Database Migrations

Terrareg can be deployed via Docker and scaled out to support high-availability and load requirements.

However, whilst perform upgrades with database migrations, it's important to ensure that only one container performs the database migration step.

This can be accomplished in two ways:
 * Scale down to 1 container when performing upgrades that container a database migration
 * Run a single dedicated container that performs the database upgrades.

In either situation, when performing a database upgrade, it is highly recommended that any containers serving the web-application are stopped.

### On-the-fly DB migrations

To enable database migrations in all containers (assuming the service will be scaled to a single container during upgrade), set the environment variable `MIGRATE_DATABASE` to `True`.

### Dedicated DB migration container

To dedicate a single container to DB migrations, set `MIGRATE_DATABASE` to `False` on all containers running the web application and create a new container

## Allowing Terrareg to Communicate with itself

During module extraction/analysis, Terrareg will need to communicate with itself, which is required during cost analysis and graph generation.

To configure this, set the [DOMAIN_NAME](./CONFIG.md#domain_name) configuration.

To ensure terraform does not generate unecessary analytics in the module, terrareg must manage the .terraformrc file in the user's home directory.
This functionality is enabled by default in the docker container, by disabled outside of it. To enable/disable this functionality, see [MANAGE_TERRAFORM_RC_FILE](./CONFIG.md#manage_terraform_rc_file)

## Docker storage

### Module data

If [module hosting](#module-hosting) is being used, ensure that a directory is mounted into the container for storing module data.
This path can be customised by setting [DATA_DIRECTORY](./CONFIG.md#data_directory)

## Database URL

A database URL should be configured, otherwise Terrareg will default to a local sqlite database (though this _could_ be mounted via a docker volume, it certainly shouldn't be used for multiple containers).

This should be configured by setting [DATABASE_URL](./CONFIG.md#database_url)

## Listen port

The port that Terrareg listens on can be configured with [LISTEN_PORT](./CONFIG.md#listen_port)

## SSL

Although Terrareg can be deployed without SSL - this is only recommended for testing and local development.
Aside from the usualy reasons for using SSL, it is also required for Terraform to communicate with the registry to obtain modules as a registry provider. If SSL is not used, Terrareg will fall-back to providing Terrareform examples using a 'http' download URL for Terraform.

Terrareg must be configured with the URL that the registry is accessible. To configure this, please see [PUBLIC_URL](./CONFIG.md#public_url)


### Enabling SSL on the application

SSL can be enabled on Terrareg itself - the certificates must be mounted inside the container (or be available on the filesystem, if running outside of a docker container) and the absolute path can be provided using the environment variables [SSL_CERT_PRIVATE_KEY](./CONFIG.md#ssl_cert_private_key) and [SSL_CERT_PUBLIC_KEY](./CONFIG.md#ssl_cert_public_key).

If Terrareg is being run outside of a docker container, these can be provided as command line arguments `--ssl-cert-private-key` and `--ssl-cert-public-key`.

### Offloading SSL using a reverse proxy

SSL can also be provided by a reverse proxy in front of Terrareg and traffic to the Terrareg container can be served via HTTP.


# Security

Terrareg must be configured with a secret key, which is used for encrypting sessions, to enable authentication. This can be configured via [SECRET_KEY](./CONFIG.md#secret_key).

By default, Terrareg administration is protected by an [ADMIN_AUTHENTICATION_TOKEN](./CONFIG.md#admin_authentication_token), which is set via environment variable, which becomes the password used for authentication on the login page.

Authentication is required to perform tasks such as: creating namespaces, creating/deleting modules and managing additional user permissions, with a few exceptions:

### Securing module creation/indexing

By default, indexing and publishing modules does *not* require authentication. To disable unauthorised indexing/publishing of modules, set up dedicated API keys for these functions, see [UPLOAD_API_KEYS](./CONFIG.MD#upload_api_keys) and [PUBLISH_API_KEYS](./CONFIG.MD#Ppublish_api_keys).

API keys can be specified in requests to API endpoing with an 'X-Terrareg-ApiKey' header.

Enabling [access controls](#access-controls) will disable unauthenticated indexing/publishing of modules, requiring authentication via SSO or Admin password or, once set, the upload/publish API keys.

### Auto creation of namespaces/modules

By default, the module upload endpoint will create a namespace/module if it does not exist.

For a secure installation, it is recommended to disable this, forcing all namespaces/modules to be created by authenticated users.

The disable this functionality, set [AUTO_CREATE_NAMESPACE](./CONFIG.md#auto_create_namespace) and [AUTO_CREATE_MODULE_PROVIDER](./CONFIG.md#auto_create_module_provider)

## API authentication

Each protected API endpoint (generally all API endpoints that are non read-only actions) require API key authentication, with the exception of module upload/publish endpoints (see [Security module creationg/indexing](#securing-module-creationindexing)).

To authenticate to an API endpoint, you must either:
 * Pass the approrpriate API key for the endpoint (specifically for module upload/publishing), in a `X-Terrareg-ApiKey` header;
 * Pass the admin authentication token to the endpoint in a `X-Terrareg-ApiKey` header;
 * Authenticate via the authentication API endpoint and authenticate using session (not a programatic approach).

There will be future work to allow API key authentication for all endpoints in future: https://gitlab.dockstudios.co.uk/pub/terrareg/-/issues/353

## Access controls

By default, all authenticated users will have global admin permissions, allowing the creation/deletion of all namespaces, modules and module versions.

Role based access controls (RBAC) can be enabled by setting [ENABLE_ACCESS_CONTROLS](./CONFIG.md#enable_access_controls).

This will remove any default privilges given to SSO users. The admin user will still retain global admin privileges.

For information about user groups, see [Single Sign-on](#single-sign-on).

## Single Sign-on

Single sign-on can be used to allow authentication via external authentication providers.

SSO groups can be assigned global admin permissions, or per-namespace permissions, allowing creation/deletion of modules and versions.

User groups and permissions can be configured on the 'User Groups' (/user-groups) page.

Once single sign-on has been setup, the [ADMIN_AUTHENTICATION_TOKEN](./CONFIG.md#admin_authentication_token) can be removed, stopping further sign-in via password authentication.

*NOTE* OpenID and SAML2 authentication is currently experimental.

It is advised to use with caution and avoid using on publicly hosted/accessible instances.

### OpenID Connect

To configure OpenID connect, setup an application in an identity provider (IdP) with the following:

| | |
| --- | --- |
| Grant type | Authorization Code |
| Sign-in redirect URI | `https://<terrareg-instance-domain>/openid/callback` |
| Sign-out URI | `https://<terrareg-instance-domain>/logout` |
| Initiate login URI | `https://<terrareg-instance-domain>/openid/login` |
| Login flow | Redirect to app to initiate login |

Obtain the client ID, client secret and issuer URL from the IdP provider and populate the following environment variables:

 * [PUBLIC_URL](./CONFIG.md#public_url)
 * [OPENID_CONNECT_CLIENT_ID](./CONFIG.md#openid_connect_client_id)
 * [OPENID_CONNECT_CLIENT_SECRET](./CONFIG.md#openid_connect_client_secret)
 * [OPENID_CONNECT_ISSUER](./CONFIG.md#openid_connect_issuer)

Note: Most IdP providers will require the terrareg installation to be accessed via https.
The instance should be configured with SSL certificates ([SSL_CERT_PRIVATE_KEY](./CONFIG.md#ssl_cert_private_key)/[SSL_CERT_PUBLIC_KEY](./CONFIG.md#ssl_cert_public_key)) or be hosted behind a reverse-proxy/load balancer.

The text displayed on the login button can be customised by setting [OPENID_CONNECT_LOGIN_TEXT](./CONFIG.md#openid_connect_login_text)

### SAML2

Generate a public and a private key, using:

    openssl genrsa -out private.key 4096
    openssl req -new -x509 -key private.key -out publickey.cer -days 365

Set the folllowing environment variables:

* [PUBLIC_URL](./CONFIG.md#public_url) (required)
* [SAML2_IDP_METADATA_URL](./CONFIG.md#saml2_idp_metadata_url) (required)
* [SAML2_ENTITY_ID](./CONFIG.md#saml2_entity_id) (required)
* [SAML2_PRIVATE_KEY](./CONFIG.md#saml2_private_key) (required) (See above)
* [SAML2_PUBLIC_KEY](./CONFIG.md#saml2_public_key) (required) (See above)
* [SAML2_ISSUER_ENTITY_ID](./CONFIG.md#saml2_issuer_entity_id) (optional)

In the IdP:

* configure the Single signin URL to `https://{terrareg_installation_domain}/saml/login?sso`;
* configure the request and response to be signed;
* ensure at least one attribute is assigned.


### Github Authentication

Terrareg can be configured to authenticate users via Github.

This makes several changes to the workflow for users:

The following configurations must be configured (see CONFIG.md for details about how to generate them):
 * GITHUB_APP_CLIENT_ID
 * GITHUB_APP_CLIENT_SECRET
 * GITHUB_APP_WEBHOOK_SECRET
 * GITHUB_APP_PRIVATE_KEY_PATH

For self-hosted Github Enterprise installations, `GITHUB_URL` and `GITHUB_API_URL` can be set

Github users are automatically mapped to any groups in Terrareg that are named after organisations that they are named after.
E.g. User `testuser`, who is a member of the `testorg` organisation would be mapped to the groups `testuser` and `testorg` Terrareg group, if they exist.


#### Setup Github app

To setup the application for Github, goto `https://github.com/settings/apps/new` (or `https://github.com/organizations/<organisation>/settings/apps`)

Configure the following:

 * Homepage URL: https://my-terrareg-installation.example.com
 * Identifying and authorizing users
   * Callback URL: https://my-terrareg-installation.example.com/github/callback
   * Expire user authorization tokens: Check
   * Request user authorization (OAuth) during installation: Check
   * Enable Device Flow: Uncheck
 * Post installation
   * Setup URL: `Not set`
   * Redirect on update: Uncheck
 * Webhook
   * Active: Check
   * Webhook URL: https://my-terrareg-installation.example.com/github/webhook
   * Webhook secret: `Value from GITHUB_APP_WEBHOOK_SECRET`
 * Permissions:
   * None
 * Where can this GitHub App be installed?
   * Select based on whether you with to limit who can authenticate to Terrareg (publicly or members of the organisation)

Generate client ID and client secret:

 * On the same page, obtain the client ID from the "about" section of the page. Use this to populate `GITHUB_APP_CLIENT_ID`
 * Click "Generate a new client secret"
 * Copy the value from the client secret to populate `GITHUB_APP_CLIENT_SECRET`


Once created, generate a private key (this is not currently used, but will be in future):

 * On the same page, find `Private keys` and click "Generate a private key"
 * Download the file and make the file accessible to the Terrareg installation, set in `GITHUB_APP_PRIVATE_KEY_PATH`

#### Github namespace mapping

If `AUTO_GENERATE_GITHUB_ORGANISATION_NAMESPACES` is enabled, namespaces are automatically created for the logged in Github user and all organisations that they are an owner of.

The user has full permissions to each of these namespaces, allowing them to create modules in them.


## Disable Unauthenticated Access

Unauthenticated access can be disabled, enforcing all users to authenticate to use Terrareg, by disabling the setting [ALLOW_UNAUTHENTICATED_ACCESS](./CONFIG.md#allow_unauthenticated_access).

Disabling unauthenticated access requires Terraform to authenticate to obtain modules, see below.

### Enabling Terraform authentication

Terraform can authenticate to the registry via the built-in authentication mechanisms of Terrareg. That is, Terraform attempts to authenticate to Terrareg by redirecting the user to Terrareg. Terrareg ensures the user is authenticated and provides a redirect back to Terraform to finalise authentication using OpenIDC.

For Terrareg to be able to authenticate users, several configurations must be provided:
 
 * [TERRAFORM_OIDC_IDP_SUBJECT_ID_HASH_SALT](./CONFIG.md#terraform_oidc_idp_subject_id_hash_salt)
 * [TERRAFORM_OIDC_IDP_SIGNING_KEY_PATH](./CONFIG.md#terraform_oidc_idp_signing_key_path)
 * [PUBLIC_URL](./CONFIG.md#public_url)
 * [TERRAFORM_PRESIGNED_URL_SECRET](./CONFIG.md#terraform_presigned_url_secret)

Consider reviewing the default values for:
 * [TERRAFORM_OIDC_IDP_SESSION_EXPIRY](./CONFIG.md#terraform_oidc_idp_session_expiry)
 * [TERRAFORM_PRESIGNED_URL_EXPIRY](./CONFIG.md#terraform_presigned_url_expiry)

Once these are configured, users can authenticate via terraform using:
```
terraform login <registry-fqdn>
# e.g. terraform login my-registry.example.com
```


# Module best practices

Terrareg has been extensively tested with Terraform modules of all shapes and sizes, meaning that it should be able to provide valuable information without any modification to modules before indexing.

However, to get the _most_ out of Terrareg, there are some practices/guides that will help.

## Terrareg metadata file

A metadata file can be provided each an uploaded module's archive to provide additional metadata to terrareg.

This should be called `terrareg.json` or `.terrareg.json` and be placed in the root of the module.

For an example, please see: [docs/example-terrareg-module-metadata.json](./example-terrareg-module-metadata.json)

The following attributes are available at the root of the JSON object:

|Key |Description|
--- | --- |
|owner|Name of the owner of the module|
|description|Description of the module.|
|variable_template|Structure holding required input variables for module, used for 'Usage Builder'. See table below|
|repo_clone_url|Url to clone the repository. Optional placeholders `{namespace}`, `{module}` and `{provider}` can be used. E.g. `ssh://gitlab.corporate.com/scm/{namespace}/{module}.git`|
|repo_base_url|Formatted base URL for project's repo. E.g. `https://gitlab.corporate.com/{namespace}/{module}`|
|repo_browse_url|Formatted URL for user-viewable source code. Must contain `{tag}` and `{path}` placeholders. E.g. `https://github.com/{namespace}/{module}-{provider}/blob/{tag}/{path}`|

For information on the repo URLs, see [Git Providers](#git-providers)

Each of these attributes can be enforced in modules uploaded to the registry by setting [REQUIRED_MODULE_METADATA_ATTRIBUTES](./CONFIG.md#required_module_metadata_attributes).

## Description

If a metadata file is not present or a description is not provided, Terrareg will attempt to automatically generate a description of the module, using the README.md from the module.

This functionality can be disabled by setting [AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION](CONFIG.md#autogenerate_module_provider_description).

### Usage builder configuration

The usage builder requires an array of objects, which define the name, type and description of the variable.

In the following the 'config input' refers to the HTML inputs that provide the user with the ability to select/enter values. The 'terraform input' refers to the value used for the variable in the outputted terraform example.

There are common attributes that can be added to each of variable objects, which include:

|Attribute |Description |Default|
--- | --- | ---|
|name|The name of the 'config input'. This is also used as the module variable in the 'terraform input'.|Required|
|type|The type of the input variable, see table below.|Required|
|required|Boolean flag to determine whether the variable is Required.|`true`|
|quote_value|Boolean flag to determine whether the value generated is quoted for the 'terraform input'.|`false`|
|additional_help|A description that is provided, along with the terraform variable description in the usage builder|`""`|
|default_value|The default value if required is false|`null`|



|Variable type|Description|Type specific attributes|
--- | --- | ---|
|text|A plain input text box for users to provide a value that it directly used as the 'terraform input'||
|boolean|Provides a checkbox that results in a true/false value as the 'terraform input'||
|static|This does not appear in the 'Usage Builder' 'config input' table, but provides a static value in the 'terraform input'||
|select|Provides a dropdown for the user to select from a list of choices|"choices" must be added to the object, which may either be a list of strings, or a list of objects. If using a list of objects, a "name" and "value" must be provided. Optionally an "additional_content" attribute can be added to the choice, which provides additional terraform to be added to the top of the terraform example. The main variable object may also contain a "allow_custom" boolean attribute, which allows the user to enter a custom text input.|

Terrareg will automatically generated usage builder inputs based on discovered variables in the module. This functionality can be disabled by setting [AUTOGENERATE_USAGE_BUILDER_VARIABLES](./CONFIG.md#autogenerate_usage_builder_variables)

## Submodules


By default, sub-modules are located in individual sub-directories of the `modules` directory of the module, e.g.:

```
 <Root of Module>
 |
 | -> modules
 |    |
 |    | -> s3
 |    |    |
 |    |     -> main.tf
 |    |
 |    | -> route53
 |    |    |
 |    |     -> main.tf
     
```

This directory can be changed on a global level with [MODULES_DIRECTORY](./CONFIG.md#modules_directory)

## Examples

By default, examples are located in individual sub-directories of the `examples` directory of the module, e.g.:

```
 <Root of Module>
 |
 | -> examples
 |    |
 |    | -> basic
 |    |    |
 |    |     -> main.tf
 |    |
 |    | -> complete
 |    |    |
 |    |     -> main.tf
     
```

This directory can be changed on a global level with [EXAMPLES_DIRECTORY](./CONFIG.md#examples_directory)

### Variable defaults

During indexing, cost analysis checks are performed against each example.

To perform this accurately, it is best to ensure examples do not have any required variables - either with no variables present in the example or ensuring all variables have a 'default' value.


### Usage of main.tf

Each of the Terraform files in the example is shown in the UI in alphabetical order, exception for `main.tf`, which is displayed first.

It is recommended to put the 'main' functionality of the example (e.g. the call to the root module and other crucial code to demonstrate) in the main.tf, putting any 'supporting' terraform (state configuration etc.) into seperate files.

### Relative module calls

We recommend using relative paths in the source of the "module blocks" in the examples (that call the local module's root/submodules).

The Terrareg automatically converts this before displaying to users in the web interface, replacing the relative source path with a URL to the module's path within the registry and adds a version constraint.

E.g., for an example with the code:
```
/examples/basic_vpc/main.tf:

module "network" {
  source = "../../"

  vpc = module.vpc.vpc_id
}

module "vpc" {
  source = "../../modules/vpc"

  cidr = "10.0.0.0/24"
}
```

will be rewritten in the UI to:
```
/examples/basic_vpc/main.tf:

module "network" {
  source  = "my-registry.example.com/mynamespace/mymodule/myprovider"
  version = ">= 1.5.0, < 2.0.0"

  vpc = module.vpc.vpc_id
}

module "vpc" {
  source  = "my-registry.example.com/mynamespace/mymodule/myprovider//modules/vpc"
  version = ">= 1.5.0, < 2.0.0"

  cidr = "10.0.0.0/24"
}
```

Note: We also recommend using a single line break before any variables being passing into the module call, as this results in a more consistent styling of the rewritten code.

The version constraint template can be modified by setting [TERRAFORM_EXAMPLE_VERSION_TEMPLATE](./CONFIG.md#terraform_example_version_template).

# Uploading modules

## Indexing

Modules can be indexed in several ways, depending on how modules are being stored.

If modules source is being uploaded directly without a git SCM `git_clone_url` configured, the module source code must be uploaded as a zip file, using the 'upload' endpoint.

If a git-based uploading and the git_clone_url has been configured for a module, the module can be indexed via:
 * Calling the 'import' endpoint.
 * Configuring a Bitbucket/Github webhook to call on push/release.
 * Using the form in the module page, on the 'integrations' tab.

The URL for API endpoints (upload, import, Bitbucket/Github hooks) are dynamic and the URL for a module can be found on the 'integrations' tab of the module.

Note: If the [UPLOAD_API_KEYS](./CONFIG.MD#upload_api_keys) has been configured, this means:
 * Indexing via the web UI is disabled
 * The API key must be configured in any webhooks in Github/Bitbucket
 * The API key must be provided in a 'X-Terrareg-ApiKey' header to requests to the upload/import endpoints.

Once a module has been indexed, the module can be viewed by:
 * Navigating directly to the URL of the module version (`/modules/<namespace>/<module>/<provider>/<version>`)
 * Nagivate to the module by going to the `Registry -> Modules` tab, selecting the namespace, module and provider (though some of these steps maybe be skipped, as the UI will automatically forward you if there is: only one namespace, only one module in a namespace or only one provider in a module)
   * If the UI shows that there are 'No versions of the module have been indexed/published', you can enable the displayed of 'Un-published versions' in the 'Preferences' panel (upper-right).

Once you are happy with the module, continue to [publishing](#publishing) the module.

## Publishing

Until publishing, by default:
 * A module will not appear in search results until a version has been indexed *and* published.
 * The version of the module will not be available to Terraform using the module.
 * The module version is not visible in the module page, unless the user preferences have been configured to show unpublished versions.

To publish a module, the 'publish' API endpoint must be called. The URL for this can be found in the 'integrations' tab of the module.

If [PUBLISH_API_KEYS](./CONFIG.MD#Ppublish_api_keys) has been configured, a valid publish API must be provided as the value of a 'X-Terrareg-ApiKey' header.

## Re-indexing a module version

Once a module version has been uploaded, the behavoir of how Terrareg handles requests to re-upload the module can be configured with [MODULE_VERSION_REINDEX_MODULE](./CONFIG.md#module_version_reindex_mode), which can be used to disable re-indexing of existing module versions or enable/disable re-publishing of previously published versions.


# Security Scanning

Security scanning of modules is performed automatically, without any additional setup.

To disable security scanning results, set the [ENABLE_SECURITY_SCANNING](./CONFIG.md#enable_security_scanning) configuration.

This configuration does not change whether security scans are performing during module indexing, instead, it disables the display of security vulnerabilities in the UI. This means that if the configuration is reverted in future, the security issues are immediately displayed without having to re-index modules.


# Cost Analysis

Example cost analysis is performed using infracost.

A valid API key must be provided to enable this functionality.

Terrareg supports both:
 * Hosted Infracost solution (see [INFRACOST_API_KEY](./CONFIG.md#infracost_api_key) to setup)
 * Locally hosted Infracost API (see [INFRACOST_PRICING_API_ENDPOINT](./CONFIG.md#infracost_pricing_api_endpoint), [INFRACOST_API_KEY](./CONFIG.md#infracost_api_key)).

To disable TLS verification for a locally hosted Infracost pricing API, see [INFRACOST_TLS_INSECURE_SKIP_VERIFY](./CONFIG.md#infracost_tls_insecure_skip_verify)

# Module storage

Terrareg can work in one of two ways:
 * Modules hosting
 * Git-based

If only one of these methods is to be used, the other can be disabled:

 * [ALLOW_MODULE_HOSTING](./CONFIG.md#allow_module_hosting) - enables/disables hosting module source code
 * [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](./CONFIG.md#allow_custom_git_url_module_provider) and [ALLOW_CUSTOM_GIT_URL_MODULE_VERSION](./CONFIG.md#allow_custom_git_url_module_version) - enables disables setting git URLs on modules/within metadata file.


## Module hosting

Modules are indexed by uploading a zip/tar archive of the code to the upload endpoint.

Modules are stored in Terrareg in the [DATA_DIRECTORY](./CONFIG.md#data_directory) - it is important to ensure this path is mounted to persistent storage, so that the data is not lost between container rebuilds/upgrades.

When a module is used in Terraform, Terraform obtains the module source code directly from Terrareg.

Since there is not currently any global authentication to access modules in the registry, this means modules can be downloaded anonymously.

## Git-based

Git can be used as the source for modules.

Each module can be configured with a repository clone url.

When a module version is indexed, Terrareg clones the git repository and indexes the code.
Terrareg will only store meta-data about the module version and store terraform files from each of the 'examples', to be displayed in the UI.

When a module is used in Terraform, Terraform communicates with the registry, which provides a redirect URL to the original git repository. Terraform will then download the source code from the original git repository. If the git repository is authenticated via SSH, Terraform will automatically authenticate using the end-user's SSH key.

This means modules remain secured, protected by the SCM tool.

Modules can stil be uploaded via uploading a zip file, however, if module hosting is disabled, the module must be also configured with a git clone URL for Terraform to download the module, either via Git Provider, custom clone URL in the module or defined in the terrareg metadata file of the module.

The git URLs can be set in several places:
 * Globally using 'Git Providers'
 * In the setting of a module
 * In the terrareg metadata file within a module (though the module must be uploaded via source archive)

These URLs can be provided in the terrareg metadata file within the repository, setting/overriding the URLs displayed for the version of the module.

Terrareg will retain an archive of indexed modules, this can be disabled using [DELETE_EXTERNALLY_HOSTED_ARTIFACTS](./CONFIG.md#delete_externally_hosted_artifacts)

# Git Providers

To fully utilise the features that Terraform provides for git, each module can be provided with three URLs:
 * Repository base URL - the base URL of the repository
 * Repository clone URL - the url for cloning the repository
 * Repository source browse URL - the URL for browsing a single file, using a placeholders for git tag and file path

The avoid having to set this up in each module, Git Providers collapse these configurations into an entity, which can be selected during module creation.

Using git providers, the namespace, module name and provider of the module can be used in the templates of the URLs, meaning that the namespace can be used to determine the organisation/user of the repository (when using Github for example).

Details of the format of this configuration and some examples can be found in the configuration documentation: [GIT_PROVIDER_CONFIG](./CONFIG.md#git_provider_config).

Once enabled, users can be limited from providing custom URLs by disabling: [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](./CONFIG.md#allow_custom_git_url_module_provider) and [ALLOW_CUSTOM_GIT_URL_MODULE_VERSION](./CONFIG.md#allow_custom_git_url_module_version)


# Create modules in the registry

## Terminology

Each "terraform module" in the registry is identified by 3 components:
 * Namespace (see (Namespace)[#namespace] documentation)
 * Module name
   * This is the descriptive name of the module
 * Provider
   * This refers to the 'primary' Terraform provider that is used within the module (e.g. 'aws' would be used for modules that only/primarily manage aws resources). A `null` provider can be used for any modules that do not use providers.

We use the terminology 'Module' for a single module, where `namespace/module/provider1` and `namespace/module/provider2` are different modules.

## Namespaces

Namespaces form part of the module path and are created indepedently of modules.

These can be used in several ways:
 * A logical way to seperate modules
 * Used in conjunction with [Git Providers](#git-providers) to determine part of the SCM path (e.g. the organisation/user for github or projects for stash)
 * To manage user group permissions, which are set on a namespace level (see [Single Sign-on](#single-sign-on))

Namespaces are managed by administrators of the registry using the 'Create Namespace' page (in the 'Create' menu drop-down)

## Create a module

Once logged in, as an admin or a user with 'Full' namespace permissions, to to 'Create -> Module' in the menu drop-down.

The list of namespaces is provided (for non-admin users, only the namespaces that the user has 'Full' permissions to are displayed).

Enter a module name and provider - these must adhere to Terraform's naming restrictions (using only lower-case characters and dashes)

The git provider can be selected - for information on this see [Git Providers](#git-providers).

If a [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](./CONFIG.md#allow_custom_git_url_module_provider) is enabled, the custom URLs (as docuented in the Git Providers section) can be entered specifically for the module.

### Git tag format

The git tag format is used for bi-directional conversion of the Semantic Version used for module versions and the tag for the version in Git.

The git tag format is provided as a templated string of the naming convension that the git repository uses.

The primary placeholder that should be used is `{version}`, which will contain the full semantic version (e.g `1.0.0` or `1.0.0-beta`).

If the tags are prefixed or suffixed with any static string, this can be added to the git tag format, e.g. `release/{version}` or `v{version}` for tags such as `release/1.2.3` or `v1.2.3`, respectively.

#### Non-semantic versioned git tags

If the git tagging format of the repository does not contain the full semantic version, specific placeholders are availale to designate how to extract the individual major, minor and patch parts of the release.
**However**, using this methology is strongly discouraged as it _could_ lead to a many-to-one mapping between module versions and git tags. This is because Terraform (and the associated Terraform module registry APIs) must adhere to Semantic versioning. If a *part* of the semantic version is not provided in the "git tag format", it will be set to 0, i.e. if a git tag format of `{major}.{minor}` is used, a git tag 1.1 would be used for all `1.1.X` semantic versions.

To attempt to avoid this, the module version import API endpoints do not allow indexing by **version** if these placeholders are uesd and, instead, a git tag must be supplied to the [import api](./API.md#apimoduleversionimport). Importing/indexing modules using the Github/Gitlab/Bitbucket hooks are unaffected by this restriction.

Placeholders for non-semantic versions:
 * `{major}` - Specifies the major part of the semantic version
 * `{minor}` - Specifies the minor part of the semantic version
 * `{patch}` - Specifies the patch part of the semantic version

**NOTE:** The preference of these should always be in the above order, i.e.:
 * If you version using `v1`, you should use the placeholder `v{major}`
 * If you version using `v1.1`, you should use the placeholder `v{major}.{minor}`

## Restricting providers

The names of providers that can be used in the registry can be restricted using [ALLOWED_PROVIDERS](./CONFIG.md#allowed_providers).

# Module Usage Analytics

Module analaytics allow each use of a module to be tracked - giving statistics such as:
 * Identifing IDs of the consumers of the modules
 * The latest version of the module that is being used by the consumer
 * The version of Terraform used by the consumer
 * The 'highest' deployment environment that the consumer has deployed the module to


## Identifying deployment environments

When Terraform (that is using a module) is deployed a particular environment, this environment can be captured in the analytics and show in the 'analytics' tab to display how the deployment of the instance of the module has progressed.

Environments are defined in a heirarchy, e.g. dev->production. In this example, if a module is deployed to 'dev', it is show with the environment 'dev' in the Terrareg UI. If it's then deployed to 'production', it is shown with 'production'.

This identification is performed using Terraform authentication.

The keys used for each environment can be defined in [ANALYTICS_AUTH_KEYS](./CONFIG.md#analytics_auth_keys).

The authentication keys must then be configured on the nodes that perform the deployment, setting up `~/.terraformrc` using the following example:
```
credentials "terrareg.my.domain" {
  token = "<environment auth key>"
}
```

If deployment nodes are shared between environments, this configuration will need to be dynamically regenerated based on the environment being deployed to.

## Customising the analytics token

The label used to describe analytics tokens and the example analytics token displayed in the UI can be customised by setting:
 * [ANALYTICS_TOKEN_PHRASE](./CONFIG.md#analytics_token_phrase) - the phrase used to describe the analytics token
 * [EXAMPLE_ANALYTICS_TOKEN](./CONFIG.md#example_analytics_token) - a noun to describe the analytics token or .
 * [ANALYTICS_TOKEN_DESCRIPTION](./CONFIG.md#analytics_token_description) - description of the token

For example, to rebrand the analytics token as someone's first name, you could set:
 * EXAMPLE_ANALYTICS_TOKEN - "john"
 * ANALYTICS_TOKEN_PHRASE - "first name"
 * ANALYTICS_TOKEN_DESCRIPTION - "Set to your first name"

The usage example would then read:
```
Ensure the "john" placeholder must be replaced with your 'first name',
...
module "terraform-aws-rds" {
  source  = "terrareg.my.domain/john__namespace/module/null"

```
The 'usage bulder' table would a variable: "first name" with a description "Set to your first name"


## Enforcing analytics tokens

By default, analytics token must be provided to use a module.

Analytics enforced can be disabled, allowing module usage with an analytics key being passed.

To disable this, see [ALLOW_UNIDENTIFIED_DOWNLOADS](./CONFIG.md#allow_unidentified_downloads)

## Disable analytics token enforcement for some users

Where an example in a module requires another module from the registry to work, it is only useful to use a source URL without an analytics token, to avoid end-user copy+pasting the example and leaving an example analytics token.
However, during the development of modules, to easily test the examples, the analayics token enforcement check can be disabled for the user by using a Terraform auth token (configured in the user's .terraformrc file) configured in the registry.

The configure this, see [IGNORE_ANALYTICS_TOKEN_AUTH_KEYS](./CONFIG.md#ignore_analytics_token_auth_keys).

# Customising Terrareg UI

## Rebranding

The name of the application (in headings and titles) can be customised by setting [APPLICATION_NAME](./CONFIG.md#application_name)

The logo displayed in the UI can be customised by setting [LOGO_URL](./CONFIG.md#logo_url) to a URL of an externally hosted image.

## Module page

### Version constraint

The version constraint shown in the UI can be customised to display a default version that matches how you'd like users to use the module.

For example, it can be used to provide an example to pin the current version or to pin the current minor version or the pin the current major version.

This can be set using [TERRAFORM_EXAMPLE_VERSION_TEMPLATE](./CONFIG.md#terraform_example_version_template), which contains the available placeholders and some examples.

### Labels

Modules can be labelled in two ways:
 * Trusted - this is applied on a per-namespace basis and all modules within the namespace are labeled as 'Trusted'
 * Contributed - this label is applied to any module that is not within a 'Trusted' namespace.
 * Verified - this can be applied on a per-module basis and can be set by anyone with MODIFY privileges of the namespace that contains the module.

These labels are available as filters in the search results.

The list of trusted namespaces can be configured by setting [TRUSTED_NAMESPACES](./CONFIG.md#trusted_namespaces)

The textual representation of these labels can be modified by setting [TRUSTED_NAMESPACE_LABEL](./CONFIG.md#trusted_namespace_label), [VERIFIED_MODULE_LABEL](./CONFIG.md#verified_module_label) and [CONTRIBUTED_NAMESPACE_LABEL](./CONFIG.md#contributed_namespace_label).


# Diagnosing issues

## The domain shown in Terraform snippets contains the wrong URL

This can happen if the following are true:
 * [PUBLIC_URL](./CONFIG.md#public_url) has not been set
 * [DOMAIN_NAME](./CONFIG.md#domain_name) (deprecated) has not been set
 * The domain is rewritten by a reverse proxy before traffic reaches Terrareg

To fix this, set [PUBLIC_URL](./CONFIG.md#public_url) to the URL that users access the Terrareg instance.
