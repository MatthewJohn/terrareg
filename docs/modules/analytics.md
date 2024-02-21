# Module Usage Analytics

Module analaytics allow each use of a module to be tracked - giving statistics such as:

 * Identifing IDs of the consumers of the modules
 * The latest version of the module that is being used by the consumer
 * The version of Terraform used by the consumer
 * The 'highest' deployment environment that the consumer has deployed the module to


## Identifying deployment environments

When Terraform (that is using a module) is deployed a particular environment, this environment can be captured in the analytics and show in the 'analytics' tab to display how the deployment of the instance of the module has progressed.

Environments are defined in a heirarchy, e.g. dev->production. In this example, if a module is deployed to 'dev', it is show with the environment 'dev' in the Terrareg UI. If it's then deployed to 'production', it is shown with 'production'.

This identification is performed using Terraform authentication.

The keys used for each environment can be defined in [ANALYTICS_AUTH_KEYS](../CONFIG.md#analytics_auth_keys).

The authentication keys must then be configured on the nodes that perform the deployment, setting up `~/.terraformrc` using the following example:
```
credentials "terrareg.my.domain" {
  token = "<environment auth key>"
}
```

If deployment nodes are shared between environments, this configuration will need to be dynamically regenerated based on the environment being deployed to.

## Customising the analytics token

The label used to describe analytics tokens and the example analytics token displayed in the UI can be customised by setting:

 * [ANALYTICS_TOKEN_PHRASE](../CONFIG.md#analytics_token_phrase) - the phrase used to describe the analytics token
 * [EXAMPLE_ANALYTICS_TOKEN](../CONFIG.md#example_analytics_token) - a noun to describe the analytics token or .
 * [ANALYTICS_TOKEN_DESCRIPTION](../CONFIG.md#analytics_token_description) - description of the token

For example, to rebrand the analytics token as someone's first name, you could set:

 * EXAMPLE_ANALYTICS_TOKEN - "john"
 * ANALYTICS_TOKEN_PHRASE - "first name"
 * ANALYTICS_TOKEN_DESCRIPTION - "Set to your first name"

The usage example would then read:
```
Ensure the "john" placeholder must be replaced with your 'first name',
...
module "terraform-aws-rds" {
  source  = "terrareg.my.domain/john__namespace/module/null"

```
The 'usage bulder' table would a variable: "first name" with a description "Set to your first name"


## Enforcing analytics tokens

By default, analytics token must be provided to use a module.

Analytics enforced can be disabled, allowing module usage with an analytics key being passed.

To disable this, see [ALLOW_UNIDENTIFIED_DOWNLOADS](../CONFIG.md#allow_unidentified_downloads)

## Disable analytics token enforcement for some users

Where an example in a module requires another module from the registry to work, it is only useful to use a source URL without an analytics token, to avoid end-user copy+pasting the example and leaving an example analytics token.
However, during the development of modules, to easily test the examples, the analayics token enforcement check can be disabled for the user by using a Terraform auth token (configured in the user's .terraformrc file) configured in the registry.

The configure this, see [IGNORE_ANALYTICS_TOKEN_AUTH_KEYS](../CONFIG.md#ignore_analytics_token_auth_keys).
