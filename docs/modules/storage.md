
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

Optionally, a default `git_path` can be supplied with placeholders, which provides the default `git_path` (the directory within a repository where the module is present) value when creating a module. If a common repository with multiple modules is used, the base/clone/source URL templates may not be unique per module. The templates are validated to ensure that placeholders are present in each of the templates, but can be included in the `git_path` instead.
For example, the repository clone URL could be `git@git.example.com/{namespace}/modules.git` with a git path `modules/{module}-{provider}`, meaning that an example module (Namespace: `infrastructure`, Module: `s3-bucket`, Provider: `aws`) would be present in the repository `git@git.example.com/infrastructure/modules.git` in the directory `modules/s3-bucket-aws`.

Details of the format of this configuration and some examples can be found in the configuration documentation: [GIT_PROVIDER_CONFIG](../CONFIG.md#git_provider_config).

Once enabled, users can be limited from providing custom URLs by disabling: [ALLOW_CUSTOM_GIT_URL_MODULE_PROVIDER](../CONFIG.md#allow_custom_git_url_module_provider) and [ALLOW_CUSTOM_GIT_URL_MODULE_VERSION](../CONFIG.md#allow_custom_git_url_module_version)

# Git tag security

By default, when using Git repositories for a module in Terareg, the git tag will be provided to Terraform, when initialising the module.
Terraform will then clone and checkout the tag from the git repository.

This can lead to security issues, where the tag is rewritten, in a form of supply-chain attack.

Terrareg can be configured to return the git commit SHA of the tag (at the time of indexing).

If the tag is changed, the original commit will be used by Terraform (assuming it is still available, otherwise Terraform will fail).

To enable this feature, see [MODULE_VERSION_USE_GIT_COMMIT](../CONFIG.md#module_version_use_git_commit)
