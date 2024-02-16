
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
