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


### Provider sources

Provider sources are a method of integrating with a Git provider for authentication and creating providers.

Provider sources allow users to authenticate to Terrareg.
Information about their associated projects/organisations and repositories are used to allow providers to be created.

Currently, only Github is supported as a provider source.

NOTE: These may eventually be used for modules and eventually replace GIT_PROVIDERS.


#### Confguring provider sources

Provider sources are configured using [PROVIDER_SOURCES](./CONFIG.md#provider_sources) configuration.

Each provider source must be configured with:
 * `name` - The name of the provider source, shown the users when selecting the source for creating a provider.
 * `type` - Must be one of the supported platforms. (See the configuration for a list of supported values)
 * `login_button_text` - The text displayed in the login button on the login page for the provider
 * `auto_generate_namespaces` - Determines if the namespaces are created for each of the organisations that the user is an owner of, when authenticating.
   * If this is disabled, namespaces must be created by a site admin.

##### Github

The following configurations must be configured (see CONFIG.md for details about how to generate them):
 - `base_url` - Base public URL, e.g. `https://github.com`
 - `api_url` - API URL, e.g. `https://api.github.com`
 - `app_id` - Github app ID
 - `client_id` - Github app client ID for Github authentication.
 - `client_secret` - Github App client secret for Github authentication.
 - `private_key_path` - Path to private key generated for Github app.
 - `webhook_secret` (optional) - Web hook secret that Github will provide to Terrareg when calling the webhook endpoint. This cannot be used for non-publicly accessible Terrareg installations.
 - `default_access_key` (optional) - Github access token, used for perform Github API requests for providers that are not authenticated via Github App.
 - `default_installation_id` (optional) - A default installation that provides access to Github APIs for providers that are not authenticated via Github App.

For self-hosted Github Enterprise installations, `base_url` and `api_url` can be set to match the github installation.

Github users are automatically mapped to any groups in Terrareg that are named after organisations that they are named after.
E.g. User `testuser`, who is a member of the `testorg` organisation would be mapped to the groups `testuser` and `testorg` Terrareg group, if they exist.


###### Setup Github app

To setup the application for Github, goto `https://github.com/settings/apps/new` (or `https://github.com/organizations/<organisation>/settings/apps`)

Configure the following:

In this example, the "provider source" "name" will be "github-terra-example", be sure to replace this with the chosen provider source name.

 * Homepage URL: https://my-terrareg-installation.example.com
 * Identifying and authorizing users
   * Callback URL: https://my-terrareg-installation.example.com/github-terra-example/callback
   * Expire user authorization tokens: Check
   * Request user authorization (OAuth) during installation: Check
   * Enable Device Flow: Uncheck
 * Post installation
   * Setup URL: `Not set`
   * Redirect on update: Uncheck
 * Webhook (only select if your Terrareg installation is accessible from the internet, meaning Github can call it's APIs)
   * Active: Check
   * Webhook URL: https://my-terrareg-installation.example.com/github-terra-example/webhook
   * Webhook secret: `Value from webhook_secret`
 * Permissions:
   * None
 * Where can this GitHub App be installed?
   * Select based on whether you with to limit who can authenticate to Terrareg (publicly or members of the organisation)

From the created app landing page, obtain the App ID and Client ID from the "about" section of the page to populate the `app_id` and `client_id` configurations.

Generate a client secret:

 * Click "Generate a new client secret"
 * Copy the value from the client secret to populate `client_secret`


Once created, generate a private key (this is not currently used, but will be in future):

 * On the same page, find `Private keys` and click "Generate a private key"
 * Download the file and make the file accessible to the Terrareg installation, set in `private_key_path`

###### Github namespace mapping

If `auto_generate_namespaces` is enabled, namespaces are automatically created for the logged in Github user and all organisations that they are an owner of.

The user has full permissions to each of these namespaces, allowing them to create modules/providers in them.

###### Default Github authentication

Terrareg can be used in two use cases for using Github as a provider source:
 1. Owners of Terraform providers authenticate to Terrareg.
   * These users add their own providers to Terrareg and manage them.
 2. Terrareg administrators add providers from Github that they do not maintain.

Both of these can be used simulatiously and Terrareg will authenticate using the Github App authentication, when a user authenticates via Github and uses the default authentication for the second use case.

For the second case, authenticating to Terrareg via Github will not provide the necessary permissions to create providers and index versions.
To achive this, Terrareg must be provided with a means to authenticate to Github to query the APIs.

This can be performed by setting one of the two provider source configurations:
 * `default_installation_id` - The created Github app for Terrareg can be 'installed' to a user/organisations profile. Once this has been performed, the "installation ID" can be provided to Terrareg.
 * `default_access_token` - A personal access token can be provided to Terrareg, which will be used.


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
