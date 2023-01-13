# User Guide

## Contents

* [Deployment](#deployment)
  * [Docker Environment Variables](#docker-environment-variables)
  * [Database Migrations](#database-migrations)
* [Security](#security)
   * [Single Single-On](#single-sign-on)
* [Uploading Modules](#uploading-modules)
* [Security Scanning](#security-scanning)
* [Cost Analysis](#cost-analysis)
* [Module Usage Analytics](#module-usage-analytics)


## Deployment

### Docker environment variables

The following environment variables are available to configure the docker container

#### MIGRATE_DATABASE

Whether to perform a database migration on container startup.

Set to `True` to enable database migration

*Note:* Be very careful when scaling the application. There should never be more than one instance of Terrareg running with `MIGRATE_DATABASE` set to `True` during an upgrade.

When upgrading, scale the application to a single instance before upgrading to a newer version.

Alternatively, set `MIGRATE_DATABASE` to `False` and run a dedicated instance for performing database upgrades.
Use `MIGRATE_DATABASE_ONLY` to run an instance that will perform the necessary database migrations and immediately exit.

Default: `False`

#### MIGRATE_DATABASE_ONLY

Whether to perform database migration and exit immediately.

This is useful for scheduling migrations by starting a 'migration' instance of the application.

Set to `True` to exit after migration.

The `MIGRATE_DATABASE` environment variable must also be set to `True` to perform the migration, otherwise nothing will be performed and the container will exit.

Default: `False`

#### SSH_PRIVATE_KEY

Provide the contents of the SSH key to perform git clones.

This is an alternative to mounting the '.ssh' directory of the root user.

Default: ''


### Application environment variables

For a list of available application configuration environment variables, please see [docs/CONFIG.md](./CONFIG.md)


### Database Migrations

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


## Security

By default, Terrareg administration is protected by an [ADMIN_AUTHENTICATION_TOKEN](./CONFIG.md#admin_authentication_token), which is set via environment variable, which becomes the password used for authentication on the login page.

Authentication is required to perform tasks such as: creating namespaces, creating/deleting modules and managing additional user permissions.

However, indexing and publishing modules does *not* require authentication. To disable unauthorised indexing/publishing of modules, set up dedicated API keys for these functions, see [UPLOAD_API_KEYS](./CONFIG.MD#upload_api_keys) and [PUBLISH_API_KEYS](./CONFIG.MD#Ppublish_api_keys)


### Single Sign-on

Single sign-on can be used to allow authentication via external authentication providers.

SSO groups can be assigned global admin permissions, or per-namespace permissions, allowing creation/deletion of modules and versions.

User groups and permissions can be configured on the 'User Groups' (/user-groups) page.

Once single sign-on has been setup, the [ADMIN_AUTHENTICATION_TOKEN](./CONFIG.md#adminauthenticationtoken) can be disabled, stopping further sign-in via password authentication.

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

 * [DOMAIN_NAME](./CONFIG.md#domainname)
 * [OPENID_CONNECT_CLIENT_ID](./CONFIG.md#openidconnectclientid)
 * [OPENID_CONNECT_CLIENT_SECRET](./CONFIG.md#openidconnectclientsecret)
 * [OPENID_CONNECT_ISSUER](./CONFIG.md#openidconnectissuer)

Note: Most IdP providers will require the terrareg installation to be accessed via https.
The instance should be configured with SSL certificates ([SSL_CERT_PRIVATE_KEY](./CONFIG.md#sslcertprivatekey)/[SSL_CERT_PUBLIC_KEY](./CONFIG.md#sslcertpublickey)) or be hosted behind a reverse-proxy/load balancer.

The text displayed on the login button can be customised by setting [OPENID_CONNECT_LOGIN_TEXT](./CONFIG.md#openidconnectlogintext)

### SAML2

Generate a public and a private key, using:

    openssl genrsa -out private.key 4096
    openssl req -new -x509 -key private.key -out publickey.cer -days 365

Set the folllowing environment variables:

* [SAML2_IDP_METADATA_URL](./CONFIG.md#saml2idpmetadataurl) (required)
* [SAML2_ENTITY_ID](./CONFIG.md#saml2entityid) (required)
* [SAML2_PRIVATE_KEY](./CONFIG.md#saml2privatekey) (required) (See above)
* [SAML2_PUBLIC_KEY](./CONFIG.md#saml2publickey) (required) (See above)
* [SAML2_ENTITY_ID](./CONFIG.md#saml2entityid) (optional)

In the IdP:

* configure the Single signin URL to `https://{terrareg_installation_domain}/saml/login?sso`;
* configure the request and response to be signed;
* ensure at least one attribute is assigned.


## Uploading Modules

Terrareg has been extensively tested with Terraform modules of all shapes and sizes, meaning that it should be able to provide valuable information without any modification to modules before indexing.

However, to get the _most_ out of Terrareg, there are some practices/guides that will help.

### Terrareg metadata file

A metadata file can be provided each an uploaded module's archive to provide additional metadata to terrareg.

This should be called `terrareg.json` and be placed in the root of the module.

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

### Description

If a metadata file is not present or a description is not provided, Terrareg will attempt to automatically generate a description of the module, using the README.md from the module.

This functionality can be disabled by setting [AUTOGENERATE_MODULE_PROVIDER_DESCRIPTION](CONFIG.md#autogeneratemoduleproviderdescription)

#### Usage builder configuration

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


### Submodules


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

This directory can be changed on a global level with [EXAMPLES_DIRECTORY](./CONFIG.md#examplesdirectory)

### Examples

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

This directory can be changed on a global level with [EXAMPLES_DIRECTORY](./CONFIG.md#examplesdirectory)

#### Variable defaults

During indexing, cost analysis checks are performed against each example.

To perform this accurately, it is best to ensure examples do not have any required variables - either with no variables present in the example or ensuring all variables have a 'default' value.


#### Usage of main.tf

Each of the Terraform files in the example is shown in the UI in alphabetical order, exception for `main.tf`, which is displayed first.

It is recommended to put the 'main' functionality of the example (e.g. the call to the root module and other crucial code to demonstrate) in the main.tf, putting any 'supporting' terraform (state configuration etc.) into seperate files.

#### Relative module calls

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


## Security Scanning

Security scanning of modules is performed automatically, without any additional setup.

To disable security scanning results, set the [ENABLE_SECURITY_SCANNING](./CONFIG.md#enable_security_scanning) configuration.

This configuration does not change whether security scans are performing during module indexing, instead, it disables the display of security vulnerabilities in the UI. This means that if the configuration is reverted in future, the security issues are immediately displayed without having to re-index modules.


## Cost Analysis

Example cost analysis is performed using infracost.

A valid API key must be provided to enable this functionality.

Terrareg supports both:
 * Hosted Infracost solution (see [INFRACOST_API_KEY](./CONFIG.md#infracostapikey) to setup)
 * Locally hosted Infracost API (see [INFRACOST_PRICING_API_ENDPOINT](./CONFIG.md#infracostpricingapiendpoint), [INFRACOST_API_KEY](./CONFIG.md#infracostapikey)).

To disable TLS verification for a locally hosted Infracost pricing API, see [INFRACOST_TLS_INSECURE_SKIP_VERIFY](./CONFIG.md#infracosttlsinsecureskipverify)

## Create modules in the registry

### Terminology

Each "terraform module" in the registry is identified by 3 components:
 * Namespace (see (Namespace)[#namespace] documentation)
 * Module name
   * This is the descriptive name of the module
 * Provider
   * This refers to the 'primary' Terraform provider that is used within the module (e.g. 'aws' would be used for modules that only/primarily manage aws resources). A `null` provider can be used for any modules that do not use providers.

We use the terminology 'Module' for a single module, where `namespace/module/provider1` and `namespace/module/provider2` are different modules.

### Namespaces

Namespaces form part of the module path and are created indepedently of modules.

These can be used in several ways:
 * Just a logical way to seperate modules
 * Used in conjunction with [Git Providers](#git-providers) to determine part of the SCM path (e.g. the organisation/user for github or projects for stash)
 * To manage user group permissions, which are set on a namespace level (see [Single Sign-on](#single-sign-on))

Namespaces are managed by administrators of the registry using the 'Create Namespace' page (in the 'Create' menu drop-down)

### Create a module

Once logged in, as an admin or a user with 'Full' namespace permissions, to to 'Create -> Module' in the menu drop-down.

The list of namespaces is provided (for non-admin users, only the namespaces that the user has 'Full' permissions to are displayed).

Enter a module name and provider - these must adhere to Terraform's naming restrictions (using only lower-case characters and dashes)

The git provider can be selected - for information on this see [Git Providers](#git-providers).

If a [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](./CONFIG.md#allow_custom_git_url_module_provider) is enabled, the custom URLs (as docuented in the Git Providers section) can be entered specifically for the module.


## Module Usage Analytics

Module analaytics allow each use of a module to be tracked - giving statistics such as:
 * Identifing IDs of the consumers of the modules
 * The latest version of the module that is being used by the consumer
 * The version of Terraform used by the consumer
 * The 'highest' deployment environment that the consumer has deployed the module to


### Identifying deployment environments

### Customising the analytics token

### Enforcing analytics tokens

By default, modules can be downloaded without an analytics token.

Analytics can be enforced by denying module downloads without an analytics token being passed.

To enable this, see [ALLOW_UNIDENTIFIED_DOWNLOADS](./CONFIG.md#allowunidentifieddownloads)

