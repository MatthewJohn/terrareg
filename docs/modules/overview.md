# Create modules

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

