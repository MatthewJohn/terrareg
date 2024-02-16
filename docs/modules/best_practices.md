

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
