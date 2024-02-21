
# Module storage

Terrareg can work in one of two ways:

 * Modules hosting
 * Git-based

If only one of these methods is to be used, the other can be disabled:

 * [ALLOW_MODULE_HOSTING](../CONFIG.md#allow_module_hosting) - enables/disables hosting module source code
 * [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](../CONFIG.md#allow_custom_git_url_module_provider) and [ALLOW_CUSTOM_GIT_URL_MODULE_VERSION](../CONFIG.md#allow_custom_git_url_module_version) - enables disables setting git URLs on modules/within metadata file.


## Module hosting

Modules are indexed by uploading a zip/tar archive of the code to the upload endpoint.

Modules are stored in Terrareg in the [DATA_DIRECTORY](../CONFIG.md#data_directory) - it is important to ensure this path is mounted to persistent storage, so that the data is not lost between container rebuilds/upgrades.

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

Terrareg will retain an archive of indexed modules, this can be disabled using [DELETE_EXTERNALLY_HOSTED_ARTIFACTS](../CONFIG.md#delete_externally_hosted_artifacts)

# Git Providers

To fully utilise the features that Terraform provides for git, each module can be provided with three URLs:

 * Repository base URL - the base URL of the repository
 * Repository clone URL - the url for cloning the repository
 * Repository source browse URL - the URL for browsing a single file, using a placeholders for git tag and file path

The avoid having to set this up in each module, Git Providers collapse these configurations into an entity, which can be selected during module creation.

Using git providers, the namespace, module name and provider of the module can be used in the templates of the URLs, meaning that the namespace can be used to determine the organisation/user of the repository (when using Github for example).

Details of the format of this configuration and some examples can be found in the configuration documentation: [GIT_PROVIDER_CONFIG](../CONFIG.md#git_provider_config).

Once enabled, users can be limited from providing custom URLs by disabling: [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](../CONFIG.md#allow_custom_git_url_module_provider) and [ALLOW_CUSTOM_GIT_URL_MODULE_VERSION](../CONFIG.md#allow_custom_git_url_module_version)


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

If a [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](../CONFIG.md#allow_custom_git_url_module_provider) is enabled, the custom URLs (as docuented in the Git Providers section) can be entered specifically for the module.

### Git tag format

The git tag format is used for bi-directional conversion of the Semantic Version used for module versions and the tag for the version in Git.

The git tag format is provided as a templated string of the naming convension that the git repository uses.

The primary placeholder that should be used is `{version}`, which will contain the full semantic version (e.g `1.0.0` or `1.0.0-beta`).

If the tags are prefixed or suffixed with any static string, this can be added to the git tag format, e.g. `release/{version}` or `v{version}` for tags such as `release/1.2.3` or `v1.2.3`, respectively.

#### Non-semantic versioned git tags

If the git tagging format of the repository does not contain the full semantic version, specific placeholders are availale to designate how to extract the individual major, minor and patch parts of the release.
**However**, using this methology is strongly discouraged as it _could_ lead to a many-to-one mapping between module versions and git tags. This is because Terraform (and the associated Terraform module registry APIs) must adhere to Semantic versioning. If a *part* of the semantic version is not provided in the "git tag format", it will be set to 0, i.e. if a git tag format of `{major}.{minor}` is used, a git tag 1.1 would be used for all `1.1.X` semantic versions.

To attempt to avoid this, the module version import API endpoints do not allow indexing by **version** if these placeholders are uesd and, instead, a git tag must be supplied to the [import api](../API.md#apimoduleversionimport). Importing/indexing modules using the Github/Gitlab/Bitbucket hooks are unaffected by this restriction.

Placeholders for non-semantic versions:

 * `{major}` - Specifies the major part of the semantic version
 * `{minor}` - Specifies the minor part of the semantic version
 * `{patch}` - Specifies the patch part of the semantic version

**NOTE:** The preference of these should always be in the above order, i.e.:

 * If you version using `v1`, you should use the placeholder `v{major}`
 * If you version using `v1.1`, you should use the placeholder `v{major}.{minor}`

