
# API Docs




## ApiTerraformWellKnown

`/.well-known/terraform.json`

Terraform .well-known discovery


### GET

Return wellknown JSON


## PrometheusMetrics

`/metrics`

Provide usage anayltics for Prometheus scraper


### GET


Return Prometheus metrics for global statistics and module provider statistics



## ApiModuleList

`/v1/modules`

`/v1/modules/`




### GET

Return list of modules.


## ApiModuleSearch

`/v1/modules/search`

`/v1/modules/search/`




### GET

Search for modules, given query string, namespace or provider.


## ApiNamespaceModules

`/v1/modules/<string:namespace>`

`/v1/modules/<string:namespace>/`

Interface to obtain list of modules in namespace.


### GET

Return list of modules in namespace


## ApiModuleDetails

`/v1/modules/<string:namespace>/<string:name>`

`/v1/modules/<string:namespace>/<string:name>/`




### GET

Return latest version for each module provider.


## ApiModuleProviderDetails

`/v1/modules/<string:namespace>/<string:name>/<string:provider>`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/`




### GET

Return list of version.


## ApiModuleVersions

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/versions/`




### GET

Return list of version.


## ApiModuleVersionDetails

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/`




### GET

Return list of version.


## ApiModuleVersionDownload

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/download`

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/download`

Provide download endpoint.


### GET

Provide download header for location to download source.


## ApiModuleProviderDownloadsSummary

`/v1/modules/<string:namespace>/<string:name>/<string:provider>/downloads/summary`

Provide download summary for module provider.


### GET

Return list of download counts for module provider.


## ApiTerraregGraphData

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data`

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data/submodule/<path:submodule_path>`

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/graph/data/example/<path:example_path>`

Interface to obtain module verison graph data.


### GET

Return graph data for module version.


## ApiOpenIdInitiate

`/openid/login`

Interface to initiate authentication via OpenID connect


### GET

Generate session for storing OpenID state token and redirect to openid login provider.


## ApiOpenIdCallback

`/openid/callback`

Interface to handle callback from authorization flow from OpenID connect


### GET

Handle response from OpenID callback


## ApiSamlInitiate

`/saml/login`

Interface to initiate authentication via OpenID connect


### GET

Setup authentication request to redirect user to SAML provider.
### POST

Handle POST request.


## ApiSamlMetadata

`/saml/metadata`

Meta-data endpoint for SAML


### GET

Return SAML SP metadata.


## ApiTerraregConfig

`/v1/terrareg/config`

Endpoint to return config used by UI.


### GET

Return config.


## ApiTerraregGitProviders

`/v1/terrareg/git_providers`

Interface to obtain git provider configurations.


### GET

Return list of git providers


## ApiTerraregGlobalStatsSummary

`/v1/terrareg/analytics/global/stats_summary`

Provide global download stats for homepage.


### GET

Return number of namespaces, modules, module versions and downloads


## ApiTerraregMostRecentlyPublishedModuleVersion

`/v1/terrareg/analytics/global/most_recently_published_module_version`

Return data for most recently published module version.


### GET

Return number of namespaces, modules, module versions and downloads


## ApiTerraregGlobalUsageStats

`/v1/terrareg/analytics/global/usage_stats`

Provide interface to obtain statistics about global module usage.


### GET


Return stats on total module providers,
total unique analytics tokens per module
(with and without auth token).



## ApiTerraregModuleProviderAnalyticsTokenVersions

`/v1/terrareg/analytics/<string:namespace>/<string:name>/<string:provider>/token_versions`

Provide download summary for module provider.


### GET

Return list of download counts for module provider.


## ApiTerraregMostDownloadedModuleProviderThisWeek

`/v1/terrareg/analytics/global/most_downloaded_module_provider_this_week`

Return data for most downloaded module provider this week.


### GET

Return most downloaded module this week


## ApiTerraregInitialSetupData

`/v1/terrareg/initial_setup`

Interface to provide data to the initial setup page.


### GET

Return information for steps for setting up Terrareg.


## ApiTerraregNamespaces

`/v1/terrareg/namespaces`

Provide interface to obtain namespaces.


### GET

Return list of namespaces.
### POST

Create namespace.


## ApiTerraregNamespaceDetails

`/v1/terrareg/namespaces/<string:namespace>`

Interface to obtain custom terrareg namespace details.


### GET

Return custom terrareg config for namespace.
### POST

Edit name/display name of a namespace


## ApiTerraregNamespaceModules

`/v1/terrareg/modules/<string:namespace>`

Interface to obtain list of modules in namespace.


### GET

Return list of modules in namespace


## ApiTerraregModuleProviderDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>`

Interface to obtain module provider details.


### GET

Return details about module version.


## ApiTerraregModuleVersionDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>`

Interface to obtain module verison details.


### GET

Return details about module version.


## ApiTerraregModuleProviderVersions

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/versions`

Interface to obtain module provider versions


### GET

Return list of module versions for module provider


## ApiTerraregModuleProviderCreate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/create`

Provide interface to create module provider.


### POST

Handle update to settings.


## ApiTerraregModuleProviderDelete

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/delete`

Provide interface to delete module provider.


### DELETE

Delete module provider.


## ApiTerraregModuleProviderSettings

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/settings`

Provide interface to update module provider settings.


### POST

Handle update to settings.


## ApiTerraregModuleProviderIntegrations

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/integrations`

Interface to provide list of integration URLs


### GET

Return list of integration URLs


## ApiModuleVersionCreateBitBucketHook

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/bitbucket`

Provide interface for bitbucket hook to detect pushes of new tags.


### POST

Create new version based on bitbucket hooks.


## ApiModuleVersionCreateGitHubHook

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/hooks/github`

Provide interface for GitHub hook to detect new and changed releases.


### POST

Create, update or delete new version based on GitHub release hooks.


## ApiTerraregModuleVersionVariableTemplate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/variable_template`

Provide variable template for module version.


### GET

Return variable template.


## ApiTerraregModuleVersionFile

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/files/<string:path>`

Interface to obtain content of module version file.


### GET

Return conent of module version file.


## ApiTerraregModuleVersionReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/readme_html`

Provide variable template for module version.


### GET

Return variable template.


## ApiModuleVersionUpload

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/upload`




### POST

Handle module version upload.


## ApiModuleVersionCreate

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/import`


Provide interface to create release for git-backed modules.

**DEPRECATION NOTICE**

This API maybe removed in future.
This deprecation is still in discussion.

Consider migrating to '/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/import'



### POST

Handle creation of module version.


## ApiModuleVersionImport

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/import`


Provide interface to import/index version for git-backed modules.



### POST

Handle creation of module version.
#### Arguments

| Argument | Location (JSON POST body or query string argument) | Type | Required | Default | Help |
|----------|----------------------------------------------------|------|----------|---------|------|
| version | json | str | False | `None` | The semantic version number of the module to be imported. This can only be used if the git tag format of the module provider contains a {version} placeholder. Conflicts with git_tag |
| git_tag | json | str | False | `None` | The git tag of the module to be imported. Conflicts with version. |



## ApiModuleVersionSourceDownload

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/source.zip`

Return source package of module version


### GET

Return static file.


## ApiTerraregModuleVersionPublish

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/publish`

Provide interface to publish module version.


### POST

Publish module.


## ApiTerraregModuleVersionDelete

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/delete`

Provide interface to delete module version.


### DELETE

Delete module version.


## ApiTerraregModuleVerisonSubmodules

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules`

Interface to obtain list of submodules in module version.


### GET

Return list of submodules.


## ApiTerraregSubmoduleDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/details/<path:submodule>`

Interface to obtain submodule details.


### GET

Return details of submodule.


## ApiTerraregSubmoduleReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/submodules/readme_html/<path:submodule>`

Interface to obtain submodule REAMDE in HTML format.


### GET

Return HTML formatted README of submodule.


## ApiTerraregModuleVersionExamples

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples`

Interface to obtain list of examples in module version.


### GET

Return list of examples.


## ApiTerraregExampleDetails

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/details/<path:example>`

Interface to obtain example details.


### GET

Return details of example.


## ApiTerraregExampleReadmeHtml

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/readme_html/<path:example>`

Interface to obtain example REAMDE in HTML format.


### GET

Return HTML formatted README of example.


## ApiTerraregExampleFileList

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/filelist/<path:example>`

Interface to obtain list of example files.


### GET

Return list of files available in example.


## ApiTerraregExampleFile

`/v1/terrareg/modules/<string:namespace>/<string:name>/<string:provider>/<string:version>/examples/file/<path:example_file>`

Interface to obtain content of example file.


### GET

Return conent of example file in example module.


## ApiTerraregProviderLogos

`/v1/terrareg/provider_logos`

Provide interface to obtain all provider logo details


### GET

Return all details about provider logos.


## ApiTerraregModuleSearchFilters

`/v1/terrareg/search_filters`

Return list of filters availabe for search.


### GET

Return list of available filters and filter counts for search query.


## ApiTerraregAuditHistory

`/v1/terrareg/audit-history`

Interface to obtain audit history


### GET

Obtain audit history events


## ApiTerraregAuthUserGroups

`/v1/terrareg/user-groups`

Interface to list and create user groups.


### GET

Obtain list of user groups.
### POST

Create user group


## ApiTerraregAuthUserGroup

`/v1/terrareg/user-groups/<string:user_group>`

Interface to interact with single user group.


### DELETE

Delete user group.


## ApiTerraregAuthUserGroupNamespacePermissions

`/v1/terrareg/user-groups/<string:user_group>/permissions/<string:namespace>`

Interface to create user groups namespace permissions.


### POST

Create user group namespace permission
### DELETE

Delete user group namespace permission


## ApiTerraregAdminAuthenticate

`/v1/terrareg/auth/admin/login`

Interface to perform authentication as an admin and set appropriate cookie.


### POST

Handle POST requests to the authentication endpoint.


## ApiTerraregIsAuthenticated

`/v1/terrareg/auth/admin/is_authenticated`

Interface to teturn whether user is authenticated as an admin.


### GET

Return information about current user.


## ApiTerraregHealth

`/v1/terrareg/health`

Endpoint to return 200 when healthy.


### GET

Return static 200