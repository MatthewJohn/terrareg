

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
