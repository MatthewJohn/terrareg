
# API Docs




## ApiTerraformWellKnown

`/.well-known/terraform.json`

Terraform .well-known discovery


#### GET

Return well-known JSON


## PrometheusMetrics

`/metrics`

Provide usage analytics for Prometheus scraper


#### GET


Return Prometheus metrics for global statistics and module provider statistics



## ApiModuleList

`/v1/modules`

`/v1/modules/`




#### GET

Return list of modules.


## ApiModuleSearch

`/v1/modules/search`

`/v1/modules/search/`




#### GET

Search for modules, given query string, namespace or provider.


## ApiNamespaceModules

`/v1/modules/<string:namespace>`

`/v1/modules/<string:namespace>/`

Interface to obtain list of modules in namespace.


#### GET

Return list of modules in namespace


## ApiModuleDetails

`/v1/modules/<string:namespace>/<string:name>`

`/v1/modules/<string:namespace>/<string:name>/`




#### GET

Return latest version for each module provider.


## ApiModuleProviderDetails

`/v1/modules/<string:namespace>/<string:name>/<string:provider>`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/`




#### GET

Return list of version.


## ApiModuleVersions

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions/`




#### GET

Return list of version.


## ApiModuleVersionDetails

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/`




#### GET

Return list of version.


## ApiModuleVersionDownload

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/download`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/download`

Provide download endpoint.


#### GET

Provide download header for location to download source.


## ApiModuleProviderDownloadsSummary

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/downloads/summary`

Provide download summary for module provider.


#### GET

Return list of download counts for module provider.


## ApiProviderList

`/v1/providers`

`/v1/providers/`

Interface to list all providers


#### GET

Return list of modules.


## ApiNamespaceProviders

`/v1/providers/<string:namespace>`

`/v1/providers/<string:namespace>/`

`/v1/terrareg/providers/<string:namespace>`

Interface to obtain list of providers in namespace.


#### GET

Return list of providers in namespace


## ApiProvider

`/v1/providers/<string:namespace>/<string:provider>`

`/v1/providers/<string:namespace>/<string:provider>/<string:version>`




#### GET

Return provider details.


## ApiProviderVersions

`/v1/providers/<string:namespace>/<string:provider>/versions`




#### GET

Return provider version details.


## ApiProviderVersionDownload

`/v1/providers/<string:namespace>/<string:provider>/<string:version>/download/<string:os>/<string:arch>`




#### GET

Return provider details.


## ApiProviderSearch

`/v1/providers/search`




#### GET

Search for modules, given query string, namespace or provider.


## ApiV2Provider

`/v2/providers/<string:namespace>/<string:provider>`

Interface for providing provider details


#### GET

Return provider details.


## ApiProviderProviderDownloadSummary

`/v2/providers/<int:provider_id>/downloads/summary`

Interface for providing download summary for providers


#### GET

Return download summary.


## ApiV2ProviderDocs

`/v2/provider-docs`

Interface for querying provider docs


#### GET


Query provider version documentation.

This API is very static and requires all arguments to be passed.
Page size, is effectively unused, as the query filters will result in 0 or 1 result.

##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| filter[provider-version] | args | int | True | `None` | Provider version ID to query documenation from |
| filter[category] | args | str | True | `None` | Provider documentation category |
| filter[slug] | args | str | True | `None` | Slug of documentation to query for |
| filter[language] | args | str | True | `None` | Documentation language to filter results |
| page[size] | args | int | True | `None` | Result page size |



## ApiV2ProviderDoc

`/v2/provider-docs/<int:doc_id>`

Interface for obtain provider doc details


#### GET


Obtain details about provider document

##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| output | args | str | False | `md` | Content output type, either "html" or "md" |



## ApiGpgKeys

`/v2/gpg-keys`

Provide interface to create GPG Keys.


#### GET

Lists GPG keys for given namespaces
##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| filter[namespace] | args | str | True | `None` | Comma-separated list of namespaces to obtain GPG keys for |

#### POST


Handle creation of GPG key.

POST Body must be JSON, in the format:
```
{
    "data": {
        "type": "gpg-keys",
        "attributes": {
            "namespace": "my-namespace",
            "ascii-armor": "-----BEGIN PGP PUBLIC KEY BLOCK-----
...
-----END PGP PUBLIC KEY BLOCK-----
"
        }
    },
    "csrf_token": "xxxaaabbccc"
}
```



## ApiGpgKey

`/v2/gpg-keys/<string:namespace>/<string:key_id>`

Provide interface to create GPG Keys.


#### GET

Get details for a given GPG key for a namespace
#### DELETE


Perform deletion of GPG key

##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| csrf_token | json | str | False | `None` | CSRF token |



## ApiProviderCategories

`/v2/categories`

Interface to obtain list of provider categories.


#### GET

Return list of all provider categories


## ApiTerraregGraphData

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data`

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data/submodule/<path:submodule_path>`

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data/example/<path:example_path>`

Interface to obtain module version graph data.


#### GET

Return graph data for module version.


## ApiOpenIdInitiate

`/openid/login`

Interface to initiate authentication via OpenID connect


#### GET

Generate session for storing OpenID state token and redirect to OpenID login provider.


## ApiOpenIdCallback

`/openid/callback`

Interface to handle callback from authorization flow from OpenID connect


#### GET

Handle response from OpenID callback


## ApiSamlInitiate

`/saml/login`

Interface to initiate authentication via OpenID connect


#### GET

Setup authentication request to redirect user to SAML provider.
#### POST

Handle POST request.


## ApiSamlMetadata

`/saml/metadata`

Meta-data endpoint for SAML


#### GET

Return SAML SP metadata.


## GithubLoginInitiate

`/<string:provider_source>/login`

Interface to initiate authentication via Github


#### GET

Redirect to github login.


## GithubLoginCallback

`/<string:provider_source>/callback`

Interface to handle call-back from Github login


#### GET

Handle callback from Github auth.


## GithubAuthStatus

`/<string:provider_source>/auth/status`

Interface to provide details about current authentication status with Github


#### GET

Provide authentication status.


## GithubOrganisations

`/<string:provider_source>/organizations`

Interface to provide details about current Github organisations for the logged in user


#### GET

Provide organisation details.


## GithubRepositories

`/<string:provider_source>/repositories`

Interface to provide details about current Github repositories for the logged in user


#### GET

Provide organisation details.


## GithubRefreshNamespace

`/<string:provider_source>/refresh-namespace`

Interface to refresh repositories for a namespaces from a provider source


#### POST

Refresh repositories for given namespace.
##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| namespace | json | str | True | `None` | Namespace to refresh repositories for |
| csrf_token | json | str | True | `None` | CSRF token |



## GithubRepositoryPublishProvider

`/<string:provider_source>/repositories/<int:repository_id>/publish-provider`

Interface publish a repository as a provider


#### POST

Publish repository provider.
##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| category_id | form | int | True | `None` | Provider category ID for provider |
| csrf_token | form | str | False | `None` | CSRF Token |



## ApiTerraregConfig

`/v1/terrareg/config`

Endpoint to return config used by UI.


#### GET

Return config.


## ApiTerraregGitProviders

`/v1/terrareg/git_providers`

Interface to obtain git provider configurations.


#### GET

Return list of git providers


## ApiTerraregGlobalStatsSummary

`/v1/terrareg/analytics/global/stats_summary`

Provide global download stats for homepage.


#### GET

Return number of namespaces, modules, module versions and downloads


## ApiTerraregMostRecentlyPublishedModuleVersion

`/v1/terrareg/analytics/global/most_recently_published_module_version`

Return data for most recently published module version.


#### GET

Return number of namespaces, modules, module versions and downloads


## ApiTerraregGlobalUsageStats

`/v1/terrareg/analytics/global/usage_stats`

Provide interface to obtain statistics about global module usage.


#### GET


Return stats on total module providers,
total unique analytics tokens per module
(with and without auth token).



## ApiTerraregModuleProviderAnalyticsTokenVersions

`/v1/terrareg/analytics/<string:namespace>/<string:name>/<string:provider>/token_versions`

Provide download summary for module provider.


#### GET

Return list of download counts for module provider.


## ApiTerraregMostDownloadedModuleProviderThisWeek

`/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week`

Return data for most downloaded module provider this week.


#### GET

Return most downloaded module this week


## ApiTerraregInitialSetupData

`/v1/terrareg/initial_setup`

Interface to provide data to the initial setup page.


#### GET

Return information for steps for setting up Terrareg.


## ApiTerraregNamespaces

`/v1/terrareg/namespaces`

Provide interface to obtain namespaces.


#### GET


Return list of namespaces.

The offset/limit arguments are currently optional.
Without them, all namespaces will be returned in a list (legacy response format).
Providing these values will return an object with a meta object and a list of namespaces.

##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| only_published | args | boolean | False | `False` | Whether to only show namespaces with published modules or providers |
| type | args | str | False | `module` | Type of namespace to show results for. Either "provider" or "module" |
| offset | args | int | False | `0` | Pagination offset |
| limit | args | int | False | `None` | Pagination limit |

#### POST

Create namespace.


## ApiTerraregNamespaceDetails

`/v1/terrareg/namespaces/<string:namespace>`

Interface to obtain custom terrareg namespace details.


#### GET

Return custom terrareg config for namespace.
#### POST

Edit name/display name of a namespace
#### DELETE


Delete namespace

JSON body:
 * csrf_token - CSRF token required for session-based authentication



## ApiTerraregNamespaceModules

`/v1/terrareg/modules/<string:namespace>`

Interface to obtain list of modules in namespace.


#### GET

Return list of modules in namespace


## ApiTerraregModuleProviders

`/v1/terrareg/modules/<string:namespace>/<string:name>`

Interface to obtain list of providers for module.


#### GET

Return list of modules in namespace


## ApiTerraregModuleProviderDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>`

Interface to obtain module provider details.


#### GET

Return details about module version.


## ApiTerraregModuleVersionDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>`

Interface to obtain module version details.


#### GET

Return details about module version.


## ApiTerraregModuleProviderVersions

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/versions`

Interface to obtain module provider versions


#### GET

Return list of module versions for module provider


## ApiTerraregModuleProviderCreate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/create`

Provide interface to create module provider.


#### POST

Handle update to settings.


## ApiTerraregModuleProviderDelete

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/delete`

Provide interface to delete module provider.


#### DELETE

Delete module provider.


## ApiTerraregModuleProviderSettings

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/settings`

Provide interface to update module provider settings.


#### POST

Handle update to settings.


## ApiTerraregModuleProviderIntegrations

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/integrations`

Interface to provide list of integration URLs


#### GET

Return list of integration URLs


## ApiTerraregModuleProviderRedirects

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/redirects`

Provide interface to delete module provider redirect.


#### GET

Delete module provider.


## ApiTerraregModuleProviderRedirectDelete

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/redirects/<string:module_provider_redirect_id>`

Provide interface to delete module provider redirect.


#### DELETE

Delete module provider.


## ApiModuleVersionCreateBitBucketHook

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/bitbucket`

Provide interface for Bitbucket hook to detect pushes of new tags.


#### POST

Create new version based on Bitbucket hooks.


## ApiModuleVersionCreateGitHubHook

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/github`

Provide interface for GitHub hook to detect new and changed releases.


#### POST

Create, update or delete new version based on GitHub release hooks.


## ApiTerraregModuleVersionVariableTemplate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/variable_template`

Provide variable template for module version.


#### GET

Return variable template.


## ApiTerraregModuleVersionFile

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/files/<string:path>`

Interface to obtain content of module version file.


#### GET

Return content of module version file.


## ApiTerraregModuleVersionReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/readme_html`

Provide variable template for module version.


#### GET

Return variable template.


## ApiModuleVersionUpload

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload`




#### POST

Handle module version upload.


## ApiModuleVersionCreate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/import`


Provide interface to create release for git-backed modules.

**DEPRECATION NOTICE**

This API maybe removed in future.
This deprecation is still in discussion.

Consider migrating to '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/import'



#### POST

Handle creation of module version.


## ApiModuleVersionImport

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/import`


Provide interface to import/index version for git-backed modules.



#### POST

Handle creation of module version.
##### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| version | json | str | False | `None` | The semantic version number of the module to be imported. This can only be used if the git tag format of the module provider contains a {version} placeholder. Conflicts with git_tag |
| git_tag | json | str | False | `None` | The git tag of the module to be imported. Conflicts with version. |



## ApiModuleVersionSourceDownload

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/source.zip`

Return source package of module version


#### GET

Return static file.


## ApiTerraregModuleVersionPublish

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/publish`

Provide interface to publish module version.


#### POST

Publish module.


## ApiTerraregModuleVersionDelete

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/delete`

Provide interface to delete module version.


#### DELETE

Delete module version.


## ApiTerraregModuleVerisonSubmodules

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules`

Interface to obtain list of submodules in module version.


#### GET

Return list of submodules.


## ApiTerraregSubmoduleDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/details/<path:submodule>`

Interface to obtain submodule details.


#### GET

Return details of submodule.


## ApiTerraregSubmoduleReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/readme_html/<path:submodule>`

Interface to obtain submodule REAMDE in HTML format.


#### GET

Return HTML formatted README of submodule.


## ApiTerraregModuleVersionExamples

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples`

Interface to obtain list of examples in module version.


#### GET

Return list of examples.


## ApiTerraregExampleDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/details/<path:example>`

Interface to obtain example details.


#### GET

Return details of example.


## ApiTerraregExampleReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/readme_html/<path:example>`

Interface to obtain example REAMDE in HTML format.


#### GET

Return HTML formatted README of example.


## ApiTerraregExampleFileList

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/filelist/<path:example>`

Interface to obtain list of example files.


#### GET

Return list of files available in example.


## ApiTerraregExampleFile

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/file/<path:example_file>`

Interface to obtain content of example file.


#### GET

Return content of example file in example module.


## ApiTerraregProviderLogos

`/v1/terrareg/provider_logos`

Provide interface to obtain all provider logo details


#### GET

Return all details about provider logos.


## ApiTerraregModuleSearchFilters

`/v1/terrareg/search_filters`

`/v1/terrareg/modules/search/filters`


Return list of filters available for search.

*Deprecation*: The `/v1/terrareg/search_filters` endpoint has been deprecated in favor of `/v1/terrareg/modules/search/filters`

The previous endpoint will be removed in a future major release.



#### GET

Return list of available filters and filter counts for search query.


## ApiTerraregProviderSearchFilters

`/v1/terrareg/providers/search/filters`


Return list of filters available for provider search.



#### GET

Return list of available filters and filter counts for search query.


## ApiTerraregAuditHistory

`/v1/terrareg/audit-history`

Interface to obtain audit history


#### GET

Obtain audit history events


## ApiTerraregAuthUserGroups

`/v1/terrareg/user-groups`

Interface to list and create user groups.


#### GET

Obtain list of user groups.
#### POST

Create user group


## ApiTerraregAuthUserGroup

`/v1/terrareg/user-groups/<string:user_group>`

Interface to interact with single user group.


#### DELETE

Delete user group.


## ApiTerraregAuthUserGroupNamespacePermissions

`/v1/terrareg/user-groups/<string:user_group>/permissions/<string:namespace>`

Interface to create user groups namespace permissions.


#### POST

Create user group namespace permission
#### DELETE

Delete user group namespace permission


## ApiTerraregAdminAuthenticate

`/v1/terrareg/auth/admin/login`

Interface to perform authentication as an admin and set appropriate cookie.


#### POST

Handle POST requests to the authentication endpoint.


## ApiTerraregIsAuthenticated

`/v1/terrareg/auth/admin/is_authenticated`

Interface to return whether user is authenticated as an admin.


#### GET

Return information about current user.


## ApiTerraregHealth

`/v1/terrareg/health`

Endpoint to return 200 when healthy.


#### GET

Return static 200